import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { UserUseCase } from '../../../core/usecases/user.usecase';
import { RoleUseCase } from '../../../core/usecases/role.usecase';
import { SubscriptionUseCase } from '../../../core/usecases/subscription.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { AdminUser, ImpersonationResult, StaffNote } from '../../../core/domain/entities/user.model';
import { AdminRole } from '../../../core/domain/entities/role.model';
import { SubscriptionPlan } from '../../../core/domain/entities/subscription.model';

export type UserDetailTab = 'PROFILE' | 'SERVICES' | 'APPOINTMENTS' | 'WALLET' | 'DISPUTES' | 'LOGS' | 'NOTES';

@Component({
  selector: 'app-user-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './user-detail.component.html',
  styleUrl: './user-detail.component.css'
})
export class UserDetailComponent implements OnInit {
  userId = signal<string>('');
  user = signal<AdminUser | null>(null);
  gdprData = signal<any | null>(null);
  roles = signal<AdminRole[]>([]);
  subscriptionPlans = signal<SubscriptionPlan[]>([]);
  staffNotes = signal<StaffNote[]>([]);
  
  activeTab = signal<UserDetailTab>('PROFILE');
  isLoading = signal(true);
  isActioning = signal(false);
  
  // Note Form
  newNoteContent = '';
  newNoteCategory = 'GENERAL';
  newNotePinned = false;

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);
  
  // Modals & Action States
  showRoleModal = signal(false);
  selectedRoleId = '';
  impersonationData = signal<ImpersonationResult | null>(null);
  riskEvaluation = signal<any | null>(null);
  showRiskModal = signal(false);

  // Computed Properties
  userTypeBadge = computed(() => {
    const t = this.user()?.userType?.toLowerCase();
    return t === 'provider' ? 'badge-info' : 'badge-purple';
  });

  isProvider = computed(() => this.user()?.userType?.toLowerCase() === 'provider');

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private userUC: UserUseCase,
    private roleUC: RoleUseCase,
    private subUC: SubscriptionUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const id = params.get('id');
      if (id) {
        this.userId.set(id);
        this.loadFullUserDetails(id);
        this.loadStaffNotes(id);
      }
    });
    this.loadRoles();
    this.loadSubscriptionPlans();
  }

  loadSubscriptionPlans(): void {
    this.subUC.listPlans().subscribe({
      next: (plans) => {
        // Filter out inactive plans and free / basic starter tiers
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

  loadFullUserDetails(id: string): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.userUC.getUserById(id).subscribe({
      next: (userData) => {
        this.user.set(userData);
        this.selectedRoleId = userData.roleId || '';
        this.isLoading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to fetch user account profile.');
        this.isLoading.set(false);
      }
    });

    // Also fetch deep relational telemetry (Appointments, Wallet, Disputes, Reviews, Messages)
    this.userUC.getGDPRUserData(id).subscribe({
      next: (data) => {
        this.gdprData.set(data);
      },
      error: () => {}
    });
  }

  loadRoles(): void {
    this.roleUC.listRoles().subscribe({
      next: (data) => this.roles.set(data),
      error: () => {}
    });
  }

  loadStaffNotes(id: string): void {
    this.userUC.listStaffNotes(id).subscribe({
      next: (notes) => this.staffNotes.set(notes || []),
      error: () => {}
    });
  }

  submitStaffNote(): void {
    const id = this.userId();
    if (!id || !this.newNoteContent.trim()) return;

    this.isActioning.set(true);
    this.userUC.createStaffNote(id, {
      content: this.newNoteContent.trim(),
      category: this.newNoteCategory,
      is_pinned: this.newNotePinned
    }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.newNoteContent = '';
        this.newNotePinned = false;
        this.successMessage.set('Staff internal note recorded successfully.');
        this.loadStaffNotes(id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to save staff note.');
      }
    });
  }

  deleteStaffNote(noteId: string): void {
    const id = this.userId();
    if (!id || !noteId) return;

    this.userUC.deleteStaffNote(id, noteId).subscribe({
      next: () => {
        this.loadStaffNotes(id);
      },
      error: () => {}
    });
  }

  async toggleAccountStatus(): Promise<void> {
    const u = this.user();
    if (!u) return;

    if (!u.isActive) {
      this.isActioning.set(true);
      this.errorMessage.set(null);
      this.userUC.updateUser(u.id, { isActive: true }).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.successMessage.set(`Account ${u.email} has been activated.`);
          this.loadFullUserDetails(u.id);
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || err?.error?.error || 'Failed to activate account.');
        }
      });
    } else {
      const confirmed = await this.dialog.dangerConfirm(
        'Suspend User Account',
        `Are you sure you want to suspend account ${u.email}? The user will immediately be logged out and prohibited from platform actions.`,
        'Suspend Account'
      );
      if (!confirmed) return;

      this.isActioning.set(true);
      this.errorMessage.set(null);
      this.userUC.updateUser(u.id, { isActive: false }).subscribe({
        next: () => {
          this.isActioning.set(false);
          this.successMessage.set(`Account ${u.email} has been suspended.`);
          this.loadFullUserDetails(u.id);
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isActioning.set(false);
          this.errorMessage.set(err?.error?.message || err?.error?.error || 'Failed to suspend account.');
        }
      });
    }
  }

  async deleteUserAccount(): Promise<void> {
    const u = this.user();
    if (!u) return;

    const fullName = `${u.firstName || ''} ${u.lastName || ''}`.trim();
    const displayName = fullName ? `${fullName} (${u.email})` : u.email;
    const confirmed = await this.dialog.dangerConfirm(
      'Delete User Account',
      `Are you sure you want to permanently delete user account "${displayName}"? This will terminate active sessions and remove the account.`,
      'Delete User Account'
    );
    if (!confirmed) return;

    this.isActioning.set(true);
    this.errorMessage.set(null);
    this.userUC.deleteUser(u.id).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.router.navigate(['/admin/users']);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || err?.error?.error || 'Failed to delete user account.');
      }
    });
  }

  impersonate(): void {
    const u = this.user();
    if (!u) return;

    this.isActioning.set(true);
    this.errorMessage.set(null);

    this.userUC.impersonateUser(u.id).subscribe({
      next: (res) => {
        this.isActioning.set(false);
        this.impersonationData.set(res);
        this.successMessage.set(`Impersonation bearer session token generated for ${u.email}!`);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to generate impersonation token.');
      }
    });
  }

  exportGDPRArchive(): void {
    const u = this.user();
    if (!u) return;

    this.userUC.getGDPRUserData(u.id).subscribe({
      next: (data) => {
        const jsonStr = JSON.stringify(data, null, 2);
        const blob = new Blob([jsonStr], { type: 'application/json' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `user_archive_${u.email.replace(/[@.]/g, '_')}_${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
        this.successMessage.set(`Exported GDPR archive for ${u.email}`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to export GDPR data archive.');
      }
    });
  }

  openRoleModal(): void {
    this.selectedRoleId = this.user()?.roleId || '';
    this.showRoleModal.set(true);
  }

  closeRoleModal(): void {
    this.showRoleModal.set(false);
  }

  saveRoleAssignment(): void {
    const u = this.user();
    if (!u || !this.selectedRoleId) return;

    this.isActioning.set(true);
    this.userUC.assignUserRole(u.id, this.selectedRoleId).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.showRoleModal.set(false);
        this.successMessage.set('RBAC security role assigned successfully.');
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to assign security role.');
      }
    });
  }

  // Edit Profile Modal
  showEditModal = signal(false);
  editForm = {
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    password: '',
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

  openEditModal(): void {
    const u = this.user();
    if (!u) return;
    const defaultTier = this.subscriptionPlans()[0]?.tier || 'SILVER';
    const currentTier = u.subscriptionTier && u.subscriptionTier !== 'BASIC' && u.subscriptionTier !== 'NONE' && u.subscriptionTier !== 'Free'
      ? u.subscriptionTier
      : defaultTier;

    this.editForm = {
      firstName: u.firstName || '',
      lastName: u.lastName || '',
      email: u.email || '',
      phone: u.phone || '',
      password: '',
      gender: u.gender || '',
      bio: u.bio || '',
      address: u.address || '',
      city: u.city || '',
      state: u.state || '',
      zipCode: u.zipCode || '',
      country: u.country || 'United States',
      service: u.service || '',
      userType: u.userType || 'seeker',
      isStaff: u.isStaff || false,
      isActive: u.isActive ?? true,
      isVerified: u.isVerified || false,
      isIdentityVerified: u.isIdentityVerified || false,
      subscriptionTier: currentTier,
      subscriptionInterval: u.subscriptionInterval || 'month',
      preferredPaymentMode: u.preferredPaymentMode || 'IN_APP',
      roleId: u.roleId || ''
    };
    this.showEditModal.set(true);
  }

  closeEditModal(): void {
    this.showEditModal.set(false);
  }

  saveUserDetails(): void {
    const u = this.user();
    if (!u || !this.editForm.email) return;

    this.isActioning.set(true);
    this.errorMessage.set(null);

    const payload: any = {
      email: this.editForm.email,
      firstName: this.editForm.firstName,
      lastName: this.editForm.lastName,
      phone: this.editForm.phone,
      gender: this.editForm.gender,
      bio: this.editForm.bio,
      address: this.editForm.address,
      city: this.editForm.city,
      state: this.editForm.state,
      zipCode: this.editForm.zipCode,
      country: this.editForm.country,
      service: this.editForm.service,
      userType: this.editForm.userType,
      isStaff: this.editForm.isStaff,
      isActive: this.editForm.isActive,
      isVerified: this.editForm.isVerified,
      isIdentityVerified: this.editForm.isIdentityVerified,
      subscriptionTier: this.editForm.subscriptionTier,
      subscriptionInterval: this.editForm.subscriptionInterval,
      preferredPaymentMode: this.editForm.preferredPaymentMode,
      roleId: this.editForm.roleId
    };
    if (this.editForm.password) {
      payload.password = this.editForm.password;
    }

    this.userUC.updateUser(u.id, payload).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.showEditModal.set(false);
        this.successMessage.set('User account details updated successfully.');
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || err?.error?.error || 'Failed to update user details.');
      }
    });
  }

  toggleEmailVerified(): void {
    const u = this.user();
    if (!u) return;
    const target = !u.isVerified;
    this.isActioning.set(true);
    this.userUC.updateUser(u.id, { isVerified: target }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Email verification status updated to ${target ? 'VERIFIED' : 'UNVERIFIED'}.`);
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update email verification.');
      }
    });
  }

  toggleIdentityVerified(): void {
    const u = this.user();
    if (!u) return;
    const target = !u.isIdentityVerified;
    this.isActioning.set(true);
    this.userUC.updateUser(u.id, { isIdentityVerified: target }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Identity verification status updated to ${target ? 'VERIFIED ID' : 'UNVERIFIED'}.`);
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update identity verification.');
      }
    });
  }

  toggleStaff(): void {
    const u = this.user();
    if (!u) return;
    const target = !u.isStaff;
    this.isActioning.set(true);
    this.userUC.updateUser(u.id, { isStaff: target }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Staff access updated to ${target ? 'STAFF OPERATOR' : 'STANDARD MEMBER'}.`);
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update staff access.');
      }
    });
  }

  changeTier(tier: string): void {
    const u = this.user();
    if (!u) return;
    this.isActioning.set(true);
    this.userUC.updateUser(u.id, { subscriptionTier: tier }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Subscription tier updated to ${tier}.`);
        this.loadFullUserDetails(u.id);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update tier.');
      }
    });
  }

  // ─── DIRECT MESSAGING ───────────────────────────────────────────────────

  async sendDirectMessage(): Promise<void> {
    const u = this.user();
    if (!u) return;

    const message = await this.dialog.prompt({
      title: `Send Direct Message to ${u.firstName || u.email}`,
      message: 'This message will be dispatched directly to the user\'s real-time notification feed and mobile push inbox.',
      placeholder: 'Type your message to the user here...',
      confirmText: 'Dispatch Direct Message',
      cancelText: 'Cancel'
    });

    if (!message || !message.trim()) return;

    this.isActioning.set(true);
    this.userUC.sendDirectMessage(u.id, 'Notice from Administration', message.trim()).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Direct message successfully dispatched to ${u.firstName || u.email}.`);
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to dispatch direct message.');
      }
    });
  }

  // ─── RISK & FRAUD EVALUATION ────────────────────────────────────────────

  evaluateRisk(): void {
    const u = this.user();
    if (!u) return;

    this.isActioning.set(true);
    this.userUC.evaluateUserRisk(u.id).subscribe({
      next: (res) => {
        this.isActioning.set(false);
        this.riskEvaluation.set(res);
        this.showRiskModal.set(true);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to evaluate user fraud & risk score.');
      }
    });
  }

  closeRiskModal(): void {
    this.showRiskModal.set(false);
  }
}
