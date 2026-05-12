INSERT INTO users (
  id, email, password, name, phone, is_active
) VALUES (
  uuid_generate_v4(),
  'admin@conduit-th.dev',
  '$2a$12$wAxNeKZUO2hsecVNSDOBzO7pebqeH8YbUChjXPzzhqlkU/CuFOqVW',
  'System Admin',
  '',
  true
) ON CONFLICT (email) DO NOTHING;
