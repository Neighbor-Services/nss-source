import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { LegalUseCase } from '../../../core/usecases/legal.usecase';
import { LegalDocument } from '../../../core/domain/entities/legal.model';

@Component({
  selector: 'app-legal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './legal.component.html',
  styleUrl: './legal.component.css'
})
export class LegalComponent implements OnInit {
  documents = signal<LegalDocument[]>([]);
  selectedType = signal<string>('ALL');
  selectedStatus = signal<'ALL' | 'ACTIVE' | 'INACTIVE'>('ALL');

  isLoading = signal(false);
  isSaving = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Modal State
  isModalOpen = signal(false);
  isEditing = signal(false);
  currentDoc = signal<Partial<LegalDocument>>({
    document_type: 'TERMS',
    title: '',
    slug: '',
    content: '',
    version: '1.0',
    is_active: true
  });

  // KPI Metrics
  totalDocs = computed(() => this.documents().length);
  activeDocs = computed(() => this.documents().filter(d => d.is_active).length);
  docTypes = computed(() => Array.from(new Set(this.documents().map(d => d.document_type))));

  filteredDocuments = computed(() => {
    return this.documents().filter(d => {
      const matchType = this.selectedType() === 'ALL' || d.document_type === this.selectedType();
      const matchStatus = this.selectedStatus() === 'ALL' || 
        (this.selectedStatus() === 'ACTIVE' && d.is_active) || 
        (this.selectedStatus() === 'INACTIVE' && !d.is_active);
      return matchType && matchStatus;
    });
  });

  constructor(private legalUC: LegalUseCase) {}

  ngOnInit(): void {
    this.loadDocuments();
  }

  loadDocuments(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    let isActive: boolean | undefined = undefined;
    if (this.selectedStatus() === 'ACTIVE') isActive = true;
    if (this.selectedStatus() === 'INACTIVE') isActive = false;

    this.legalUC.listDocuments({
      type: this.selectedType() !== 'ALL' ? this.selectedType() : undefined,
      is_active: isActive
    }).subscribe({
      next: (res) => {
        this.documents.set(res.results || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load legal documents.');
      }
    });
  }

  setTypeFilter(type: string): void {
    this.selectedType.set(type);
    this.loadDocuments();
  }

  setStatusFilter(status: 'ALL' | 'ACTIVE' | 'INACTIVE'): void {
    this.selectedStatus.set(status);
    this.loadDocuments();
  }

  openCreateModal(): void {
    this.isEditing.set(false);
    this.currentDoc.set({
      document_type: 'TERMS',
      title: '',
      slug: '',
      content: '',
      version: '1.0',
      is_active: true
    });
    this.isModalOpen.set(true);
  }

  openEditModal(doc: LegalDocument): void {
    this.isEditing.set(true);
    this.currentDoc.set({ ...doc });
    this.isModalOpen.set(true);
  }

  closeModal(): void {
    this.isModalOpen.set(false);
  }

  generateSlug(): void {
    const title = this.currentDoc().title || '';
    const slug = title
      .toLowerCase()
      .trim()
      .replace(/[^\w\s-]/g, '')
      .replace(/[\s_-]+/g, '-')
      .replace(/^-+|-+$/g, '');
    this.currentDoc.update(d => ({ ...d, slug }));
  }

  saveDocument(): void {
    const doc = this.currentDoc();
    if (!doc.title || !doc.slug || !doc.content) {
      this.errorMessage.set('Title, slug, and content are required.');
      return;
    }

    this.isSaving.set(true);
    this.errorMessage.set(null);

    if (this.isEditing() && doc.id) {
      this.legalUC.updateDocument(doc.id, doc).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isModalOpen.set(false);
          this.successMessage.set(`Document "${doc.title}" successfully updated.`);
          setTimeout(() => this.successMessage.set(null), 4000);
          this.loadDocuments();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update document.');
        }
      });
    } else {
      this.legalUC.createDocument(doc).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isModalOpen.set(false);
          this.successMessage.set(`Document "${doc.title}" successfully published.`);
          setTimeout(() => this.successMessage.set(null), 4000);
          this.loadDocuments();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create document.');
        }
      });
    }
  }

  deleteDocument(doc: LegalDocument): void {
    if (!confirm(`Are you sure you want to permanently delete "${doc.title}" (v${doc.version})?`)) {
      return;
    }

    this.legalUC.deleteDocument(doc.id).subscribe({
      next: () => {
        this.successMessage.set(`Document "${doc.title}" deleted.`);
        setTimeout(() => this.successMessage.set(null), 4000);
        this.loadDocuments();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete legal document.');
      }
    });
  }
}
