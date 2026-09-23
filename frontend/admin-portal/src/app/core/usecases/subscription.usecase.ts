import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { SubscriptionRepository } from '../repositories/subscription.repository';
import { SubscriptionItem, SubscriptionPlan } from '../domain/entities/subscription.model';

@Injectable({
  providedIn: 'root'
})
export class SubscriptionUseCase {
  constructor(private subRepo: SubscriptionRepository) {}

  listSubscriptions(isActive?: boolean): Observable<SubscriptionItem[]> {
    return this.subRepo.listSubscriptions(isActive);
  }

  toggleSubscription(id: string, isActive: boolean): Observable<{ success: boolean; isActive: boolean }> {
    return this.subRepo.toggleSubscription(id, isActive);
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
    return this.subRepo.assignSubscription(data);
  }

  listPlans(): Observable<SubscriptionPlan[]> {
    return this.subRepo.listPlans();
  }

  createPlan(plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan> {
    return this.subRepo.createPlan(plan);
  }

  updatePlan(id: string, plan: Partial<SubscriptionPlan>): Observable<SubscriptionPlan> {
    return this.subRepo.updatePlan(id, plan);
  }

  deletePlan(id: string): Observable<{ status: string }> {
    return this.subRepo.deletePlan(id);
  }
}
