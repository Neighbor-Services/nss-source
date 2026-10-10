import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CatalogUseCase } from '../../../core/usecases/catalog.usecase';
import { CategoryItem, CatalogServiceItem } from '../../../core/domain/entities/catalog.model';

import { DialogService } from '../../../core/services/dialog.service';

@Component({
  selector: 'app-catalog',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './catalog.component.html',
  styleUrl: './catalog.component.css'
})
export class CatalogComponent implements OnInit {
  categories = signal<CategoryItem[]>([]);
  services = signal<CatalogServiceItem[]>([]);
  isLoading = signal(false);
  isSubmitting = signal(false);

  selectedCategoryTab = signal<string>('all');
  serviceSearchQuery = signal<string>('');
  categorySearchQuery = signal<string>('');
  activeSection = signal<'CATEGORIES' | 'SERVICES' | 'TAXONOMY' | 'AI_SYNONYMS'>('CATEGORIES');

  // Bulk Selection States
  selectedCategoryIds = signal<string[]>([]);
  selectedServiceIds = signal<string[]>([]);
  isBatchDeletingCategories = signal(false);
  isBatchDeletingServices = signal(false);

  // Computed KPIs
  totalCategories = computed(() => this.categories().length);
  totalServices = computed(() => this.services().length);
  featuredServicesCount = computed(() => this.services().filter(s => s.isPopular).length);
  activeCategoriesCount = computed(() => this.categories().filter(c => c.isActive).length);

  // Batch Import / Export State
  showBatchImportModal = signal(false);
  batchImportText = signal('');
  isImporting = signal(false);
  importResultMsg = signal<string | null>(null);

  // AI Synonym Manager State
  aiSynonyms = signal<Record<string, string[]>>({});
  aiSynonymsCount = signal(0);
  newConceptKey = signal('');
  newConceptSynonyms = signal('');
  isAddingConcept = signal(false);

  // Category Modal State
  showCategoryModal = signal(false);
  isEditingCategory = signal(false);
  categoryForm: Partial<CategoryItem> = {
    name: '',
    slug: '',
    description: '',
    icon: 'star',
    isActive: true
  };


