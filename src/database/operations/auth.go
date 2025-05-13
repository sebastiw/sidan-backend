package database

import (
	// "database/sql"
	// "errors"
	// "fmt"

	// . "github.com/sebastiw/sidan-backend/src/database"
	// . "github.com/sebastiw/sidan-backend/src/database/models"
)

// func NewResourceOperation(db *sql.DB) ResourceOperation {
// 	return ResourceOperation{db}
// }

// type ResourceOperation struct {
// 	db *sql.DB
// }

// func (o ResourceOperation) GetResourceByResourceIdentifier(resourceIdentifier string) Resource {
// 	var r = Resource{}

// 		q := `
// SELECT
//   id, created_at, updated_at, resource_identifier, description
// FROM cl2025_resources
// WHERE resource_identifier=?
// ORDER BY created_at,id
// LIMIT 1
// `

// 	err := o.db.QueryRow(q, resourceIdentifier).Scan(
// 		&r.Id,
// 		&r.CreatedAt,
// 		&r.UpdatedAt,
// 		&r.ResourceIdentifier,
// 		&r.Description)

// 	switch {
// 	case err == sql.ErrNoRows:
// 	case err != nil:
// 		ErrorCheck(err)
// 	default:
// 	}
// 	return r
// }

// func NewPermissionOperation(db *sql.DB) PermissionOperation {
// 	return PermissionOperation{db}
// }

// type PermissionOperation struct {
// 	db *sql.DB
// }

// func (o PermissionOperation) GetPermissionsByResourceId(resourceId int64) []Permission {
// 	l := make([]Permission, 0)

// 		q := `
// SELECT
//   id, created_at, updated_at, permission_identifier, description, resource_id, resource
// FROM cl2025_permissions
// WHERE resource_id=?
// ORDER BY created_at,id
// `

// 	rows, err := o.db.Query(q, resourceId)
// 	ErrorCheck(err)
// 	defer rows.Close()

// 	for rows.Next() {
// 		var p = Permission{}
// 		err = rows.Scan(
// 			&p.Id,
// 			&p.CreatedAt,
// 			&p.UpdatedAt,
// 			&p.PermissionIdentifier,
// 			&p.Description,
// 			&p.ResourceId,
// 			&p.Resource)
// 		switch {
// 		case err == sql.ErrNoRows:
// 		case err != nil:
// 			ErrorCheck(err)
// 		default:
// 		}
// 		l = append(l, p)
// 	}

// 	return l
// }

// func NewClientOperation(db *sql.DB) ClientOperation {
// 	return ClientOperation{db}
// }

// type ClientOperation struct {
// 	db *sql.DB
// }

// func (o ClientOperation) GetClientByClientIdentifier(clientIdentifier string) Client {
// 	var c = Client{}

// 	q := `
// SELECT
//   id, created_at, updated_at, client_identifier, client_secret_encrypted,
//   description, enabled, consent_required, is_public, authorization_code_enabled,
//   client_credentials_enabled, token_expiration_in_seconds,
//   refresh_token_offline_idle_timeout_in_seconds,
//   refresh_token_offline_max_lifetime_in_seconds,
//   include_open_id_connect_claims_in_access_token, default_acr_level
// FROM cl2025_clients
// WHERE client_identifier=?
// ORDER BY created_at,id
// LIMIT 1
// `
// 	err := o.db.QueryRow(q, clientIdentifier).Scan(
// 		&c.Id,
// 		&c.CreatedAt,
// 		&c.UpdatedAt,
// 		&c.ClientIdentifier,
// 		&c.ClientSecretEncrypted,
// 		&c.Description,
// 		&c.Enabled,
// 		&c.ConsentRequired,
// 		&c.IsPublic,
// 		&c.AuthorizationCodeEnabled,
// 		&c.ClientCredentialsEnabled,
// 		&c.TokenExpirationInSeconds,
// 		&c.RefreshTokenOfflineIdleTimeoutInSeconds,
// 		&c.RefreshTokenOfflineMaxLifetimeInSeconds,
// 		&c.IncludeOpenIDConnectClaimsInAccessToken,
// 		&c.DefaultAcrLevel)

// 	switch {
// 	case err == sql.ErrNoRows:
// 	case err != nil:
// 		ErrorCheck(err)
// 	default:
// 	}
// 	return c

// }

// func (o ClientOperation) ClientLoadRedirectURIs(client *Client) {
// 	l := make([]RedirectURI, 0)

// 	q := `
// SELECT
//   id, created_at, uri, client_id
// FROM cl2025_redirect_uris
// WHERE client_id=?
// ORDER BY created_at,id
// `

