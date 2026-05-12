-- materials (price list ต่อบริษัท)
CREATE TABLE materials (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id   UUID NOT NULL REFERENCES companies(id),
  code         VARCHAR(20) NOT NULL,   -- เช่น 1.1.20, 6.17.2
  name         TEXT NOT NULL,          -- เช่น ท่อ PB ขนาด 20 มม.
  unit         VARCHAR(20) NOT NULL,   -- เมตร, ตัว, ชุด, หัว
  unit_price   DECIMAL(12, 2) NOT NULL DEFAULT 0,
  labor_cost   DECIMAL(12, 2) NOT NULL DEFAULT 0,
  k_factor     VARCHAR(20),            -- K NO, K 5.1.2 (เก็บไว้แสดง)
  is_active    BOOLEAN NOT NULL DEFAULT true,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, code)
);

CREATE INDEX idx_materials_company ON materials(company_id);
CREATE INDEX idx_materials_code    ON materials(company_id, code);

-- material price history
CREATE TABLE material_price_history (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  material_id  UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
  changed_by   UUID NOT NULL REFERENCES users(id),
  old_price    DECIMAL(12, 2) NOT NULL,
  new_price    DECIMAL(12, 2) NOT NULL,
  changed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_price_history_material ON material_price_history(material_id);

-- job materials (วัสดุที่ใช้ในงาน)
CREATE TABLE job_materials (
  id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id         UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  assignment_id  UUID REFERENCES job_assignments(id), -- ของทีมไหน
  material_id    UUID REFERENCES materials(id),        -- NULL ถ้าเป็นรายการพิเศษ
  code           VARCHAR(20),   -- snapshot ณ เวลาที่บันทึก
  name           TEXT NOT NULL,
  unit           VARCHAR(20) NOT NULL,
  quantity       DECIMAL(10, 3) NOT NULL DEFAULT 0,
  unit_price     DECIMAL(12, 2) NOT NULL DEFAULT 0,  -- snapshot ราคา
  labor_cost     DECIMAL(12, 2) NOT NULL DEFAULT 0,  -- snapshot ค่าแรง
  total          DECIMAL(12, 2) GENERATED ALWAYS AS
                   ((unit_price + labor_cost) * quantity) STORED,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_materials_job        ON job_materials(job_id);
CREATE INDEX idx_job_materials_assignment ON job_materials(assignment_id);

-- photo stage enum
CREATE TYPE photo_stage AS ENUM (
  'before',   -- ก่อนดำเนินการ
  'during',   -- ระหว่างดำเนินการ
  'after'     -- หลังดำเนินการ
);

-- job photos
CREATE TABLE job_photos (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id       UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  uploaded_by  UUID NOT NULL REFERENCES users(id),
  s3_key       TEXT NOT NULL,       -- key ใน MinIO
  url          TEXT NOT NULL,       -- presigned URL (regenerate ได้)
  stage        photo_stage NOT NULL,
  is_selected  BOOLEAN NOT NULL DEFAULT false, -- office เลือกใส่ PDF
  caption      TEXT,
  uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photos_job ON job_photos(job_id);

-- document type enum
CREATE TYPE document_type AS ENUM (
  'repair_photo',   -- หน้า 1: รูปงานซ่อมท่อ
  'repair_form_1',  -- หน้า 2: ใบแจ้งซ่อม (checkbox + แผนที่)
  'repair_form_2',  -- หน้า 3: ใบแจ้งซ่อม (ราคา + วัสดุ)
  'memo',           -- หน้า 4: บันทึกข้อความ
  'delivery'        -- หน้า 5: ใบส่งมอบงาน
);

-- document status enum
CREATE TYPE document_status AS ENUM (
  'pending',
  'generating',
  'done',
  'failed'
);

-- documents
CREATE TABLE documents (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id        UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  generated_by  UUID NOT NULL REFERENCES users(id),
  doc_type      document_type NOT NULL,
  s3_key        TEXT,
  url           TEXT,
  status        document_status NOT NULL DEFAULT 'pending',
  error_msg     TEXT,
  generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_job ON documents(job_id);
