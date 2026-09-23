import { Observable } from 'rxjs';
import { PlatformSettings } from '../domain/entities/settings.model';

export abstract class SettingsRepository {
  abstract getPlatformSettings(): Observable<PlatformSettings>;
  abstract updatePlatformSettings(settings: Partial<PlatformSettings>): Observable<PlatformSettings>;
  abstract clearCache(): Observable<{ success: boolean; message: string }>;
  abstract broadcastNotification(targetUserType: string, title: string, message: string): Observable<{ success: boolean; recipients: number }>;
}
