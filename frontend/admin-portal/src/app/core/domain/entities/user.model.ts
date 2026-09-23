export interface AdminUser {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phone: string;
  userType: 'provider' | 'seeker' | string;
  isActive: boolean;
  isStaff: boolean;
  isVerified: boolean;
  isIdentityVerified: boolean;
  subscriptionTier: string;
  subscriptionInterval?: string;
  preferredPaymentMode?: string;
  checkrStatus: string;
  bio?: string;
  city?: string;
  state?: string;
  zipCode?: string;
  address?: string;
  country?: string;
  gender?: string;
  service?: string;
  profilePicture?: string;
  createdAt: string;
  updatedAt?: string;
  walletBalance?: number;
  roleId?: string;
  roleName?: string;
}

export interface ImpersonationResult {
  accessToken: string;
  expiresIn: number;
  targetUserId: string;
  targetEmail: string;
}

export interface StaffNote {
  id: string;
  user_id: string;
  author_admin_id: string;
  author_email: string;
  author_name?: string;
  content: string;
  category: 'GENERAL' | 'FRAUD' | 'COMPLIANCE' | 'BILLING' | string;
  is_pinned: boolean;
  created_at: string;
  updated_at?: string;
}

export interface ProviderFunnelData {
  total_signed_up: number;
  id_uploaded: number;
  checkr_completed: number;
  stripe_connected: number;
  first_booking_done: number;
  conversion_rate_pct: number;
}

export interface AdminNotificationItem {
  id: string;
  type: string;
  severity: 'HIGH' | 'MEDIUM' | 'INFO';
  title: string;
  message: string;
  resource_id?: string;
  route?: string;
  is_read: boolean;
  created_at: string;
}
