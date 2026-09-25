import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { UserRepository, ListUsersParams } from '../../core/repositories/user.repository';
import { AdminUser, ImpersonationResult } from '../../core/domain/entities/user.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class UserRepositoryImpl implements UserRepository {
  constructor(private http: HttpClient) {}

  listUsers(params?: ListUsersParams): Observable<{ results: AdminUser[]; count: number }> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}?page=${params?.page || 1}&page_size=${params?.pageSize || 20}`;
    if (params?.search) url += `&search=${encodeURIComponent(params.search)}`;
    if (params?.userType) url += `&user_type=${encodeURIComponent(params.userType)}`;
    if (params?.isStaff !== undefined) url += `&is_staff=${params.isStaff}`;
    if (params?.isActive !== undefined) url += `&is_active=${params.isActive}`;
    if (params?.isVerified !== undefined) url += `&is_verified=${params.isVerified}`;
    if (params?.isIdentityVerified !== undefined) url += `&is_identity_verified=${params.isIdentityVerified}`;
    if (params?.subscriptionTier) url += `&subscription_tier=${encodeURIComponent(params.subscriptionTier)}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        const results: AdminUser[] = raw.map((u: any) => ({
          id: u.id?.toString() || u.uuid || '',
          email: u.email || '',
          firstName: u.first_name || (u.profile?.first_name) || u.firstName || '',
          lastName: u.last_name || (u.profile?.last_name) || u.lastName || '',
          phone: u.phone_number || (u.profile?.phone) || u.phone || '',
          userType: (u.user_type || (u.profile?.user_type) || u.userType || 'seeker').toLowerCase(),
          isActive: u.is_active ?? true,
          isStaff: u.is_staff ?? false,
          isVerified: u.is_verified ?? false,
          isIdentityVerified: u.is_identity_verified ?? (u.profile?.is_identity_verified) ?? false,
          subscriptionTier: (u.subscription_tier || (u.profile?.subscription_tier) || 'NONE').toUpperCase(),
          checkrStatus: (u.checkr_status || (u.background_check?.status) || 'none').toLowerCase(),
          createdAt: u.created_at || u.createdAt || new Date().toISOString(),
          walletBalance: u.wallet_balance ?? (u.wallet?.balance) ?? 0,
          roleId: u.role_id || u.admin_role_id || '',
          roleName: u.role_name || (u.admin_role?.name) || (u.is_staff ? 'Admin Staff' : 'Member')
        }));
        return {
          results,
          count: res.count ?? results.length
        };
      })
    );
  }

  getUserById(id: string): Observable<AdminUser> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}`).pipe(
      map(u => ({
        id: u.id?.toString() || id,
        email: u.email || '',
        firstName: u.first_name || (u.profile?.first_name) || '',
        lastName: u.last_name || (u.profile?.last_name) || '',
        phone: u.phone_number || u.phone || (u.profile?.phone) || '',
        userType: (u.user_type || (u.profile?.user_type) || 'seeker').toLowerCase(),
        isActive: u.is_active ?? true,
        isStaff: u.is_staff ?? false,
        isVerified: u.is_verified ?? false,
        isIdentityVerified: u.is_identity_verified ?? (u.profile?.is_identity_verified) ?? false,
        subscriptionTier: (u.subscription_tier || (u.profile?.subscription_tier) || 'NONE').toUpperCase(),
        subscriptionInterval: (u.subscription_interval || (u.profile?.subscription_interval) || 'month').toLowerCase(),
        preferredPaymentMode: u.preferred_payment_mode || (u.profile?.preferred_payment_mode) || 'ON_SITE',
        checkrStatus: (u.checkr_status || (u.background_check?.status) || 'none').toLowerCase(),
        bio: u.bio || (u.profile?.bio) || '',
        city: u.city || (u.profile?.city) || '',
        state: u.state || (u.profile?.state) || '',
        zipCode: u.zip_code || (u.profile?.zip_code) || '',
        address: u.address || (u.profile?.address) || '',
        country: u.country || (u.profile?.country) || '',
        gender: u.gender || (u.profile?.gender) || '',
        service: u.service || (u.profile?.service) || '',
        profilePicture: u.profile_picture || (u.profile?.profile_picture) || '',
        createdAt: u.created_at || new Date().toISOString(),
        walletBalance: u.wallet_balance ?? (u.wallet?.balance) ?? 0,
        roleId: u.role_id || u.admin_role_id || '',
        roleName: u.role_name || (u.admin_role?.name) || (u.is_staff ? 'Admin Staff' : 'Member')
      }))
    );
  }

  createUser(data: any): Observable<AdminUser> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}`, {
      email: data.email,
      password: data.password,
      first_name: data.firstName || data.first_name,
      last_name: data.lastName || data.last_name,
      phone: data.phone || data.phone_number,
      user_type: data.userType || data.user_type || 'seeker',
      is_active: data.isActive ?? data.is_active ?? true,
      is_staff: data.isStaff ?? data.is_staff ?? false,
      is_verified: data.isVerified ?? data.is_verified ?? false,
      is_identity_verified: data.isIdentityVerified ?? data.is_identity_verified ?? false,
      subscription_tier: data.subscriptionTier || data.subscription_tier || 'BASIC',
      subscription_interval: data.subscriptionInterval || data.subscription_interval || 'month',
      preferred_payment_mode: data.preferredPaymentMode || data.preferred_payment_mode || 'ON_SITE',
      bio: data.bio || '',
      city: data.city || '',
      state: data.state || '',
      zip_code: data.zipCode || data.zip_code || '',
      address: data.address || '',
      country: data.country || '',
      gender: data.gender || '',
      service: data.service || '',
      role_id: data.roleId || data.role_id || null
    }).pipe(
      map(u => ({
        id: u.id?.toString() || '',
        email: u.email || '',
        firstName: u.first_name || (u.profile?.first_name) || '',
        lastName: u.last_name || (u.profile?.last_name) || '',
        phone: u.phone_number || u.phone || (u.profile?.phone) || '',
        userType: u.user_type || (u.profile?.user_type) || 'seeker',
        isActive: u.is_active ?? true,
        isStaff: u.is_staff ?? false,
        isVerified: u.is_verified ?? false,
        isIdentityVerified: u.is_identity_verified ?? (u.profile?.is_identity_verified) ?? false,
        subscriptionTier: u.subscription_tier || (u.profile?.subscription_tier) || 'Free',
        subscriptionInterval: u.subscription_interval || (u.profile?.subscription_interval) || 'month',
        preferredPaymentMode: u.preferred_payment_mode || (u.profile?.preferred_payment_mode) || 'ON_SITE',
        checkrStatus: u.checkr_status || 'none',
        bio: u.bio || (u.profile?.bio) || '',
        city: u.city || (u.profile?.city) || '',
        state: u.state || (u.profile?.state) || '',
        zipCode: u.zip_code || (u.profile?.zip_code) || '',
        address: u.address || (u.profile?.address) || '',
        country: u.country || (u.profile?.country) || '',
        gender: u.gender || (u.profile?.gender) || '',
        service: u.service || (u.profile?.service) || '',
        createdAt: u.created_at || new Date().toISOString(),
        walletBalance: u.wallet_balance || (u.wallet?.balance) || 0,
        roleId: u.role_id || u.admin_role_id || '',
        roleName: u.role_name || (u.is_staff ? 'Admin Staff' : 'Member')
      }))
    );
  }

  updateUser(id: string, data: any): Observable<AdminUser> {
    const payload: any = {};
    if (data.email !== undefined) payload.email = data.email;
    if (data.firstName !== undefined || data.first_name !== undefined) payload.first_name = data.firstName ?? data.first_name;
    if (data.lastName !== undefined || data.last_name !== undefined) payload.last_name = data.lastName ?? data.last_name;
    if (data.phone !== undefined || data.phone_number !== undefined) payload.phone = data.phone ?? data.phone_number;
    if (data.password !== undefined && data.password !== '') payload.password = data.password;
    if (data.userType !== undefined || data.user_type !== undefined) payload.user_type = data.userType ?? data.user_type;
    if (data.isActive !== undefined || data.is_active !== undefined) payload.is_active = data.isActive ?? data.is_active;
    if (data.isStaff !== undefined || data.is_staff !== undefined) payload.is_staff = data.isStaff ?? data.is_staff;
    if (data.isVerified !== undefined || data.is_verified !== undefined) payload.is_verified = data.isVerified ?? data.is_verified;
    if (data.isIdentityVerified !== undefined || data.is_identity_verified !== undefined) payload.is_identity_verified = data.isIdentityVerified ?? data.is_identity_verified;
    if (data.subscriptionTier !== undefined || data.subscription_tier !== undefined) payload.subscription_tier = data.subscriptionTier ?? data.subscription_tier;
    if (data.subscriptionInterval !== undefined || data.subscription_interval !== undefined) payload.subscription_interval = data.subscriptionInterval ?? data.subscription_interval;
    if (data.preferredPaymentMode !== undefined || data.preferred_payment_mode !== undefined) payload.preferred_payment_mode = data.preferredPaymentMode ?? data.preferred_payment_mode;
    if (data.bio !== undefined) payload.bio = data.bio;
    if (data.city !== undefined) payload.city = data.city;
    if (data.state !== undefined) payload.state = data.state;
    if (data.zipCode !== undefined || data.zip_code !== undefined) payload.zip_code = data.zipCode ?? data.zip_code;
    if (data.address !== undefined) payload.address = data.address;
    if (data.country !== undefined) payload.country = data.country;
    if (data.gender !== undefined) payload.gender = data.gender;
    if (data.service !== undefined) payload.service = data.service;
    if (data.roleId !== undefined || data.role_id !== undefined) payload.role_id = data.roleId ?? data.role_id;

    return this.http.patch<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}`, payload).pipe(
      map(u => ({
        id: u.id?.toString() || id,
        email: u.email || '',
        firstName: u.first_name || (u.profile?.first_name) || '',
        lastName: u.last_name || (u.profile?.last_name) || '',
        phone: u.phone_number || u.phone || (u.profile?.phone) || '',
        userType: u.user_type || (u.profile?.user_type) || 'seeker',
        isActive: u.is_active ?? true,
        isStaff: u.is_staff ?? false,
        isVerified: u.is_verified ?? false,
        isIdentityVerified: u.is_identity_verified ?? (u.profile?.is_identity_verified) ?? false,
        subscriptionTier: u.subscription_tier || (u.profile?.subscription_tier) || 'Free',
        subscriptionInterval: u.subscription_interval || (u.profile?.subscription_interval) || 'month',
        preferredPaymentMode: u.preferred_payment_mode || (u.profile?.preferred_payment_mode) || 'ON_SITE',
        checkrStatus: u.checkr_status || 'none',
        bio: u.bio || (u.profile?.bio) || '',
        city: u.city || (u.profile?.city) || '',
        state: u.state || (u.profile?.state) || '',
        zipCode: u.zip_code || (u.profile?.zip_code) || '',
        address: u.address || (u.profile?.address) || '',
        country: u.country || (u.profile?.country) || '',
        gender: u.gender || (u.profile?.gender) || '',
        service: u.service || (u.profile?.service) || '',
        createdAt: u.created_at || new Date().toISOString(),
        walletBalance: u.wallet_balance || (u.wallet?.balance) || 0,
        roleId: u.role_id || u.admin_role_id || '',
        roleName: u.role_name || (u.is_staff ? 'Admin Staff' : 'Member')
      }))
    );
  }

  deleteUser(id: string): Observable<{ success: boolean }> {
    return this.http.delete<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}`).pipe(
      map(() => ({ success: true }))
    );
  }

  restoreUser(id: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}/restore`, {}).pipe(
      map(() => ({ success: true }))
    );
  }

  impersonateUser(id: string): Observable<ImpersonationResult> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}/impersonate`, {}).pipe(
      map(res => ({
        accessToken: res.access_token || res.token || '',
        expiresIn: res.expires_in || 3600,
        targetUserId: res.target_user_id || id,
        targetEmail: res.target_email || ''
      }))
    );
  }

  getGDPRUserData(id: string): Observable<any> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${id}/gdpr-export`).pipe(
      map(res => {
        const services = res.services || res.profile?.catalog_services || res.profile?.catalogServices || [];
        return {
          ...res,
          services
        };
      })
    );
  }

  assignUserRole(userId: string, roleId: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${userId}/role`, { role_id: roleId }).pipe(
      map(() => ({ success: true }))
    );
  }

  listStaffNotes(userId: string): Observable<any[]> {
    return this.http.get<any[]>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${userId}/notes`);
  }

  createStaffNote(userId: string, data: { content: string; category?: string; is_pinned?: boolean }): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${userId}/notes`, data);
  }

  deleteStaffNote(userId: string, noteId: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${userId}/notes/${noteId}`);
  }

  getProviderFunnel(): Observable<any> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.providerFunnel}`);
  }

  getNotificationFeed(): Observable<{ notifications: any[]; count: number; unread_count: number }> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.notificationFeed}`);
  }

  sendDirectMessage(userId: string, title: string, message: string, channel: string = 'NOTIFICATION'): Observable<{ status: string; sent_at: string }> {
    return this.http.post<any>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.users}/${userId}/message`,
      { title, message, channel }
    );
  }

  evaluateUserRisk(userId: string): Observable<any> {
    return this.http.post<any>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.fraudEvaluate}`,
      { user_id: userId }
    );
  }
}
