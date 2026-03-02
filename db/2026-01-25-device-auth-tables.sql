CREATE TABLE IF NOT EXISTS `device_auth_states` (
  `device_code` varchar(64) NOT NULL,
  `user_code` varchar(32) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'pending',
  `member_number` bigint(20) DEFAULT NULL,
  `scopes` text,
  `created_at` datetime(3) DEFAULT NULL,
  `expires_at` datetime(3) NOT NULL,
  `last_checked` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`device_code`),
  UNIQUE KEY `idx_device_auth_states_user_code` (`user_code`),
  KEY `idx_device_auth_states_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
