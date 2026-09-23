import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';

export interface RealtimeServerEvent {
  type: string;
  title?: string;
  body?: string;
  data?: any;
  created_at?: string;
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

  public connect(url = 'ws://localhost:8000/ws'): void {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    try {
      this.socket = new WebSocket(url);

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
        this.scheduleReconnect(url);
      };

      this.socket.onerror = (err) => {
        console.error('[WebSocketService] WebSocket error:', err);
        this.socket?.close();
      };
    } catch (err) {
      console.error('[WebSocketService] Failed to establish connection:', err);
      this.scheduleReconnect(url);
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
