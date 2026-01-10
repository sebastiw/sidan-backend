-- Authentication System Tables
-- JWT-based authentication with OAuth2 support
-- Only stores temporary CSRF state during OAuth2 flow

-- OAuth2 state tracking for CSRF protection (temporary, expires in 10 minutes)
CREATE TABLE IF NOT EXISTS auth_states (
    id VARCHAR(64) PRIMARY KEY COMMENT 'Random state ID for CSRF protection',
    provider VARCHAR(32) NOT NULL COMMENT 'OAuth2 provider name (google, github, etc)',
    nonce VARCHAR(64) NOT NULL COMMENT 'Additional random value for security',
    pkce_verifier VARCHAR(128) COMMENT 'PKCE code verifier for authorization code flow',
    redirect_uri TEXT COMMENT 'Client redirect URI after auth',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL COMMENT 'State expires after 10 minutes',
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
