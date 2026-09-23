import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PromoRepository } from '../../core/repositories/promo.repository';
import { AdminPromoCode } from '../../core/domain/entities/promo.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class PromoRepositoryImpl implements PromoRepository {
  private readonly baseUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.promoCodes;

  constructor(private http: HttpClient) {}

  listPromoCodes(): Observable<AdminPromoCode[]> {
    return this.http.get<AdminPromoCode[]>(this.baseUrl);
  }

  createPromoCode(promo: Partial<AdminPromoCode>): Observable<AdminPromoCode> {
    return this.http.post<AdminPromoCode>(this.baseUrl, promo);
  }

  updatePromoCode(id: string, promo: Partial<AdminPromoCode>): Observable<AdminPromoCode> {
    return this.http.put<AdminPromoCode>(`${this.baseUrl}/${id}`, promo);
  }

  deletePromoCode(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.baseUrl}/${id}`);
  }
}
