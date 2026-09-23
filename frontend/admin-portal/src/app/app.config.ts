import { ApplicationConfig, provideZonelessChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';
import { routes } from './app.routes';
import { authInterceptor } from './data/interceptors/auth.interceptor';

// Feature Repository Contracts
import { DashboardRepository } from './core/repositories/dashboard.repository';
import { UserRepository } from './core/repositories/user.repository';
import { VerificationRepository } from './core/repositories/verification.repository';
import { BackgroundCheckRepository } from './core/repositories/background-check.repository';
import { DisputeRepository } from './core/repositories/dispute.repository';
import { PayoutRepository } from './core/repositories/payout.repository';
import { SubscriptionRepository } from './core/repositories/subscription.repository';
import { CatalogRepository } from './core/repositories/catalog.repository';
import { ReportRepository } from './core/repositories/report.repository';
import { FraudRepository } from './core/repositories/fraud.repository';
import { TemplateRepository } from './core/repositories/template.repository';
import { RoleRepository } from './core/repositories/role.repository';
import { SettingsRepository } from './core/repositories/settings.repository';
import { FeatureFlagRepository } from './core/repositories/feature-flag.repository';
import { SystemRepository } from './core/repositories/system.repository';
import { AppointmentRepository } from './core/repositories/appointment.repository';
import { ReviewRepository } from './core/repositories/review.repository';
import { PromoRepository } from './core/repositories/promo.repository';
import { LegalRepository } from './core/repositories/legal.repository';
import { SupportRepository } from './core/repositories/support.repository';
import { CmsRepository } from './core/repositories/cms.repository';

// Feature Repository Implementations
import { DashboardRepositoryImpl } from './data/repositories/dashboard.repository.impl';
import { UserRepositoryImpl } from './data/repositories/user.repository.impl';
import { VerificationRepositoryImpl } from './data/repositories/verification.repository.impl';
import { BackgroundCheckRepositoryImpl } from './data/repositories/background-check.repository.impl';
import { DisputeRepositoryImpl } from './data/repositories/dispute.repository.impl';
import { PayoutRepositoryImpl } from './data/repositories/payout.repository.impl';
import { SubscriptionRepositoryImpl } from './data/repositories/subscription.repository.impl';
import { CatalogRepositoryImpl } from './data/repositories/catalog.repository.impl';
import { ReportRepositoryImpl } from './data/repositories/report.repository.impl';
import { FraudRepositoryImpl } from './data/repositories/fraud.repository.impl';
import { TemplateRepositoryImpl } from './data/repositories/template.repository.impl';
import { RoleRepositoryImpl } from './data/repositories/role.repository.impl';
import { SettingsRepositoryImpl } from './data/repositories/settings.repository.impl';
import { FeatureFlagRepositoryImpl } from './data/repositories/feature-flag.repository.impl';
import { SystemRepositoryImpl } from './data/repositories/system.repository.impl';
import { AppointmentRepositoryImpl } from './data/repositories/appointment.repository.impl';
import { ReviewRepositoryImpl } from './data/repositories/review.repository.impl';
import { PromoRepositoryImpl } from './data/repositories/promo.repository.impl';
import { LegalRepositoryImpl } from './data/repositories/legal.repository.impl';
import { SupportRepositoryImpl } from './data/repositories/support.repository.impl';
import { CmsRepositoryImpl } from './data/repositories/cms.repository.impl';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZonelessChangeDetection(),
    provideRouter(
      routes,
      withComponentInputBinding(),
      withInMemoryScrolling({ scrollPositionRestoration: 'top' })
    ),
    provideHttpClient(
      withFetch(),
      withInterceptors([authInterceptor])
    ),
    { provide: DashboardRepository, useClass: DashboardRepositoryImpl },
    { provide: UserRepository, useClass: UserRepositoryImpl },
    { provide: VerificationRepository, useClass: VerificationRepositoryImpl },
    { provide: BackgroundCheckRepository, useClass: BackgroundCheckRepositoryImpl },
    { provide: DisputeRepository, useClass: DisputeRepositoryImpl },
    { provide: PayoutRepository, useClass: PayoutRepositoryImpl },
    { provide: SubscriptionRepository, useClass: SubscriptionRepositoryImpl },
    { provide: CatalogRepository, useClass: CatalogRepositoryImpl },
    { provide: ReportRepository, useClass: ReportRepositoryImpl },
    { provide: FraudRepository, useClass: FraudRepositoryImpl },
    { provide: TemplateRepository, useClass: TemplateRepositoryImpl },
    { provide: RoleRepository, useClass: RoleRepositoryImpl },
    { provide: SettingsRepository, useClass: SettingsRepositoryImpl },
    { provide: FeatureFlagRepository, useClass: FeatureFlagRepositoryImpl },
    { provide: SystemRepository, useClass: SystemRepositoryImpl },
    { provide: AppointmentRepository, useClass: AppointmentRepositoryImpl },
    { provide: ReviewRepository, useClass: ReviewRepositoryImpl },
    { provide: PromoRepository, useClass: PromoRepositoryImpl },
    { provide: LegalRepository, useClass: LegalRepositoryImpl },
    { provide: SupportRepository, useClass: SupportRepositoryImpl },
    { provide: CmsRepository, useClass: CmsRepositoryImpl }
  ]
};
