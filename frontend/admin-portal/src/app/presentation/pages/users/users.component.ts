import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';
import { UserUseCase } from '../../../core/usecases/user.usecase';
import { RoleUseCase } from '../../../core/usecases/role.usecase';
import { SubscriptionUseCase } from '../../../core/usecases/subscription.usecase';
import { DashboardUseCase } from '../../../core/usecases/dashboard.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { AdminUser, ImpersonationResult, ProviderFunnelData } from '../../../core/domain/entities/user.model';
import { AdminRole } from '../../../core/domain/entities/role.model';
import { SubscriptionPlan } from '../../../core/domain/entities/subscription.model';
import { DashboardStats } from '../../../core/domain/entities/dashboard.model';

export type UserTabFilter = 'ALL' | 'SEEKER' | 'PROVIDER' | 'STAFF' | 'VERIFIED' | 'SUSPENDED';

@Component({
  selector: 'app-users',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './users.component.html',
  styleUrl: './users.component.css'
})
export class UsersComponent implements OnInit {
  users = signal<AdminUser[]>([]);
  roles = signal<AdminRole[]>([]);
  subscriptionPlans = signal<SubscriptionPlan[]>([]);
  funnel = signal<ProviderFunnelData | null>(null);
  dashboardStats = signal<DashboardStats | null>(null);
  showFunnel = signal(true);
  totalCount = signal(0);
  
  // Search & Filter state
  searchQuery = '';
  selectedUserType = '';
  activeTab = signal<UserTabFilter>('ALL');
  
  // Pagination state
  currentPage = signal(1);
  pageSize = signal(10);
  
  // Modal & Selection state
  selectedUser = signal<AdminUser | null>(null);
  selectedRoleId = '';
  activeModalTab = signal<'profile' | 'rbac' | 'impersonation'>('profile');
  
  // Async states
  isLoading = signal(false);
  isActioning = signal(false);
  actionMessage = signal<string | null>(null);
  isSuccess = signal(true);
  impersonationData = signal<ImpersonationResult | null>(null);
  impersonationReason = '';

  // Multi-selection batch operations state
  selectedUserIds = signal<string[]>([]);
  isAllSelected = computed(() => {
    const list = this.filteredUsers();
    return list.length > 0 && list.every(u => this.selectedUserIds().includes(u.id));
  });
  selectedCount = computed(() => this.selectedUserIds().length);

  // Global KPIs from Dashboard Telemetry (stable across filter changes)
  totalUsers = computed(() => this.dashboardStats()?.totalUsers ?? this.totalCount() ?? this.users().length);
  seekersCount = computed(() => this.dashboardStats()?.activeSeekers ?? 0);
  providersCount = computed(() => this.dashboardStats()?.activeProviders ?? 0);
  staffCount = computed(() => this.dashboardStats()?.totalStaff ?? 0);
  verifiedCount = computed(() => this.dashboardStats()?.totalVerified ?? 0);
  suspendedCount = computed(() => this.dashboardStats()?.totalSuspended ?? 0);
  
  totalPages = computed(() => Math.max(1, Math.ceil(this.totalCount() / this.pageSize())));

  // Results loaded from server
  filteredUsers = computed(() => this.users());

