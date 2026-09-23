import { ApplicationConfig, provideZonelessChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { routes } from './app.routes';

// Feature Repository Contracts
import { CmsRepository } from './core/repositories/cms.repository';
import { CatalogRepository } from './core/repositories/catalog.repository';
import { ContactRepository } from './core/repositories/contact.repository';
import { ResolutionRepository } from './core/repositories/resolution.repository';
import { LegalRepository } from './core/repositories/legal.repository';

// Feature Repository Implementations
import { CmsRepositoryImpl } from './data/repositories/cms.repository.impl';
import { CatalogRepositoryImpl } from './data/repositories/catalog.repository.impl';
import { ContactRepositoryImpl } from './data/repositories/contact.repository.impl';
import { ResolutionRepositoryImpl } from './data/repositories/resolution.repository.impl';
import { LegalRepositoryImpl } from './data/repositories/legal.repository.impl';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZonelessChangeDetection(),
    provideRouter(
      routes,
      withComponentInputBinding(),
      withInMemoryScrolling({ scrollPositionRestoration: 'top', anchorScrolling: 'enabled' })
    ),
    provideHttpClient(withFetch()),
    { provide: CmsRepository, useClass: CmsRepositoryImpl },
    { provide: CatalogRepository, useClass: CatalogRepositoryImpl },
    { provide: ContactRepository, useClass: ContactRepositoryImpl },
    { provide: ResolutionRepository, useClass: ResolutionRepositoryImpl },
    { provide: LegalRepository, useClass: LegalRepositoryImpl }
  ]
};
