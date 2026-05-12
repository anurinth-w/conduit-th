-- job type enum
CREATE TYPE job_type AS ENUM (
  'repair',        -- งานซ่อม
  'expansion',     -- งานขยาย
  'meter',         -- งานมาตร
  'meter_clean',   -- งานล้างมาตร
  'install',       -- งานติดตั้ง (locked)
  'deposit_refund' -- ขอคืนเงินค้ำ (locked)
);

-- job status enum
CREATE TYPE job_status AS ENUM (
  'open',            -- สร้างใหม่
  'assigned',        -- จ่ายงานแล้ว
  'in_progress',     -- กำลังทำ
  'pending_surface', -- ค้างคืนผิว
  'done',            -- เสร็จแล้ว
  'duplicate'        -- งานซ้ำ
);

-- assignment type enum
CREATE TYPE assignment_type AS ENUM (
  'main',    -- ทีมหลัก
  'surface'  -- ทีมคืนผิว
);

-- assignment status enum
CREATE TYPE assignment_status AS ENUM (
  'pending',    -- รอรับงาน
  'accepted',   -- รับงานแล้ว
  'in_progress',-- กำลังทำ
  'done'        -- เสร็จแล้ว
);

-- jobs
CREATE TABLE jobs (
  id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id          UUID NOT NULL REFERENCES companies(id),
  created_by          UUID NOT NULL REFERENCES users(id),

  -- เลขงาน
  job_code            VARCHAR(50),              -- ระบบ generate เช่น M05025
  ref_code            VARCHAR(50),              -- เลข R จากปะปาสาขาใหญ่ เช่น R202605012573
  report_number       VARCHAR(50),              -- เลขแจ้งจากปะปาสาขา เช่น M05025, PTY L3-01001
  water_user_code     VARCHAR(50),              -- เลขผู้ใช้น้ำ (optional)

  -- ประเภทและสถานะ
  job_type            job_type NOT NULL,
  status              job_status NOT NULL DEFAULT 'open',

  -- ข้อมูลงาน
  cause               VARCHAR(255),             -- สาเหตุ เช่น ท่อรั่ว
  location_text       TEXT,                     -- บริเวณ
  subdistrict         VARCHAR(100),             -- ตำบล
  district            VARCHAR(100),             -- อำเภอ
  province            VARCHAR(100),             -- จังหวัด
  lat                 DECIMAL(10, 7),           -- พิกัด GPS
  lng                 DECIMAL(10, 7),

  -- รายละเอียดท่อ
  pipe_type           VARCHAR(20),              -- PB, PVC, HDPE, GS
  pipe_size_mm        INTEGER,                  -- ขนาดท่อ มม.
  surface_condition   VARCHAR(100),             -- สภาพพื้นที่
  surface_area_sqm    DECIMAL(8, 2),            -- ขนาดคืนผิว ตร.ม.
  work_method         VARCHAR(100),             -- วิธีดำเนินงาน เช่น แรงงานคน

  -- ที่มาของงาน
  job_source          VARCHAR(100),             -- งานรับแจ้งจากผู้ใช้น้ำ
  contact_technician  VARCHAR(255),             -- ช่างเชน 0831342716
  contact_coordinator VARCHAR(255),             -- ผู้ประสานงาน

  -- ค่าใช้จ่าย (คำนวณจาก job_materials)
  cost_main           DECIMAL(12, 2) DEFAULT 0, -- ค่างานทีมหลัก
  cost_surface        DECIMAL(12, 2) DEFAULT 0, -- ค่างานทีมคืนผิว

  -- เวลา
  notified_at         TIMESTAMPTZ,              -- เวลาที่ปะปาแจ้งมา
  started_at          TIMESTAMPTZ,              -- เริ่มซ่อม
  ended_at            TIMESTAMPTZ,              -- เสร็จ

  -- duplicate
  duplicate_ref       VARCHAR(50),              -- อ้างอิงว่าซ้ำกับงานไหน
  duplicate_note      TEXT,

  -- line
  line_message_id     UUID,                     -- linked line message

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_company    ON jobs(company_id);
CREATE INDEX idx_jobs_status     ON jobs(status);
CREATE INDEX idx_jobs_job_type   ON jobs(job_type);
CREATE INDEX idx_jobs_job_code   ON jobs(job_code);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);

-- job sequence counters (สำหรับ generate job_code)
CREATE TABLE job_code_sequences (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID NOT NULL REFERENCES companies(id),
  job_type    job_type NOT NULL,
  period      VARCHAR(10) NOT NULL, -- เช่น "2505" (ปี+เดือน) หรือ "250512" (ปี+เดือน+วัน)
  last_seq    INTEGER NOT NULL DEFAULT 0,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, job_type, period)
);

-- job assignments (หัวหน้าทีมที่รับงาน)
CREATE TABLE job_assignments (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  technician_id   UUID NOT NULL REFERENCES users(id),
  assigned_by     UUID NOT NULL REFERENCES users(id),
  assignment_type assignment_type NOT NULL DEFAULT 'main',
  status          assignment_status NOT NULL DEFAULT 'pending',
  cost_share      DECIMAL(12, 2) DEFAULT 0, -- ค่างานที่ทีมนี้ได้รับ
  note            TEXT,
  assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignments_job         ON job_assignments(job_id);
CREATE INDEX idx_assignments_technician  ON job_assignments(technician_id);
