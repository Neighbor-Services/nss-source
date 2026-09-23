import { Observable } from 'rxjs';
import { AdminPromoCode } from '../domain/entities/promo.model';

export abstract class PromoRepository {
  abstract listPromoCodes(): Observable<AdminPromoCode[]>;
  abstract createPromoCode(promo: Partial<AdminPromoCode>): Observable<AdminPromoCode>;
  abstract updatePromoCode(id: string, promo: Partial<AdminPromoCode>): Observable<AdminPromoCode>;
  abstract deletePromoCode(id: string): Observable<{ status: string }>;
}
