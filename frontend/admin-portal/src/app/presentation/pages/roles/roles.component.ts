import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RoleUseCase } from '../../../core/usecases/role.usecase';
import { AdminRole, CreateRoleDTO } from '../../../core/domain/entities/role.model';

export type RolesSectionTab = 'ROLES_GRID' | 'PERMISSION_SCOPES' | 'SESSION_SECURITY';

export interface PermissionScope {
  key: string;
  label: string;
  desc: string;
  category: string;
}

export const COMPREHENSIVE_PERMISSIONS: PermissionScope[] = [
  // 1. User Management & Impersonation
  { key: 'users:read', label: 'View Users & Profiles', desc: 'Browse seekers and providers directory, profile data, and history', category: 'User Management' },
  { key: 'users:write', label: 'Edit & Update Users', desc: 'Modify account details, update verification flags, and change status', category: 'User Management' },
  { key: 'users:delete', label: 'Delete User Records', desc: 'Permanently remove customer or provider accounts from the platform', category: 'User Management' },
  { key: 'users:impersonate', label: 'Impersonate Staff / Users', desc: 'Generate one-time bypass access tokens to view user sessions', category: 'User Management' },

  // 2. Bookings & Appointments
  { key: 'appointments:read', label: 'View Bookings & Orders', desc: 'Access service request orders, appointments, and fulfillment history', category: 'Bookings & Orders' },
  { key: 'appointments:manage', label: 'Manage & Force Cancel Bookings', desc: 'Override appointment status, cancel orders, and reassign jobs', category: 'Bookings & Orders' },

  // 3. Provider Identity & Background Checks
  { key: 'verifications:read', label: 'View ID Verification Submissions', desc: 'Inspect driver licenses, passports, and utility bills', category: 'Identity & Safety' },
  { key: 'verifications:manage', label: 'Approve & Reject ID Credentials', desc: 'Review, approve, or reject provider identity documents', category: 'Identity & Safety' },
  { key: 'background_checks:read', label: 'View Checkr Screenings', desc: 'Access candidate criminal history and DMV driving records', category: 'Identity & Safety' },
  { key: 'background_checks:manage', label: 'Manage Background Checks', desc: 'Order candidate screenings, invite providers, and override reports', category: 'Identity & Safety' },

  // 4. Marketplace Catalog & Categories
  { key: 'categories:manage', label: 'Manage Service Taxonomies', desc: 'Create, organize, and delete service category taxonomy groups', category: 'Catalog & Services' },
  { key: 'catalog_services:manage', label: 'Standardized Catalog Services', desc: 'Publish standardized service templates and price corridor baselines', category: 'Catalog & Services' },

  // 5. Moderation, Reviews & Disputes
  { key: 'reviews:manage', label: 'Moderate Customer Reviews', desc: 'Audit ratings, hide defamatory reviews, and manage feedback', category: 'Trust & Moderation' },
  { key: 'disputes:manage', label: 'Dispute Center & Refunds', desc: 'Triage customer disputes, mediate evidence, and issue refunds', category: 'Trust & Moderation' },
  { key: 'reports:manage', label: 'Resolve Abuse & Incident Reports', desc: 'Triage user reports against abuse, harassment, or safety violations', category: 'Trust & Moderation' },

  // 6. Financials, Wallets & Billing
  { key: 'payouts:manage', label: 'Approve Payout Disbursements', desc: 'Authorize Stripe Connect provider payouts and manage transfer holds', category: 'Financials & Payouts' },
  { key: 'wallets:manage', label: 'Manage Escrow & Wallet Balances', desc: 'Debit, credit, and adjust user wallet balances and ledger transactions', category: 'Financials & Payouts' },
  { key: 'subscriptions:manage', label: 'Manage Provider Subscriptions', desc: 'Configure tier plans, subscription perks, and billing cycles', category: 'Financials & Payouts' },
  { key: 'reports:financial', label: 'Financial Reports & Tax Exports', desc: 'Access platform revenue analytics, escrow ledger, and 1099-K tax exports', category: 'Financials & Payouts' },
  { key: 'promos:manage', label: 'Manage Promo Codes & Coupons', desc: 'Create discount codes, specify usage limits, and expire vouchers', category: 'Financials & Payouts' },

  // 7. Security, Fraud & Audit
  { key: 'fraud:read', label: 'Inspect Risk Alerts & Scores', desc: 'Monitor heuristic fraud triggers, card velocity, and anomaly flags', category: 'Security & Governance' },
  { key: 'fraud:manage', label: 'Enforce Risk Rules & Lockouts', desc: 'Quarantine high-risk accounts and configure Stripe Radar parameters', category: 'Security & Governance' },
  { key: 'roles:manage', label: 'Manage RBAC Security Roles', desc: 'Create, update, and assign roles and granular permission scopes', category: 'Security & Governance' },
  { key: 'audit_logs:read', label: 'View Immutable Audit Trails', desc: 'Inspect administrative audit trails and operator activity records', category: 'Security & Governance' },

  // 8. Infrastructure & Operations
  { key: 'feature_flags:manage', label: 'Manage Dynamic Feature Flags', desc: 'Toggle platform operational killswitches and canary features', category: 'System & Infrastructure' },
  { key: 'system:health', label: 'System Health & Metrics', desc: 'Inspect database latency, WebSocket hub status, and memory load', category: 'System & Infrastructure' },
  { key: 'system:backup', label: 'System Snapshots & Backups', desc: 'Trigger manual database backups and restore point snapshots', category: 'System & Infrastructure' },

  // 9. Content & Communications
  { key: 'cms:manage', label: 'Manage Public Website CMS', desc: 'Update hero banner copy, testimonials, FAQs, and stats counters', category: 'Content & Support' },
  { key: 'legal:manage', label: 'Legal & Compliance Documents', desc: 'Publish Terms of Service, Privacy Policies, and liability waivers', category: 'Content & Support' },
  { key: 'templates:manage', label: 'Notification & Email Templates', desc: 'Customize transactional emails, SMS alerts, and push notifications', category: 'Content & Support' },
  { key: 'support:manage', label: 'Support Inquiries & Contact Desk', desc: 'Review and reply to contact messages and public inquiries', category: 'Content & Support' }
];

