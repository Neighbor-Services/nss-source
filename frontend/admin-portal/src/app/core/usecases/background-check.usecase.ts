import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { BackgroundCheckRepository } from '../repositories/background-check.repository';
import { BackgroundCheckItem } from '../domain/entities/background-check.model';

@Injectable({
  providedIn: 'root'
})
export class BackgroundCheckUseCase {
  constructor(private bgRepo: BackgroundCheckRepository) {}

  listBackgroundChecks(): Observable<BackgroundCheckItem[]> {
    return this.bgRepo.listBackgroundChecks();
  }

  overrideBackgroundCheck(id: string, newStatus: string, reason: string): Observable<{ success: boolean }> {
    return this.bgRepo.overrideBackgroundCheck(id, newStatus, reason);
  }
}
