"use client";

import { use, useState } from "react";
import { useJobMaterials, useAddMaterial, useDeleteMaterial } from "@/hooks/useJobMaterials";
import { useAuthStore } from "@/stores/auth";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import Link from "next/link";
import { ArrowLeft, Plus, Trash2 } from "lucide-react";

export default function MaterialsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { activeMembership } = useAuthStore();
  const canSeePrice = (activeMembership?.Role === "admin" || activeMembership?.Role === "manager");

  const { data: materials, isLoading } = useJobMaterials(id);
  const { mutate: addMaterial, isPending: adding } = useAddMaterial(id);
  const { mutate: deleteMaterial } = useDeleteMaterial(id);

  const [form, setForm] = useState({ name: "", unit: "", quantity: "1", unit_price: "0", labor_cost: "0" });

  function handleAdd() {
    if (!form.name || !form.unit || !form.quantity) {
      toast.error("กรอกข้อมูลให้ครบ");
      return;
    }
    addMaterial({
      name: form.name,
      unit: form.unit,
      quantity: parseFloat(form.quantity),
      unit_price: canSeePrice ? parseFloat(form.unit_price) : 0,
      labor_cost: canSeePrice ? parseFloat(form.labor_cost) : 0,
    }, {
      onSuccess: () => {
        toast.success("เพิ่มวัสดุแล้ว");
        setForm({ name: "", unit: "", quantity: "1", unit_price: "0", labor_cost: "0" });
      },
      onError: () => toast.error("เพิ่มวัสดุไม่สำเร็จ"),
    });
  }

  if (isLoading) {
    return <div className="p-8 space-y-3">{[...Array(4)].map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}</div>;
  }

  const total = materials?.reduce((sum, m) => sum + m.Total, 0) ?? 0;

  return (
    <div className="p-8 max-w-4xl">
      <div className="flex items-center gap-3 mb-6">
        <Link href={`/jobs/${id}`}>
          <Button variant="ghost" size="icon"><ArrowLeft className="w-4 h-4" /></Button>
        </Link>
        <h1 className="text-xl font-medium">วัสดุที่ใช้</h1>
      </div>

      {/* Add form */}
      <div className="border rounded-lg p-4 mb-6">
        <h2 className="text-sm font-medium mb-3">เพิ่มวัสดุ</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div className="col-span-2 space-y-1">
            <Label>ชื่อวัสดุ</Label>
            <Input placeholder="เช่น ท่อ PVC 15 มม." value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
          </div>
          <div className="space-y-1">
            <Label>หน่วย</Label>
            <Input placeholder="เมตร / หัว" value={form.unit}
              onChange={e => setForm(f => ({ ...f, unit: e.target.value }))} />
          </div>
          <div className="space-y-1">
            <Label>จำนวน</Label>
            <Input type="number" value={form.quantity}
              onChange={e => setForm(f => ({ ...f, quantity: e.target.value }))} />
          </div>
          {canSeePrice && (
            <>
              <div className="space-y-1">
                <Label>ราคา/หน่วย</Label>
                <Input type="number" value={form.unit_price}
                  onChange={e => setForm(f => ({ ...f, unit_price: e.target.value }))} />
              </div>
              <div className="space-y-1">
                <Label>ค่าแรง/หน่วย</Label>
                <Input type="number" value={form.labor_cost}
                  onChange={e => setForm(f => ({ ...f, labor_cost: e.target.value }))} />
              </div>
            </>
          )}
        </div>
        <Button className="mt-3" onClick={handleAdd} disabled={adding}>
          <Plus className="w-4 h-4 mr-2" />
          {adding ? "กำลังเพิ่ม..." : "เพิ่มวัสดุ"}
        </Button>
      </div>

      {/* Table */}
      {!materials || materials.length === 0 ? (
        <p className="text-muted-foreground">ยังไม่มีวัสดุ</p>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ชื่อวัสดุ</TableHead>
                <TableHead>หน่วย</TableHead>
                <TableHead className="text-right">จำนวน</TableHead>
                {canSeePrice && <TableHead className="text-right">ราคา/หน่วย</TableHead>}
                {canSeePrice && <TableHead className="text-right">ค่าแรง/หน่วย</TableHead>}
                {canSeePrice && <TableHead className="text-right">รวม</TableHead>}
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {materials.map((m) => (
                <TableRow key={m.ID}>
                  <TableCell>{m.Name}</TableCell>
                  <TableCell>{m.Unit}</TableCell>
                  <TableCell className="text-right">{m.Quantity}</TableCell>
                  {canSeePrice && <TableCell className="text-right">฿{m.UnitPrice.toLocaleString()}</TableCell>}
                  {canSeePrice && <TableCell className="text-right">฿{m.LaborCost.toLocaleString()}</TableCell>}
                  {canSeePrice && <TableCell className="text-right">฿{m.Total.toLocaleString()}</TableCell>}
                  <TableCell>
                    <Button variant="ghost" size="icon"
                      onClick={() => deleteMaterial(m.ID, { onSuccess: () => toast.success("ลบแล้ว") })}>
                      <Trash2 className="w-4 h-4 text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {canSeePrice && (
            <div className="px-4 py-3 border-t text-right text-sm font-medium">
              รวมทั้งหมด: ฿{total.toLocaleString()}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
