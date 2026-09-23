import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { SettingsRepository } from '../repositories/settings.repository';
import { PlatformSettings } from '../domain/entities/settings.model';

@Injectable({
  providedIn: 'root'
})
export class SettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  getPlatformSettings(): Observable<PlatformSettings> {
    return this.settingsRepo.getPlatformSettings();
  }

  updatePlatformSettings(settings: Partial<PlatformSettings>): Observable<PlatformSettings> {
    return this.settingsRepo.updatePlatformSettings(settings);
  }

  clearCache(): Observable<{ success: boolean; message: string }> {
    return this.settingsRepo.clearCache();
  }

  broadcastNotification(targetUserType: string, title: string, message: string): Observable<{ success: boolean; recipients: number }> {
    return this.settingsRepo.broadcastNotification(targetUserType, title, message);
  }
}
