import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ResolutionReport, ResolutionResponse } from '../domain/entities/resolution.model';
import { ResolutionRepository } from '../repositories/resolution.repository';

@Injectable({
  providedIn: 'root'
})
export class SubmitResolutionUseCase {
  constructor(private resolutionRepo: ResolutionRepository) {}

  execute(report: ResolutionReport): Observable<ResolutionResponse> {
    return this.resolutionRepo.submitResolutionReport(report);
  }
}
