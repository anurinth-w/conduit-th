# Phase 2 — Development Log
**วันที่:** 15 พฤษภาคม 2026  
**สถานะ:** ✅ เสร็จสมบูรณ์

---

## Services ที่สร้างใน Phase 2

| Service | Port | หน้าที่ |
|---|---|---|
| Material | 8004 | รายการวัสดุต่อบริษัท, ประวัติราคา |
| Media | 8005 | อัปโหลดรูป, presigned URL, MinIO |
| Document | 8006 | generate PDF จาก HTML template ผ่าน Gotenberg |
| Notify | 8007 | ส่ง LINE push message ผ่าน LINE OA ของลูกค้า |
| Gateway | 8000 | JWT auth, rate limit, reverse proxy |

---

## Database Migrations ที่เพิ่ม

### Migration 006 — LINE Groups + Payment System
```
line_groups              — กลุ่ม LINE ทั้งหมดของบริษัท
job_line_groups          — เชื่อมงาน × กลุ่ม (source/destination)
job_payments             — รับเงินจากลูกค้า (receivable)
worker_compensations     — จ่ายเงินให้ช่าง (payable)
worker_compensation_jobs — งานที่อยู่ใน compensation batch
jobs.payment_status      — summary status (unpaid/pending/paid)
```

### Migration 007 — Document Template System
```
document_templates       — HTML template แต่ละหน้า
document_bundles         — ชุดเอกสารต่อบริษัท × ประเภทงาน
document_bundle_pages    — หน้าที่อยู่ใน bundle เรียงลำดับ
```

### Migration 008 — LINE Fields
```
companies.line_channel_access_token  — LINE OA token ของบริษัท
companies.line_channel_secret        — สำหรับ verify webhook
users.line_user_id                   — LINE user ID ของแต่ละคน
```

---

## Architecture Decisions

### LINE OA เป็นภาระของบริษัทลูกค้า
- บริษัทลูกค้าสมัคร LINE OA เอง จ่ายค่าบริการเอง
- Conduit-TH รับแค่ Channel Access Token มาเก็บใน DB
- 1 บริษัท = 1 LINE OA = 1 token
- LINE OA อยู่ในทุกกลุ่มของบริษัทนั้น ทำหน้าที่ทั้งฟังและพูด

### LINE Notification Flow
```
งานสร้างใหม่     → push หา Manager 1:1 (ทุกคนที่มี line_user_id)
Assign งาน       → push ไปกลุ่ม destination (กลุ่มช่าง)
ช่างทำเสร็จ      → push หา Manager 1:1
Manager confirm  → push กลับกลุ่ม source (กลุ่มต้นทาง)
```

### Document Template System
- แต่ละหน้าเป็น HTML template แยก (Go template engine)
- Bundle กำหนดว่างานประเภทนี้ของบริษัทนี้ใช้หน้าไหนบ้าง เรียงลำดับยังไง
- Dev สร้าง template ให้บริษัท และเก็บค่าบริการ setup
- เพิ่มบริษัทใหม่ → upload HTML template + สร้าง bundle ใหม่ ไม่ต้องแก้ code

### Payment Model
```
ลูกค้า → [receivable] → บริษัท → [payable] → ช่าง
```
- **Receivable:** `job_payments` — ลูกค้าจ่ายให้บริษัท
- **Payable:** `worker_compensations` — บริษัทจ่ายให้ช่าง
  - `per_job` — จ่ายเป็นรายงาน
  - `salary` — จ่ายเป็นเงินเดือน (รวมหลายงานใน period)

### Email
- ตัดออกจาก Phase 2
- เก็บไว้ Phase 3 สำหรับ report รายเดือนให้ผู้บริหาร

---

## Infrastructure ที่เพิ่ม

### Gotenberg (port 3000)
- Microservice แปลง HTML → PDF ผ่าน Chrome headless
- เพิ่มเข้า docker-compose แล้ว
- Document service ส่ง HTML ไปแล้วได้ PDF กลับมา

---

## Git Flow

```
main       ← production (ลูกค้าใช้) — แตะได้แค่ตอน stable จริงๆ
  ↑ PR เมื่อพร้อม
develop    ← staging (เทสรวมกันก่อน) — ถ้าพังก็พังตรงนี้
  ↑ PR เมื่อ feature เสร็จ
feat/xxx   ← development (สร้าง feature ใหม่)
```

**กฎ:**
- `main` — ห้าม push ตรง ต้องผ่าน PR จาก develop เท่านั้น
- `develop` — ห้าม push ตรง ต้องผ่าน PR จาก feature branch
- `feat/xxx` — ทำงานได้เสรี

---

## สิ่งที่ยังไม่ได้ทำ (Technical Debt)

- Unit tests ทุก service
- Integration tests
- Logging / Monitoring
- ทดสอบกับ LINE API จริงๆ
- Load test gateway
- Security audit JWT และ rate limit

---

## Phase ถัดไป

**Phase 3:**
- Report service
- Config service  
- Admin panel / Frontend (React หรือ LINE LIFF)
- Google Sheets sync
- Trigger engine

**Phase 4:**
- Line webhook
- AI parser
- OCR
