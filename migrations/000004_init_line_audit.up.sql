CREATE TYPE ai_status AS ENUM (
  'pending',
  'processing',
  'done',
  'failed',
  'ignored'
);

CREATE TABLE line_messages (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID NOT NULL REFERENCES companies(id),
  line_user_id    VARCHAR(100) NOT NULL,
  line_group_id   VARCHAR(100),
  line_message_id VARCHAR(100) NOT NULL UNIQUE,
  raw_message     TEXT NOT NULL,
  extracted_data  JSONB,
  linked_job_id   UUID REFERENCES jobs(id),
  ai_status       ai_status NOT NULL DEFAULT 'pending',
  reviewed_by     UUID REFERENCES users(id),
  reviewed_at     TIMESTAMPTZ,
  received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_messages_company   ON line_messages(company_id);
CREATE INDEX idx_line_messages_ai_status ON line_messages(ai_status);
CREATE INDEX idx_line_messages_received  ON line_messages(received_at DESC);

CREATE TABLE audit_logs (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID REFERENCES companies(id),
  actor_id    UUID REFERENCES users(id),
  action      VARCHAR(100) NOT NULL,
  entity      VARCHAR(50) NOT NULL,
  entity_id   UUID NOT NULL,
  before_data JSONB,
  after_data  JSONB,
  ip_address  INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_company  ON audit_logs(company_id);
CREATE INDEX idx_audit_entity   ON audit_logs(entity, entity_id);
CREATE INDEX idx_audit_actor    ON audit_logs(actor_id);
CREATE INDEX idx_audit_created  ON audit_logs(created_at DESC);
