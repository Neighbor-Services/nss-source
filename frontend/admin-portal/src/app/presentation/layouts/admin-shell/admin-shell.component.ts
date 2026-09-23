import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { TopbarComponent } from '../topbar/topbar.component';
import { LayoutService } from '../../../core/services/layout.service';
import { IdleService } from '../../../core/services/idle.service';
import { WebSocketService } from '../../../core/services/websocket.service';

@Component({
  selector: 'app-admin-shell',
  standalone: true,
  imports: [CommonModule, RouterOutlet, SidebarComponent, TopbarComponent],
  templateUrl: './admin-shell.component.html',
  styleUrl: './admin-shell.component.css'
})
export class AdminShellComponent implements OnInit, OnDestroy {
  constructor(
    public layoutService: LayoutService,
    public idleService: IdleService,
    private ws: WebSocketService
  ) {}

  ngOnInit() {
    this.idleService.init();
    this.ws.connect();
    this.ws.getEvents().subscribe((evt) => {
      if (evt?.type === 'CRITICAL_FRAUD' || evt?.type === 'URGENT_DISPUTE') {
        this.ws.playIncidentChime();
      }
    });
  }

  ngOnDestroy() {
    this.ws.disconnect();
  }
}
