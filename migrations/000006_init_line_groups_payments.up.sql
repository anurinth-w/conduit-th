-- ============================================================
-- 000006_init_line_groups_payments.up.sql
-- ============================================================

CREATE TYPE line_group_purpose AS ENUM (
  'inbound',
  'outbound',
  'both'
);

CREATE TYPE job_line_group_role AS ENUM (
  'source',
  'destination'
);

CREATE TYPE payment_status AS ENUM (
  'unpaid',
  'pending',
  'paid',
  'cancelled'
);

CREATE TYPE compensation_type AS ENUM (
  'per_job',
  'salary'
);

CREATE TABLE line_groups (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID NOT NULL REFERENCES companies(id),
  line_group_id   TEXT NOT NULL,
  name            TEXT NOT NULL,
  purpose         line_group_purpose NOT NULL DEFAULT 'both',
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, line_group_id)
);

CREATE INDEX idx_line_groups_company ON line_groups(company_id);
CREATE INDEX idx_line_groups_active  ON line_groups(company_id, is_active);

CREATE TABLE job_line_groups (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  line_group_id   UUID NOT NULL REFERENCES line_groups(id),
  role            job_line_group_role NOT NULL,
  notified_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(job_id, line_group_id, role)
);

CREATE INDEX idx_job_line_groups_job   ON job_line_groups(job_id);
CREATE INDEX idx_job_line_groups_group ON job_line_groups(line_group_id);

ALTER TABLE jobs
  ADD COLUMN payment_status payment_status NOT NULL DEFAULT 'unpaid';

CREATE INDEX idx_jobs_payment_status ON jobs(payment_status);

CREATE TABLE job_payments (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  amount          NUMERIC(12, 2) NOT NULL DEFAULT 0,
  status          payment_status NOT NULL DEFAULT 'unpaid',
  slip_media_id   UUID,
  paid_at         TIMESTAMPTZ,
  note            TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by      UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_job_payments_job    ON job_payments(job_id);
CREATE INDEX idx_job_payments_status ON job_payments(status);

CREATE TABLE worker_compensations (
  id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id          UUID NOT NULL REFERENCES companies(id),
  worker_id           UUID NOT NULL REFERENCES users(id),
  compensation_type   compensation_type NOT NULL DEFAULT 'per_job',
  period_start        DATE,
  period_end          DATE,
  amount              NUMERIC(12, 2) NOT NULL DEFAULT 0,
  status              payment_status NOT NULL DEFAULT 'unpaid',
  slip_media_id       UUID,
  paid_at             TIMESTAMPTZ,
  note                TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by          UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_worker_comp_company ON worker_compensations(company_id);
CREATE INDEX idx_worker_comp_worker  ON worker_compensations(worker_id);
CREATE INDEX idx_worker_comp_status  ON worker_compensations(status);

CREATE TABLE worker_compensation_jobs (
  compensation_id UUID NOT NULL REFERENCES worker_compensations(id) ON DELETE CASCADE,
  job_id          UUID NOT NULL REFERENCES jobs(id),
  job_amount      NUMERIC(12, 2) NOT NULL DEFAULT 0,
  PRIMARY KEY (compensation_id, job_id)
);

CREATE INDEX idx_comp_jobs_job ON worker_compensation_jobs(job_id);
