import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PUBLIC_PATHS = ["/login"];

const OFFICE_PATHS = ["/jobs", "/materials", "/users"];
const FIELD_PATHS = ["/my-jobs"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // ถ้าเป็น public path ผ่านเลย
  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  const token = request.cookies.get("access_token")?.value;

  // ไม่มี token → redirect ไป login
  if (!token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // decode role จาก JWT (ไม่ต้อง verify เพราะ gateway จัดการแล้ว)
  try {
    const payload = JSON.parse(
      Buffer.from(token.split(".")[1], "base64").toString()
    );

    const role = payload.role as string | undefined;
    const isTechnician = role === "technician";

    // ช่างพยายามเข้า office path → redirect ไป my-jobs
    if (isTechnician && OFFICE_PATHS.some((p) => pathname.startsWith(p))) {
      return NextResponse.redirect(new URL("/my-jobs", request.url));
    }

    // office/manager พยายามเข้า field path → redirect ไป jobs
    if (!isTechnician && FIELD_PATHS.some((p) => pathname.startsWith(p))) {
      return NextResponse.redirect(new URL("/jobs", request.url));
    }

    return NextResponse.next();
  } catch {
    return NextResponse.redirect(new URL("/login", request.url));
  }
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|api).*)"],
};
