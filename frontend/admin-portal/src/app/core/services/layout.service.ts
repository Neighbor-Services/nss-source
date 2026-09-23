import { Injectable, signal } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { filter } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class LayoutService {
  isMobileSidebarOpen = signal<boolean>(false);
  isDesktopCollapsed = signal<boolean>(false);

  constructor(private router: Router) {
    // Restore collapsed preference from localStorage if on desktop
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('ns_admin_sidebar_collapsed');
      if (saved === 'true') {
        this.isDesktopCollapsed.set(true);
      }
    }

    // Automatically close sidebar drawer on mobile upon route navigation
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe(() => {
      this.closeMobileSidebar();
    });
  }

  toggleSidebar(): void {
    if (typeof window !== 'undefined' && window.innerWidth <= 1024) {
      this.isMobileSidebarOpen.update(v => !v);
    } else {
      this.toggleDesktopCollapse();
    }
  }

  toggleDesktopCollapse(): void {
    this.isDesktopCollapsed.update(v => {
      const next = !v;
      if (typeof window !== 'undefined') {
        localStorage.setItem('ns_admin_sidebar_collapsed', String(next));
      }
      return next;
    });
  }

  toggleMobileSidebar(): void {
    this.isMobileSidebarOpen.update(v => !v);
  }

  openSidebar(): void {
    if (typeof window !== 'undefined' && window.innerWidth <= 1024) {
      this.isMobileSidebarOpen.set(true);
    } else {
      this.isDesktopCollapsed.set(false);
      localStorage.setItem('ns_admin_sidebar_collapsed', 'false');
    }
  }

  closeSidebar(): void {
    this.closeMobileSidebar();
  }

  closeMobileSidebar(): void {
    this.isMobileSidebarOpen.set(false);
  }
}
