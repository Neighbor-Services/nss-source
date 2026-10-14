import { Component, OnInit, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../../data/datasources/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css'
})
export class LoginComponent implements OnInit, OnDestroy {
  email = '';
  password = '';
  totpCode = '';
  is2FAStep = signal(false);
  showPassword = signal(false);
  isLoading = signal(false);
  errorMessage = signal<string | null>(null);

  // Brute-force protection states
  isLockedOut = signal(false);
  lockoutRemainingSeconds = signal(0);
  failedAttempts = signal(0);
  private lockoutTimer: any = null;

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit() {
    // Check if redirected from authGuard due to unauthorized access
    const errorParam = this.route.snapshot.queryParams['error'];
    if (errorParam === 'unauthorized' || errorParam === 'forbidden') {
      this.errorMessage.set('Access denied: You do not have permission to access the administrator portal.');
    }
  }

  ngOnDestroy() {
    if (this.lockoutTimer) {
      clearInterval(this.lockoutTimer);
      this.lockoutTimer = null;
    }
  }

  togglePasswordVisibility() {
    this.showPassword.update(v => !v);
  }

  startLockout(seconds: number, message?: string) {
    if (this.lockoutTimer) {
      clearInterval(this.lockoutTimer);
      this.lockoutTimer = null;
    }
    this.isLockedOut.set(true);
    this.lockoutRemainingSeconds.set(seconds);
    this.errorMessage.set(message || 'Too many failed login attempts. To protect system security, login is temporarily locked.');

    this.lockoutTimer = setInterval(() => {
      const current = this.lockoutRemainingSeconds();
      if (current <= 1) {
        clearInterval(this.lockoutTimer);
        this.lockoutTimer = null;
        this.isLockedOut.set(false);
        this.lockoutRemainingSeconds.set(0);
        this.errorMessage.set(null);
      } else {
        this.lockoutRemainingSeconds.set(current - 1);
      }
    }, 1000);
  }

  onLogin(e: Event) {
    e.preventDefault();
    if (this.isLockedOut()) {
      return;
    }

    if (!this.email || !this.password) {
      this.errorMessage.set('Please enter both email and password.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.authService.login(this.email, this.password, this.totpCode).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        this.failedAttempts.set(0);

        if (res.requires2FA) {
          this.is2FAStep.set(true);
          this.errorMessage.set(null);
          return;
        }

        // Verify that the user has admin rights (Staff or Superuser)
        const isStaff = res.user?.isStaff === true || (res.user as any)?.is_staff === true;
        const isSuperuser = res.user?.isSuperuser === true || (res.user as any)?.is_superuser === true;

        if (!isStaff && !isSuperuser) {
          this.authService.logout();
          this.errorMessage.set('Access denied: You do not have permission to access the administrator portal.');
          return;
        }

        this.router.navigate(['/admin/dashboard']);
      },
      error: (err) => {
        this.isLoading.set(false);
        const attempts = this.failedAttempts() + 1;
        this.failedAttempts.set(attempts);

        // Server brute-force lockout (429 Too Many Requests)
        if (err?.status === 429 || err?.error?.code === 'TOO_MANY_FAILED_ATTEMPTS') {
          const retrySec = err?.error?.retry_after_seconds || 60;
          const msg = err?.error?.message || 'Too many failed login attempts. To protect system security, login is temporarily locked.';
          this.startLockout(retrySec, msg);
          return;
        }

        // Client-side brute force defense after 5 consecutive failures
        if (attempts >= 5) {
          this.startLockout(60, 'Too many consecutive failed attempts. Security cooldown active. Please wait 60 seconds.');
          return;
        }

        // 403 Forbidden (Non-admin user trying to log in)
        if (err?.status === 403 || err?.error?.code === 'ADMIN_ACCESS_REQUIRED' || err?.message?.includes('Access denied')) {
          this.errorMessage.set(err?.error?.message || 'Access denied: You do not have permission to access the administrator portal.');
          return;
        }

        this.errorMessage.set(err?.error?.message || err?.error?.error || err?.message || 'Invalid administrator credentials or unauthorized access.');
      }
    });
  }

  onVerify2FA(e: Event) {
    e.preventDefault();
    if (this.isLockedOut()) {
      return;
    }

    if (!this.totpCode || this.totpCode.trim().length < 6) {
      this.errorMessage.set('Please enter the 6-digit verification code from your authenticator app.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.authService.login(this.email, this.password, this.totpCode.trim()).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        this.failedAttempts.set(0);

        if (res.token) {
          const isStaff = res.user?.isStaff === true || (res.user as any)?.is_staff === true;
          const isSuperuser = res.user?.isSuperuser === true || (res.user as any)?.is_superuser === true;

          if (!isStaff && !isSuperuser) {
            this.authService.logout();
            this.errorMessage.set('Access denied: You do not have permission to access the administrator portal.');
            return;
          }

          this.router.navigate(['/admin/dashboard']);
        } else {
          this.errorMessage.set('Two-factor authentication failed. Please try again.');
        }
      },
      error: (err) => {
        this.isLoading.set(false);
        if (err?.status === 429) {
          const retrySec = err?.error?.retry_after_seconds || 60;
          this.startLockout(retrySec, 'Too many failed verification attempts. Please wait before retrying.');
          return;
        }
        this.errorMessage.set(err?.error?.error || err?.error?.message || 'Invalid 6-digit TOTP code.');
      }
    });
  }

  cancel2FA() {
    this.is2FAStep.set(false);
    this.totpCode = '';
    this.errorMessage.set(null);
  }
}
