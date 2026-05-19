"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useMembers } from "@/hooks/useMembers";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Label } from "@/components/ui/label";

interface AssignPanelProps {
  jobId: string;
  open: boolean;
  onClose: () => void;
}

export function AssignPanel({ jobId, open, onClose }: AssignPanelProps) {
  const queryClient = useQueryClient();
  const { data: members } = useMembers();
  const [technicianId, setTechnicianId] = useState("");

  const technicians = members?.filter((m) => m.Role === "technician") ?? [];

  const { mutate, isPending } = useMutation({
    mutationFn: async () => {
      await api.post(`/v1/jobs/${jobId}/assign`, {
        technician_id: technicianId,
        assignment_type: "main",
      });
    },
    onSuccess: () => {
      toast.success("มอบหมายงานสำเร็จ");
      queryClient.invalidateQueries({ queryKey: ["job", jobId] });
      onClose();
    },
    onError: () => {
      toast.error("มอบหมายงานไม่สำเร็จ");
    },
  });

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>มอบหมายงาน</SheetTitle>
        </SheetHeader>

        <div className="mt-6 space-y-4">
          <div className="space-y-1.5">
            <Label>เลือกช่าง</Label>
            {technicians.length === 0 ? (
              <p className="text-sm text-muted-foreground">ไม่มีช่างในระบบ</p>
            ) : (
              <Select value={technicianId} onValueChange={setTechnicianId}>
                <SelectTrigger>
                  <SelectValue placeholder="เลือกช่าง..." />
                </SelectTrigger>
                <SelectContent>
                  {technicians.map((t) => (
                    <SelectItem key={t.ID} value={t.ID}>
                      {t.Name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          <Button
            className="w-full"
            disabled={!technicianId || isPending}
            onClick={() => mutate()}
          >
            {isPending ? "กำลังมอบหมาย..." : "มอบหมายงาน"}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
