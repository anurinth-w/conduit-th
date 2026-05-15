CREATE TABLE document_templates (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id   UUID NOT NULL REFERENCES companies(id),
  name         TEXT NOT NULL,
  doc_type     document_type NOT NULL,
  html_content TEXT NOT NULL,
  version      INTEGER NOT NULL DEFAULT 1,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, doc_type, version)
);

CREATE INDEX idx_doc_templates_company ON document_templates(company_id);
CREATE INDEX idx_doc_templates_active  ON document_templates(company_id, is_active);

CREATE TABLE document_bundles (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id   UUID NOT NULL REFERENCES companies(id),
  job_type     job_type NOT NULL,
  name         TEXT NOT NULL,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, job_type)
);

CREATE INDEX idx_doc_bundles_company ON document_bundles(company_id);

CREATE TABLE document_bundle_pages (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  bundle_id    UUID NOT NULL REFERENCES document_bundles(id) ON DELETE CASCADE,
  template_id  UUID NOT NULL REFERENCES document_templates(id),
  page_order   INTEGER NOT NULL,
  UNIQUE(bundle_id, page_order)
);

CREATE INDEX idx_bundle_pages_bundle ON document_bundle_pages(bundle_id);
