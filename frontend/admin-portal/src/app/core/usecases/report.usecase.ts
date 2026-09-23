import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ReportRepository } from '../repositories/report.repository';
import { FinancialReport, ModerationReportItem } from '../domain/entities/report.model';

@Injectable({
  providedIn: 'root'
})
export class ReportUseCase {
  constructor(private reportRepo: ReportRepository) {}

  getFinancialReport(startDate?: string, endDate?: string): Observable<FinancialReport> {
    return this.reportRepo.getFinancialReport(startDate, endDate);
  }

  auditLedger(): Observable<{ stripeInflow: number; platformEscrow: number; reconciledCount: number }> {
    return this.reportRepo.auditLedger();
  }

  export1099KPackage(year: number, quarter: string): Observable<Blob> {
    return this.reportRepo.export1099KPackage(year, quarter);
  }

  exportData(type: 'users' | 'payouts' | 'disputes', startDate?: string, endDate?: string): Observable<Blob> {
    return this.reportRepo.exportData(type, startDate, endDate);
  }

  exportUsersCSV(search?: string, userType?: string): Observable<Blob> {
    return this.reportRepo.exportUsersCSV(search, userType);
  }

  exportPayoutsCSV(status?: string): Observable<Blob> {
    return this.reportRepo.exportPayoutsCSV(status);
  }

  exportDisputesCSV(status?: string): Observable<Blob> {
    return this.reportRepo.exportDisputesCSV(status);
  }

  listModerationReports(status?: string, page?: number, pageSize?: number): Observable<ModerationReportItem[]> {
    return new Observable(observer => {
      this.reportRepo.listModerationReports(status, page, pageSize).subscribe({
        next: res => {
          observer.next(res.results);
          observer.complete();
        },
        error: err => observer.error(err)
      });
    });
  }

  resolveModerationReport(id: string, notes?: string): Observable<any> {
    return this.reportRepo.resolveModerationReport(id, 'RESOLVE', notes);
  }

  dismissModerationReport(id: string, notes?: string): Observable<any> {
    return this.reportRepo.dismissModerationReport(id, notes);
  }

  getStripeLiveBalance(): Observable<{ available: number; pending: number; reserved: number; currency: string; lastRefreshed: string }> {
    return this.reportRepo.getStripeLiveBalance();
  }
}
