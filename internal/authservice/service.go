package authservice

import (
	"crypto/x509"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/emailaddr"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/store"
)

type acsTemplateData struct {
	SignOnURL   string
	SAMLRequest string
	RelayState  string
}

var acsTemplate = template.Must(template.New("acs").Parse(`
<html>
	<body>
		<form method="POST" action="{{ .SignOnURL }}">
			<input type="hidden" name="SAMLRequest" value="{{ .SAMLRequest }}"></input>
			<input type="hidden" name="RelayState" value="{{ .RelayState }}"></input>
		</form>
		<script>
			document.forms[0].submit();
		</script>
	</body>
</html>
`))

type errorTemplateData struct {
	ErrorMessage            string
	SAMLFlowID              string
	WantAudienceRestriction string
	GotAudienceRestriction  string
}

//go:embed templates/static
var staticData embed.FS
var staticFS, _ = fs.Sub(staticData, "templates/static")

//go:embed templates/error.html
var errorTemplateContent string
var errorTemplate = template.Must(template.New("error").Parse(errorTemplateContent))

type Service struct {
	Store *store.Store
}

func (s *Service) NewHandler() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/internal/static/").Handler(http.StripPrefix("/internal/static/", http.FileServer(http.FS(staticFS))))
	r.HandleFunc("/v1/saml/{saml_conn_id}/init", s.samlInit).Methods("GET")
	r.HandleFunc("/v1/saml/{saml_conn_id}/acs", s.samlAcs).Methods("POST")
	return r
}

func (s *Service) samlInit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	samlConnID := mux.Vars(r)["saml_conn_id"]
	state := r.URL.Query().Get("state")

	slog.InfoContext(ctx, "init", "saml_connection_id", samlConnID, "state", state)

	dataRes, err := s.Store.AuthGetInitData(ctx, &store.AuthGetInitDataRequest{
		SAMLConnectionID: samlConnID,
		State:            state,
	})
	if err != nil {
		panic(err)
	}

	initRes := saml.Init(&saml.InitRequest{
		RequestID:  dataRes.RequestID,
		SPEntityID: dataRes.SPEntityID,
		Now:        time.Now(),
	})

	if err := s.Store.AuthUpsertInitiateData(ctx, &store.AuthUpsertInitiateDataRequest{
		State:           state,
		InitiateRequest: initRes.InitiateRequest,
	}); err != nil {
		panic(err)
	}

	if err := acsTemplate.Execute(w, &acsTemplateData{
		SignOnURL:   dataRes.IDPRedirectURL,
		SAMLRequest: initRes.SAMLRequest,
		RelayState:  state,
	}); err != nil {
		panic(fmt.Errorf("acsTemplate.Execute: %w", err))
	}
}

func (s *Service) samlAcs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	samlConnID := mux.Vars(r)["saml_conn_id"]

	slog.InfoContext(ctx, "acs", "saml_connection_id", samlConnID)

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "acs_form", "form", r.Form)

	dataRes, err := s.Store.AuthGetValidateData(ctx, &store.AuthGetValidateDataRequest{
		SAMLConnectionID: samlConnID,
	})

	cert, err := x509.ParseCertificate(dataRes.IDPX509Certificate)
	if err != nil {
		panic(err)
	}

	// assess the validity of the response; note that invalid requests may still have a nil err; the problem details
	// are stored in validateRes
	// todo maybe split out validateRes, validateProblems, err as the signature instead?
	validateRes, err := saml.Validate(&saml.ValidateRequest{
		SAMLResponse:   r.FormValue("SAMLResponse"),
		IDPCertificate: cert,
		IDPEntityID:    dataRes.IDPEntityID,
		SPEntityID:     dataRes.SPEntityID,
		Now:            time.Now(),
	})
	if err != nil {
		panic(err)
	}

	alreadyProcessed, err := s.Store.AuthCheckAssertionAlreadyProcessed(ctx, validateRes.RequestID)
	if err != nil {
		panic(err)
	}

	if alreadyProcessed {
		http.Error(w, "assertion previously processed", http.StatusBadRequest)
		return
	}

	var badSubjectID *string
	var email string
	subjectEmailDomain, err := emailaddr.Parse(validateRes.SubjectID)
	if err != nil {
		badSubjectID = &validateRes.SubjectID
	}
	if badSubjectID == nil {
		email = validateRes.SubjectID
	}

	var domainMismatchEmail *string
	if badSubjectID == nil {
		var domainOk bool
		for _, domain := range dataRes.OrganizationDomains {
			if domain == subjectEmailDomain {
				domainOk = true
			}
		}
		if !domainOk {
			domainMismatchEmail = &subjectEmailDomain
		}
	}

	createSAMLLoginRes, err := s.Store.AuthUpsertReceiveAssertionData(ctx, &store.AuthUpsertSAMLLoginEventRequest{
		SAMLConnectionID:                     samlConnID,
		SAMLFlowID:                           validateRes.RequestID,
		Email:                                email,
		SubjectIDPAttributes:                 validateRes.SubjectAttributes,
		SAMLAssertion:                        validateRes.Assertion,
		ErrorBadIssuer:                       validateRes.BadIssuer,
		ErrorBadAudience:                     validateRes.BadAudience,
		ErrorBadSubjectID:                    badSubjectID,
		ErrorEmailOutsideOrganizationDomains: domainMismatchEmail,
	})
	if err != nil {
		panic(err)
	}

	// present an error to the end user depending on their settings
	// todo make this pretty html
	if validateRes.BadIssuer != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage: "bad issuer",
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
	}
	if validateRes.BadAudience != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:            "Incorrect SP Entity ID in AudienceRestriction",
			SAMLFlowID:              createSAMLLoginRes.SAMLFlowID,
			WantAudienceRestriction: dataRes.SPEntityID,
			GotAudienceRestriction:  *validateRes.BadAudience,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badSubjectID != nil {
		http.Error(w, "bad subject id", http.StatusBadRequest)
		return
	}
	if domainMismatchEmail != nil {
		http.Error(w, "bad email domain", http.StatusBadRequest)
		return
	}

	// past this point, we presume the request is valid

	redirectURL, err := url.Parse(dataRes.EnvironmentRedirectURL)
	if err != nil {
		panic(err)
	}

	redirectQuery := url.Values{}
	redirectQuery.Set("saml_access_code", createSAMLLoginRes.Token)
	redirectURL.RawQuery = redirectQuery.Encode()
	redirect := redirectURL.String()

	if err := s.Store.AuthUpdateAppRedirectURL(ctx, &store.AuthUpdateAppRedirectURLRequest{
		SAMLFlowID:     createSAMLLoginRes.SAMLFlowID,
		AppRedirectURL: redirect,
	}); err != nil {
		panic(err)
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