// 	if client == nil {
// 		return
// 	}

// 	rows, err := o.db.Query(q, client.Id)
// 	ErrorCheck(err)
// 	defer rows.Close()

// 	for rows.Next() {
// 		var r = RedirectURI{}
// 		err = rows.Scan(
// 			&r.Id,
// 			&r.CreatedAt,
// 			&r.URI,
// 			&r.ClientId,
// 		)
// 		switch {
// 		case err == sql.ErrNoRows:
// 		case err != nil:
// 			ErrorCheck(err)
// 		default:
// 		}
// 		l = append(l, r)
// 	}

// 	client.RedirectURIs = l
// }

// func NewUserSessionOperation(db *sql.DB) UserSessionOperation {
// 	return UserSessionOperation{db}
// }

// type UserSessionOperation struct {
// 	db *sql.DB
// }

// func (o UserSessionOperation) GetUserSessionBySessionIdentifier(sessionIdentifier string) UserSession {
// 	var u = UserSession{}

// 	q := `
// SELECT
//   id, created_at, updated_at, session_identifier, started, last_accessed,
//   auth_methods, acr_level, auth_time, ip_address, device_name, device_type, device_os,
//   level2_auth_config_has_changed, user_id
// FROM cl2025_user_sessions
// WHERE session_identifier=?
// ORDER BY created_at,id
// LIMIT 1
// `
// 	err := o.db.QueryRow(q, sessionIdentifier).Scan(
// 		&u.Id,
// 		&u.CreatedAt,
// 		&u.UpdatedAt,
// 		&u.SessionIdentifier,
// 		&u.Started,
// 		&u.LastAccessed,
// 		&u.AuthMethods,
// 		&u.AcrLevel,
// 		&u.AuthTime,
// 		&u.IpAddress,
// 		&u.DeviceName,
// 		&u.DeviceType,
// 		&u.DeviceOS,
// 		&u.Level2AuthConfigHasChanged,
// 		&u.UserId,
// 	)

// 	switch {
// 	case err == sql.ErrNoRows:
// 	case err != nil:
// 		ErrorCheck(err)
// 	default:
// 	}
// 	return u
// }

// func (o UserSessionOperation) UserSessionLoadUser(userSession *UserSession) {
// 	if userSession == nil {
// 		return
// 	}

// 	uo := NewOauthUserOperation(o.db)
// 	user := uo.GetUserById(userSession.UserId)
// 	userSession.User = user
// }

// func (o UserSessionOperation) UserSessionsLoadUsers(userSessions []UserSession) {
// 	if userSessions == nil {
// 		return
// 	}

// 	userIds := make([]int64, 0, len(userSessions))
// 	for _, userSession := range userSessions {
// 		userIds = append(userIds, userSession.UserId)
// 	}

// 	uo := NewOauthUserOperation(o.db)
// 	users := uo.GetUsersByIds(userIds)

// 	usersById := make(map[int64]OauthUser)
// 	for _, user := range users {
// 		usersById[user.Id] = user
// 	}

// 	for i, userSession := range userSessions {
// 		user, ok := usersById[userSession.UserId]
// 		if !ok {
// 			ErrorCheck(errors.New(fmt.Sprintf("unable to find user with id %v", userSession.Id)))
// 		}
// 		userSessions[i].User = user
// 	}
// }

// func NewOauthUserOperation(db *sql.DB) OauthUserOperation {
// 	return OauthUserOperation{db}
// }

// type OauthUserOperation struct {
// 	db *sql.DB
// }

// func (o OauthUserOperation) GetUserById(userId int64) OauthUser {
// 	u := OauthUser{}

// 		q := `
// SELECT
//   id, created_at, updated_at, enabled, subject, username, given_name,
//   middle_name, family_name, nickname, website, gender, email, email_verified,
//   email_verification_code_encrypted, email_verification_code_issued_at,
//   zone_info_country_name, zone_info, locale, birth_date,
//   phone_number_country_uniqueid, phone_number_country_callingcode, phone_number,
//   phone_number_verified, phone_number_verification_code_encrypted,
//   phone_number_verification_code_issued_at, address_line1, address_line2,
//   address_locality, address_region, address_postal_code, address_country,
//   password_hash, otp_secret, otp_enabled, forgot_password_code_encrypted,
//   forgot_password_code_issued_at
// FROM cl2025_user
// WHERE id=?
// ORDER BY created_at,id
// LIMIT 1
// `

