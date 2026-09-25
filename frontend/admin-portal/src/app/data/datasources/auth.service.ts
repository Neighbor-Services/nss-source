import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap, map } from 'rxjs';
import { ADMIN_API_CONFIG } from './admin-api.config';
import { AdminAuthResponse } from '../../core/domain/entities/auth.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly TOKEN_KEY = 'ns_admin_token';
  private readonly REFRESH_KEY = 'ns_admin_refresh_token';
  private readonly USER_KEY = 'ns_admin_user';

  currentUser = signal<any>(this.getStoredUser());
  isAuthenticated = signal<boolean>(!!this.getToken());

  constructor(
    private http: HttpClient,
    private router: Router
  ) {}

  login(email: string, password: string, totpCode?: string): Observable<AdminAuthResponse> {
    const payload: any = { email, password };
    if (totpCode) {
      payload.code = totpCode;
      payload.totp_code = totpCode;
    }

    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.login}`, payload).pipe(
      map(res => {
        if (res.requires_2fa || res.requires2fa) {
          return {
            token: '',
            requires2FA: true,
            tempToken: res.temp_token || '',
            user: { id: '', email, firstName: '', lastName: '', isStaff: true }
          };
        }

        const token = res.access || res.tokens?.access || res.token || '';
        const refreshToken = res.refresh || res.tokens?.refresh || '';
        const user = {
          id: res.user?.id || '',
          email: res.user?.email || email,
          firstName: res.user?.first_name || '',
          lastName: res.user?.last_name || '',
          isStaff: res.user?.is_staff ?? true,
          isSuperuser: res.user?.is_superuser ?? false
        };

        if (token) {
          this.setSession(token, refreshToken, user);
        }

        return { token, refreshToken, user, requires2FA: false };
      })
    );
  }

  refreshToken(): Observable<string> {
    const refresh = this.getRefreshToken();
    if (!refresh) {
      throw new Error('No refresh token available');
    }

    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}/accounts/token/refresh/`, { refresh }).pipe(
      map(res => {
        const newToken = res.access || res.token || '';
        if (newToken) {
          localStorage.setItem(this.TOKEN_KEY, newToken);
          return newToken;
        }
        throw new Error('Failed to refresh token');
      })
    );
  }

  changePassword(oldPassword: string, newPassword: string): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.changePassword}`, {
      old_password: oldPassword,
      new_password: newPassword
    });
  }

  logout() {
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_KEY);
    localStorage.removeItem(this.USER_KEY);
    this.currentUser.set(null);
    this.isAuthenticated.set(false);
    this.router.navigate(['/login']);
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_KEY);
  }

  private setSession(token: string, refreshToken: string, user: any) {
    localStorage.setItem(this.TOKEN_KEY, token);
    if (refreshToken) {
      localStorage.setItem(this.REFRESH_KEY, refreshToken);
    }
    localStorage.setItem(this.USER_KEY, JSON.stringify(user));
    this.currentUser.set(user);
    this.isAuthenticated.set(true);
  }

  private getStoredUser(): any {
    const raw = localStorage.getItem(this.USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }
}
