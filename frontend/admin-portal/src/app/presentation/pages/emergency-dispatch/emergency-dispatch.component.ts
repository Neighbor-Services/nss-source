import { Component, OnInit, OnDestroy, signal, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { WebSocketService } from '../../../core/services/websocket.service';
import { AuditService } from '../../../core/services/audit.service';
import { PiiMaskComponent } from '../../../core/components/pii-mask/pii-mask.component';
import { HasPermissionDirective } from '../../../core/directives/has-permission.directive';

export type IncidentStatus = 'BROADCASTING' | 'ACCEPTED' | 'EN_ROUTE' | 'COMPLETED' | 'CANCELLED_FRAUD';

export interface IncidentProviderCandidate {
  id: string;
  name: string;
  phone: string;
  trade: string;
  distanceKm: number;
  etaMins: number;
  rating: number;
  isEnRoute?: boolean;
}

export interface EmergencyIncident {
  id: string;
  dispatchNumber: string;
  tradeConcept: string;
  title: string;
  description: string;
  seekerName: string;
  seekerPhone: string;
  address: string;
  latitude: number;
  longitude: number;
  broadcastRadiusKm: number;
  status: IncidentStatus;
  createdAt: Date;
  elapsedSeconds: number;
  assignedProvider?: IncidentProviderCandidate;
  candidateProviders: IncidentProviderCandidate[];
  isFlaggedFraud?: boolean;
}

@Component({
  selector: 'app-emergency-dispatch',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, PiiMaskComponent, HasPermissionDirective],
  templateUrl: './emergency-dispatch.component.html',
  styleUrl: './emergency-dispatch.component.css'
})
export class EmergencyDispatchComponent implements OnInit, OnDestroy {
  private ws = inject(WebSocketService);
  private audit = inject(AuditService);

  private timerInterval: any;

  selectedIncident = signal<EmergencyIncident | null>(null);
  filterStatus = signal<string>('ALL');
  searchQuery = signal<string>('');

  incidents = signal<EmergencyIncident[]>([
    {
      id: 'DISP-9021',
      dispatchNumber: 'SOS-2026-9021',
      tradeConcept: 'Plumbing',
      title: 'Burst Main Pipe Flooding Living Room',
      description: 'Water gushing from ceiling pipe near main breaker panel. Urgent shutoff & repair required.',
      seekerName: 'Marcus Sterling',
      seekerPhone: '+1 555-234-8921',
      address: '742 Evergreen Terrace, Springfield',
      latitude: 40.7128,
      longitude: -74.006,
      broadcastRadiusKm: 5,
      status: 'BROADCASTING',
      createdAt: new Date(Date.now() - 4 * 60 * 1000),
      elapsedSeconds: 240,
      candidateProviders: [
        {
          id: 'PROV-101',
          name: 'David Vance (Master Plumber)',
          phone: '+1 555-888-1290',
          trade: 'Plumbing',
          distanceKm: 1.8,
          etaMins: 6,
          rating: 4.95
        },
        {
          id: 'PROV-102',
          name: 'Rapid Response Rooter LLC',
          phone: '+1 555-777-3401',
          trade: 'Plumbing',
          distanceKm: 3.4,
          etaMins: 11,
          rating: 4.88
        }
      ]
    },
    {
      id: 'DISP-9020',
      dispatchNumber: 'SOS-2026-9020',
      tradeConcept: 'Electrical',
      title: 'Electrical Sparking & Smoke Behind Outlet',
      description: 'Heavy smoke and arcing sound in kitchen wall. Main breaker turned off.',
      seekerName: 'Elena Rostova',
      seekerPhone: '+1 555-612-4490',
      address: '108 West 42nd St, Suite 4B',
      latitude: 40.7580,
      longitude: -73.9855,
      broadcastRadiusKm: 10,
      status: 'EN_ROUTE',
      createdAt: new Date(Date.now() - 14 * 60 * 1000),
      elapsedSeconds: 840,
      assignedProvider: {
        id: 'PROV-205',
        name: 'Apex Electric Pros (Jordan K.)',
        phone: '+1 555-432-8811',
        trade: 'Electrical',
        distanceKm: 0.9,
        etaMins: 3,
        rating: 4.98,
        isEnRoute: true
      },
      candidateProviders: []
    },
    {
      id: 'DISP-9019',
      dispatchNumber: 'SOS-2026-9019',
      tradeConcept: 'Locksmith',
      title: 'Elderly Resident Locked Outside in Cold',
      description: 'Keys locked inside with stove running. Immediate unlock required.',
      seekerName: 'Grace Hopper',
      seekerPhone: '+1 555-901-7722',
      address: '350 5th Ave, Floor 12',
      latitude: 40.7484,
      longitude: -73.9857,
      broadcastRadiusKm: 5,
      status: 'ACCEPTED',
      createdAt: new Date(Date.now() - 8 * 60 * 1000),
      elapsedSeconds: 480,
      assignedProvider: {
        id: 'PROV-310',
        name: 'SafeLock Solutions (Ken M.)',
        phone: '+1 555-321-9988',
        trade: 'Locksmith',
        distanceKm: 2.1,
        etaMins: 7,
        rating: 4.91
      },
      candidateProviders: []
    }
  ]);

