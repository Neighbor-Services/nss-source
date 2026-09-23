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
  activeSection = signal<'CATEGORIES' | 'SERVICES' | 'TAXONOMY'>('CATEGORIES');

  // Computed KPIs
  totalCategories = computed(() => this.categories().length);
  totalServices = computed(() => this.services().length);
  featuredServicesCount = computed(() => this.services().filter(s => s.isPopular).length);
  activeCategoriesCount = computed(() => this.categories().filter(c => c.isActive).length);

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
      }
    });
  }

  filteredServices(): CatalogServiceItem[] {
    let list = this.services();
    const tab = this.selectedCategoryTab();
    if (tab !== 'all') list = list.filter(s => s.categoryId === tab);

    const q = this.serviceSearchQuery().trim().toLowerCase();
    if (q) {
      list = list.filter(s =>
        s.name.toLowerCase().includes(q) ||
        (s.description || '').toLowerCase().includes(q)
      );
    }
    return list;
  }

  getServiceCountForCategory(catId: string): number {
    return this.services().filter(s => s.categoryId === catId).length;
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
        this.successMessage.set(`Catalog service "${srv.name}" deleted.`);
        this.loadCatalog();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to delete service');
      }
    });
  }
}
