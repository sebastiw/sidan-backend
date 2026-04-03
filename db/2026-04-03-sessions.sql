CREATE TABLE IF NOT EXISTS `oauth2_sessions` (
    `token`         VARCHAR(64)  NOT NULL,
    `member_number` BIGINT       NOT NULL,
    `email`         VARCHAR(255) NOT NULL,
    `provider`      VARCHAR(32)  NOT NULL,
    `expires_at`    DATETIME     NOT NULL,
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`token`),
    INDEX `idx_member_number` (`member_number`),
    INDEX `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
