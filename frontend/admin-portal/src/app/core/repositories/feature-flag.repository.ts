import { Observable } from 'rxjs';
import { FeatureFlagItem } from '../domain/entities/feature-flag.model';

export abstract class FeatureFlagRepository {
  abstract listFeatureFlags(): Observable<FeatureFlagItem[]>;
  abstract setFeatureFlag(key: string, enabled: boolean): Observable<{ success: boolean }>;
}
