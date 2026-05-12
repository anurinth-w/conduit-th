-- extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- สำหรับ fuzzy search

-- companies
CREATE TABLE companies (
  id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name               VARCHAR(255) NOT NULL,
  tax_id             VARCHAR(20),
  address            TEXT,
  phone              VARCHAR(20),
  email              VARCHAR(255),
  logo_url           TEXT,
  job_code_formats   JSONB NOT NULL DEFAULT '{}',
  trigger_rules      JSONB NOT NULL DEFAULT '[]',
  doc_templates      JSONB NOT NULL DEFAULT '{}',
  active_job_types   JSONB NOT NULL DEFAULT '[]',
  is_active          BOOLEAN NOT NULL DEFAULT true,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN companies.job_code_formats IS
  'format per job_type e.g. {"repair": "PTY L3-{MM}{XXX}", "meter": "M{MM}{XXX}"}';
COMMENT ON COLUMN companies.trigger_rules IS
  'OR conditions e.g. [{"type":"amount","threshold":50000},{"type":"job_count","threshold":20}]';
COMMENT ON COLUMN companies.active_job_types IS
  'enabled job types for this company e.g. ["repair","expansion","meter"]';

-- users (ไม่มี company_id ตรงๆ เพราะ 1 user มีได้หลาย company)
CREATE TABLE users (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email        VARCHAR(255) NOT NULL UNIQUE,
  password     VARCHAR(255) NOT NULL,
  name         VARCHAR(255) NOT NULL,
  phone        VARCHAR(20),
  avatar_url   TEXT,
  is_active    BOOLEAN NOT NULL DEFAULT true,
  last_login   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- roles enum
CREATE TYPE user_role AS ENUM ('admin', 'manager', 'office', 'technician');

-- user <-> company memberships (many-to-many)
CREATE TABLE user_company_memberships (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  role            user_role NOT NULL,
  job_type_scope  JSONB NOT NULL DEFAULT '[]',
  is_active       BOOLEAN NOT NULL DEFAULT true,
  joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, company_id)
);

COMMENT ON COLUMN user_company_memberships.job_type_scope IS
  'for office role: which job types they can see e.g. ["repair","meter"]
   empty array = see all types (for manager/admin)';

CREATE INDEX idx_memberships_user    ON user_company_memberships(user_id);
CREATE INDEX idx_memberships_company ON user_company_memberships(company_id);

-- refresh tokens
CREATE TABLE refresh_tokens (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  VARCHAR(255) NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