// 	err := o.db.QueryRow(q, userId).Scan(
// 		&u.Id,
// 		&u.CreatedAt,
// 		&u.UpdatedAt,
// 		&u.Enabled,
// 		&u.Subject,
// 		&u.Username,
// 		&u.GivenName,
// 		&u.MiddleName,
// 		&u.FamilyName,
// 		&u.Nickname,
// 		&u.Website,
// 		&u.Gender,
// 		&u.Email,
// 		&u.EmailVerified,
// 		&u.EmailVerificationCodeEncrypted,
// 		&u.EmailVerificationCodeIssuedAt,
// 		&u.ZoneInfoCountryName,
// 		&u.ZoneInfo,
// 		&u.Locale,
// 		&u.BirthDate,
// 		&u.PhoneNumberCountryUniqueId,
// 		&u.PhoneNumberCountryCallingCode,
// 		&u.PhoneNumber,
// 		&u.PhoneNumberVerified,
// 		&u.PhoneNumberVerificationCodeEncrypted,
// 		&u.PhoneNumberVerificationCodeIssuedAt,
// 		&u.AddressLine1,
// 		&u.AddressLine2,
// 		&u.AddressLocality,
// 		&u.AddressRegion,
// 		&u.AddressPostalCode,
// 		&u.AddressCountry,
// 		&u.PasswordHash,
// 		&u.OTPSecret,
// 		&u.OTPEnabled,
// 		&u.ForgotPasswordCodeEncrypted,
// 		&u.ForgotPasswordCodeIssuedAt,
// 	)
// 	//			&u.Groups,
// 	//			&u.Permissions,
// 	//			&u.Attributes)
// 	switch {
// 	case err == sql.ErrNoRows:
// 	case err != nil:
// 		ErrorCheck(err)
// 	default:
// 	}

// 	return u
// }

// func (o OauthUserOperation) GetUsersByIds(userIds []int64) map[int64]OauthUser {
// 	l := make(map[int64]OauthUser)

// 		q := `
// SELECT
//   id, created_at, updated_at, enabled, subject, username, given_name,
//   middle_name, family_name, nickname, website, gender, email, email_verified,
//   email_verification_code_encrypted, email_verification_code_issued_at,
//   zone_info_country_name, zone_info, locale, birth_date,
//   phone_number_country_uniqueid, phone_number_country_callingcode, phone_number,
//   phone_number_verified, phone_number_verification_code_encrypted,
//   phone_number_verification_code_issued_at, address_line1, address_line2,
//   address_locality, address_region, address_postal_code, address_country,
//   password_hash, otp_secret, otp_enabled, forgot_password_code_encrypted,
//   forgot_password_code_issued_at
// FROM cl2025_user
// WHERE id in ?
// ORDER BY created_at,id
// `

// 	rows, err := o.db.Query(q, userIds)
// 	ErrorCheck(err)
// 	defer rows.Close()

// 	for rows.Next() {
// 		var u = OauthUser{}
// 		err = rows.Scan(
// 			&u.Id,
// 			&u.CreatedAt,
// 			&u.UpdatedAt,
// 			&u.Enabled,
// 			&u.Subject,
// 			&u.Username,
// 			&u.GivenName,
// 			&u.MiddleName,
// 			&u.FamilyName,
// 			&u.Nickname,
// 			&u.Website,
// 			&u.Gender,
// 			&u.Email,
// 			&u.EmailVerified,
// 			&u.EmailVerificationCodeEncrypted,
// 			&u.EmailVerificationCodeIssuedAt,
// 			&u.ZoneInfoCountryName,
// 			&u.ZoneInfo,
// 			&u.Locale,
// 			&u.BirthDate,
// 			&u.PhoneNumberCountryUniqueId,
// 			&u.PhoneNumberCountryCallingCode,
// 			&u.PhoneNumber,
// 			&u.PhoneNumberVerified,
// 			&u.PhoneNumberVerificationCodeEncrypted,
// 			&u.PhoneNumberVerificationCodeIssuedAt,
// 			&u.AddressLine1,
// 			&u.AddressLine2,
// 			&u.AddressLocality,
// 			&u.AddressRegion,
// 			&u.AddressPostalCode,
// 			&u.AddressCountry,
// 			&u.PasswordHash,
// 			&u.OTPSecret,
// 			&u.OTPEnabled,
// 			&u.ForgotPasswordCodeEncrypted,
// 			&u.ForgotPasswordCodeIssuedAt,
// 		)
// 		//			&u.Groups,
// 		//			&u.Permissions,
// 		//			&u.Attributes)
// 		switch {
// 		case err == sql.ErrNoRows:
// 		case err != nil:
// 			ErrorCheck(err)
// 		default:
// 		}
// 		l[u.Id] = u
// 	}

// 	return l
// }