@Component({
  selector: 'app-roles',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './roles.component.html',
  styleUrl: './roles.component.css'
})
export class RolesComponent implements OnInit {
  roles = signal<AdminRole[]>([]);
  loading = signal<boolean>(false);
  savingRole = signal<boolean>(false);
  deletingRoleId = signal<string | null>(null);
  
  showRoleModal = signal<boolean>(false);
  isEditMode = signal<boolean>(false);
  editingRoleId = signal<string | null>(null);
  
  showDeleteModal = signal<boolean>(false);
  roleToDelete = signal<AdminRole | null>(null);

  mainSection = signal<RolesSectionTab>('ROLES_GRID');
  permissionSearchQuery = signal<string>('');
  selectedCategoryFilter = signal<string>('ALL');

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  availablePermissions = COMPREHENSIVE_PERMISSIONS;

  // Computed KPIs
  totalRoles = computed(() => this.roles().length);
  superAdminRoles = computed(() => this.roles().filter(r => r.slug === 'super-admin' || r.permissions?.length >= 20).length);
  totalPermissionsDefined = computed(() => this.availablePermissions.length);
  totalCategories = computed(() => this.permissionCategories.length);

  // Grouped permissions for picker
  permissionCategories = [
    'User Management',
    'Bookings & Orders',
    'Identity & Safety',
    'Catalog & Services',
    'Trust & Moderation',
    'Financials & Payouts',
    'Security & Governance',
    'System & Infrastructure',
    'Content & Support'
  ];

  filteredPermissions = computed(() => {
    const q = this.permissionSearchQuery().toLowerCase().trim();
    const cat = this.selectedCategoryFilter();

    return this.availablePermissions.filter(p => {
      const matchCat = cat === 'ALL' || p.category === cat;
      const matchQ = !q || p.key.toLowerCase().includes(q) || p.label.toLowerCase().includes(q) || p.desc.toLowerCase().includes(q) || p.category.toLowerCase().includes(q);
      return matchCat && matchQ;
    });
  });

  currentRoleForm: CreateRoleDTO = {
    name: '',
    slug: '',
    description: '',
    permissions: []
  };

  constructor(private roleUseCase: RoleUseCase) {}

  ngOnInit(): void {
    this.fetchRoles();
  }

  fetchRoles(): void {
    this.loading.set(true);
    this.errorMessage.set(null);
    this.roleUseCase.listRoles().subscribe({
      next: (data) => {
        this.roles.set(data);
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load RBAC roles');
        this.loading.set(false);
      }
    });
  }

