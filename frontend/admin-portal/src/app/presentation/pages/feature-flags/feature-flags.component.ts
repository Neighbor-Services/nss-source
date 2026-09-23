import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FeatureFlagUseCase } from '../../../core/usecases/feature-flag.usecase';
import { FeatureFlagItem } from '../../../core/domain/entities/feature-flag.model';

export type FeatureFlagMainTab = 'FLAGS' | 'KILLSWITCH' | 'CANARY';

@Component({
  selector: 'app-feature-flags',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './feature-flags.component.html',
  styleUrl: './feature-flags.component.css'
})
export class FeatureFlagsComponent implements OnInit {
  flags = signal<FeatureFlagItem[]>([]);
  isLoading = signal(false);
  isTogglingKey = signal<string | null>(null);
  searchFilter = '';
  activeTab = signal<FeatureFlagMainTab>('FLAGS');

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalFlags = computed(() => this.flags().length);
  enabledCount = computed(() => this.flags().filter(f => f.enabled).length);
  disabledCount = computed(() => this.flags().filter(f => !f.enabled).length);
  killswitchCount = computed(() => this.flags().filter(f => f.key.includes('kill') || f.key.includes('disable') || f.key.includes('payout') || f.key.includes('maintenance')).length);


  constructor(private flagUC: FeatureFlagUseCase) {}

  ngOnInit() {
    this.loadFlags();
  }

  loadFlags() {
    this.isLoading.set(true);
    this.errorMessage.set(null);
    this.flagUC.listFeatureFlags().subscribe({
      next: (f) => {
        this.flags.set(f);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to load runtime feature flags');
      }
    });
  }

  filteredFlags(): FeatureFlagItem[] {
    const q = this.searchFilter.toLowerCase().trim();
    if (!q) return this.flags();
    return this.flags().filter(f =>
      f.key.toLowerCase().includes(q) ||
      f.name.toLowerCase().includes(q) ||
      f.description.toLowerCase().includes(q)
    );
  }

  toggleFlag(f: FeatureFlagItem) {
    const newState = !f.enabled;
    this.isTogglingKey.set(f.key);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.flagUC.setFeatureFlag(f.key, newState).subscribe({
      next: () => {
        this.isTogglingKey.set(null);
        this.flags.update(list => list.map(item => item.key === f.key ? { ...item, enabled: newState } : item));
        this.successMessage.set(`Feature flag "${f.key}" is now ${newState ? 'ENABLED' : 'DISABLED'}.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isTogglingKey.set(null);
        this.errorMessage.set(err.error?.message || 'Failed to update feature flag');
      }
    });
  }
}
