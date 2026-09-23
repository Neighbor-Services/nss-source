import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { LegalRepository } from '../../core/repositories/legal.repository';
import { LegalDocument } from '../../core/domain/entities/legal.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class LegalRepositoryImpl implements LegalRepository {
  private readonly baseUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.legalDocuments;

  constructor(private http: HttpClient) {}

  listDocuments(params?: { type?: string; is_active?: boolean }): Observable<{ results: LegalDocument[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.type) httpParams = httpParams.set('type', params.type);
    if (params?.is_active !== undefined) httpParams = httpParams.set('is_active', params.is_active.toString());

    return this.http.get<{ results: LegalDocument[]; count: number }>(this.baseUrl, { params: httpParams });
  }

  getDocument(id: string): Observable<LegalDocument> {
    return this.http.get<LegalDocument>(`${this.baseUrl}/${id}`);
  }

  createDocument(doc: Partial<LegalDocument>): Observable<LegalDocument> {
    return this.http.post<LegalDocument>(this.baseUrl, doc);
  }

  updateDocument(id: string, doc: Partial<LegalDocument>): Observable<LegalDocument> {
    return this.http.put<LegalDocument>(`${this.baseUrl}/${id}`, doc);
  }

  deleteDocument(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.baseUrl}/${id}`);
  }
}
