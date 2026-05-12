-- seed system admin (คุณ)
-- password: เปลี่ยนก่อน deploy จริง (นี่คือ bcrypt hash ของ "changeme")
INSERT INTO users (
  id, email, password, name, phone, is_active
) VALUES (
  uuid_generate_v4(),
  'admin@conduit-th.dev',
  '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj2NbuyHB6Iq',
  'System Admin',
  '',
  true
) ON CONFLICT (email) DO NOTHING;
