import { Routes } from '@angular/router';
import { LoginComponent } from './presentation/pages/login/login.component';
import { AdminShellComponent } from './presentation/layouts/admin-shell/admin-shell.component';
import { DashboardComponent } from './presentation/pages/dashboard/dashboard.component';
import { UsersComponent } from './presentation/pages/users/users.component';
import { UserDetailComponent } from './presentation/pages/user-detail/user-detail.component';
import { VerificationsComponent } from './presentation/pages/verifications/verifications.component';
import { BackgroundChecksComponent } from './presentation/pages/background-checks/background-checks.component';
import { DisputesComponent } from './presentation/pages/disputes/disputes.component';
import { PayoutsComponent } from './presentation/pages/payouts/payouts.component';
import { CatalogComponent } from './presentation/pages/catalog/catalog.component';
import { FinancialComponent } from './presentation/pages/financial/financial.component';
import { ReportsComponent } from './presentation/pages/reports/reports.component';
import { FraudComponent } from './presentation/pages/fraud/fraud.component';
import { TemplatesComponent } from './presentation/pages/templates/templates.component';
import { RolesComponent } from './presentation/pages/roles/roles.component';
import { SettingsComponent } from './presentation/pages/settings/settings.component';
import { SubscriptionsComponent } from './presentation/pages/subscriptions/subscriptions.component';
import { FeatureFlagsComponent } from './presentation/pages/feature-flags/feature-flags.component';
import { SystemComponent } from './presentation/pages/system/system.component';
import { AppointmentsComponent } from './presentation/pages/appointments/appointments.component';
import { ReviewsComponent } from './presentation/pages/reviews/reviews.component';
import { PromosComponent } from './presentation/pages/promos/promos.component';
import { LegalComponent } from './presentation/pages/legal/legal.component';
import { SupportComponent } from './presentation/pages/support/support.component';
import { CmsComponent } from './presentation/pages/cms/cms.component';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  { path: 'login', component: LoginComponent },
  {
    path: 'admin',
    component: AdminShellComponent,
    canActivate: [authGuard],
    children: [
      { path: 'dashboard', component: DashboardComponent },
      { path: 'users', component: UsersComponent },
      { path: 'users/:id', component: UserDetailComponent },
      { path: 'appointments', component: AppointmentsComponent },
      { path: 'verifications', component: VerificationsComponent },
      { path: 'background-checks', component: BackgroundChecksComponent },
      { path: 'reviews', component: ReviewsComponent },
      { path: 'disputes', component: DisputesComponent },
      { path: 'reports', component: ReportsComponent },
      { path: 'financial', component: FinancialComponent },
      { path: 'finance', redirectTo: 'financial' },
      { path: 'payouts', component: PayoutsComponent },
      { path: 'promos', component: PromosComponent },
      { path: 'subscriptions', component: SubscriptionsComponent },
      { path: 'catalog', component: CatalogComponent },
      { path: 'legal', component: LegalComponent },
      { path: 'support', component: SupportComponent },
      { path: 'contact-messages', component: SupportComponent },
      { path: 'resolutions', component: SupportComponent },
      { path: 'cms', component: CmsComponent },
      { path: 'fraud', component: FraudComponent },
      { path: 'templates', component: TemplatesComponent },
      { path: 'roles', component: RolesComponent },
      { path: 'settings', component: SettingsComponent },
      { path: 'feature-flags', component: FeatureFlagsComponent },
      { path: 'system', component: SystemComponent },
      { path: '', redirectTo: 'dashboard', pathMatch: 'full' }
    ]
  },
  { path: '', redirectTo: 'admin/dashboard', pathMatch: 'full' },
  { path: '**', redirectTo: 'admin/dashboard' }
];
