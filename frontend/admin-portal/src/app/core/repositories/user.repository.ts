import { Observable } from 'rxjs';
import { AdminUser, ImpersonationResult, StaffNote, ProviderFunnelData, AdminNotificationItem } from '../domain/entities/user.model';

export interface ListUsersParams {
  search?: string;
  userType?: string;
  isStaff?: boolean;
  isActive?: boolean;
  isVerified?: boolean;
  isIdentityVerified?: boolean;
  subscriptionTier?: string;
  page?: number;
  pageSize?: number;
}

export abstract class UserRepository {
  abstract listUsers(params?: ListUsersParams): Observable<{ results: AdminUser[]; count: number }>;
  abstract getUserById(id: string): Observable<AdminUser>;
  abstract createUser(data: any): Observable<AdminUser>;
  abstract updateUser(id: string, data: any): Observable<AdminUser>;
  abstract deleteUser(id: string): Observable<{ success: boolean }>;
  abstract restoreUser(id: string): Observable<{ success: boolean }>;
  abstract impersonateUser(id: string): Observable<ImpersonationResult>;
  abstract getGDPRUserData(id: string): Observable<any>;
  abstract assignUserRole(userId: string, roleId: string): Observable<{ success: boolean }>;
  abstract listStaffNotes(userId: string): Observable<StaffNote[]>;
  abstract createStaffNote(userId: string, data: { content: string; category?: string; is_pinned?: boolean }): Observable<StaffNote>;
  abstract deleteStaffNote(userId: string, noteId: string): Observable<{ status: string }>;
  abstract getProviderFunnel(): Observable<ProviderFunnelData>;
  abstract getNotificationFeed(): Observable<{ notifications: AdminNotificationItem[]; count: number; unread_count: number }>;
  abstract sendDirectMessage(userId: string, title: string, message: string, channel?: string): Observable<{ status: string; sent_at: string }>;
  abstract evaluateUserRisk(userId: string): Observable<any>;
}