  // Service Modal State
  showServiceModal = signal(false);
  isEditingService = signal(false);
  serviceForm: Partial<CatalogServiceItem> = {
    categoryId: '',
    name: '',
    description: '',
    minPrice: 50,
    maxPrice: 150,
    pricingType: 'hourly',
    isPopular: false,
    isActive: true
  };

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  constructor(
    private catalogUC: CatalogUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit() {
    this.loadCatalog();
  }

  loadCatalog() {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.catalogUC.listCategories().subscribe({
      next: (cats) => {
        this.categories.set(cats);
        // Clean up any stale selected category IDs that no longer exist
        const validIds = new Set(cats.map(c => c.id));
        this.selectedCategoryIds.update(ids => ids.filter(id => validIds.has(id)));
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to load categories');
      }
    });

    this.catalogUC.listCatalogServices().subscribe({
      next: (srvs) => {
        this.services.set(srvs);
        // Clean up any stale selected service IDs that no longer exist
        const validIds = new Set(srvs.map(s => s.id));
        this.selectedServiceIds.update(ids => ids.filter(id => validIds.has(id)));
      }
    });
  }

  filteredCategories(): CategoryItem[] {
    let list = this.categories();
    const q = this.categorySearchQuery().trim().toLowerCase();
    if (q) {
      list = list.filter(c =>
        c.name.toLowerCase().includes(q) ||
        (c.slug || '').toLowerCase().includes(q) ||
        (c.description || '').toLowerCase().includes(q)
      );
    }
    return list;
  }

  filteredServices(): CatalogServiceItem[] {
    let list = this.services();
    const tab = this.selectedCategoryTab();
    if (tab !== 'all') list = list.filter(s => s.categoryId === tab);

    const q = this.serviceSearchQuery().trim().toLowerCase();
    if (q) {
      list = list.filter(s =>
        s.name.toLowerCase().includes(q) ||
        (s.description || '').toLowerCase().includes(q) ||
        (s.categoryName || '').toLowerCase().includes(q)
      );
    }
    return list;
  }

  getServiceCountForCategory(catId: string): number {
    return this.services().filter(s => s.categoryId === catId).length;
  }

  // ─── CATEGORY SELECTION & BULK ACTIONS ──────────────────────────────────────

  isAllCategoriesSelected(): boolean {
    const list = this.filteredCategories();
    if (list.length === 0) return false;
    const selected = this.selectedCategoryIds();
    return list.every(c => selected.includes(c.id));
  }

  toggleSelectAllCategories() {
    const list = this.filteredCategories();
    if (this.isAllCategoriesSelected()) {
      const listIds = new Set(list.map(c => c.id));
      this.selectedCategoryIds.update(ids => ids.filter(id => !listIds.has(id)));
    } else {
      const current = new Set(this.selectedCategoryIds());
      list.forEach(c => current.add(c.id));
      this.selectedCategoryIds.set(Array.from(current));
    }
  }

  isCategorySelected(id: string): boolean {
    return this.selectedCategoryIds().includes(id);
  }

  toggleCategorySelect(id: string) {
    this.selectedCategoryIds.update(ids =>
      ids.includes(id) ? ids.filter(item => item !== id) : [...ids, id]
    );
  }

  clearCategorySelection() {
    this.selectedCategoryIds.set([]);
  }

  async bulkDeleteCategories() {
    const ids = this.selectedCategoryIds();
    if (ids.length === 0) return;

    const confirmed = await this.dialog.dangerConfirm(
      'Bulk Delete Categories',
      `Are you sure you want to delete ${ids.length} selected categories? All catalog services under these categories will also be removed. This action cannot be undone.`,
      `Delete ${ids.length} Categories`
    );
    if (!confirmed) return;

    this.isBatchDeletingCategories.set(true);
    this.errorMessage.set(null);

    this.catalogUC.bulkDeleteCategories(ids).subscribe({
      next: (res) => {
        this.isBatchDeletingCategories.set(false);
        this.selectedCategoryIds.set([]);
        this.successMessage.set(`Successfully deleted ${res.deleted_count || ids.length} categories.`);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isBatchDeletingCategories.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to bulk delete categories');
      }
    });
  }

  // ─── SERVICE SELECTION & BULK ACTIONS ───────────────────────────────────────

  isAllServicesSelected(): boolean {
    const list = this.filteredServices();
    if (list.length === 0) return false;
    const selected = this.selectedServiceIds();
    return list.every(s => selected.includes(s.id));
  }

  toggleSelectAllServices() {
    const list = this.filteredServices();
    if (this.isAllServicesSelected()) {
      const listIds = new Set(list.map(s => s.id));
      this.selectedServiceIds.update(ids => ids.filter(id => !listIds.has(id)));
    } else {
      const current = new Set(this.selectedServiceIds());
      list.forEach(s => current.add(s.id));
      this.selectedServiceIds.set(Array.from(current));
    }
  }

  isServiceSelected(id: string): boolean {
    return this.selectedServiceIds().includes(id);
  }

  toggleServiceSelect(id: string) {
    this.selectedServiceIds.update(ids =>
      ids.includes(id) ? ids.filter(item => item !== id) : [...ids, id]
    );
  }

  clearServiceSelection() {
    this.selectedServiceIds.set([]);
  }

  async bulkDeleteServices() {
    const ids = this.selectedServiceIds();
    if (ids.length === 0) return;

    const confirmed = await this.dialog.dangerConfirm(
      'Bulk Delete Catalog Services',
      `Are you sure you want to delete ${ids.length} selected catalog services? This action cannot be undone.`,
      `Delete ${ids.length} Services`
    );
    if (!confirmed) return;

    this.isBatchDeletingServices.set(true);
    this.errorMessage.set(null);

    this.catalogUC.bulkDeleteCatalogServices(ids).subscribe({
      next: (res) => {
        this.isBatchDeletingServices.set(false);
        this.selectedServiceIds.set([]);
        this.successMessage.set(`Successfully deleted ${res.deleted_count || ids.length} catalog services.`);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isBatchDeletingServices.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to bulk delete catalog services');
      }
    });
  }

  // ─── CATEGORY ACTIONS ───────────────────────────────────────────────────────

  openAddCategory() {
    this.categoryForm = {
      name: '',
      slug: '',
      description: '',
      icon: 'star',
      isActive: true
    };
    this.isEditingCategory.set(false);
    this.showCategoryModal.set(true);
  }

  openEditCategory(cat: CategoryItem) {
    this.categoryForm = { ...cat };
    this.isEditingCategory.set(true);
    this.showCategoryModal.set(true);
  }

  closeCategoryModal() {
    this.showCategoryModal.set(false);
  }

  onCategoryNameChange() {
    if (!this.isEditingCategory()) {
      this.categoryForm.slug = (this.categoryForm.name || '')
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
    }
  }

  saveCategory() {
    if (!this.categoryForm.name?.trim()) {
      this.errorMessage.set('Category name is required');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    if (this.isEditingCategory() && this.categoryForm.id) {
      this.catalogUC.updateCategory(this.categoryForm.id, this.categoryForm).subscribe({
        next: () => {
          this.isSubmitting.set(false);
          this.showCategoryModal.set(false);
          this.successMessage.set(`Category "${this.categoryForm.name}" updated successfully!`);
          this.loadCatalog();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isSubmitting.set(false);
          this.errorMessage.set(err.error?.message || 'Failed to update category');
        }
      });
    } else {
      this.catalogUC.createCategory(this.categoryForm).subscribe({
        next: () => {
          this.isSubmitting.set(false);
          this.showCategoryModal.set(false);
          this.successMessage.set(`Category "${this.categoryForm.name}" created successfully!`);
          this.loadCatalog();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isSubmitting.set(false);
          this.errorMessage.set(err.error?.message || 'Failed to create category');
        }
      });
    }
  }

  async deleteCategory(cat: CategoryItem) {
    const confirmed = await this.dialog.dangerConfirm(
      'Delete Category',
      `Are you sure you want to delete category "${cat.name}"? This action cannot be undone.`,
      'Delete Category'
    );
    if (!confirmed) return;

    this.catalogUC.deleteCategory(cat.id).subscribe({
      next: () => {
        this.selectedCategoryIds.update(ids => ids.filter(id => id !== cat.id));
        this.successMessage.set(`Category "${cat.name}" deleted.`);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to delete category');
      }
    });
  }

  // ─── SERVICE ACTIONS ────────────────────────────────────────────────────────

  openAddService() {
    this.serviceForm = {
      categoryId: this.categories()[0]?.id || '',
      name: '',
      description: '',
      minPrice: 50,
      maxPrice: 150,
      pricingType: 'hourly',
      isPopular: false,
      isActive: true
    };
    this.isEditingService.set(false);
    this.showServiceModal.set(true);
  }

  openEditService(srv: CatalogServiceItem) {
    this.serviceForm = { ...srv };
    this.isEditingService.set(true);
    this.showServiceModal.set(true);
  }

  closeServiceModal() {
    this.showServiceModal.set(false);
  }

  saveService() {
    if (!this.serviceForm.name?.trim() || !this.serviceForm.categoryId) {
      this.errorMessage.set('Service name and category are required');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    if (this.isEditingService() && this.serviceForm.id) {
      this.catalogUC.updateCatalogService(this.serviceForm.id, this.serviceForm).subscribe({
        next: () => {
          this.isSubmitting.set(false);
          this.showServiceModal.set(false);
          this.successMessage.set(`Catalog service "${this.serviceForm.name}" updated successfully!`);
          this.loadCatalog();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isSubmitting.set(false);
          this.errorMessage.set(err.error?.message || 'Failed to update catalog service');
        }
      });
    } else {
      this.catalogUC.createCatalogService(this.serviceForm).subscribe({
        next: () => {
          this.isSubmitting.set(false);
          this.showServiceModal.set(false);
          this.successMessage.set(`Catalog service "${this.serviceForm.name}" created successfully!`);
          this.loadCatalog();
          setTimeout(() => this.successMessage.set(null), 4000);
        },
        error: (err) => {
          this.isSubmitting.set(false);
          this.errorMessage.set(err.error?.message || 'Failed to create catalog service');
        }
      });
    }
  }

  async deleteService(srv: CatalogServiceItem) {
    const confirmed = await this.dialog.dangerConfirm(
      'Delete Catalog Service',
      `Are you sure you want to delete service "${srv.name}"? This action cannot be undone.`,
      'Delete Service'
    );
    if (!confirmed) return;

    this.catalogUC.deleteCatalogService(srv.id).subscribe({
      next: () => {
        this.selectedServiceIds.update(ids => ids.filter(id => id !== srv.id));
        this.successMessage.set(`Catalog service "${srv.name}" deleted.`);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to delete service');
      }
    });
  }

  // ─── BATCH IMPORT / EXPORT ──────────────────────────────────────────────────

  openBatchImport() {
    this.batchImportText.set('');
    this.importResultMsg.set(null);
    this.showBatchImportModal.set(true);
  }

  closeBatchImport() {
    this.showBatchImportModal.set(false);
  }

  downloadSampleCSV() {
    const csvHeader = 'category_name,category_description,service_name,service_description,default_service_location,specialties\n';
    const sampleRows =
      'Home Services,General home maintenance,Electrical Wiring,Indoor and outdoor wiring & outlet repair,CUSTOMER_LOCATION,Panel Upgrades|Lighting|Outlets\n' +
      'Automotive,Mobile auto repairs,Brake Pad Replacement,On-site brake pad & rotor service,CUSTOMER_LOCATION,Brakes|Rotors|Inspections\n' +
      'Health & Wellness,Personal fitness and therapy,Deep Tissue Massage,Certified massage therapy,PROVIDER_LOCATION,Deep Tissue|Sports Massage\n';

    const blob = new Blob([csvHeader + sampleRows], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'catalog_batch_template.csv';
    a.click();
    URL.revokeObjectURL(url);
  }

  handleCSVFileSelected(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      this.batchImportText.set(text);
    };
    reader.readAsText(file);
  }

  submitBatchImport() {
    const text = this.batchImportText().trim();
    if (!text) {
      this.importResultMsg.set('Please paste CSV/JSON or select a file to import.');
      return;
    }

    this.isImporting.set(true);
    this.importResultMsg.set(null);

    let items: any[] = [];
    if (text.startsWith('[') || text.startsWith('{')) {
      try {
        const parsed = JSON.parse(text);
        items = Array.isArray(parsed) ? parsed : [parsed];
      } catch (e: any) {
        this.isImporting.set(false);
        this.importResultMsg.set('Invalid JSON format: ' + e.message);
        return;
      }
    } else {
      // Parse CSV
      const lines = text.split('\n').filter(l => l.trim().length > 0);
      if (lines.length <= 1) {
        this.isImporting.set(false);
        this.importResultMsg.set('CSV must contain a header and at least one data row.');
        return;
      }

      const headers = lines[0].split(',').map(h => h.trim().toLowerCase().replace(/['"]+/g, ''));
      for (let i = 1; i < lines.length; i++) {
        const values = lines[i].split(',').map(v => v.trim().replace(/^["']|["']$/g, ''));
        const row: any = {};
        headers.forEach((h, idx) => {
          row[h] = values[idx] || '';
        });

        const specialties = (row.specialties || '').split(/[|,]/).map((s: string) => s.trim()).filter(Boolean);

        items.push({
          category_name: row.category_name || row.category || '',
          category_description: row.category_description || '',
          service_name: row.service_name || row.service || row.name || '',
          service_description: row.service_description || row.description || '',
          default_service_location: row.default_service_location || 'CUSTOMER_LOCATION',
          specialties: specialties
        });
      }
    }

    this.catalogUC.importCatalogBatch(items).subscribe({
      next: (res) => {
        this.isImporting.set(false);
        this.successMessage.set(`Successfully imported ${res.imported_count} catalog items!`);
        this.showBatchImportModal.set(false);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.isImporting.set(false);
        this.importResultMsg.set(err.error?.message || 'Failed to import catalog batch.');
      }
    });
  }

  exportCatalogCSV() {
    this.catalogUC.exportCatalogBatch().subscribe({
      next: (res) => {
        const items = res.items || [];
        if (items.length === 0) {
          alert('No catalog services found to export.');
          return;
        }

        const headers = ['CategoryID', 'CategoryName', 'ServiceID', 'ServiceName', 'Description', 'Location', 'Specialties'];
        const csvRows = [headers.join(',')];

        items.forEach(item => {
          const row = [
            `"${item.category_id}"`,
            `"${(item.category_name || '').replace(/"/g, '""')}"`,
            `"${item.service_id}"`,
            `"${(item.service_name || '').replace(/"/g, '""')}"`,
            `"${(item.service_description || '').replace(/"/g, '""')}"`,
            `"${item.default_service_location || ''}"`,
            `"${(item.specialties || []).join('|')}"`
          ];
          csvRows.push(row.join(','));
        });

        const blob = new Blob([csvRows.join('\n')], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `neighbor_catalog_export_${new Date().toISOString().slice(0, 10)}.csv`;
        a.click();
        URL.revokeObjectURL(url);
      }
    });
  }

  exportCatalogJSON() {
    this.catalogUC.exportCatalogBatch().subscribe({
      next: (res) => {
        const blob = new Blob([JSON.stringify(res.items || [], null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `neighbor_catalog_export_${new Date().toISOString().slice(0, 10)}.json`;
        a.click();
        URL.revokeObjectURL(url);
      }
    });
  }

  // ─── AI SYNONYM ONTOLOGY MANAGER ────────────────────────────────────────────

  loadAISynonyms() {
    this.catalogUC.getAISynonyms().subscribe({
      next: (res) => {
        this.aiSynonyms.set(res.synonyms || {});
        this.aiSynonymsCount.set(res.count || Object.keys(res.synonyms || {}).length);
      }
    });
  }

  getSynonymKeys(): string[] {
    return Object.keys(this.aiSynonyms()).sort();
  }

  addConcept() {
    const key = this.newConceptKey().trim().toLowerCase();
    const synStr = this.newConceptSynonyms().trim();

    if (!key || !synStr) {
      this.errorMessage.set('Concept key and at least one synonym are required.');
      return;
    }

    const synonyms = synStr.split(/[,|\n]/).map(s => s.trim().toLowerCase()).filter(Boolean);
    if (synonyms.length === 0) {
      this.errorMessage.set('Please provide valid synonyms separated by commas.');
      return;
    }

    this.isAddingConcept.set(true);
    this.catalogUC.addAISynonym(key, synonyms).subscribe({
      next: () => {
        this.isAddingConcept.set(false);
        this.newConceptKey.set('');
        this.newConceptSynonyms.set('');
        this.successMessage.set(`Concept "${key}" added with ${synonyms.length} synonyms!`);
        this.loadAISynonyms();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isAddingConcept.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to add concept synonym.');
      }
    });
  }

  async removeConcept(key: string) {
    const confirmed = await this.dialog.dangerConfirm(
      'Remove AI Synonym Concept',
      `Are you sure you want to remove the concept "${key}" from the AI matcher ontology?`,
      'Remove Concept'
    );
    if (!confirmed) return;

    this.catalogUC.deleteAISynonym(key).subscribe({
      next: () => {
        this.successMessage.set(`Concept "${key}" removed.`);
        this.loadAISynonyms();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to remove concept.');
      }
    });
  }
}

