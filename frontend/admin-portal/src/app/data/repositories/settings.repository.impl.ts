import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { SettingsRepository } from '../../core/repositories/settings.repository';
import { PlatformSettings } from '../../core/domain/entities/settings.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class SettingsRepositoryImpl implements SettingsRepository {
  constructor(private http: HttpClient) {}

  getPlatformSettings(): Observable<PlatformSettings> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.settings}`).pipe(
      map(s => ({
        backgroundCheckPaymentMode: s.background_check_payment_mode || 'PLATFORM_PAYS',
        backgroundCheckFee: s.background_check_fee ?? 29.99,
        broadcastRadiusKm: s.broadcast_radius_km ?? 25.0,
        matchRadiusKm: s.match_radius_km ?? 15.0
      }))
    );
  }

  updatePlatformSettings(settings: Partial<PlatformSettings>): Observable<PlatformSettings> {
    return this.http.patch<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.settings}`, {
      background_check_payment_mode: settings.backgroundCheckPaymentMode,
      background_check_fee: settings.backgroundCheckFee,
      broadcast_radius_km: settings.broadcastRadiusKm,
      match_radius_km: settings.matchRadiusKm
    }).pipe(
      map(s => ({
        backgroundCheckPaymentMode: s.background_check_payment_mode || 'PLATFORM_PAYS',
        backgroundCheckFee: s.background_check_fee ?? 29.99,
        broadcastRadiusKm: s.broadcast_radius_km ?? 25.0,
        matchRadiusKm: s.match_radius_km ?? 15.0
      }))
    );
  }

  clearCache(): Observable<{ success: boolean; message: string }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.cacheClear}`, {}).pipe(
      map(res => ({ success: true, message: res.status || 'Application cache flushed successfully' }))
    );
  }

  broadcastNotification(targetUserType: string, title: string, message: string): Observable<{ success: boolean; recipients: number }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.broadcastNotifications}`, {
      target_user_type: targetUserType,
      title,
      message
    }).pipe(
      map(res => ({ success: true, recipients: res.recipients ?? 0 }))
    );
  }
}
