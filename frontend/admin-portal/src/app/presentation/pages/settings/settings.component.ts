import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SettingsUseCase } from '../../../core/usecases/settings.usecase';
import { PlatformSettings } from '../../../core/domain/entities/settings.model';

export type SettingsMainTab = 'PARAMETERS' | 'BROADCAST' | 'INFRA_MESH';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.css'
})
export class SettingsComponent implements OnInit {
  activeTab = signal<SettingsMainTab>('PARAMETERS');
  settings = signal<PlatformSettings>({
    backgroundCheckPaymentMode: 'PLATFORM_PAYS',
    backgroundCheckFee: 29.99,
    broadcastRadiusKm: 25.0,
    matchRadiusKm: 15.0
  });

  // Broadcast state
  broadcastTarget = 'ALL';
  broadcastTitle = '';
  broadcastMessage = '';

  loadingSettings = signal<boolean>(false);
  savingSettings = signal<boolean>(false);
  broadcasting = signal<boolean>(false);
  clearingCache = signal<boolean>(false);

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  constructor(private settingsUseCase: SettingsUseCase) {}

  ngOnInit(): void {
    this.fetchSettings();
  }

  fetchSettings(): void {
    this.loadingSettings.set(true);
    this.settingsUseCase.getPlatformSettings().subscribe({
      next: (data) => {
        this.settings.set(data);
        this.loadingSettings.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load platform settings');
        this.loadingSettings.set(false);
      }
    });
  }

  saveSettings(): void {
    this.savingSettings.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.settingsUseCase.updatePlatformSettings(this.settings()).subscribe({
      next: (data) => {
        this.settings.set(data);
        this.savingSettings.set(false);
        this.successMessage.set('Platform operational parameters saved successfully!');
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to update settings');
        this.savingSettings.set(false);
      }
    });
  }

  sendBroadcast(): void {
    if (!this.broadcastTitle.trim() || !this.broadcastMessage.trim()) {
      this.errorMessage.set('Please provide both broadcast title and message body');
      return;
    }

    this.broadcasting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const target = this.broadcastTarget === 'ALL' ? '' : this.broadcastTarget;

    this.settingsUseCase.broadcastNotification(target, this.broadcastTitle, this.broadcastMessage).subscribe({
      next: (res) => {
        this.broadcasting.set(false);
        this.successMessage.set(`Broadcast notification dispatched to ${res.recipients} active device recipients!`);
        this.broadcastTitle = '';
        this.broadcastMessage = '';
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to dispatch broadcast notification');
        this.broadcasting.set(false);
      }
    });
  }

  flushCache(): void {
    this.clearingCache.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.settingsUseCase.clearCache().subscribe({
      next: (res) => {
        this.clearingCache.set(false);
        this.successMessage.set(res.message);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to clear application cache');
        this.clearingCache.set(false);
      }
    });
  }
}
