export interface User {
  uid: string;
  email: string;
  name: string;
}

export interface Membership {
  ID: string;
  UserID: string;
  CompanyID: string;
  CompanyName: string;
  Role: "admin" | "manager" | "office" | "technician";
  JobTypeScope: string[];
  IsActive: boolean;
}

export interface AuthState {
  user: User | null;
  memberships: Membership[];
  activeMembership: Membership | null;
  accessToken: string | null;
  refreshToken: string | null;
}
export interface Job {
  ID: string;
  CompanyID: string;
  CreatedBy: string;
  JobCode: string;
  RefCode: string;
  JobType: string;
  Status: "pending" | "assigned" | "in_progress" | "done" | "cancelled";
  Cause: string;
  LocationText: string;
  Subdistrict: string;
  District: string;
  Province: string;
  ContactTechnician: string;
  CostMain: number;
  CostSurface: number;
  NotifiedAt: string | null;
  StartedAt: string | null;
  EndedAt: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
