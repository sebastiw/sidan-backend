-- Authentication System Tables
-- Phase 1: Token Storage and State Management

-- OAuth2 state tracking for CSRF protection
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

-- OAuth2 tokens (encrypted at rest)
CREATE TABLE IF NOT EXISTS auth_tokens (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT NOT NULL COMMENT 'Reference to cl2007_members.id',
    provider VARCHAR(32) NOT NULL COMMENT 'OAuth2 provider name',
    access_token TEXT NOT NULL COMMENT 'Encrypted access token',
    refresh_token TEXT COMMENT 'Encrypted refresh token (nullable)',
    token_type VARCHAR(32) DEFAULT 'Bearer' COMMENT 'Token type (usually Bearer)',
    expires_at TIMESTAMP NULL COMMENT 'Token expiration time',
    scopes JSON COMMENT 'Array of granted scopes',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_member_provider (member_id, provider),
    INDEX idx_member (member_id),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Provider account linking (which emails belong to which member)
CREATE TABLE IF NOT EXISTS auth_provider_links (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT NOT NULL COMMENT 'Reference to cl2007_members.id',
    provider VARCHAR(32) NOT NULL COMMENT 'OAuth2 provider name',
    provider_user_id VARCHAR(255) NOT NULL COMMENT 'User ID from provider',
    provider_email VARCHAR(255) NOT NULL COMMENT 'Email from provider',
    email_verified BOOLEAN DEFAULT FALSE COMMENT 'Whether provider verified the email',
    linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_provider_user (provider, provider_user_id),
    INDEX idx_member (member_id),
    INDEX idx_provider_email (provider, provider_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Session management (DB-backed sessions)
CREATE TABLE IF NOT EXISTS auth_sessions (
    id VARCHAR(128) PRIMARY KEY COMMENT 'Session ID (stored in cookie)',
    member_id BIGINT NOT NULL COMMENT 'Reference to cl2007_members.id',
    data JSON COMMENT 'Session data (scopes, metadata)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL COMMENT 'Session expiration',
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_member (member_id),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
