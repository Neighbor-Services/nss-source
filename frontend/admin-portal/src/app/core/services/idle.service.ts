import { Injectable, signal, inject, NgZone } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../data/datasources/auth.service';

@Injectable({
  providedIn: 'root'
})
export class IdleService {
  private readonly IDLE_TIMEOUT_MS = 30 * 60 * 1000; // 30 minutes
  private readonly WARNING_TIMEOUT_MS = 28 * 60 * 1000; // 28 minutes warning

  private timerId: any = null;
  private warningTimerId: any = null;

  isLocked = signal(false);
  showWarning = signal(false);
  secondsRemaining = signal(120);

  private auth = inject(AuthService);
  private router = inject(Router);
  private ngZone = inject(NgZone);

  init() {
    this.setupEventListeners();
    this.resetTimer();
  }

  private setupEventListeners() {
    const events = ['mousemove', 'keydown', 'mousedown', 'touchstart', 'scroll'];
    this.ngZone.runOutsideAngular(() => {
      events.forEach(evt => {
        window.addEventListener(evt, () => {
          if (!this.isLocked()) {
            this.ngZone.run(() => this.resetTimer());
          }
        }, { passive: true });
      });
    });
  }

  resetTimer() {
    if (!this.auth.isAuthenticated()) return;

    this.showWarning.set(false);
    this.clearTimers();

    this.warningTimerId = setTimeout(() => {
      this.showWarning.set(true);
      this.startCountdown();
    }, this.WARNING_TIMEOUT_MS);

    this.timerId = setTimeout(() => {
      this.lockSession();
    }, this.IDLE_TIMEOUT_MS);
  }

  private startCountdown() {
    this.secondsRemaining.set(120);
    const interval = setInterval(() => {
      const remaining = this.secondsRemaining() - 1;
      this.secondsRemaining.set(remaining);
      if (remaining <= 0 || !this.showWarning()) {
        clearInterval(interval);
      }
    }, 1000);
  }

  lockSession() {
    this.isLocked.set(true);
    this.showWarning.set(false);
    this.clearTimers();
    this.auth.logout();
  }

  unlockSession() {
    this.isLocked.set(false);
    this.resetTimer();
  }

  private clearTimers() {
    if (this.timerId) clearTimeout(this.timerId);
    if (this.warningTimerId) clearTimeout(this.warningTimerId);
  }
}