  filteredIncidents = computed(() => {
    const status = this.filterStatus();
    const query = this.searchQuery().toLowerCase().trim();

    return this.incidents().filter((inc) => {
      const matchStatus = status === 'ALL' || inc.status === status;
      const matchQuery =
        !query ||
        inc.dispatchNumber.toLowerCase().includes(query) ||
        inc.title.toLowerCase().includes(query) ||
        inc.seekerName.toLowerCase().includes(query) ||
        inc.tradeConcept.toLowerCase().includes(query);
      return matchStatus && matchQuery;
    });
  });

  broadcastingCount = computed(() => this.incidents().filter(i => i.status === 'BROADCASTING').length);
  enRouteCount = computed(() => this.incidents().filter(i => i.status === 'EN_ROUTE' || i.status === 'ACCEPTED').length);

  ngOnInit() {
    if (this.incidents().length > 0) {
      this.selectedIncident.set(this.incidents()[0]);
    }

    // Tick elapsed times every second
    this.timerInterval = setInterval(() => {
      this.incidents.update(list =>
        list.map(i => {
          if (i.status === 'COMPLETED' || i.status === 'CANCELLED_FRAUD') return i;
          return { ...i, elapsedSeconds: i.elapsedSeconds + 1 };
        })
      );
    }, 1000);
  }

  ngOnDestroy() {
    if (this.timerInterval) clearInterval(this.timerInterval);
  }

  selectIncident(inc: EmergencyIncident) {
    this.selectedIncident.set(inc);
  }

  formatElapsed(secs: number): string {
    const mins = Math.floor(secs / 60);
    const s = secs % 60;
    return `${mins}m ${s < 10 ? '0' : ''}${s}s`;
  }

  // Admin Override Actions
  extendRadius(inc: EmergencyIncident) {
    const newRadius = inc.broadcastRadiusKm + 5;
    this.incidents.update(list =>
      list.map(i => (i.id === inc.id ? { ...i, broadcastRadiusKm: newRadius } : i))
    );
    if (this.selectedIncident()?.id === inc.id) {
      this.selectedIncident.update(i => i ? { ...i, broadcastRadiusKm: newRadius } : null);
    }
    this.audit.logAction({
      actionType: 'SOS_EXTEND_RADIUS',
      targetEntity: 'EMERGENCY_DISPATCH',
      targetId: inc.dispatchNumber,
      details: `Admin expanded broadcast radius from ${inc.broadcastRadiusKm}km to ${newRadius}km`
    });
  }

  reassignProvider(inc: EmergencyIncident, prov: IncidentProviderCandidate) {
    this.incidents.update(list =>
      list.map(i =>
        i.id === inc.id
          ? {
              ...i,
              status: 'EN_ROUTE',
              assignedProvider: { ...prov, isEnRoute: true }
            }
          : i
      )
    );
    if (this.selectedIncident()?.id === inc.id) {
      this.selectedIncident.update(i =>
        i ? { ...i, status: 'EN_ROUTE', assignedProvider: { ...prov, isEnRoute: true } } : null
      );
    }
    this.audit.logAction({
      actionType: 'SOS_MANUAL_REASSIGN',
      targetEntity: 'EMERGENCY_DISPATCH',
      targetId: inc.dispatchNumber,
      details: `Admin manually assigned provider ${prov.name} (${prov.id})`
    });
  }

  cancelFraud(inc: EmergencyIncident) {
    if (!confirm(`Are you sure you want to cancel and flag ${inc.dispatchNumber} as fraudulent?`)) {
      return;
    }
    this.incidents.update(list =>
      list.map(i => (i.id === inc.id ? { ...i, status: 'CANCELLED_FRAUD', isFlaggedFraud: true } : i))
    );
    if (this.selectedIncident()?.id === inc.id) {
      this.selectedIncident.update(i => i ? { ...i, status: 'CANCELLED_FRAUD', isFlaggedFraud: true } : null);
    }
    this.audit.logAction({
      actionType: 'SOS_CANCEL_FRAUD',
      targetEntity: 'EMERGENCY_DISPATCH',
      targetId: inc.dispatchNumber,
      details: `Admin cancelled emergency dispatch and flagged user as suspicious`
    });
  }
}
