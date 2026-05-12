# CI Troubleshooting — Go Workspace + Docker Build

บันทึกปัญหาที่เจอระหว่าง setup CI สำหรับ Go monorepo
ใช้เวลาแก้นานมากเพราะแต่ละปัญหาซ่อนปัญหาถัดไปไว้

---

## ภาพรวมของปัญหา

ระบบใช้ Go Workspace (go.work) สำหรับ monorepo
ซึ่งมีความซับซ้อนในการ build ด้วย Docker และ CI
เพราะ go.work ต้องการ context ที่ครบถ้วนจากทุก module

---

## ปัญหาที่ 1 — go.sum ไม่มีใน repo

### อาการ
Restore cache failed: Dependencies file is not found
Supported file pattern: go.sum
### สาเหตุ
services ส่วนใหญ่ยังไม่เคยรัน `go mod tidy`
ทำให้ไม่มี go.sum เลย และ `actions/setup-go` หา go.sum
ที่ root ไม่เจอจึง fail

### วิธีแก้
```bash
for svc in auth job user ...; do
  cd services/$svc && go mod tidy && cd ../..
done
```

### บทเรียน
ทุกครั้งที่สร้าง service ใหม่ต้องรัน `go mod tidy` ทันที
ก่อน commit ครั้งแรก

---

## ปัญหาที่ 2 — go.work.sum ไม่มี

### อาการ
go: cannot load module services/auth listed in go.work file
### สาเหตุ
`go work sync` ไม่ได้สร้าง go.work.sum ให้อัตโนมัติ
ต้องใช้ `go mod download all` แทน

### วิธีแก้
```bash
go mod download all
# จะสร้าง go.work.sum ให้อัตโนมัติ
```

### บทเรียน
`go work sync` vs `go mod download all` ต่างกัน
- `go work sync` — sync workspace modules เท่านั้น
- `go mod download all` — download dependencies และสร้าง go.work.sum

---

## ปัญหาที่ 3 — vendor folder มี .gitignore ซ้อนอยู่

### อาการ
```bash
git add vendor/
git status  # ไม่มีอะไรใน staging area เลย
```

### สาเหตุ
packages บางตัวเช่น gin, jwt, pgx มี `.gitignore`
ของตัวเองซ้อนอยู่ใน vendor/ ทำให้ git ไม่ track files ข้างใน
แม้จะ `git add -f` ก็ยังไม่ได้

### วิธีแก้
ไม่ใช้ vendor approach สำหรับ CI ครับ
ใช้ `go mod download` ใน Dockerfile และ CI แทน

### บทเรียน
vendor mode เหมาะกับ single module ไม่ใช่ Go workspace
เพราะ packages มักมี .gitignore ของตัวเองที่ขัดกับ git add

---

## ปัญหาที่ 4 — go version ใน go.mod ไม่ตรงกับ dependencies

### อาการ
go: github.com/gin-gonic/gin@v1.12.0 requires go >= 1.25.0
go: github.com/jackc/pgx/v5@v5.9.2 requires go >= 1.25.0
(running go 1.23.12; GOTOOLCHAIN=local)
### สาเหตุ
ตั้งต้น lock go version ไว้ที่ 1.23 แต่ dependencies
ที่ติดตั้งไป (gin, pgx, crypto) version ล่าสุดต้องการ
go 1.24-1.25 ทำให้ build fail ทุก environment

### วิธีแก้
```bash
# อัปเกรด go version ทั้งหมดให้ตรงกับ dependency requirement
sed -i 's/^go 1\.23/go 1.25/' go.work
for svc in auth job ...; do
  sed -i 's/^go 1\.23/go 1.25/' services/$svc/go.mod
done
# อัปเดต CI และ Dockerfile ด้วย
```

### บทเรียน
เมื่อติดตั้ง dependency ใหม่ต้องเช็คก่อนว่า
dependency นั้นต้องการ Go version เท่าไหร่
วิธีง่ายที่สุดคือดูที่ go.mod ของ package นั้นใน pkg.go.dev

---

## ปัญหาที่ 5 — go.work version ไม่ sync กับ go.mod

### อาการ
go: module services/auth requires go >= 1.25.0,
but go.work lists go 1.23
### สาเหตุ
`go work use` จะ update go.work version ให้ match
กับ go.mod ที่สูงที่สุดใน workspace อัตโนมัติ
แต่เราแก้ go.mod ด้วย sed แล้ว go.work ไม่ได้ update ตาม

### วิธีแก้
```bash
sed -i 's/^go 1\.23/go 1.25/' go.work
# หรือรัน go work use เพื่อให้ auto-detect
go work use ./services/...
```

### บทเรียน
go.work version ต้อง >= go.mod version ที่สูงสุดใน workspace
เสมอ ถ้าเปลี่ยน go.mod ต้องเปลี่ยน go.work ด้วย

---

## สรุป Root Cause ทั้งหมด

ปัญหาทั้งหมดมาจากจุดเดียวกัน คือ

**Go version ที่เลือกตั้งแต่แรก (1.23) ต่ำกว่าที่ dependencies ต้องการ**

ส่งผลเป็น chain reaction:
1. `go mod tidy` ดึง dependency version ใหม่มา (ที่ต้องการ 1.25)
2. go.mod บอกว่าต้องการ 1.25
3. go.work ยังบอก 1.23 → conflict
4. Docker build ใช้ go 1.23 → ไม่สามารถ download dependencies ได้
5. CI ก็ fail ด้วยเหตุผลเดียวกัน

---

## Checklist สำหรับ Service ใหม่

ทุกครั้งที่สร้าง service ใหม่ใน monorepo

```bash
# 1. เช็ค go version ของ dependencies ก่อนติดตั้ง
# ดูที่ https://pkg.go.dev/<package>

# 2. รัน go mod tidy ทันทีหลัง go get
cd services/<new-service>
go get <package>
go mod tidy

# 3. sync go.work
cd ../..
go work use ./services/<new-service>

# 4. สร้าง go.work.sum
go mod download all

# 5. ตรวจ version ทุกอัน
head -2 go.work
head -2 services/<new-service>/go.mod
```

---

## Key Takeaway

> **เลือก Go version ให้สูงพอตั้งแต่แรก**
> ดู dependency ที่จะใช้ก่อน แล้วเลือก version ที่รองรับทุกตัว
> อย่า lock version ต่ำโดยไม่มีเหตุผล
> การ downgrade ทีหลังยากกว่า upgrade มาก
