-- ============================================================
-- 000008_add_line_fields.down.sql
-- ============================================================

DROP INDEX IF EXISTS idx_users_line_user_id;

ALTER TABLE users
  DROP COLUMN IF EXISTS line_user_id;

ALTER TABLE companies
  DROP COLUMN IF EXISTS line_channel_secret,
  DROP COLUMN IF EXISTS line_channel_access_token;
