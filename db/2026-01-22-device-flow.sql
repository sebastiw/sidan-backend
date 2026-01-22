-- Device Flow Authentication Table
-- OAuth2 device flow for CLI and limited-input devices
-- Stores temporary device codes that expire in 10 minutes

CREATE TABLE IF NOT EXISTS device_codes (
    device_code VARCHAR(128) PRIMARY KEY COMMENT 'Device code for token polling',
    user_code VARCHAR(16) UNIQUE NOT NULL COMMENT 'Human-friendly code (e.g., ABCD-1234)',
    verification_uri TEXT NOT NULL COMMENT 'URL for user verification',
    expires_at TIMESTAMP NOT NULL COMMENT 'Device code expires after 10 minutes',
    `interval` INT NOT NULL DEFAULT 5 COMMENT 'Polling interval in seconds',
    
    -- Approval state (set when user authorizes)
    approved BOOLEAN DEFAULT FALSE COMMENT 'Whether user has authorized',
    member_number BIGINT COMMENT 'Member number after approval',
    email VARCHAR(255) COMMENT 'Member email after approval',
    scopes TEXT COMMENT 'JSON array of granted scopes',
    provider VARCHAR(32) COMMENT 'OAuth2 provider used for verification',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_code (user_code),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
