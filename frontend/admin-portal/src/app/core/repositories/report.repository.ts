import { Observable } from 'rxjs';
import { FinancialReport, ModerationReportItem } from '../domain/entities/report.model';

export abstract class ReportRepository {
  abstract getFinancialReport(startDate?: string, endDate?: string): Observable<FinancialReport>;
  abstract auditLedger(): Observable<{ stripeInflow: number; platformEscrow: number; reconciledCount: number }>;
  abstract export1099KPackage(year: number, quarter: string): Observable<Blob>;
  abstract exportData(type: 'users' | 'payouts' | 'disputes', startDate?: string, endDate?: string): Observable<Blob>;
  abstract exportUsersCSV(search?: string, userType?: string): Observable<Blob>;
  abstract exportPayoutsCSV(status?: string): Observable<Blob>;
  abstract exportDisputesCSV(status?: string): Observable<Blob>;
  abstract listModerationReports(status?: string, page?: number, pageSize?: number): Observable<{ results: ModerationReportItem[]; count: number }>;
  abstract resolveModerationReport(id: string, action: string, notes?: string): Observable<any>;
  abstract dismissModerationReport(id: string, notes?: string): Observable<any>;
  abstract getStripeLiveBalance(): Observable<{ available: number; pending: number; reserved: number; currency: string; lastRefreshed: string }>;
}
