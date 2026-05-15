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