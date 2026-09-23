export interface AdminAuthResponse {
  token: string;
  refreshToken?: string;
  requires2FA?: boolean;
  tempToken?: string;
  user: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    isStaff: boolean;
    isSuperuser?: boolean;
  };
}
