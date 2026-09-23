import { Observable } from 'rxjs';
import { SubscriptionItem, SubscriptionPlan } from '../domain/entities/subscription.model';

export abstract class SubscriptionRepository {
  abstract listSubscriptions(isActive?: boolean): Observable<SubscriptionItem[]>;
  abstract toggleSubscription(id: string, isActive: boolean): Observable<{ success: boolean; isActive: boolean }>;
  abstract assignSubscription(data: {
    userId: string;
    planId?: string;
    tier?: string;
    interval?: string;
    isActive?: boolean;
    nextPayment?: string;
    notes?: string;
  }): Observable<SubscriptionItem>;
  abstract listPlans(): Observable<SubscriptionPlan[]>;
  abstract createPlan(plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan>;
  abstract updatePlan(id: string, plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan>;
  abstract deletePlan(id: string): Observable<{ status: string }>;
}
