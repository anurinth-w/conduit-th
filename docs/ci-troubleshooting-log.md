# CI Troubleshooting Log — Initial Setup

บันทึกปัญหาที่เจอและวิธีแก้ระหว่าง setup CI pipeline ครั้งแรก
เขียนไว้เพื่อให้เข้าใจ "ทำไม" ไม่ใช่แค่ "ทำอะไร"

---

## ปัญหาที่ 1 — GitHub Actions matrix condition ใช้ matrix.* ใน if ไม่ได้

### อาการ
Invalid workflow file
(Line: 118, Col: 9): Unrecognized named-value: 'matrix'.
Located at position 1 within expression: matrix.changed == 'true'
### สาเหตุ
GitHub Actions มีข้อจำกัดว่า expression `matrix.*` ใช้ได้แค่ใน
step level เท่านั้น ไม่สามารถใช้ใน `if` ระดับ job ได้
เพราะ job-level `if` ถูก evaluate ก่อนที่ matrix จะถูก expand

### วิธีคิด
ถ้า syntax ไม่รองรับ ต้องเปลี่ยน approach
แทนที่จะใช้ matrix + if ให้แยกเป็น job ย่อยแต่ละ service
แล้วใช้ `needs.changes.outputs.<service>` โดยตรงแทน

### วิธีแก้
แยก test job ออกเป็น `test-auth`, `test-job`, `test-user` ฯลฯ
แต่ละ job มี if condition ของตัวเองที่ reference output จาก changes job โดยตรง

```yaml
# ❌ ไม่ได้
test:
  if: ${{ matrix.changed == 'true' }}  # matrix ยังไม่ถูก expand ตอนนี้

# ✅ ได้
test-auth:
  if: needs.changes.outputs.auth == 'true'  # reference โดยตรง
```

---

## ปัญหาที่ 2 — Lint fail: expected 'package', found 'EOF'

### อาการ
expected 'package', found 'EOF' (typecheck)
### สาเหตุ
`main.go` ของทุก service ถูกสร้างด้วย `touch` ซึ่งสร้างไฟล์เปล่า
Go linter และ compiler ต้องการ `package` declaration เป็นบรรทัดแรกเสมอ
ไฟล์เปล่า = invalid Go file

### วิธีคิด
Go ทุกไฟล์ต้องขึ้นต้นด้วย `package <name>` เสมอ
ถึงแม้จะยังไม่มี code ก็ต้องมี declaration นี้

### วิธีแก้
```go
package main

func main() {}
```
ใส่ขั้นต่ำนี้ไว้ก่อน แล้วค่อยเพิ่ม code จริงทีหลัง

---

## ปัญหาที่ 3 — Docker build ไม่เจอ shared/ และ go.work

### อาการ
failed to solve: process "/bin/sh -c cd services/auth && go build -o /bin/service ."
did not complete successfully: exit code: 1
### สาเหตุ
Docker build context คือชุดไฟล์ที่ Docker มองเห็นระหว่าง build
Dockerfile เดิม copy เฉพาะ `services/auth/` แต่ service นั้น
import `shared/` และต้องการ `go.work` เพื่อ resolve workspace
Docker ไม่มีไฟล์เหล่านั้นเลยทำให้ build fail

### วิธีคิด
วาด dependency graph ของ service ก่อน
services/auth/main.go
└── import shared/pkg/...  ← ต้องการ shared/
└── go.work                ← ต้องการ go.work
Docker ต้องมีครบทุกอย่างที่ Go ต้องการ

### วิธีแก้
```dockerfile
# copy ตาม dependency order
COPY go.work ./          # workspace definition
COPY shared/ shared/     # shared packages
COPY services/auth/ services/auth/  # target service
RUN go build -o /bin/service ./services/auth/
```

### บทเรียน
เรียง COPY จาก "เปลี่ยนน้อย → เปลี่ยนบ่อย"
เพื่อให้ Docker layer cache ทำงานได้ดีที่สุด
`go.work` เปลี่ยนน้อยสุด → `shared/` → `services/<svc>/` เปลี่ยนบ่อยสุด

---

## ปัญหาที่ 4 — go.work ต้องการ go.mod ของทุก service

### อาการ
go: cannot load module services/auth listed in go.work file:
open services/auth/go.mod: no such file or directory
### สาเหตุ
`go.work` list ทุก service ไว้ เมื่อ Go เริ่ม build มันจะ
load go.work ก่อนแล้ว verify ว่าทุก module ที่ list ไว้มี go.mod จริง
แม้จะ build แค่ service เดียว Go ก็ยังต้องการ go.mod ของทุก service
เพื่อ resolve dependency graph ทั้งหมด

### วิธีคิด
go.work ทำงานแบบ "all or nothing"
ถ้า list service ไว้ใน go.work ต้องมี go.mod ของ service นั้นเสมอ

### วิธีแก้
Copy go.mod ของทุก service เข้า Docker แต่ copy source code
เฉพาะ service ที่จะ build เท่านั้น

```dockerfile
# go.mod ทุก service (go.work ต้องการ)
COPY services/auth/go.mod services/auth/go.mod
COPY services/job/go.mod services/job/go.mod
# ... ทุก service

# source code เฉพาะ service นี้
COPY services/auth/ services/auth/
```

---

## ปัญหาที่ 5 — Push ไป GHCR ไม่ได้: installation not allowed to Create organization package

### อาการ
denied: installation not allowed to Create organization package
### สาเหตุ
GitHub Actions ใช้ `GITHUB_TOKEN` เพื่อ authenticate กับ GHCR
default permission ของ token คือ "Read" เท่านั้น
การ push image ต้องการ "Write" permission

### วิธีคิด
Permission error มักมาจาก 2 ที่คือ
1. Token ไม่มี permission พอ
2. Repo settings ไม่อนุญาต

ในกรณีนี้เป็นข้อ 2 — ต้องแก้ที่ repo settings

### วิธีแก้
Settings → Actions → General → Workflow permissions
เปลี่ยนเป็น "Read and write permissions"

---

## สรุปวิธีคิดในการ debug CI

1. **อ่าน error message ทั้งหมด** อย่าดูแค่บรรทัดสุดท้าย
   error จริงมักอยู่กลาง log ไม่ใช่ท้าย log

2. **วาด dependency** ก่อนแก้ เช่น service ต้องการอะไร
   Docker เห็นอะไร มีครบไหม

3. **แก้ทีละปัญหา** อย่า commit หลายอย่างพร้อมกัน
   จะได้รู้ว่าอะไรแก้แล้วได้ผล

4. **อ่าน log ดิบ** อย่าดูแค่ summary
   กด expand แต่ละ step เพื่อดู error จริง

5. **Permission error** มักแก้ที่ settings ไม่ใช่ code
