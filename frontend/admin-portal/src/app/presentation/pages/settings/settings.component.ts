import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { SettingsUseCase } from '../../../core/usecases/settings.usecase';
import { AuthService } from '../../../data/datasources/auth.service';
import { PlatformSettings } from '../../../core/domain/entities/settings.model';

export type SettingsMainTab = 'PARAMETERS' | 'BROADCAST' | 'INFRA_MESH' | 'ADMIN_SECURITY';

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

  // Password Change State
  adminOldPassword = signal('');
  adminNewPassword = signal('');
  adminConfirmPassword = signal('');
  showAdminOldPassword = signal(false);
  showAdminNewPassword = signal(false);
  showAdminConfirmPassword = signal(false);
  changingPassword = signal(false);
  changePasswordSuccess = signal<string | null>(null);
  changePasswordError = signal<string | null>(null);

  loadingSettings = signal<boolean>(false);
  savingSettings = signal<boolean>(false);
  broadcasting = signal<boolean>(false);
  clearingCache = signal<boolean>(false);

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  constructor(
    private settingsUseCase: SettingsUseCase,
    public authService: AuthService,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      if (params['tab'] === 'ADMIN_SECURITY') {
        this.activeTab.set('ADMIN_SECURITY');
      }
    });
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

  updateAdminPassword(): void {
    const oldP = this.adminOldPassword().trim();
    const newP = this.adminNewPassword().trim();
    const confP = this.adminConfirmPassword().trim();

    this.changePasswordError.set(null);
    this.changePasswordSuccess.set(null);

    if (!oldP) {
      this.changePasswordError.set('Current password is required.');
      return;
    }
    if (!newP) {
      this.changePasswordError.set('New password is required.');
      return;
    }
    if (newP.length < 6) {
      this.changePasswordError.set('New password must be at least 6 characters long.');
      return;
    }
    if (newP !== confP) {
      this.changePasswordError.set('New password and confirmation do not match.');
      return;
    }
    if (oldP === newP) {
      this.changePasswordError.set('New password must be different from current password.');
      return;
    }

    this.changingPassword.set(true);

    this.authService.changePassword(oldP, newP).subscribe({
      next: (res: any) => {
        this.changingPassword.set(false);
        this.changePasswordSuccess.set(res?.detail || 'Administrator password updated successfully!');
        this.adminOldPassword.set('');
        this.adminNewPassword.set('');
        this.adminConfirmPassword.set('');
        setTimeout(() => this.changePasswordSuccess.set(null), 5000);
      },
      error: (err: any) => {
        this.changingPassword.set(false);
        const msg = err?.error?.detail || err?.error?.message || err?.error?.error || err?.message || 'Failed to change password. Please check your current password.';
        this.changePasswordError.set(typeof msg === 'string' ? msg : JSON.stringify(msg));
      }
    });
  }
}
