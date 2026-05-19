"use client";

import { useJobs } from "@/hooks/useJobs";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import Link from "next/link";

const STATUS_LABEL: Record<string, string> = {
  pending: "รอดำเนินการ",
  assigned: "มอบหมายแล้ว",
  in_progress: "กำลังดำเนินการ",
  done: "เสร็จแล้ว",
  cancelled: "ยกเลิก",
};

const STATUS_COLOR: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  pending: "outline",
  assigned: "secondary",
  in_progress: "default",
  done: "secondary",
  cancelled: "destructive",
};

export default function JobsPage() {
  const { data: jobs, isLoading, isError } = useJobs();

  if (isLoading) {
    return (
      <div className="p-8 space-y-3">
        <Skeleton className="h-8 w-48" />
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8">
        <p className="text-destructive">โหลดข้อมูลไม่สำเร็จ</p>
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-medium mb-6">รายการงาน</h1>

      {!jobs || jobs.length === 0 ? (
        <p className="text-muted-foreground">ไม่มีรายการงาน</p>
      ) : (
        <div className="space-y-3">
          {jobs.map((job) => (
            <Link
              key={job.ID}
              href={`/jobs/${job.ID}`}
              className="block border rounded-lg p-4 hover:bg-muted/50 transition-colors"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium text-sm">{job.JobCode}</span>
                    <Badge variant={STATUS_COLOR[job.Status]}>
                      {STATUS_LABEL[job.Status] ?? job.Status}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground truncate">{job.Cause}</p>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate">{job.LocationText}</p>
                </div>
                <span className="text-xs text-muted-foreground shrink-0">
                  {new Date(job.CreatedAt).toLocaleDateString("th-TH")}
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
