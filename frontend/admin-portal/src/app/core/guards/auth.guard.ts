import { inject } from '@angular/core';
import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { AuthService } from '../../data/datasources/auth.service';

export const authGuard: CanActivateFn = (): boolean | UrlTree => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (authService.isAuthenticated() && authService.isAdminUser()) {
    return true;
  }

  // If authenticated as a standard user without admin rights, purge session
  if (authService.isAuthenticated()) {
    authService.logout();
    return router.createUrlTree(['/login'], {
      queryParams: { error: 'unauthorized' }
    });
  }

  return router.createUrlTree(['/login']);
};
