-- Remove redundant auth tables (JWT tokens are stateless)
-- We only need auth_states for OAuth2 CSRF protection
-- Email verification is done directly against cl2007_members.email

DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS auth_provider_links;
