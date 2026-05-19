"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";

interface ConfirmPanelProps {
  jobId: string;
  open: boolean;
  onClose: () => void;
}

export function ConfirmPanel({ jobId, open, onClose }: ConfirmPanelProps) {
  const queryClient = useQueryClient();

  const { mutate, isPending } = useMutation({
    mutationFn: async () => {
      await api.patch(`/v1/jobs/${jobId}/status`, { status: "done" });
    },
    onSuccess: () => {
      toast.success("ยืนยันงานเสร็จสิ้นแล้ว");
      queryClient.invalidateQueries({ queryKey: ["job", jobId] });
      queryClient.invalidateQueries({ queryKey: ["jobs"] });
      onClose();
    },
    onError: () => {
      toast.error("เกิดข้อผิดพลาด กรุณาลองใหม่");
    },
  });

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>ยืนยันงานเสร็จ</SheetTitle>
        </SheetHeader>

        <div className="mt-6 space-y-4">
          <p className="text-sm text-muted-foreground">
            กดยืนยันเพื่อเปลี่ยนสถานะงานเป็น "เสร็จแล้ว" การดำเนินการนี้ไม่สามารถย้อนกลับได้
          </p>
          <Button
            className="w-full"
            disabled={isPending}
            onClick={() => mutate()}
          >
            {isPending ? "กำลังบันทึก..." : "ยืนยันงานเสร็จ"}
          </Button>
          <Button variant="outline" className="w-full" onClick={onClose}>
            ยกเลิก
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
