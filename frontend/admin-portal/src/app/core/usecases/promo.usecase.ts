import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { PromoRepository } from '../repositories/promo.repository';
import { AdminPromoCode } from '../domain/entities/promo.model';

@Injectable({
  providedIn: 'root'
})
export class PromoUseCase {
  constructor(private promoRepo: PromoRepository) {}

  listPromoCodes(): Observable<AdminPromoCode[]> {
    return this.promoRepo.listPromoCodes();
  }

  createPromoCode(promo: Partial<AdminPromoCode>): Observable<AdminPromoCode> {
    return this.promoRepo.createPromoCode(promo);
  }

  updatePromoCode(id: string, promo: Partial<AdminPromoCode>): Observable<AdminPromoCode> {
    return this.promoRepo.updatePromoCode(id, promo);
  }

  deletePromoCode(id: string): Observable<{ status: string }> {
    return this.promoRepo.deletePromoCode(id);
  }
}
