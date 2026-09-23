import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SubscriptionUseCase } from '../../../core/usecases/subscription.usecase';
import { UserUseCase } from '../../../core/usecases/user.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { SubscriptionItem, SubscriptionPlan } from '../../../core/domain/entities/subscription.model';
import { AdminUser } from '../../../core/domain/entities/user.model';

export type SubTabFilter = 'ALL' | 'ACTIVE' | 'PRO' | 'ELITE' | 'INACTIVE';
export type SubSectionTab = 'SUBSCRIBERS' | 'TIER_MATRIX' | 'BILLING_ENGINE';

@Component({
  selector: 'app-subscriptions',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './subscriptions.component.html',
  styleUrl: './subscriptions.component.css'
})
export class SubscriptionsComponent implements OnInit {
  subscriptions = signal<SubscriptionItem[]>([]);
  plans = signal<SubscriptionPlan[]>([]);
  isLoading = signal(false);
  isLoadingPlans = signal(false);
  isActioning = signal(false);
  togglingId = signal<string | null>(null);
  selectedFilter: 'all' | 'active' | 'inactive' = 'all';
  activeTab = signal<SubTabFilter>('ALL');
  mainSection = signal<SubSectionTab>('SUBSCRIBERS');

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Manual Subscription Grant Modal State
  showManualModal = signal(false);
  providersList = signal<AdminUser[]>([]);
  isLoadingProviders = signal(false);
  providerSearch = signal('');
  manualSubForm = {
    userId: '',
    planId: '',
    tier: 'PRO',
    interval: 'month',
    durationPreset: '1m',
    customEndDate: '',
    notes: '',
    isActive: true
  };

  // Plan Modal State
  showPlanModal = signal(false);
  editingPlan = signal<SubscriptionPlan | null>(null);
  planForm: Partial<SubscriptionPlan> = {
    name: '',
    tier: 'SILVER',
    interval: 'month',
    description: '',
    price: 29.99,
    currency: 'USD',
    features: [],
    appleProductId: '',
    googleProductId: '',
    maxCatalogServices: 5,
    isActive: true,
    displayOrder: 1
  };
  featuresText = '';

  // Computed KPIs
  totalCount = computed(() => this.subscriptions().length);
  activeCount = computed(() => this.subscriptions().filter(s => s.isActive).length);
  proCount = computed(() => this.subscriptions().filter(s => s.tier?.toUpperCase() === 'PRO' || s.tier?.toUpperCase() === 'SILVER').length);
  eliteCount = computed(() => this.subscriptions().filter(s => s.tier?.toUpperCase() === 'ELITE' || s.tier?.toUpperCase() === 'GOLD' || s.tier?.toUpperCase() === 'PLATINUM').length);
  plansCount = computed(() => this.plans().length);
  estimatedMRR = computed(() => (this.proCount() * 29) + (this.eliteCount() * 79));

  filteredProviders = computed(() => {
    const q = this.providerSearch().toLowerCase().trim();
    if (!q) return this.providersList();
    return this.providersList().filter(p =>
      (p.firstName && p.firstName.toLowerCase().includes(q)) ||
      (p.lastName && p.lastName.toLowerCase().includes(q)) ||
      (p.email && p.email.toLowerCase().includes(q)) ||
      (p.id && p.id.toLowerCase().includes(q))
    );
  });

  filteredSubscriptions = computed(() => {
    const tab = this.activeTab();
    if (tab === 'ACTIVE') return this.subscriptions().filter(s => s.isActive);
    if (tab === 'PRO') return this.subscriptions().filter(s => s.tier?.toUpperCase() === 'PRO' || s.tier?.toUpperCase() === 'SILVER');
    if (tab === 'ELITE') return this.subscriptions().filter(s => s.tier?.toUpperCase() === 'ELITE' || s.tier?.toUpperCase() === 'GOLD' || s.tier?.toUpperCase() === 'PLATINUM');
    if (tab === 'INACTIVE') return this.subscriptions().filter(s => !s.isActive);
    return this.subscriptions();
  });

