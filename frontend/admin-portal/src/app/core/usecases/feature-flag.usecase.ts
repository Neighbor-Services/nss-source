import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { FeatureFlagRepository } from '../repositories/feature-flag.repository';
import { FeatureFlagItem } from '../domain/entities/feature-flag.model';

@Injectable({
  providedIn: 'root'
})
export class FeatureFlagUseCase {
  constructor(private flagRepo: FeatureFlagRepository) {}

  listFeatureFlags(): Observable<FeatureFlagItem[]> {
    return this.flagRepo.listFeatureFlags();
  }

  setFeatureFlag(key: string, enabled: boolean): Observable<{ success: boolean }> {
    return this.flagRepo.setFeatureFlag(key, enabled);
  }
}
