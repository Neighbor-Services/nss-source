import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { SubscriptionRepository } from '../../core/repositories/subscription.repository';
import { SubscriptionItem, SubscriptionPlan } from '../../core/domain/entities/subscription.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class SubscriptionRepositoryImpl implements SubscriptionRepository {
  constructor(private http: HttpClient) {}

  listSubscriptions(isActive?: boolean): Observable<SubscriptionItem[]> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptions}`;
    if (isActive !== undefined) {
      url += `?is_active=${isActive}`;
    }

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((s: any) => {
          const user = s.user_details || s.user || {};
          const profile = user.profile || {};
          const first = profile.first_name || user.first_name || '';
          const last = profile.last_name || user.last_name || '';
          const fullName = [first, last].filter(Boolean).join(' ');
          const userName = fullName || user.email || 'Provider User';
          const plan = s.plan_details || s.plan || {};

          return {
            id: s.id?.toString() || '',
            userId: s.user_id || user.id || s.user || '',
            userName,
            userEmail: user.email || s.user_email || '',
            tier: plan.tier || profile.subscription_tier || s.tier || 'PRO',
            interval: plan.interval || profile.subscription_interval || s.interval || 'month',
            isActive: s.is_active ?? true,
            stripeSubscriptionId: s.store_transaction_id || s.stripe_subscription_id || '',
            currentPeriodEnd: s.next_payment || s.current_period_end || '',
            createdAt: s.created_at || new Date().toISOString()
          };
        });
      })
    );
  }

  toggleSubscription(id: string, isActive: boolean): Observable<{ success: boolean; isActive: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptions}/${id}/toggle`, {
      is_active: isActive
    }).pipe(
      map(res => ({
        success: true,
        isActive: res.is_active ?? isActive
      }))
    );
  }

  assignSubscription(data: {
    userId: string;
    planId?: string;
    tier?: string;
    interval?: string;
    isActive?: boolean;
    nextPayment?: string;
    notes?: string;
  }): Observable<SubscriptionItem> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}/admin/subscriptions/assign`, {
      user_id: data.userId,
      plan_id: data.planId || null,
      tier: data.tier || 'SILVER',
      interval: data.interval || 'month',
      is_active: data.isActive ?? true,
      next_payment: data.nextPayment || null,
      notes: data.notes || ''
    }).pipe(
      map((s: any) => {
        const user = s.user_details || s.user || {};
        const profile = user.profile || {};
        const first = profile.first_name || user.first_name || '';
        const last = profile.last_name || user.last_name || '';
        const fullName = [first, last].filter(Boolean).join(' ');
        const userName = fullName || user.email || 'Provider User';
        const plan = s.plan_details || s.plan || {};

        return {
          id: s.id?.toString() || '',
          userId: s.user_id || user.id || s.user || '',
          userName,
          userEmail: user.email || '',
          tier: plan.tier || profile.subscription_tier || s.tier || 'PRO',
          interval: plan.interval || profile.subscription_interval || s.interval || 'month',
          isActive: s.is_active ?? true,
          stripeSubscriptionId: s.store_transaction_id || '',
          currentPeriodEnd: s.next_payment || '',
          createdAt: s.created_at || new Date().toISOString()
        };
      })
    );
  }

  listPlans(): Observable<SubscriptionPlan[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptionPlans}`).pipe(
      map(res => {
        const list = Array.isArray(res) ? res : (res.results || []);
        return list.map((p: any) => ({
          id: p.id,
          name: p.name,
          tier: p.tier,
          interval: p.interval,
          description: p.description || '',
          price: typeof p.price === 'string' ? parseFloat(p.price) : (p.price || 0),
          currency: p.currency || 'USD',
          features: Array.isArray(p.features) ? p.features : [],
          appleProductId: p.apple_product_id || p.appleProductId || '',
          googleProductId: p.google_product_id || p.googleProductId || '',
          maxCatalogServices: p.max_catalog_services ?? p.maxCatalogServices ?? 1,
          isActive: p.is_active ?? p.isActive ?? true,
          displayOrder: p.display_order ?? p.displayOrder ?? 0,
          created_at: p.created_at
        }));
      })
    );
  }

  createPlan(plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan> {
    return this.http.post<SubscriptionPlan>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptionPlans}`, {
      name: plan.name,
      tier: plan.tier,
      interval: plan.interval,
      description: plan.description,
      price: plan.price,
      currency: plan.currency || 'USD',
      features: plan.features || [],
      apple_product_id: plan.appleProductId || '',
      google_product_id: plan.googleProductId || '',
      max_catalog_services: plan.maxCatalogServices ?? 1,
      is_active: plan.isActive ?? true,
      display_order: plan.displayOrder ?? 0
    });
  }

  updatePlan(id: string, plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan> {
    return this.http.put<SubscriptionPlan>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptionPlans}/${id}`, {
      name: plan.name,
      tier: plan.tier,
      interval: plan.interval,
      description: plan.description,
      price: plan.price,
      currency: plan.currency || 'USD',
      features: plan.features || [],
      apple_product_id: plan.appleProductId || '',
      google_product_id: plan.googleProductId || '',
      max_catalog_services: plan.maxCatalogServices ?? 1,
      is_active: plan.isActive ?? true,
      display_order: plan.displayOrder ?? 0
    });
  }

  deletePlan(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.subscriptionPlans}/${id}`);
  }
}
