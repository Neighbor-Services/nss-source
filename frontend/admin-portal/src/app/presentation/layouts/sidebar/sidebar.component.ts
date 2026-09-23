import { Component, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../../data/datasources/auth.service';
import { LayoutService } from '../../../core/services/layout.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.css'
})
export class SidebarComponent {
  private authService = inject(AuthService);
  public layoutService = inject(LayoutService);
  currentUser = this.authService.currentUser;

  displayName = computed(() => {
    const u = this.currentUser();
    if (u?.firstName || u?.lastName) {
      return `${u.firstName || ''} ${u.lastName || ''}`.trim();
    }
    return u?.email?.split('@')[0] || 'Administrator';
  });

  userEmail = computed(() => {
    return this.currentUser()?.email || 'admin@neighborservice.com';
  });

  userInitial = computed(() => {
    const name = this.displayName();
    return name.charAt(0).toUpperCase() || 'A';
  });

  logout() {
    this.authService.logout();
  }
}
