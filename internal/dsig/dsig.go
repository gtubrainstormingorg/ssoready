package dsig

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ssoready/ssoready/internal/c14n"
	"github.com/ssoready/ssoready/internal/samlres"
	"github.com/ssoready/ssoready/internal/uxml"
)

var (
	ErrUnsigned       = fmt.Errorf("dsig: unsigned saml assertion")
	ErrNoRSAPublicKey = fmt.Errorf("dsig: cert does not contain *rsa.PublicKey")
	ErrBadDigest      = fmt.Errorf("dsig: digest mismatch in saml assertion")
)

func Verify(cert *x509.Certificate, data string) error {
	var res samlres.SAMLResponse
	if err := xml.Unmarshal([]byte(data), &res); err != nil {
		return err
	}

	if res.Assertion.Signature.SignatureValue == "" {
		return ErrUnsigned
	}

	digestData, err := responseDigestData(res, data)
	if err != nil {
		return err
	}

	digestHash := sha256.Sum256(digestData)
	digestHashBase64 := base64.StdEncoding.EncodeToString(digestHash[:])

	if res.Assertion.Signature.SignedInfo.Reference.DigestValue != digestHashBase64 {
		return ErrBadDigest
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return ErrNoRSAPublicKey
	}

	signatureData, err := responseSignatureData(data)
	if err != nil {
		return err
	}

	signatureHash := sha256.Sum256(signatureData)

	signatureBase64 := res.Assertion.Signature.SignatureValue
	signatureBase64 = strings.ReplaceAll(signatureBase64, " ", "")
	signatureBase64 = strings.ReplaceAll(signatureBase64, "\n", "")
	expectedSignature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, signatureHash[:], expectedSignature); err != nil {
		return fmt.Errorf("verify signature: %w", err)
	}

	return nil
}

func responseDigestData(res samlres.SAMLResponse, data string) ([]byte, error) {
	doc, err := uxml.Parse(data)
	if err != nil {
		return nil, err
	}

	assertion, _ := onlyPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
	}, doc.Root)

	nosig := exceptPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
	}, assertion)

	var inclusiveNamespaces []string
	for _, t := range res.Assertion.Signature.SignedInfo.Reference.Transforms.Transform {
		if t.Algorithm == "http://www.w3.org/2001/10/xml-exc-c14n#" {
			inclusiveNamespaces = strings.Split(t.InclusiveNamespaces.PrefixList, " ")
		}
	}

	return c14n.Canonicalize(nosig, inclusiveNamespaces)
}

func responseSignatureData(data string) ([]byte, error) {
	doc, err := uxml.Parse(data)
	if err != nil {
		return nil, err
	}

	// todo remove ok?
	n, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
	}, doc.Root)

	return c14n.Canonicalize(n, nil)
}