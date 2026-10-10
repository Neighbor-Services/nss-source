import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';
import { ADMIN_API_CONFIG } from '../../data/datasources/admin-api.config';

export interface RealtimeServerEvent {
  type: string;
  title?: string;
  body?: string;
  data?: any;
  created_at?: string;
}

function getWsUrl(): string {
  const base = ADMIN_API_CONFIG.baseUrl;
  const isSsl = base.startsWith('https://');
  const host = base.replace(/^https?:\/\//, '').split('/api')[0];
  const protocol = isSsl ? 'wss://' : 'ws://';
  const token = typeof localStorage !== 'undefined' ? (localStorage.getItem('admin_access_token') || localStorage.getItem('token') || '') : '';
  return `${protocol}${host}/ws${token ? '?token=' + encodeURIComponent(token) : ''}`;
}

@Injectable({
  providedIn: 'root'
})
export class WebSocketService {
  private socket: WebSocket | null = null;
  private eventsSubject = new Subject<RealtimeServerEvent>();
  private reconnectTimeout: any;
  private isConnected = false;

  constructor() {}

  public connect(url?: string): void {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const token = typeof localStorage !== 'undefined' ? (localStorage.getItem('admin_access_token') || localStorage.getItem('token') || '') : '';
    if (!token) {
      // Delay connection until user has logged in
      this.scheduleReconnect(url || getWsUrl());
      return;
    }

    const targetUrl = url || getWsUrl();

    try {
      this.socket = new WebSocket(targetUrl);

      this.socket.onopen = () => {
        this.isConnected = true;
        console.log('[WebSocketService] Connected to real-time events hub.');
        // Join admin room
        this.send({ action: 'join', room: 'admin' });
      };

      this.socket.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          this.eventsSubject.next(parsed);
        } catch (e) {
          console.warn('[WebSocketService] Failed to parse message', event.data);
        }
      };

      this.socket.onclose = () => {
        this.isConnected = false;
        console.log('[WebSocketService] Disconnected. Reconnecting in 5s...');
        this.scheduleReconnect(targetUrl);
      };

      this.socket.onerror = (err) => {
        console.warn('[WebSocketService] WebSocket connection retry:', err);
        this.socket?.close();
      };
    } catch (err) {
      console.warn('[WebSocketService] Failed to establish connection:', err);
      this.scheduleReconnect(targetUrl);
    }
  }

  public getEvents(): Observable<RealtimeServerEvent> {
    return this.eventsSubject.asObservable();
  }

  public send(payload: any): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(payload));
    }
  }

  private scheduleReconnect(url: string): void {
    clearTimeout(this.reconnectTimeout);
    this.reconnectTimeout = setTimeout(() => {
      this.connect(url);
    }, 5000);
  }

  public playIncidentChime(): void {
    try {
      const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext;
      if (!AudioContextClass) return;
      const ctx = new AudioContextClass();
      
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
      osc.frequency.setValueAtTime(880, ctx.currentTime + 0.1); // A5

      gain.gain.setValueAtTime(0.15, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.35);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start();
      osc.stop(ctx.currentTime + 0.35);
    } catch (_) {}
  }

  public disconnect(): void {
    clearTimeout(this.reconnectTimeout);
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}
