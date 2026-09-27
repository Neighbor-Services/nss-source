import { inject } from '@angular/core';
import { CanActivateFn, Router, ActivatedRouteSnapshot } from '@angular/router';
import { AuthService } from '../../data/datasources/auth.service';
import { AdminPermission, AdminRole } from '../directives/has-permission.directive';

export const permissionGuard = (requiredPermission: AdminPermission): CanActivateFn => {
  return (route: ActivatedRouteSnapshot) => {
    const auth = inject(AuthService);
    const router = inject(Router);

    const user = auth.currentUser();
    if (!user) {
      return router.createUrlTree(['/login']);
    }

    if (user.isSuperuser || user.is_superuser) {
      return true;
    }

    const role: AdminRole = (user.role as AdminRole) || 'SUPER_ADMIN';
    if (role === 'SUPER_ADMIN') {
      return true;
    }

    // Role specific access check
    if (requiredPermission === 'payouts:read' && role === 'KYC_REVIEWER') {
      return router.createUrlTree(['/admin/dashboard']);
    }

    return true;
  };
};
