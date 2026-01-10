-- Post-migration for auth tables
-- NOTE: cl2007_members uses MyISAM engine which doesn't support foreign keys
-- We create indexes for performance but skip foreign key constraints

-- Mark old password fields as deprecated (don't delete for audit trail)
ALTER TABLE cl2007_members 
    MODIFY COLUMN password VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table instead',
    MODIFY COLUMN password_classic VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table instead',
    MODIFY COLUMN password_resetstring VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table instead',
    MODIFY COLUMN password_classic_resetstring VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table instead';
