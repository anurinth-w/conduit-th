"use client";

import { use } from "react";
import { useJob } from "@/hooks/useJob";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { AssignPanel } from "@/components/jobs/AssignPanel";
import { useState } from "react";
import { ConfirmPanel } from "@/components/jobs/ConfirmPanel";



const STATUS_LABEL: Record<string, string> = {
  pending: "รอดำเนินการ",
  assigned: "มอบหมายแล้ว",
  in_progress: "กำลังดำเนินการ",
  done: "เสร็จแล้ว",
  cancelled: "ยกเลิก",
};

function Row({ label, value }: { label: string; value?: string | number | null }) {
  if (!value) return null;
  return (
    <div className="flex gap-4 py-2 border-b last:border-0">
      <span className="text-sm text-muted-foreground w-40 shrink-0">{label}</span>
      <span className="text-sm">{value}</span>
    </div>
  );
}

export default function JobDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [assignOpen, setAssignOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { data: job, isLoading, isError } = useJob(id);
  
  if (isLoading) {
    return (
      <div className="p-8 space-y-3">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-4 w-32" />
        {[...Array(8)].map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    );
  }

  if (isError || !job) {
    return (
      <div className="p-8">
        <p className="text-destructive">โหลดข้อมูลไม่สำเร็จ</p>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-3xl">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <Link href="/jobs">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-medium">{job.JobCode}</h1>
            <Badge>{STATUS_LABEL[job.Status] ?? job.Status}</Badge>
          </div>
          <p className="text-sm text-muted-foreground mt-0.5">{job.JobType}</p>
        </div>
      </div>

      {/* Detail */}
      <div className="border rounded-lg px-4 mb-6">
        <Row label="สาเหตุ" value={job.Cause} />
        <Row label="สถานที่" value={job.LocationText} />
        <Row label="ตำบล" value={job.Subdistrict} />
        <Row label="อำเภอ" value={job.District} />
        <Row label="จังหวัด" value={job.Province} />
        <Row label="ช่างผู้รับผิดชอบ" value={job.ContactTechnician} />
        <Row label="ผู้ประสานงาน" value={job.ContactCoordinator} />
        <Row label="ประเภทท่อ" value={job.PipeType} />
        <Row label="ขนาดท่อ (มม.)" value={job.PipeSizeMM || null} />
        <Row label="วิธีทำงาน" value={job.WorkMethod} />
        <Row label="สภาพผิวจราจร" value={job.SurfaceCondition} />
        <Row label="ค่าใช้จ่ายหลัก" value={job.CostMain ? `฿${job.CostMain.toLocaleString()}` : null} />
        <Row label="ค่าผิวจราจร" value={job.CostSurface ? `฿${job.CostSurface.toLocaleString()}` : null} />
        <Row label="วันที่สร้าง" value={new Date(job.CreatedAt).toLocaleDateString("th-TH")} />
      </div>

      {/* Actions */}
      <div className="flex gap-3 flex-wrap">
        <Link href={`/jobs/${id}/photos`}>
          <Button variant="outline">รูปถ่าย</Button>
        </Link>
        <Link href={`/jobs/${id}/materials`}>
          <Button variant="outline">วัสดุ</Button>
        </Link>
        <Link href={`/jobs/${id}/document`}>
          <Button variant="outline">สร้าง PDF</Button>
        </Link>
        <Button onClick={() => setAssignOpen(true)}>มอบหมายงาน</Button>
        <Button variant="outline" onClick={() => setConfirmOpen(true)}>ยืนยันงานเสร็จ</Button>
      </div>
      <AssignPanel jobId={id} open={assignOpen} onClose={() => setAssignOpen(false)} />
      <ConfirmPanel jobId={id} open={confirmOpen} onClose={() => setConfirmOpen(false)} />
    </div>
  );
}
