-- AI status enum
CREATE TYPE ai_status AS ENUM (
  'pending',   -- รอประมวลผล
  'processing',
  'done',      -- extract สำเร็จ
  'failed',    -- extract ไม่ได้
  'ignored'    -- ไม่ใช่ข้อความงาน
);

-- line messages
CREATE TABLE line_messages (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID NOT NULL REFERENCES companies(id),
  line_user_id    VARCHAR(100) NOT NULL,
  line_group_id   VARCHAR(100),
  line_message_id VARCHAR(100) NOT NULL UNIQUE, -- idempotency
  raw_message     TEXT NOT NULL,
  extracted_data  JSONB,          -- AI extracted fields
  linked_job_id   UUID REFERENCES jobs(id),
  ai_status       ai_status NOT NULL DEFAULT 'pending',
  reviewed_by     UUID REFERENCES users(id),
  reviewed_at     TIMESTAMPTZ,
  received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_messages_company    ON line_messages(company_id);
CREATE INDEX idx_line_messages_ai_status  ON line_messages(ai_status);
CREATE INDEX idx_line_messages_received   ON line_messages(received_at DESC);

-- audit log (immutable — ห้าม DELETE, UPDATE)
CREATE TABLE audit_logs (
  id          UUID PRIMARY KEY DEFAUL
cat > migrations/000004_init_line_audit.down.sql << 'EOF'
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS line_messages;
DROP TYPE  IF EXISTS ai_status;
