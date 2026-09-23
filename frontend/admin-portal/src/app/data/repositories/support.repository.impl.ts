import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { SupportRepository } from '../../core/repositories/support.repository';
import { ContactMessage, ResolutionReport } from '../../core/domain/entities/support.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class SupportRepositoryImpl implements SupportRepository {
  private readonly messagesUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.supportMessages;
  private readonly resolutionsUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.supportResolutions;

  constructor(private http: HttpClient) {}

  listMessages(params?: { is_resolved?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ContactMessage[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.is_resolved !== undefined) httpParams = httpParams.set('is_resolved', params.is_resolved.toString());
    if (params?.search) httpParams = httpParams.set('search', params.search);
    if (params?.page) httpParams = httpParams.set('page', params.page.toString());
    if (params?.pageSize) httpParams = httpParams.set('page_size', params.pageSize.toString());

    return this.http.get<{ results: ContactMessage[]; count: number }>(this.messagesUrl, { params: httpParams });
  }

  toggleMessageResolved(id: string, is_resolved: boolean): Observable<{ status: string }> {
    return this.http.post<{ status: string }>(`${this.messagesUrl}/${id}/toggle-resolved`, { is_resolved });
  }

  deleteMessage(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.messagesUrl}/${id}`);
  }

  listResolutions(params?: { is_reviewed?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ResolutionReport[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.is_reviewed !== undefined) httpParams = httpParams.set('is_reviewed', params.is_reviewed.toString());
    if (params?.search) httpParams = httpParams.set('search', params.search);
    if (params?.page) httpParams = httpParams.set('page', params.page.toString());
    if (params?.pageSize) httpParams = httpParams.set('page_size', params.pageSize.toString());

    return this.http.get<{ results: ResolutionReport[]; count: number }>(this.resolutionsUrl, { params: httpParams });
  }

  toggleResolutionReviewed(id: string, is_reviewed: boolean): Observable<{ status: string }> {
    return this.http.post<{ status: string }>(`${this.resolutionsUrl}/${id}/toggle-reviewed`, { is_reviewed });
  }

  deleteResolution(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.resolutionsUrl}/${id}`);
  }
}