  openCreateModal(): void {
    this.isEditMode.set(false);
    this.editingRoleId.set(null);
    this.currentRoleForm = {
      name: '',
      slug: '',
      description: '',
      permissions: ['users:read']
    };
    this.showRoleModal.set(true);
  }

  openEditModal(role: AdminRole): void {
    this.isEditMode.set(true);
    this.editingRoleId.set(role.id);
    this.currentRoleForm = {
      name: role.name,
      slug: role.slug,
      description: role.description || '',
      permissions: [...role.permissions]
    };
    this.showRoleModal.set(true);
  }

  closeRoleModal(): void {
    this.showRoleModal.set(false);
    this.editingRoleId.set(null);
  }

  confirmDeleteRole(role: AdminRole): void {
    this.roleToDelete.set(role);
    this.showDeleteModal.set(true);
  }

  closeDeleteModal(): void {
    this.showDeleteModal.set(false);
    this.roleToDelete.set(null);
  }

  onNameChange(): void {
    if (!this.isEditMode()) {
      this.currentRoleForm.slug = this.currentRoleForm.name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
    }
  }

  togglePermission(key: string): void {
    const idx = this.currentRoleForm.permissions.indexOf(key);
    if (idx > -1) {
      this.currentRoleForm.permissions.splice(idx, 1);
    } else {
      this.currentRoleForm.permissions.push(key);
    }
  }

  isPermissionSelected(key: string): boolean {
    return this.currentRoleForm.permissions.includes(key);
  }

  toggleCategory(category: string, selectAll: boolean): void {
    const categoryPermKeys = this.availablePermissions
      .filter(p => p.category === category)
      .map(p => p.key);

    if (selectAll) {
      categoryPermKeys.forEach(key => {
        if (!this.currentRoleForm.permissions.includes(key)) {
          this.currentRoleForm.permissions.push(key);
        }
      });
    } else {
      this.currentRoleForm.permissions = this.currentRoleForm.permissions.filter(
        key => !categoryPermKeys.includes(key)
      );
    }
  }

  isCategoryAllSelected(category: string): boolean {
    const categoryPermKeys = this.availablePermissions
      .filter(p => p.category === category)
      .map(p => p.key);
    return categoryPermKeys.every(k => this.currentRoleForm.permissions.includes(k));
  }

  submitRoleForm(): void {
    if (!this.currentRoleForm.name.trim() || !this.currentRoleForm.slug.trim()) {
      this.errorMessage.set('Role name and slug are required');
      return;
    }

    if (this.currentRoleForm.permissions.length === 0) {
      this.errorMessage.set('Please select at least one permission scope');
      return;
    }

    this.savingRole.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    if (this.isEditMode() && this.editingRoleId()) {
      this.roleUseCase.updateRole(this.editingRoleId()!, this.currentRoleForm).subscribe({
        next: () => {
          this.savingRole.set(false);
          this.showRoleModal.set(false);
          this.successMessage.set(`Role "${this.currentRoleForm.name}" updated successfully!`);
          this.fetchRoles();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.errorMessage.set(err.error?.message || 'Failed to update role');
          this.savingRole.set(false);
        }
      });
    } else {
      this.roleUseCase.createRole(this.currentRoleForm).subscribe({
        next: () => {
          this.savingRole.set(false);
          this.showRoleModal.set(false);
          this.successMessage.set(`Role "${this.currentRoleForm.name}" created successfully!`);
          this.fetchRoles();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.errorMessage.set(err.error?.message || 'Failed to create role');
          this.savingRole.set(false);
        }
      });
    }
  }

  executeDeleteRole(): void {
    const role = this.roleToDelete();
    if (!role) return;

    this.deletingRoleId.set(role.id);
    this.roleUseCase.deleteRole(role.id).subscribe({
      next: () => {
        this.deletingRoleId.set(null);
        this.showDeleteModal.set(false);
        this.roleToDelete.set(null);
        this.successMessage.set(`Role "${role.name}" deleted successfully.`);
        this.fetchRoles();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to delete role');
        this.deletingRoleId.set(null);
      }
    });
  }

  getPermissionsForCategory(category: string): PermissionScope[] {
    return this.availablePermissions.filter(p => p.category === category);
  }
}
