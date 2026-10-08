import { Component, OnInit, signal, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ADMIN_API_CONFIG } from '../../../data/datasources/admin-api.config';

export interface StorageObject {
  name: string;
  bucket: string;
  size: number;
  updated: string;
  content_type?: string;
  url: string;
  hash?: string;
}

@Component({
  selector: 'app-media-browser',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './media-browser.component.html',
  styleUrls: ['./media-browser.component.css']
})
export class MediaBrowserComponent implements OnInit {
  private http = inject(HttpClient);

  readonly buckets = signal<string[]>([
    'neighborservice',
    'profiles',
    'chat',
    'evidence',
    'verifications',
    'reports',
    'portfolio'
  ]);

  readonly activeBucket = signal<string>('neighborservice');
  readonly searchQuery = signal<string>('');
  readonly selectedFilter = signal<'all' | 'images' | 'documents' | 'audio'>('all');
  readonly viewMode = signal<'grid' | 'list'>('grid');

  readonly objects = signal<StorageObject[]>([]);
  readonly isLoading = signal<boolean>(false);
  readonly errorMessage = signal<string | null>(null);

  readonly previewObject = signal<StorageObject | null>(null);
  readonly copyFeedback = signal<string | null>(null);

  // Computed statistics
  readonly filteredObjects = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const filter = this.selectedFilter();
    return this.objects().filter(obj => {
      const matchesSearch = !q || obj.name.toLowerCase().includes(q) || (obj.content_type && obj.content_type.toLowerCase().includes(q));
      if (!matchesSearch) return false;

      if (filter === 'images') {
        return this.isImage(obj.name) || (obj.content_type && obj.content_type.startsWith('image/'));
      }
      if (filter === 'documents') {
        return obj.name.endsWith('.pdf') || obj.name.endsWith('.doc') || obj.name.endsWith('.docx') || obj.name.endsWith('.txt');
      }
      if (filter === 'audio') {
        return obj.name.endsWith('.mp3') || obj.name.endsWith('.wav') || obj.name.endsWith('.m4a');
      }
      return true;
    });
  });

  readonly totalSize = computed(() => {
    return this.objects().reduce((acc, obj) => acc + (obj.size || 0), 0);
  });

  ngOnInit(): void {
    this.fetchObjects();
  }

  selectBucket(bucket: string): void {
    this.activeBucket.set(bucket);
    this.fetchObjects();
  }

  fetchObjects(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    const bucket = this.activeBucket();
    const apiUrl = `${ADMIN_API_CONFIG.baseUrl}/admin/media/browser?bucket=${encodeURIComponent(bucket)}`;

    this.http.get<{ items?: StorageObject[]; data?: any; source?: string }>(apiUrl).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        if (res.items) {
          this.objects.set(res.items);
        } else if (res.data && Array.isArray(res.data.items)) {
          this.objects.set(res.data.items.map((i: any) => ({
            name: i.name || i.path,
            bucket: bucket,
            size: i.size || 0,
            updated: i.updated_at || i.updated || new Date().toISOString(),
            content_type: i.content_type || i.contentType,
            url: this.resolveUrl(i.name || i.path),
            hash: i.hash || i.etag
          })));
        } else {
          this.objects.set([]);
        }
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load storage objects from GoStore appliance.');
      }
    });
  }

  resolveUrl(path: string): string {
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path;
    }
    const clean = path.replace(/^\/+/, '');
    return `${ADMIN_API_CONFIG.baseUrl}/media/${clean}`;
  }

  isImage(name: string): boolean {
    const ext = name.toLowerCase().split('.').pop();
    return ['jpg', 'jpeg', 'png', 'webp', 'gif', 'svg'].includes(ext || '');
  }

  formatBytes(bytes: number): string {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  openPreview(obj: StorageObject): void {
    this.previewObject.set(obj);
  }

  closePreview(): void {
    this.previewObject.set(null);
  }

  copyLink(url: string): void {
    navigator.clipboard.writeText(url);
    this.copyFeedback.set('URL copied to clipboard!');
    setTimeout(() => this.copyFeedback.set(null), 2500);
  }
}
