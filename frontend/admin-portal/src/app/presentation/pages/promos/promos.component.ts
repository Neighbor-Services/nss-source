import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { PromoUseCase } from '../../../core/usecases/promo.usecase';
import { AdminPromoCode } from '../../../core/domain/entities/promo.model';

@Component({
  selector: 'app-promos',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './promos.component.html',
  styleUrl: './promos.component.css'
})
export class PromosComponent implements OnInit {
  promos = signal<AdminPromoCode[]>([]);
  
  // Modal state
  showCreateModal = signal(false);
  isEditing = signal(false);
  editingId: string | null = null;
  
  // Form Model
  formCode = '';
  formDiscountType: 'PERCENTAGE' | 'FIXED' = 'PERCENTAGE';
  formDiscountValue = 10;
  formMinSpend = 0;
  formMaxDiscount = 50;
  formMaxUses = 100;
  formIsActive = true;
  formExpiresAt = '';

  // Async states
  isLoading = signal(false);
  isActioning = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalPromos = computed(() => this.promos().length);
  activePromosCount = computed(() => this.promos().filter(p => p.is_active).length);
  totalRedemptions = computed(() => this.promos().reduce((acc, p) => acc + (p.uses_count || 0), 0));

  constructor(private promoUC: PromoUseCase) {}

  ngOnInit(): void {
    this.loadPromos();
  }

  loadPromos(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.promoUC.listPromoCodes().subscribe({
      next: (data) => {
        this.promos.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load promo codes.');
      }
    });
  }

  openCreateModal(): void {
    this.isEditing.set(false);
    this.editingId = null;
    this.formCode = '';
    this.formDiscountType = 'PERCENTAGE';
    this.formDiscountValue = 15;
    this.formMinSpend = 0;
    this.formMaxDiscount = 50;
    this.formMaxUses = 100;
    this.formIsActive = true;
    this.formExpiresAt = '';
    this.showCreateModal.set(true);
  }

  openEditModal(promo: AdminPromoCode): void {
    this.isEditing.set(true);
    this.editingId = promo.id;
    this.formCode = promo.code;
    this.formDiscountType = (promo.discount_type as any) || 'PERCENTAGE';
    this.formDiscountValue = promo.discount_value;
    this.formMinSpend = promo.min_spend;
    this.formMaxDiscount = promo.max_discount;
    this.formMaxUses = promo.max_uses;
    this.formIsActive = promo.is_active;
    this.formExpiresAt = promo.expires_at ? promo.expires_at.split('T')[0] : '';
    this.showCreateModal.set(true);
  }

  closeModal(): void {
    this.showCreateModal.set(false);
  }

  savePromo(): void {
    if (!this.formCode.trim()) return;

    this.isActioning.set(true);
    const payload: Partial<AdminPromoCode> = {
      code: this.formCode.trim().toUpperCase(),
      discount_type: this.formDiscountType,
      discount_value: Number(this.formDiscountValue),
      min_spend: Number(this.formMinSpend),
      max_discount: Number(this.formMaxDiscount),
      max_uses: Number(this.formMaxUses),
      is_active: this.formIsActive,
      expires_at: this.formExpiresAt ? new Date(this.formExpiresAt).toISOString() : undefined
    };

    if (this.isEditing() && this.editingId) {
      this.promoUC.updatePromoCode(this.editingId, payload).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.closeModal();
          this.successMessage.set(`Promo code ${payload.code} updated successfully.`);
          this.loadPromos();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update promo code.');
        }
      });
    } else {
      this.promoUC.createPromoCode(payload).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.closeModal();
          this.successMessage.set(`Promo code ${payload.code} created successfully.`);
          this.loadPromos();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create promo code.');
        }
      });
    }
  }

  toggleActive(promo: AdminPromoCode): void {
    const updated = !promo.is_active;
    this.isActioning.set(true);

    this.promoUC.updatePromoCode(promo.id, { ...promo, is_active: updated }).subscribe({
      next: () => {
        this.isActioning.set(false);
        promo.is_active = updated;
        this.successMessage.set(`Promo code ${promo.code} is now ${updated ? 'ACTIVE' : 'PAUSED'}.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to toggle promo code status.');
      }
    });
  }

  deletePromo(promo: AdminPromoCode): void {
    if (!confirm(`Are you sure you want to delete promo code ${promo.code}?`)) return;

    this.isActioning.set(true);
    this.promoUC.deletePromoCode(promo.id).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Promo code ${promo.code} deleted.`);
        this.loadPromos();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to delete promo code.');
      }
    });
  }
}