  constructor(
    private subUC: SubscriptionUseCase,
    private userUC: UserUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit(): void {
    this.loadSubscriptions();
    this.loadPlans();
  }

  loadSubscriptions(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    let activeParam: boolean | undefined = undefined;
    if (this.selectedFilter === 'active') activeParam = true;
    if (this.selectedFilter === 'inactive') activeParam = false;

    this.subUC.listSubscriptions(activeParam).subscribe({
      next: (subs) => {
        this.subscriptions.set(subs);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to load provider subscriptions');
      }
    });
  }

  loadPlans(): void {
    this.isLoadingPlans.set(true);
    this.subUC.listPlans().subscribe({
      next: (plansList) => {
        this.plans.set(plansList);
        this.isLoadingPlans.set(false);
      },
      error: () => {
        this.isLoadingPlans.set(false);
      }
    });
  }

  openCreatePlanModal(): void {
    this.editingPlan.set(null);
    this.planForm = {
      name: '',
      tier: 'SILVER',
      interval: 'month',
      description: '',
      price: 29.99,
      currency: 'USD',
      features: [],
      appleProductId: 'nsapp_pro_monthly',
      googleProductId: 'nsapp_pro_monthly',
      maxCatalogServices: 5,
      isActive: true,
      displayOrder: (this.plans().length + 1)
    };
    this.featuresText = 'Direct Client Messaging\nInstant Broadcast Alerts\n10 Portfolio Showcases';
    this.showPlanModal.set(true);
  }

  openEditPlanModal(plan: SubscriptionPlan): void {
    this.editingPlan.set(plan);
    this.planForm = {
      id: plan.id,
      name: plan.name,
      tier: plan.tier,
      interval: plan.interval,
      description: plan.description,
      price: plan.price,
      currency: plan.currency || 'USD',
      features: plan.features || [],
      appleProductId: plan.appleProductId || '',
      googleProductId: plan.googleProductId || '',
      maxCatalogServices: plan.maxCatalogServices ?? 1,
      isActive: plan.isActive ?? true,
      displayOrder: plan.displayOrder ?? 0
    };
    this.featuresText = (plan.features || []).join('\n');
    this.showPlanModal.set(true);
  }

  closePlanModal(): void {
    this.showPlanModal.set(false);
  }

  savePlan(): void {
    if (!this.planForm.name || !this.planForm.tier) {
      this.errorMessage.set('Plan Name and Tier level are required.');
      return;
    }

    this.isActioning.set(true);
    this.errorMessage.set(null);

    const parsedFeatures = this.featuresText
      .split('\n')
      .map(f => f.trim())
      .filter(f => f.length > 0);

    const payload: Partial<SubscriptionPlan> = {
      name: this.planForm.name,
      tier: this.planForm.tier,
      interval: this.planForm.interval || 'month',
      description: this.planForm.description || '',
      price: Number(this.planForm.price) || 0,
      currency: this.planForm.currency || 'USD',
      features: parsedFeatures,
      appleProductId: this.planForm.appleProductId || '',
      googleProductId: this.planForm.googleProductId || '',
      maxCatalogServices: Number(this.planForm.maxCatalogServices) ?? 1,
      isActive: this.planForm.isActive ?? true,
      displayOrder: Number(this.planForm.displayOrder) || 0
    };

    const currentEditing = this.editingPlan();
    if (currentEditing && currentEditing.id) {
      this.subUC.updatePlan(currentEditing.id, payload).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.showPlanModal.set(false);
          this.successMessage.set(`Subscription plan "${payload.name}" updated successfully.`);
          this.loadPlans();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update plan.');
        }
      });
    } else {
      this.subUC.createPlan(payload).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.showPlanModal.set(false);
          this.successMessage.set(`Subscription plan "${payload.name}" created successfully.`);
          this.loadPlans();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create plan.');
        }
      });
    }
  }

  async deletePlan(plan: SubscriptionPlan): Promise<void> {
    if (!plan.id) return;
    const confirmed = await this.dialog.confirm({
      title: 'Delete Subscription Plan',
      message: `Are you sure you want to delete tier plan "${plan.name}" (${plan.tier})? Existing active subscribers will maintain their current billing cycle.`,
      confirmText: 'Delete Plan',
      isDanger: true,
      cancelText: 'Cancel'
    });
    if (!confirmed) return;

    this.isActioning.set(true);
    this.subUC.deletePlan(plan.id).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Subscription plan "${plan.name}" deleted.`);
        this.loadPlans();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to delete plan.');
      }
    });
  }

  async toggleActive(sub: SubscriptionItem): Promise<void> {
    const targetState = !sub.isActive;
    const actionName = targetState ? 'Activate' : 'Deactivate';

    const confirmed = await this.dialog.confirm({
      title: `${actionName} Subscription`,
      message: `Are you sure you want to ${actionName.toLowerCase()} the ${sub.tier} tier subscription for ${sub.userName}?`,
      confirmText: actionName,
      isDanger: !targetState,
      cancelText: 'Cancel'
    });

    if (!confirmed) return;

    this.togglingId.set(sub.id);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.subUC.toggleSubscription(sub.id, targetState).subscribe({
      next: (res) => {
        this.togglingId.set(null);
        this.successMessage.set(`Subscription for ${sub.userName} is now ${res.isActive ? 'Active' : 'Inactive'}.`);
        this.loadSubscriptions();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.togglingId.set(null);
        this.errorMessage.set(err.error?.message || 'Failed to toggle subscription status');
      }
    });
  }

  getTierBadge(tier: string): string {
    switch (tier?.toUpperCase()) {
      case 'PLATINUM':
        return 'badge-purple';
      case 'GOLD':
      case 'ELITE':
        return 'badge-warning';
      case 'SILVER':
      case 'PRO':
        return 'badge-info';
      default:
        return 'badge-secondary';
    }
  }

  openManualSubscribeModal(): void {
    this.providerSearch.set('');
    this.manualSubForm = {
      userId: '',
      planId: this.plans().length > 0 ? this.plans()[0].id || '' : '',
      tier: this.plans().length > 0 ? this.plans()[0].tier : 'PRO',
      interval: this.plans().length > 0 ? (this.plans()[0].interval || 'month') : 'month',
      durationPreset: '1m',
      customEndDate: '',
      notes: '',
      isActive: true
    };
    this.showManualModal.set(true);

    if (this.providersList().length === 0) {
      this.isLoadingProviders.set(true);
      this.userUC.listUsers({ userType: 'PROVIDER', pageSize: 100 }).subscribe({
        next: (res) => {
          this.providersList.set(res.results || []);
          this.isLoadingProviders.set(false);
        },
        error: () => {
          this.isLoadingProviders.set(false);
        }
      });
    }
  }

  closeManualModal(): void {
    this.showManualModal.set(false);
  }

  onManualPlanChange(planId: string): void {
    const selected = this.plans().find(p => p.id === planId);
    if (selected) {
      this.manualSubForm.tier = selected.tier;
      this.manualSubForm.interval = selected.interval || 'month';
    }
  }

  saveManualSubscription(): void {
    if (!this.manualSubForm.userId) {
      this.errorMessage.set('Please select a provider to grant subscription access.');
      return;
    }

    this.isActioning.set(true);
    this.errorMessage.set(null);

    let nextPaymentDate: Date;
    const now = new Date();
    switch (this.manualSubForm.durationPreset) {
      case '1m':
        nextPaymentDate = new Date(now.setMonth(now.getMonth() + 1));
        break;
      case '3m':
        nextPaymentDate = new Date(now.setMonth(now.getMonth() + 3));
        break;
      case '6m':
        nextPaymentDate = new Date(now.setMonth(now.getMonth() + 6));
        break;
      case '1y':
        nextPaymentDate = new Date(now.setFullYear(now.getFullYear() + 1));
        break;
      case 'lifetime':
        nextPaymentDate = new Date(now.setFullYear(now.getFullYear() + 100));
        break;
      case 'custom':
        nextPaymentDate = this.manualSubForm.customEndDate ? new Date(this.manualSubForm.customEndDate) : new Date(now.setMonth(now.getMonth() + 1));
        break;
      default:
        nextPaymentDate = new Date(now.setMonth(now.getMonth() + 1));
    }

    const payload = {
      userId: this.manualSubForm.userId,
      planId: this.manualSubForm.planId || undefined,
      tier: this.manualSubForm.tier || 'PRO',
      interval: this.manualSubForm.interval || 'month',
      isActive: this.manualSubForm.isActive,
      nextPayment: nextPaymentDate.toISOString(),
      notes: this.manualSubForm.notes?.trim() || undefined
    };

    this.subUC.assignSubscription(payload).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.showManualModal.set(false);
        this.successMessage.set('Provider subscription successfully granted and active.');
        this.loadSubscriptions();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to assign subscription to provider.');
      }
    });
  }

  exportCSV(): void {
    const list = this.filteredSubscriptions();
    if (!list.length) return;

    import('../../../core/utils/export.util').then(({ exportToCsv }) => {
      exportToCsv(list, 'provider_subscriptions', [
        { key: 'id', header: 'Subscription ID' },
        { key: 'userName', header: 'Provider Name' },
        { key: 'tier', header: 'Tier Plan' },
        { key: 'isActive', header: 'Status', format: (val) => val ? 'Active' : 'Inactive' },
        { key: 'isAutoRenew', header: 'Auto-Renew', format: (val) => val ? 'Yes' : 'No' },
        { key: 'startDate', header: 'Start Date' },
        { key: 'endDate', header: 'Expiry Date' },
        { key: 'createdAt', header: 'Created At' }
      ]);
    });
  }
}
