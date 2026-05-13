package service

import (
"context"
"fmt"
"strings"
"time"

"github.com/google/uuid"
)

// GenerateJobCode สร้างรหัสงานตาม format ของบริษัท
// format tokens:
//   {MM}  = เดือน 2 หลัก
//   {DD}  = วัน 2 หลัก
//   {YY}  = ปี 2 หลัก (พ.ศ.)
//   {XXX} = เลขรัน 3 หลัก
func (s *JobService) GenerateJobCode(ctx context.Context, companyID uuid.UUID, jobType string, format string) (string, error) {
now := time.Now()

// คำนวณปี พ.ศ.
yearBE := now.Year() + 543
yy := fmt.Sprintf("%02d", yearBE%100)
mm := fmt.Sprintf("%02d", int(now.Month()))
dd := fmt.Sprintf("%02d", now.Day())

// period สำหรับ sequence counter
// ถ้า format มี {DD} ใช้ period เป็นวัน ถ้าไม่มีใช้เดือน
var period string
if strings.Contains(format, "{DD}") {
period = fmt.Sprintf("%s%s%s", yy, mm, dd)
} else {
period = fmt.Sprintf("%s%s", yy, mm)
}

seq, err := s.repo.NextSequence(ctx, companyID, jobType, period)
if err != nil {
return "", fmt.Errorf("next sequence: %w", err)
}

xxx := fmt.Sprintf("%03d", seq)

code := format
code = strings.ReplaceAll(code, "{MM}", mm)
code = strings.ReplaceAll(code, "{DD}", dd)
code = strings.ReplaceAll(code, "{YY}", yy)
code = strings.ReplaceAll(code, "{XXX}", xxx)

return code, nil
}
