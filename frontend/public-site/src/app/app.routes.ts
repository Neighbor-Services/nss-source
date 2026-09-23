import { Routes } from '@angular/router';
import { HomeComponent } from './presentation/pages/home/home.component';
import { ServicesComponent } from './presentation/pages/services/services.component';
import { AboutComponent } from './presentation/pages/about/about.component';
import { ContactComponent } from './presentation/pages/contact/contact.component';
import { ResolutionComponent } from './presentation/pages/resolution/resolution.component';
import { SupportComponent } from './presentation/pages/support/support.component';
import { LegalComponent } from './presentation/pages/legal/legal.component';

export const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'services', component: ServicesComponent },
  { path: 'about', component: AboutComponent },
  { path: 'contact', component: ContactComponent },
  { path: 'resolution', component: ResolutionComponent },
  { path: 'support', component: SupportComponent },
  { path: 'terms', component: LegalComponent, data: { type: 'terms' } },
  { path: 'privacy', component: LegalComponent, data: { type: 'privacy' } },
  { path: '**', redirectTo: '' }
];
