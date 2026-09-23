import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { FeatureFlagRepository } from '../../core/repositories/feature-flag.repository';
import { FeatureFlagItem } from '../../core/domain/entities/feature-flag.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class FeatureFlagRepositoryImpl implements FeatureFlagRepository {
  constructor(private http: HttpClient) {}

  listFeatureFlags(): Observable<FeatureFlagItem[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.featureFlags}`).pipe(
      map(res => {
        if (typeof res === 'object' && !Array.isArray(res)) {
          return Object.keys(res).map(key => ({
            key,
            name: key.replace(/_/g, ' ').toUpperCase(),
            description: `Runtime flag for ${key.replace(/_/g, ' ')}`,
            enabled: !!res[key]
          }));
        }
        if (Array.isArray(res)) {
          return res.map((f: any) => ({
            key: f.key || f.name || '',
            name: f.name || f.key || '',
            description: f.description || 'System runtime feature flag',
            enabled: f.enabled ?? f.is_enabled ?? false
          }));
        }
        return [];
      })
    );
  }

  setFeatureFlag(key: string, enabled: boolean): Observable<{ success: boolean }> {
    return this.http.put<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.featureFlags}/${key}`, { is_enabled: enabled }).pipe(
      map(() => ({ success: true }))
    );
  }
}
