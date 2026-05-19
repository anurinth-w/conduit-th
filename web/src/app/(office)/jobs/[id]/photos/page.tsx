"use client";

import { use, useRef, useState } from "react";
import { usePhotos, useTogglePhoto } from "@/hooks/usePhotos";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { ArrowLeft, CheckCircle, Circle, Upload } from "lucide-react";
import type { JobPhoto } from "@/types/api";

const STAGES = [
  { key: "before", label: "ก่อนซ่อม" },
  { key: "during", label: "ระหว่างซ่อม" },
  { key: "after",  label: "หลังซ่อม" },
];

function PhotoBlock({
  stage,
  label,
  photos,
  onToggle,
  onUpload,
  uploading,
}: {
  stage: string;
  label: string;
  photos: JobPhoto[];
  onToggle: (id: string, selected: boolean) => void;
  onUpload: (file: File, stage: string) => void;
  uploading: boolean;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="border rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <h2 className="font-medium text-sm">{label}</h2>
          <Badge variant="secondary">{photos.length} รูป</Badge>
        </div>
        <div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) onUpload(file, stage);
              e.target.value = "";
            }}
          />
          <Button
            size="sm"
            variant="outline"
            disabled={uploading}
            onClick={() => fileInputRef.current?.click()}
          >
            <Upload className="w-3.5 h-3.5 mr-1.5" />
            อัปโหลด
          </Button>
        </div>
      </div>

      {photos.length === 0 ? (
        <p className="text-sm text-muted-foreground py-4 text-center">ยังไม่มีรูป</p>
      ) : (
        <div className="grid grid-cols-3 md:grid-cols-4 gap-2">
          {photos.map((photo) => (
            <div
              key={photo.ID}
              className="relative cursor-pointer"
              onClick={() => onToggle(photo.ID, !photo.IsSelected)}
            >
              <img
                src={photo.URL}
                alt={photo.Caption || label}
                className="w-full aspect-square object-cover rounded-md border"
              />
              <div className="absolute top-1 right-1">
                {photo.IsSelected
                  ? <CheckCircle className="w-5 h-5 text-primary fill-white" />
                  : <Circle className="w-5 h-5 text-white opacity-60" />
                }
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function PhotosPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: photos, isLoading } = usePhotos(id);
  const { mutate: togglePhoto } = useTogglePhoto(id);
  const queryClient = useQueryClient();
  const [uploadingStage, setUploadingStage] = useState<string | null>(null);

  const { mutate: uploadPhoto } = useMutation({
    mutationFn: async ({ file, stage }: { file: File; stage: string }) => {
      setUploadingStage(stage);
      const form = new FormData();
      form.append("file", file);
      form.append("stage", stage);
      await api.post(`/v1/jobs/${id}/photos`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
    },
    onSuccess: () => {
      toast.success("อัปโหลดรูปสำเร็จ");
      queryClient.invalidateQueries({ queryKey: ["photos", id] });
      setUploadingStage(null);
    },
    onError: () => {
      toast.error("อัปโหลดไม่สำเร็จ");
      setUploadingStage(null);
    },
  });

  if (isLoading) {
    return (
      <div className="p-8 space-y-4">
        {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-40 w-full rounded-lg" />)}
      </div>
    );
  }

  const selectedCount = photos?.filter(p => p.IsSelected).length ?? 0;

  return (
    <div className="p-8 max-w-4xl">
      <div className="flex items-center gap-3 mb-6">
        <Link href={`/jobs/${id}`}>
          <Button variant="ghost" size="icon">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>
        <h1 className="text-xl font-medium">รูปถ่าย</h1>
        <Badge variant="secondary">{selectedCount} รูปที่เลือกแล้ว</Badge>
      </div>

      <div className="space-y-4">
        {STAGES.map(({ key, label }) => (
          <PhotoBlock
            key={key}
            stage={key}
            label={label}
            photos={photos?.filter(p => p.Stage === key) ?? []}
            onToggle={(photoId, selected) => togglePhoto({ photoId, selected })}
            onUpload={(file, stage) => uploadPhoto({ file, stage })}
            uploading={uploadingStage === key}
          />
        ))}
      </div>
    </div>
  );
}
