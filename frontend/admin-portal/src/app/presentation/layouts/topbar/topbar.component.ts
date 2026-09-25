import { Component, OnInit, OnDestroy, HostListener, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { SystemUseCase } from '../../../core/usecases/system.usecase';
import { UserUseCase } from '../../../core/usecases/user.usecase';
import { ThemeService } from '../../../core/services/theme.service';
import { AuthService } from '../../../data/datasources/auth.service';
import { LayoutService } from '../../../core/services/layout.service';
import { AdminNotificationItem } from '../../../core/domain/entities/user.model';

export interface CommandItem {
  id: string;
  title: string;
  category: 'NAVIGATION' | 'ACTIONS' | 'SECURITY';
  icon: string;
  route?: string;
  action?: () => void;
  badge?: string;
}

@Component({
  selector: 'app-topbar',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './topbar.component.html',
  styleUrl: './topbar.component.css'
})
export class TopbarComponent implements OnInit, OnDestroy {
  apiStatus = signal<'online' | 'degraded' | 'offline'>('online');
  apiLatency = signal<number>(12);
  searchQuery = '';
  showNotifications = signal(false);
  notificationCount = signal(0);
  notifications = signal<AdminNotificationItem[]>([]);

  // User Dropdown & Change Password Modal State
  showUserDropdown = signal(false);
  showChangePasswordModal = signal(false);
  oldPassword = signal('');
  newPassword = signal('');
  confirmPassword = signal('');
  showOldPassword = signal(false);
  showNewPassword = signal(false);
  showConfirmPassword = signal(false);
  passwordSubmitting = signal(false);
  passwordError = signal('');
  passwordSuccess = signal('');

  // Command Palette State
  showCommandPalette = signal(false);
  commandFilter = signal('');
  selectedIndex = signal(0);

  commandList: CommandItem[] = [
    { id: 'dash', title: 'Dashboard Overview', category: 'NAVIGATION', icon: 'DB', route: '/admin/dashboard' },
    { id: 'pwd', title: 'Change Admin Password', category: 'SECURITY', icon: 'KEY', action: () => this.openChangePasswordModal(), badge: 'Security' },
    { id: 'users', title: 'User Management & Accounts', category: 'NAVIGATION', icon: 'USR', route: '/admin/users' },
    { id: 'book', title: 'Bookings & Orders Triage', category: 'NAVIGATION', icon: 'BKG', route: '/admin/appointments' },
    { id: 'id', title: 'Provider ID Verifications', category: 'NAVIGATION', icon: 'VRF', route: '/admin/verifications', badge: 'High Priority' },
    { id: 'checkr', title: 'Checkr Background Screenings', category: 'NAVIGATION', icon: 'BGC', route: '/admin/background-checks' },
    { id: 'rev', title: 'Reviews & Feedback Moderation', category: 'NAVIGATION', icon: 'RVW', route: '/admin/reviews' },
    { id: 'disp', title: 'Dispute Center & Resolution', category: 'NAVIGATION', icon: 'DSP', route: '/admin/disputes', badge: 'Escrow' },
    { id: 'rep', title: 'Platform Moderation & Abuse Reports', category: 'NAVIGATION', icon: 'REP', route: '/admin/reports', badge: 'Reports' },
    { id: 'cat', title: 'Marketplace Categories & Catalog', category: 'NAVIGATION', icon: 'CAT', route: '/admin/catalog' },
    { id: 'fin', title: 'Financial Analytics & Tax Exports', category: 'NAVIGATION', icon: 'FIN', route: '/admin/financial' },
    { id: 'pay', title: 'Stripe Payouts & Wallets', category: 'NAVIGATION', icon: 'PAY', route: '/admin/payouts' },
    { id: 'promo', title: 'Promo Codes & Vouchers', category: 'NAVIGATION', icon: 'PRO', route: '/admin/promos' },
    { id: 'subs', title: 'Provider Subscriptions', category: 'NAVIGATION', icon: 'SUB', route: '/admin/subscriptions' },
    { id: 'fraud', title: 'Fraud Detection & Anomaly Scanner', category: 'SECURITY', icon: 'FRD', route: '/admin/fraud', badge: 'Guard Active' },
    { id: 'rbac', title: 'RBAC Roles & Team Policies', category: 'SECURITY', icon: 'ACL', route: '/admin/roles' },
    { id: 'flags', title: 'Dynamic Feature Flags', category: 'SECURITY', icon: 'FLG', route: '/admin/feature-flags' },
    { id: 'health', title: 'System Health & Telemetry', category: 'SECURITY', icon: 'SYS', route: '/admin/system' },
    { id: 'legal', title: 'Legal & Compliance Documents', category: 'NAVIGATION', icon: 'LGL', route: '/admin/legal' },
    { id: 'tpl', title: 'Email & SMS Notification Templates', category: 'NAVIGATION', icon: 'TPL', route: '/admin/templates' },
    { id: 'cms', title: 'Public Website CMS & Hero Content', category: 'NAVIGATION', icon: 'CMS', route: '/admin/cms' }
  ];

  filteredCommands = computed(() => {
    const q = this.commandFilter().toLowerCase().trim();
    if (!q) return this.commandList;
    return this.commandList.filter(c => 
      c.title.toLowerCase().includes(q) || 
      c.category.toLowerCase().includes(q) ||
      (c.badge && c.badge.toLowerCase().includes(q))
    );
  });

  private ws?: WebSocket;
  private wsReconnectTimer?: any;
  private healthInterval?: any;

  constructor(
    private systemUC: SystemUseCase,
    private userUC: UserUseCase,
    private router: Router,
    public themeService: ThemeService,
    public authService: AuthService,
    public layoutService: LayoutService
  ) {}

  ngOnInit(): void {
    this.checkHealth();
    this.loadNotifications();
    this.initWebSocket();
    this.healthInterval = setInterval(() => this.checkHealth(), 30000);
  }

  ngOnDestroy(): void {
    if (this.wsReconnectTimer) clearTimeout(this.wsReconnectTimer);
    if (this.healthInterval) clearInterval(this.healthInterval);
    if (this.ws) {
      this.ws.close();
    }
  }

  private initWebSocket(): void {
    if (typeof window === 'undefined') return;

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host;
      const wsUrl = `${protocol}//${host}/ws/admin/events`;

      this.ws = new WebSocket(wsUrl);

      this.ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data);
          if (payload && payload.type) {
            const newNotif: AdminNotificationItem = {
              id: payload.id || `ws-${Date.now()}`,
              type: payload.type,
              severity: payload.severity || 'INFO',
              title: payload.title || 'Platform Real-Time Event',
              message: payload.message || payload.content || 'New activity detected.',
              route: payload.route || '/admin/dashboard',
              is_read: false,
              created_at: new Date().toISOString()
            };
            this.notifications.update(list => [newNotif, ...list.slice(0, 19)]);
            this.notificationCount.update(c => c + 1);
          }
        } catch {
          // ignore non-json messages (e.g. heartbeat ping)
        }
      };

      this.ws.onclose = () => {
        this.wsReconnectTimer = setTimeout(() => this.initWebSocket(), 10000);
      };

      this.ws.onerror = () => {
        if (this.ws) this.ws.close();
      };
    } catch (e) {
      console.warn('Admin WebSocket connection fallback:', e);
    }
  }

  @HostListener('window:keydown', ['$event'])
  handleGlobalShortcut(event: KeyboardEvent): void {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      this.toggleCommandPalette();
    } else if (event.key === 'Escape' && this.showCommandPalette()) {
      this.closeCommandPalette();
    } else if (this.showCommandPalette()) {
      if (event.key === 'ArrowDown') {
        event.preventDefault();
        const list = this.filteredCommands();
        if (list.length > 0) {
          this.selectedIndex.update(i => (i + 1) % list.length);
        }
      } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        const list = this.filteredCommands();
        if (list.length > 0) {
          this.selectedIndex.update(i => (i - 1 + list.length) % list.length);
        }
      } else if (event.key === 'Enter') {
        event.preventDefault();
        const list = this.filteredCommands();
        if (list.length > 0) {
          const item = list[this.selectedIndex()];
          if (item) this.executeCommand(item);
        }
      }
    }
  }

  toggleCommandPalette(): void {
    this.showCommandPalette.update(v => !v);
    if (this.showCommandPalette()) {
      this.commandFilter.set('');
      this.selectedIndex.set(0);
      setTimeout(() => {
        const input = document.getElementById('paletteSearchInput');
        if (input) input.focus();
      }, 50);
    }
  }

  closeCommandPalette(): void {
    this.showCommandPalette.set(false);
  }

  executeCommand(item: CommandItem): void {
    this.closeCommandPalette();
    if (item.action) {
      item.action();
    } else if (item.route) {
      this.router.navigateByUrl(item.route);
    }
  }

  checkHealth(): void {
    const start = performance.now();
    this.systemUC.getSystemHealth().subscribe({
      next: (h) => {
        const duration = Math.max(1, Math.round(performance.now() - start));
        this.apiLatency.set(duration);
        this.apiStatus.set(h.status === 'healthy' ? 'online' : 'degraded');
      },
      error: () => {
        this.apiStatus.set('offline');
      }
    });
  }

  loadNotifications(): void {
    this.userUC.getNotificationFeed().subscribe({
      next: (res) => {
        const list = res.notifications || [];
        this.notifications.set(list);
        this.notificationCount.set(list.filter(n => !n.is_read).length);
      },
      error: () => {}
    });
  }

  toggleTheme(): void {
    this.themeService.toggleTheme();
  }

  onSearch(): void {
    this.toggleCommandPalette();
  }

  toggleNotifications(): void {
    this.showNotifications.update(v => !v);
    if (this.showNotifications()) {
      this.loadNotifications();
    }
  }

  markAllAsRead(): void {
    this.notifications.update(items => items.map(n => ({ ...n, is_read: true })));
    this.notificationCount.set(0);
  }

  navigateToAlert(item: AdminNotificationItem): void {
    item.is_read = true;
    this.notificationCount.update(c => Math.max(0, c - 1));
    this.showNotifications.set(false);
    if (item.route) {
      this.router.navigateByUrl(item.route);
    }
  }

  toggleUserDropdown(): void {
    this.showUserDropdown.update(v => !v);
    if (this.showUserDropdown()) {
      this.showNotifications.set(false);
    }
  }

  closeUserDropdown(): void {
    this.showUserDropdown.set(false);
  }

  openChangePasswordModal(): void {
    this.showUserDropdown.set(false);
    this.oldPassword.set('');
    this.newPassword.set('');
    this.confirmPassword.set('');
    this.passwordError.set('');
    this.passwordSuccess.set('');
    this.showOldPassword.set(false);
    this.showNewPassword.set(false);
    this.showConfirmPassword.set(false);
    this.showChangePasswordModal.set(true);
  }

  closeChangePasswordModal(): void {
    this.showChangePasswordModal.set(false);
    this.passwordError.set('');
    this.passwordSuccess.set('');
  }

  submitChangePassword(): void {
    const oldP = this.oldPassword().trim();
    const newP = this.newPassword().trim();
    const confP = this.confirmPassword().trim();

    this.passwordError.set('');
    this.passwordSuccess.set('');

    if (!oldP) {
      this.passwordError.set('Current password is required.');
      return;
    }
    if (!newP) {
      this.passwordError.set('New password is required.');
      return;
    }
    if (newP.length < 6) {
      this.passwordError.set('New password must be at least 6 characters long.');
      return;
    }
    if (newP !== confP) {
      this.passwordError.set('New password and confirmation do not match.');
      return;
    }
    if (oldP === newP) {
      this.passwordError.set('New password must be different from current password.');
      return;
    }

    this.passwordSubmitting.set(true);

    this.authService.changePassword(oldP, newP).subscribe({
      next: (res: any) => {
        this.passwordSubmitting.set(false);
        this.passwordSuccess.set(res?.detail || 'Password changed successfully!');
        setTimeout(() => {
          this.closeChangePasswordModal();
        }, 1600);
      },
      error: (err: any) => {
        this.passwordSubmitting.set(false);
        const msg = err?.error?.detail || err?.error?.message || err?.error?.error || err?.message || 'Failed to change password. Please check your current password.';
        this.passwordError.set(typeof msg === 'string' ? msg : JSON.stringify(msg));
      }
    });
  }

  navigateToSettings(): void {
    this.closeUserDropdown();
    this.router.navigate(['/admin/settings'], { queryParams: { tab: 'ADMIN_SECURITY' } });
  }
}
