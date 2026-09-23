import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../../data/datasources/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css'
})
export class LoginComponent {
  email = '';
  password = '';
  totpCode = '';
  is2FAStep = signal(false);
  showPassword = signal(false);
  isLoading = signal(false);
  errorMessage = signal<string | null>(null);

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  togglePasswordVisibility() {
    this.showPassword.update(v => !v);
  }

  onLogin(e: Event) {
    e.preventDefault();
    if (!this.email || !this.password) {
      this.errorMessage.set('Please enter both email and password.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.authService.login(this.email, this.password, this.totpCode).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        if (res.requires2FA) {
          this.is2FAStep.set(true);
          this.errorMessage.set(null);
        } else {
          this.router.navigate(['/admin/dashboard']);
        }
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.error || err?.error?.message || 'Invalid administrator credentials or unauthorized access.');
      }
    });
  }

  onVerify2FA(e: Event) {
    e.preventDefault();
    if (!this.totpCode || this.totpCode.trim().length < 6) {
      this.errorMessage.set('Please enter the 6-digit verification code from your authenticator app.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.authService.login(this.email, this.password, this.totpCode.trim()).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        if (res.token) {
          this.router.navigate(['/admin/dashboard']);
        } else {
          this.errorMessage.set('Two-factor authentication failed. Please try again.');
        }
      },
      error: (err) => {
        this.isLoading.set(false);
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
