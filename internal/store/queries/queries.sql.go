// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: queries.sql

package queries

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createAppOrganization = `-- name: CreateAppOrganization :one
insert into app_organizations (id, google_hosted_domain)
values ($1, $2)
returning id, google_hosted_domain
`

type CreateAppOrganizationParams struct {
	ID                 uuid.UUID
	GoogleHostedDomain *string
}

func (q *Queries) CreateAppOrganization(ctx context.Context, arg CreateAppOrganizationParams) (AppOrganization, error) {
	row := q.db.QueryRow(ctx, createAppOrganization, arg.ID, arg.GoogleHostedDomain)
	var i AppOrganization
	err := row.Scan(&i.ID, &i.GoogleHostedDomain)
	return i, err
}

const createAppSession = `-- name: CreateAppSession :one
insert into app_sessions (id, app_user_id, create_time, expire_time, token)
values ($1, $2, $3, $4, $5)
returning id, app_user_id, create_time, expire_time, token
`

type CreateAppSessionParams struct {
	ID         uuid.UUID
	AppUserID  uuid.UUID
	CreateTime time.Time
	ExpireTime time.Time
	Token      string
}

func (q *Queries) CreateAppSession(ctx context.Context, arg CreateAppSessionParams) (AppSession, error) {
	row := q.db.QueryRow(ctx, createAppSession,
		arg.ID,
		arg.AppUserID,
		arg.CreateTime,
		arg.ExpireTime,
		arg.Token,
	)
	var i AppSession
	err := row.Scan(
		&i.ID,
		&i.AppUserID,
		&i.CreateTime,
		&i.ExpireTime,
		&i.Token,
	)
	return i, err
}

const createAppUser = `-- name: CreateAppUser :one
insert into app_users (id, app_organization_id, display_name, email)
values ($1, $2, $3, $4)
returning id, app_organization_id, display_name, email
`

type CreateAppUserParams struct {
	ID                uuid.UUID
	AppOrganizationID uuid.UUID
	DisplayName       string
	Email             *string
}

func (q *Queries) CreateAppUser(ctx context.Context, arg CreateAppUserParams) (AppUser, error) {
	row := q.db.QueryRow(ctx, createAppUser,
		arg.ID,
		arg.AppOrganizationID,
		arg.DisplayName,
		arg.Email,
	)
	var i AppUser
	err := row.Scan(
		&i.ID,
		&i.AppOrganizationID,
		&i.DisplayName,
		&i.Email,
	)
	return i, err
}

const createSAMLSession = `-- name: CreateSAMLSession :one
insert into saml_sessions (id, saml_connection_id, secret_access_token, subject_id, subject_idp_attributes)
values ($1, $2, $3, $4, $5)
returning id, saml_connection_id, secret_access_token, subject_id, subject_idp_attributes
`

type CreateSAMLSessionParams struct {
	ID                   uuid.UUID
	SamlConnectionID     uuid.UUID
	SecretAccessToken    *uuid.UUID
	SubjectID            *string
	SubjectIdpAttributes []byte
}

func (q *Queries) CreateSAMLSession(ctx context.Context, arg CreateSAMLSessionParams) (SamlSession, error) {
	row := q.db.QueryRow(ctx, createSAMLSession,
		arg.ID,
		arg.SamlConnectionID,
		arg.SecretAccessToken,
		arg.SubjectID,
		arg.SubjectIdpAttributes,
	)
	var i SamlSession
	err := row.Scan(
		&i.ID,
		&i.SamlConnectionID,
		&i.SecretAccessToken,
		&i.SubjectID,
		&i.SubjectIdpAttributes,
	)
	return i, err
}

const getAPIKeyBySecretValue = `-- name: GetAPIKeyBySecretValue :one
select id, app_organization_id, secret_value
from api_keys
where secret_value = $1
`

func (q *Queries) GetAPIKeyBySecretValue(ctx context.Context, secretValue string) (ApiKey, error) {
	row := q.db.QueryRow(ctx, getAPIKeyBySecretValue, secretValue)
	var i ApiKey
	err := row.Scan(&i.ID, &i.AppOrganizationID, &i.SecretValue)
	return i, err
}

const getAppOrganizationByGoogleHostedDomain = `-- name: GetAppOrganizationByGoogleHostedDomain :one
select id, google_hosted_domain
from app_organizations
where google_hosted_domain = $1
`

func (q *Queries) GetAppOrganizationByGoogleHostedDomain(ctx context.Context, googleHostedDomain *string) (AppOrganization, error) {
	row := q.db.QueryRow(ctx, getAppOrganizationByGoogleHostedDomain, googleHostedDomain)
	var i AppOrganization
	err := row.Scan(&i.ID, &i.GoogleHostedDomain)
	return i, err
}

