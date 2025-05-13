package models


import (
       "database/sql"

       "github.com/sebastiw/sidan-backend/src/enums"
)

type Settings struct {
       Id                                        int64                `json:"id"`
       CreatedAt                                 sql.NullTime         `json:"created_at"`
       UpdatedAt                                 sql.NullTime         `json:"updated_at"`
       AppName                                   string               `json:"app_name"`
       Issuer                                    string               `json:"issuer"`
       UITheme                                   string               `json:"ui_theme"`
       PasswordPolicy                            enums.PasswordPolicy `json:"password_policy"`
       SelfRegistrationEnabled                   bool                 `json:"self_registration_enabled"`
       SelfRegistrationRequiresEmailVerification bool                 `json:"self_registration_requires_email_verification"`
       TokenExpirationInSeconds                  int                  `json:"token_expiration_in_seconds"`
       RefreshTokenOfflineIdleTimeoutInSeconds   int                  `json:"refresh_token_offline_idle_timeout_in_seconds"`
       RefreshTokenOfflineMaxLifetimeInSeconds   int                  `json:"refresh_token_offline_max_lifetime_in_seconds"`
       UserSessionIdleTimeoutInSeconds           int                  `json:"user_session_idle_timeout_in_seconds"`
       UserSessionMaxLifetimeInSeconds           int                  `json:"user_session_max_lifetime_in_seconds"`
       IncludeOpenIDConnectClaimsInAccessToken   bool                 `json:"include_open_id_connect_claims_in_access_token"`
       SessionAuthenticationKey                  []byte               `json:"session_authentication_key"`
       SessionEncryptionKey                      []byte               `json:"session_encryption_key"`
       AESEncryptionKey                          []byte               `json:"aes_encryption_key"`
       SMTPHost                                  string               `json:"smtp_host"`
       SMTPPort                                  int                  `json:"smtp_port"`
       SMTPUsername                              string               `json:"smtp_username"`
       SMTPPasswordEncrypted                     []byte               `json:"smtp_password_encrypted"`
       SMTPFromName                              string               `json:"smtp_from_name"`
       SMTPFromEmail                             string               `json:"smtp_from_email"`
       SMTPEncryption                            string               `json:"smtp_encryption"`
       SMTPEnabled                               bool                 `json:"smtp_enabled"`
}