  constructor(
    private userUC: UserUseCase,
    private roleUC: RoleUseCase,
    private subUC: SubscriptionUseCase,
    private dashboardUC: DashboardUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit() {
    this.loadStats();
    this.loadUsers();
    this.loadRoles();
    this.loadFunnel();
    this.loadSubscriptionPlans();
  }

  loadStats() {
    this.dashboardUC.getDashboardStats().subscribe({
      next: (stats) => this.dashboardStats.set(stats),
      error: () => {}
    });
  }

  loadSubscriptionPlans() {
    this.subUC.listPlans().subscribe({
      next: (plans) => {
        const paidPlans = (plans || []).filter(p =>
          p.isActive !== false &&
          p.tier?.toUpperCase() !== 'BASIC' &&
          p.tier?.toUpperCase() !== 'FREE' &&
          p.tier?.toUpperCase() !== 'NONE' &&
          (p.price || 0) > 0
        );
        this.subscriptionPlans.set(paidPlans);
      },
      error: () => {}
    });
  }

  getPlansForInterval(interval?: string): SubscriptionPlan[] {
    const all = this.subscriptionPlans();
    if (!all || all.length === 0) return [];
    if (interval) {
      const matched = all.filter(p => p.interval?.toLowerCase() === interval.toLowerCase());
      if (matched.length > 0) return matched;
    }
    return all;
  }

  loadFunnel() {
    this.userUC.getProviderFunnel().subscribe({
      next: (data) => this.funnel.set(data),
      error: () => {}
    });
  }

  setTab(tab: UserTabFilter) {
    this.activeTab.set(tab);
    if (tab === 'SEEKER') {
      this.selectedUserType = 'seeker';
    } else if (tab === 'PROVIDER') {
      this.selectedUserType = 'provider';
    } else {
      this.selectedUserType = '';
    }
    this.currentPage.set(1);
    this.loadUsers();
  }

  loadUsers() {
    this.isLoading.set(true);
    const tab = this.activeTab();

    let userType: string | undefined = this.selectedUserType || undefined;
    let isStaff: boolean | undefined = undefined;
    let isIdentityVerified: boolean | undefined = undefined;
    let isActive: boolean | undefined = undefined;

    if (tab === 'SEEKER') {
      userType = 'seeker';
    } else if (tab === 'PROVIDER') {
      userType = 'provider';
    } else if (tab === 'STAFF') {
      isStaff = true;
    } else if (tab === 'VERIFIED') {
      isIdentityVerified = true;
    } else if (tab === 'SUSPENDED') {
      isActive = false;
    }

    this.userUC.listUsers({
      search: this.searchQuery.trim() || undefined,
      userType,
      isStaff,
      isIdentityVerified,
      isActive,
      page: this.currentPage(),
      pageSize: this.pageSize()
    }).subscribe({
      next: (res) => {
        this.users.set(res.results);
        this.totalCount.set(res.count);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.actionMessage.set(err?.error?.error || 'Failed to fetch user directory.');
        this.isSuccess.set(false);
      }
    });
  }

  loadRoles() {
    this.roleUC.listRoles().subscribe({
      next: (data) => this.roles.set(data),
      error: () => {}
    });
  }

  resetFilters() {
    this.searchQuery = '';
    this.selectedUserType = '';
    this.activeTab.set('ALL');
    this.currentPage.set(1);
    this.loadUsers();
  }

  changePage(newPage: number) {
    if (newPage >= 1 && newPage <= this.totalPages()) {
      this.currentPage.set(newPage);
      this.loadUsers();
    }
  }

  changePageSize(newSize: number) {
    this.pageSize.set(newSize);
    this.currentPage.set(1);
    this.loadUsers();
  }

  openDetail(user: AdminUser, defaultTab: 'profile' | 'rbac' | 'impersonation' = 'profile') {
    this.activeModalTab.set(defaultTab);
    this.userUC.getUserById(user.id).subscribe({
      next: (fullUser) => {
        this.selectedUser.set(fullUser);
        this.selectedRoleId = fullUser.roleId || '';
      },
      error: () => {
        this.selectedUser.set(user);
        this.selectedRoleId = user.roleId || '';
      }
    });
  }

  closeDetail() {
    this.selectedUser.set(null);
    this.impersonationData.set(null);
  }

  toggleActive(user: AdminUser, makeActive: boolean) {
    this.userUC.updateUser(user.id, { isActive: makeActive }).subscribe({
      next: () => {
        this.actionMessage.set(makeActive ? `User ${user.email} has been activated successfully.` : `User ${user.email} has been suspended.`);
        this.isSuccess.set(true);
        this.loadStats();
        this.loadUsers();
      },
      error: (err) => {
        this.actionMessage.set(err?.error?.error || err?.error?.message || `Failed to ${makeActive ? 'activate' : 'suspend'} user.`);
        this.isSuccess.set(false);
      }
    });
  }

  async deleteUser(user: AdminUser): Promise<void> {
    const fullName = `${user.firstName || ''} ${user.lastName || ''}`.trim();
    const displayName = fullName ? `${fullName} (${user.email})` : user.email;
    const confirmed = await this.dialog.dangerConfirm(
      'Delete User Account',
      `Are you sure you want to permanently delete user account "${displayName}"? This will terminate active sessions and remove the account.`,
      'Delete Account'
    );
    if (!confirmed) return;

    this.isActioning.set(true);
    this.userUC.deleteUser(user.id).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.actionMessage.set(`User account ${user.email} has been deleted successfully.`);
        this.isSuccess.set(true);
        this.loadStats();
        this.loadUsers();
      },
      error: (err) => {
        this.isActioning.set(false);
        this.actionMessage.set(err?.error?.error || err?.error?.message || 'Failed to delete user account.');
        this.isSuccess.set(false);
      }
    });
  }

  toggleSelectAll(checked: boolean): void {
    if (checked) {
      this.selectedUserIds.set(this.filteredUsers().map(u => u.id));
    } else {
      this.selectedUserIds.set([]);
    }
  }

  toggleSelectUser(id: string): void {
    const set = new Set(this.selectedUserIds());
    if (set.has(id)) {
      set.delete(id);
    } else {
      set.add(id);
    }
    this.selectedUserIds.set(Array.from(set));
  }

  isUserSelected(id: string): boolean {
    return this.selectedUserIds().includes(id);
  }

  clearSelection(): void {
    this.selectedUserIds.set([]);
  }

  bulkSuspend(): void {
    const ids = this.selectedUserIds();
    if (!ids.length) return;
    
    this.isActioning.set(true);
    let completed = 0;
    let failed = 0;
    ids.forEach(id => {
      this.userUC.updateUser(id, { isActive: false }).subscribe({
        next: () => {
          completed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Successfully suspended ${completed} selected user accounts.` + (failed > 0 ? ` (${failed} failed)` : ''));
            this.isSuccess.set(true);
            this.clearSelection();
            this.loadUsers();
          }
        },
        error: () => {
          failed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Suspended ${completed} user(s), ${failed} failed.`);
            this.isSuccess.set(completed > 0);
            this.clearSelection();
            this.loadUsers();
          }
        }
      });
    });
  }

  bulkActivate(): void {
    const ids = this.selectedUserIds();
    if (!ids.length) return;
    
    this.isActioning.set(true);
    let completed = 0;
    let failed = 0;
    ids.forEach(id => {
      this.userUC.updateUser(id, { isActive: true }).subscribe({
        next: () => {
          completed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Successfully activated ${completed} selected user accounts.` + (failed > 0 ? ` (${failed} failed)` : ''));
            this.isSuccess.set(true);
            this.clearSelection();
            this.loadUsers();
          }
        },
        error: () => {
          failed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Activated ${completed} user(s), ${failed} failed.`);
            this.isSuccess.set(completed > 0);
            this.clearSelection();
            this.loadUsers();
          }
        }
      });
    });
  }

  async bulkDelete(): Promise<void> {
    const ids = this.selectedUserIds();
    if (!ids.length) return;

    const confirmed = await this.dialog.dangerConfirm(
      'Bulk Delete Users',
      `Are you sure you want to permanently delete ${ids.length} selected user accounts? This action cannot be undone.`,
      'Delete Selected Accounts'
    );
    if (!confirmed) return;

    this.isActioning.set(true);
    let completed = 0;
    let failed = 0;
    ids.forEach(id => {
      this.userUC.deleteUser(id).subscribe({
        next: () => {
          completed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Successfully deleted ${completed} user account(s).` + (failed > 0 ? ` (${failed} failed)` : ''));
            this.isSuccess.set(true);
            this.clearSelection();
            this.loadUsers();
          }
        },
        error: () => {
          failed++;
          if (completed + failed === ids.length) {
            this.isActioning.set(false);
            this.actionMessage.set(`Deleted ${completed} user(s), ${failed} failed.`);
            this.isSuccess.set(completed > 0);
            this.clearSelection();
            this.loadUsers();
          }
        }
      });
    });
  }

  exportSelectedCSV(): void {
    const ids = this.selectedUserIds();
    const data = this.users().filter(u => ids.includes(u.id));
    if (!data.length) return;

    const headers = ['ID', 'Email', 'First Name', 'Last Name', 'Phone', 'Type', 'Active', 'Staff', 'Verified ID', 'Subscription', 'Checkr Status', 'Created At'];
    const rows = data.map(u => [
      u.id,
      u.email,
      u.firstName || '',
      u.lastName || '',
      u.phone || '',
      u.userType,
      u.isActive ? 'Active' : 'Suspended',
      u.isStaff ? 'Staff' : 'Standard',
      u.isIdentityVerified ? 'Verified' : 'Unverified',
      u.subscriptionTier || 'NONE',
      u.checkrStatus || 'NOT_SUBMITTED',
      u.createdAt || ''
    ]);

    const csvContent = [headers, ...rows].map(e => e.map(cell => `"${(cell + '').replace(/"/g, '""')}"`).join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `selected_users_${new Date().toISOString().split('T')[0]}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  exportAllCSV(): void {
    const list = this.filteredUsers();
    if (!list.length) return;

    import('../../../core/utils/export.util').then(({ exportToCsv }) => {
      exportToCsv(list, 'users_directory', [
        { key: 'id', header: 'User ID' },
        { key: 'email', header: 'Email' },
        { key: 'firstName', header: 'First Name' },
        { key: 'lastName', header: 'Last Name' },
        { key: 'phone', header: 'Phone' },
        { key: 'userType', header: 'Type' },
        { key: 'isActive', header: 'Status', format: (v) => v ? 'Active' : 'Suspended' },
        { key: 'isStaff', header: 'Staff', format: (v) => v ? 'Staff' : 'Standard' },
        { key: 'isIdentityVerified', header: 'Identity Verified', format: (v) => v ? 'Yes' : 'No' },
        { key: 'subscriptionTier', header: 'Subscription' },
        { key: 'checkrStatus', header: 'Checkr Status' },
        { key: 'createdAt', header: 'Joined Date' }
      ]);
    });
  }

  impersonate(user: AdminUser) {
    if (!this.impersonationReason.trim()) {
      this.impersonationReason = 'Customer Support / Account Assistance';
    }

    this.isActioning.set(true);
    this.actionMessage.set(null);

    this.userUC.impersonateUser(user.id).subscribe({
      next: (res) => {
        this.isActioning.set(false);
        this.impersonationData.set(res);
        this.activeModalTab.set('impersonation');
        this.actionMessage.set(`Time-bound (15m) impersonation session active for ${user.email}!`);
        this.isSuccess.set(true);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.actionMessage.set(err.error?.message || 'Failed to impersonate user');
        this.isSuccess.set(false);
      }
    });
  }

  downloadGDPR(user: AdminUser) {
    this.userUC.getGDPRUserData(user.id).subscribe({
      next: (data) => {
        const jsonStr = JSON.stringify(data, null, 2);
        const blob = new Blob([jsonStr], { type: 'application/json' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `gdpr_data_${user.email.replace(/[@.]/g, '_')}_${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      },
      error: (err) => {
        this.actionMessage.set(err.error?.message || 'Failed to export GDPR data package');
        this.isSuccess.set(false);
      }
    });
  }

  exportAllUsersCSV() {
    const data = this.users();
    if (!data.length) return;

    const headers = ['ID', 'Email', 'First Name', 'Last Name', 'Phone', 'Type', 'Active', 'Staff', 'Verified ID', 'Subscription', 'Checkr Status', 'Created At'];
    const rows = data.map(u => [
      `"${u.id}"`,
      `"${u.email}"`,
      `"${u.firstName || ''}"`,
      `"${u.lastName || ''}"`,
      `"${u.phone || ''}"`,
      `"${u.userType}"`,
      u.isActive ? 'Active' : 'Suspended',
      u.isStaff ? 'Yes' : 'No',
      u.isIdentityVerified ? 'Verified' : 'Unverified',
      `"${u.subscriptionTier || 'Free'}"`,
      `"${u.checkrStatus || 'none'}"`,
      `"${u.createdAt}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `users_directory_export_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }

  // Create User Modal State
  showCreateModal = signal(false);
  createUserForm = {
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    phone: '',
    gender: '',
    bio: '',
    address: '',
    city: '',
    state: '',
    zipCode: '',
    country: 'United States',
    service: '',
    userType: 'seeker',
    isStaff: false,
    isActive: true,
    isVerified: true,
    isIdentityVerified: false,
    subscriptionTier: 'SILVER',
    subscriptionInterval: 'month',
    preferredPaymentMode: 'IN_APP',
    roleId: ''
  };

  // Edit User Modal State
  showEditModal = signal(false);
  editUserForm: any = {
    id: '',
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    phone: '',
    gender: '',
    bio: '',
    address: '',
    city: '',
    state: '',
    zipCode: '',
    country: 'United States',
    service: '',
    userType: 'seeker',
    isStaff: false,
    isActive: true,
    isVerified: false,
    isIdentityVerified: false,
    subscriptionTier: 'SILVER',
    subscriptionInterval: 'month',
    preferredPaymentMode: 'IN_APP',
    roleId: ''
  };

  openCreateModal() {
    const defaultTier = this.subscriptionPlans()[0]?.tier || 'SILVER';
    this.createUserForm = {
      email: '',
      password: '',
      firstName: '',
      lastName: '',
      phone: '',
      gender: '',
      bio: '',
      address: '',
      city: '',
      state: '',
      zipCode: '',
      country: 'United States',
      service: '',
      userType: 'seeker',
      isStaff: false,
      isActive: true,
      isVerified: true,
      isIdentityVerified: false,
      subscriptionTier: defaultTier,
      subscriptionInterval: 'month',
      preferredPaymentMode: 'IN_APP',
      roleId: ''
    };
    this.showCreateModal.set(true);
  }

  closeCreateModal() {
    this.showCreateModal.set(false);
  }

  submitCreateUser() {
    if (!this.createUserForm.email || !this.createUserForm.password) {
      this.actionMessage.set('Email and password are required.');
      this.isSuccess.set(false);
      return;
    }

    this.isActioning.set(true);
    this.userUC.createUser(this.createUserForm).subscribe({
      next: (user) => {
        this.isActioning.set(false);
        this.showCreateModal.set(false);
        this.actionMessage.set(`User ${user.email} created successfully!`);
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.isActioning.set(false);
        this.actionMessage.set(err?.error?.message || err?.error?.error || 'Failed to create user account.');
        this.isSuccess.set(false);
      }
    });
  }

  openEditModal(user: AdminUser) {
    const defaultTier = this.subscriptionPlans()[0]?.tier || 'SILVER';
    const currentTier = user.subscriptionTier && user.subscriptionTier !== 'BASIC' && user.subscriptionTier !== 'NONE' && user.subscriptionTier !== 'Free'
      ? user.subscriptionTier
      : defaultTier;

    this.editUserForm = {
      id: user.id,
      email: user.email,
      password: '',
      firstName: user.firstName,
      lastName: user.lastName,
      phone: user.phone,
      gender: user.gender || '',
      bio: user.bio || '',
      address: user.address || '',
      city: user.city || '',
      state: user.state || '',
      zipCode: user.zipCode || '',
      country: user.country || 'United States',
      service: user.service || '',
      userType: user.userType,
      isStaff: user.isStaff,
      isActive: user.isActive,
      isVerified: user.isVerified,
      isIdentityVerified: user.isIdentityVerified,
      subscriptionTier: currentTier,
      subscriptionInterval: user.subscriptionInterval || 'month',
      preferredPaymentMode: user.preferredPaymentMode || 'IN_APP',
      roleId: user.roleId || ''
    };
    this.showEditModal.set(true);
  }

  closeEditModal() {
    this.showEditModal.set(false);
  }

  submitEditUser() {
    if (!this.editUserForm.id || !this.editUserForm.email) {
      this.actionMessage.set('Email is required.');
      this.isSuccess.set(false);
      return;
    }

    this.isActioning.set(true);
    const payload: any = {
      email: this.editUserForm.email,
      firstName: this.editUserForm.firstName,
      lastName: this.editUserForm.lastName,
      phone: this.editUserForm.phone,
      gender: this.editUserForm.gender,
      bio: this.editUserForm.bio,
      address: this.editUserForm.address,
      city: this.editUserForm.city,
      state: this.editUserForm.state,
      zipCode: this.editUserForm.zipCode,
      country: this.editUserForm.country,
      service: this.editUserForm.service,
      userType: this.editUserForm.userType,
      isStaff: this.editUserForm.isStaff,
      isActive: this.editUserForm.isActive,
      isVerified: this.editUserForm.isVerified,
      isIdentityVerified: this.editUserForm.isIdentityVerified,
      subscriptionTier: this.editUserForm.subscriptionTier,
      subscriptionInterval: this.editUserForm.subscriptionInterval,
      preferredPaymentMode: this.editUserForm.preferredPaymentMode,
      roleId: this.editUserForm.roleId
    };
    if (this.editUserForm.password) {
      payload.password = this.editUserForm.password;
    }

    this.userUC.updateUser(this.editUserForm.id, payload).subscribe({
      next: (user) => {
        this.isActioning.set(false);
        this.showEditModal.set(false);
        this.actionMessage.set(`User ${user.email} updated successfully!`);
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.isActioning.set(false);
        this.actionMessage.set(err?.error?.message || err?.error?.error || 'Failed to update user.');
        this.isSuccess.set(false);
      }
    });
  }

  toggleStaff(user: AdminUser) {
    const target = !user.isStaff;
    this.userUC.updateUser(user.id, { isStaff: target }).subscribe({
      next: () => {
        this.actionMessage.set(`Staff status for ${user.email} set to ${target ? 'STAFF (Admin)' : 'STANDARD'}.`);
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.actionMessage.set(err?.error?.message || 'Failed to update staff status.');
        this.isSuccess.set(false);
      }
    });
  }

  toggleEmailVerified(user: AdminUser) {
    const target = !user.isVerified;
    this.userUC.updateUser(user.id, { isVerified: target }).subscribe({
      next: () => {
        this.actionMessage.set(`Email verification for ${user.email} set to ${target ? 'VERIFIED' : 'UNVERIFIED'}.`);
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.actionMessage.set(err?.error?.message || 'Failed to update email verification.');
        this.isSuccess.set(false);
      }
    });
  }

  toggleIdentityVerified(user: AdminUser) {
    const target = !user.isIdentityVerified;
    this.userUC.updateUser(user.id, { isIdentityVerified: target }).subscribe({
      next: () => {
        this.actionMessage.set(`Identity verification for ${user.email} set to ${target ? 'VERIFIED ID' : 'UNVERIFIED'}.`);
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.actionMessage.set(err?.error?.message || 'Failed to update identity verification.');
        this.isSuccess.set(false);
      }
    });
  }

  assignRole() {
    const user = this.selectedUser();
    if (!user || !this.selectedRoleId) return;

    this.isActioning.set(true);
    this.userUC.assignUserRole(user.id, this.selectedRoleId).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.actionMessage.set('Role assigned successfully!');
        this.isSuccess.set(true);
        this.loadUsers();
      },
      error: (err) => {
        this.isActioning.set(false);
        this.actionMessage.set(err.error?.message || 'Failed to assign role');
        this.isSuccess.set(false);
      }
    });
  }
}