const getAppSessionByToken = `-- name: GetAppSessionByToken :one
select app_sessions.app_user_id, app_users.app_organization_id
from app_sessions
         join app_users on app_sessions.app_user_id = app_users.id
where token = $1
  and expire_time > $2
`

type GetAppSessionByTokenParams struct {
	Token      string
	ExpireTime time.Time
}

type GetAppSessionByTokenRow struct {
	AppUserID         uuid.UUID
	AppOrganizationID uuid.UUID
}

func (q *Queries) GetAppSessionByToken(ctx context.Context, arg GetAppSessionByTokenParams) (GetAppSessionByTokenRow, error) {
	row := q.db.QueryRow(ctx, getAppSessionByToken, arg.Token, arg.ExpireTime)
	var i GetAppSessionByTokenRow
	err := row.Scan(&i.AppUserID, &i.AppOrganizationID)
	return i, err
}

const getAppUserByEmail = `-- name: GetAppUserByEmail :one
select id, app_organization_id, display_name, email
from app_users
where email = $1
`

func (q *Queries) GetAppUserByEmail(ctx context.Context, email *string) (AppUser, error) {
	row := q.db.QueryRow(ctx, getAppUserByEmail, email)
	var i AppUser
	err := row.Scan(
		&i.ID,
		&i.AppOrganizationID,
		&i.DisplayName,
		&i.Email,
	)
	return i, err
}

const getAppUserByID = `-- name: GetAppUserByID :one
select id, app_organization_id, display_name, email
from app_users
where app_organization_id = $1
  and id = $2
`

type GetAppUserByIDParams struct {
	AppOrganizationID uuid.UUID
	ID                uuid.UUID
}

func (q *Queries) GetAppUserByID(ctx context.Context, arg GetAppUserByIDParams) (AppUser, error) {
	row := q.db.QueryRow(ctx, getAppUserByID, arg.AppOrganizationID, arg.ID)
	var i AppUser
	err := row.Scan(
		&i.ID,
		&i.AppOrganizationID,
		&i.DisplayName,
		&i.Email,
	)
	return i, err
}

const getEnvironment = `-- name: GetEnvironment :one
select id, redirect_url, app_organization_id, display_name
from environments
where app_organization_id = $1
  and id = $2
`

type GetEnvironmentParams struct {
	AppOrganizationID uuid.UUID
	ID                uuid.UUID
}

func (q *Queries) GetEnvironment(ctx context.Context, arg GetEnvironmentParams) (Environment, error) {
	row := q.db.QueryRow(ctx, getEnvironment, arg.AppOrganizationID, arg.ID)
	var i Environment
	err := row.Scan(
		&i.ID,
		&i.RedirectUrl,
		&i.AppOrganizationID,
		&i.DisplayName,
	)
	return i, err
}

const getEnvironmentByID = `-- name: GetEnvironmentByID :one
select id, redirect_url, app_organization_id, display_name
from environments
where id = $1
`

func (q *Queries) GetEnvironmentByID(ctx context.Context, id uuid.UUID) (Environment, error) {
	row := q.db.QueryRow(ctx, getEnvironmentByID, id)
	var i Environment
	err := row.Scan(
		&i.ID,
		&i.RedirectUrl,
		&i.AppOrganizationID,
		&i.DisplayName,
	)
	return i, err
}

const getOrganization = `-- name: GetOrganization :one
select organizations.id, organizations.environment_id, organizations.external_id
from organizations
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and organizations.id = $2
`

type GetOrganizationParams struct {
	AppOrganizationID uuid.UUID
	ID                uuid.UUID
}

func (q *Queries) GetOrganization(ctx context.Context, arg GetOrganizationParams) (Organization, error) {
	row := q.db.QueryRow(ctx, getOrganization, arg.AppOrganizationID, arg.ID)
	var i Organization
	err := row.Scan(&i.ID, &i.EnvironmentID, &i.ExternalID)
	return i, err
}

const getOrganizationByID = `-- name: GetOrganizationByID :one
select id, environment_id, external_id
from organizations
where id = $1
`

func (q *Queries) GetOrganizationByID(ctx context.Context, id uuid.UUID) (Organization, error) {
	row := q.db.QueryRow(ctx, getOrganizationByID, id)
	var i Organization
	err := row.Scan(&i.ID, &i.EnvironmentID, &i.ExternalID)
	return i, err
}

const getSAMLAccessTokenData = `-- name: GetSAMLAccessTokenData :one
select saml_sessions.id, saml_sessions.saml_connection_id, saml_sessions.secret_access_token, saml_sessions.subject_id, saml_sessions.subject_idp_attributes,
       organizations.id as organization_id,
       organizations.external_id,
       environments.id  as environment_id
from saml_sessions
         join saml_connections on saml_sessions.saml_connection_id = saml_connections.id
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_sessions.secret_access_token = $2
`

