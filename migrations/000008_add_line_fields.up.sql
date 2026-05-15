-- ============================================================
-- 000008_add_line_fields.up.sql
-- เพิ่ม LINE integration fields ใน companies และ users
-- ============================================================

-- companies: เพิ่ม LINE OA token (1 บริษัท = 1 OA = 1 token)
ALTER TABLE companies
  ADD COLUMN line_channel_access_token TEXT,
  ADD COLUMN line_channel_secret       TEXT;  -- สำหรับ verify webhook signature

-- users: เพิ่ม LINE User ID (ได้มาตอน Manager เพิ่ม OA เป็นเพื่อน)
ALTER TABLE users
  ADD COLUMN line_user_id TEXT UNIQUE;

CREATE INDEX idx_users_line_user_id ON users(line_user_id);
