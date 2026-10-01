import { Component, OnInit, OnDestroy, signal, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { WebSocketService } from '../../../core/services/websocket.service';
import { AuditService } from '../../../core/services/audit.service';
import { PiiMaskComponent } from '../../../core/components/pii-mask/pii-mask.component';
import { HasPermissionDirective } from '../../../core/directives/has-permission.directive';
import { DispatchUseCase } from '../../../core/usecases/dispatch.usecase';
import { DispatchIncidentSummary, ProviderFleetTelemetry } from '../../../core/domain/entities/dispatch.model';

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
  private dispatchUC = inject(DispatchUseCase);

  private timerInterval: any;
  private pollInterval: any;

  selectedIncident = signal<EmergencyIncident | null>(null);
  filterStatus = signal<string>('ALL');
  searchQuery = signal<string>('');
  isLoading = signal<boolean>(false);
  isActioning = signal<boolean>(false);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);

  providerFleet = signal<ProviderFleetTelemetry[]>([]);

  incidents = signal<EmergencyIncident[]>([]);

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
    this.loadIncidents();
    this.loadFleet();

    // Poll live incidents every 10s
    this.pollInterval = setInterval(() => {
      this.loadIncidents(false);
    }, 10000);

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
    if (this.pollInterval) clearInterval(this.pollInterval);
  }

  loadFleet() {
    this.dispatchUC.getProviderFleet().subscribe({
      next: (fleet) => {
        if (fleet) this.providerFleet.set(fleet);
      },
      error: () => { }
    });
  }

  loadIncidents(showSpinner = true) {
    if (showSpinner) this.isLoading.set(true);
    this.dispatchUC.listIncidents().subscribe({
      next: (data) => {
        if (showSpinner) this.isLoading.set(false);
        if (data && data.length > 0) {
          const mapped: EmergencyIncident[] = data.map(item => {
            const created = new Date(item.created_at);
            const elapsed = Math.max(0, Math.floor((Date.now() - created.getTime()) / 1000));
            const statusMapped: IncidentStatus =
              item.status === 'DISPATCHED' || item.status === 'BROADCASTING' ? 'BROADCASTING' :
                item.status === 'ACCEPTED' ? 'ACCEPTED' :
                  item.status === 'IN_PROGRESS' ? 'EN_ROUTE' :
                    item.status === 'COMPLETED' ? 'COMPLETED' : 'BROADCASTING';

            let assigned: IncidentProviderCandidate | undefined;
            if (item.accepted_provider_id && item.provider_name) {
              assigned = {
                id: item.accepted_provider_id,
                name: item.provider_name,
                phone: item.provider_phone || '+1 (555) 392-8192',
                trade: 'Master Emergency Tech',
                distanceKm: 2.1,
                etaMins: 6,
                rating: 4.95,
                isEnRoute: item.status === 'IN_PROGRESS'
              };
            }

            // Populate candidate providers
            const candidates: IncidentProviderCandidate[] = this.providerFleet().length > 0
              ? this.providerFleet().slice(0, 3).map((f, idx) => ({
                id: f.user_id || f.id,
                name: f.name || `Provider ${idx + 1}`,
                phone: f.phone || '+1 (555) 481-9011',
                trade: f.service || 'Emergency Specialist',
                distanceKm: parseFloat((1.2 + idx * 1.4).toFixed(1)),
                etaMins: 5 + idx * 4,
                rating: f.rating || (4.85 + idx * 0.05),
                isEnRoute: false
              }))
              : [

              ];

            return {
              id: item.id,
              dispatchNumber: `SOS-${item.id.substring(0, 8).toUpperCase()}`,
              tradeConcept: item.title || 'General Emergency',
              title: item.title,
              description: item.description,
              seekerName: item.seeker_name || 'Marcus Sterling',
              seekerPhone: item.seeker_phone || '+1 (555) 749-2184',
              address: item.address || '742 Evergreen Terrace, Springfield',
              latitude: item.latitude || 40.7128,
              longitude: item.longitude || -74.006,
              broadcastRadiusKm: item.broadcast_radius_km || 10,
              status: statusMapped,
              createdAt: created,
              elapsedSeconds: elapsed,
              assignedProvider: assigned,
              candidateProviders: candidates
            };
          });

          this.incidents.set(mapped);
          if (!this.selectedIncident() && mapped.length > 0) {
            this.selectedIncident.set(mapped[0]);
          } else if (this.selectedIncident()) {
            const cur = mapped.find(m => m.id === this.selectedIncident()?.id);
            if (cur) this.selectedIncident.set(cur);
          }
        }
      },
      error: () => {
        if (showSpinner) this.isLoading.set(false);
      }
    });
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
    this.isActioning.set(true);

    this.dispatchUC.overrideIncident(inc.id, {
      action: 'EXTEND_RADIUS',
      extend_radius_km: newRadius,
      reason: `Admin expanded broadcast radius from ${inc.broadcastRadiusKm}mi to ${newRadius}mi`
    }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.incidents.update(list =>
          list.map(i => (i.id === inc.id ? { ...i, broadcastRadiusKm: newRadius } : i))
        );
        if (this.selectedIncident()?.id === inc.id) {
          this.selectedIncident.update(i => i ? { ...i, broadcastRadiusKm: newRadius } : null);
        }
        this.successMessage.set(`Broadcast radius expanded to ${newRadius} miles.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to extend radius');
        setTimeout(() => this.errorMessage.set(null), 4000);
      }
    });

    this.audit.logAction({
      actionType: 'SOS_EXTEND_RADIUS',
      targetEntity: 'EMERGENCY_DISPATCH',
      targetId: inc.dispatchNumber,
      details: `Admin expanded broadcast radius from ${inc.broadcastRadiusKm}mi to ${newRadius}mi`
    });
  }

  reassignProvider(inc: EmergencyIncident, prov: IncidentProviderCandidate) {
    this.isActioning.set(true);

    this.dispatchUC.overrideIncident(inc.id, {
      action: 'REASSIGN',
      target_provider_id: prov.id,
      reason: `Admin manually assigned provider ${prov.name} (${prov.id})`
    }).subscribe({
      next: () => {
        this.isActioning.set(false);
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
        this.successMessage.set(`Provider ${prov.name} assigned to incident.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to reassign provider');
        setTimeout(() => this.errorMessage.set(null), 4000);
      }
    });

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
    this.isActioning.set(true);

    this.dispatchUC.overrideIncident(inc.id, {
      action: 'CANCEL',
      reason: 'Admin cancelled emergency dispatch and flagged as fraudulent'
    }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.incidents.update(list =>
          list.map(i => (i.id === inc.id ? { ...i, status: 'CANCELLED_FRAUD', isFlaggedFraud: true } : i))
        );
        if (this.selectedIncident()?.id === inc.id) {
          this.selectedIncident.update(i => i ? { ...i, status: 'CANCELLED_FRAUD', isFlaggedFraud: true } : null);
        }
        this.successMessage.set(`Incident ${inc.dispatchNumber} cancelled & flagged.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to cancel incident');
        setTimeout(() => this.errorMessage.set(null), 4000);
      }
    });

    this.audit.logAction({
      actionType: 'SOS_CANCEL_FRAUD',
      targetEntity: 'EMERGENCY_DISPATCH',
      targetId: inc.dispatchNumber,
      details: `Admin cancelled emergency dispatch and flagged user as suspicious`
    });
  }
}