type GetSAMLAccessTokenDataParams struct {
	AppOrganizationID uuid.UUID
	SecretAccessToken *uuid.UUID
}

type GetSAMLAccessTokenDataRow struct {
	ID                   uuid.UUID
	SamlConnectionID     uuid.UUID
	SecretAccessToken    *uuid.UUID
	SubjectID            *string
	SubjectIdpAttributes []byte
	OrganizationID       uuid.UUID
	ExternalID           *string
	EnvironmentID        uuid.UUID
}

func (q *Queries) GetSAMLAccessTokenData(ctx context.Context, arg GetSAMLAccessTokenDataParams) (GetSAMLAccessTokenDataRow, error) {
	row := q.db.QueryRow(ctx, getSAMLAccessTokenData, arg.AppOrganizationID, arg.SecretAccessToken)
	var i GetSAMLAccessTokenDataRow
	err := row.Scan(
		&i.ID,
		&i.SamlConnectionID,
		&i.SecretAccessToken,
		&i.SubjectID,
		&i.SubjectIdpAttributes,
		&i.OrganizationID,
		&i.ExternalID,
		&i.EnvironmentID,
	)
	return i, err
}

const getSAMLConnectionByID = `-- name: GetSAMLConnectionByID :one
select id, organization_id, idp_redirect_url, idp_x509_certificate, idp_entity_id
from saml_connections
where id = $1
`

func (q *Queries) GetSAMLConnectionByID(ctx context.Context, id uuid.UUID) (SamlConnection, error) {
	row := q.db.QueryRow(ctx, getSAMLConnectionByID, id)
	var i SamlConnection
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.IdpRedirectUrl,
		&i.IdpX509Certificate,
		&i.IdpEntityID,
	)
	return i, err
}

const listEnvironments = `-- name: ListEnvironments :many
select id, redirect_url, app_organization_id, display_name
from environments
where app_organization_id = $1
  and id > $2
order by id
limit $3
`

type ListEnvironmentsParams struct {
	AppOrganizationID uuid.UUID
	ID                uuid.UUID
	Limit             int32
}

func (q *Queries) ListEnvironments(ctx context.Context, arg ListEnvironmentsParams) ([]Environment, error) {
	rows, err := q.db.Query(ctx, listEnvironments, arg.AppOrganizationID, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Environment
	for rows.Next() {
		var i Environment
		if err := rows.Scan(
			&i.ID,
			&i.RedirectUrl,
			&i.AppOrganizationID,
			&i.DisplayName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOrganizationDomains = `-- name: ListOrganizationDomains :many
select id, organization_id, domain
from organization_domains
where organization_id = any ($1::uuid[])
`

func (q *Queries) ListOrganizationDomains(ctx context.Context, dollar_1 []uuid.UUID) ([]OrganizationDomain, error) {
	rows, err := q.db.Query(ctx, listOrganizationDomains, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrganizationDomain
	for rows.Next() {
		var i OrganizationDomain
		if err := rows.Scan(&i.ID, &i.OrganizationID, &i.Domain); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOrganizations = `-- name: ListOrganizations :many
select id, environment_id, external_id
from organizations
where environment_id = $1
  and id > $2
order by id
limit $3
`

type ListOrganizationsParams struct {
	EnvironmentID uuid.UUID
	ID            uuid.UUID
	Limit         int32
}

func (q *Queries) ListOrganizations(ctx context.Context, arg ListOrganizationsParams) ([]Organization, error) {
	rows, err := q.db.Query(ctx, listOrganizations, arg.EnvironmentID, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Organization
	for rows.Next() {
		var i Organization
		if err := rows.Scan(&i.ID, &i.EnvironmentID, &i.ExternalID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSAMLConnections = `-- name: ListSAMLConnections :many
select id, organization_id, idp_redirect_url, idp_x509_certificate, idp_entity_id
from saml_connections
where organization_id = $1
  and id > $2
order by id
limit $3
`

type ListSAMLConnectionsParams struct {
	OrganizationID uuid.UUID
	ID             uuid.UUID
	Limit          int32
}

func (q *Queries) ListSAMLConnections(ctx context.Context, arg ListSAMLConnectionsParams) ([]SamlConnection, error) {
	rows, err := q.db.Query(ctx, listSAMLConnections, arg.OrganizationID, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SamlConnection
	for rows.Next() {
		var i SamlConnection
		if err := rows.Scan(
			&i.ID,
			&i.OrganizationID,
			&i.IdpRedirectUrl,
			&i.IdpX509Certificate,
			&i.IdpEntityID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
