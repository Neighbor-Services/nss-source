import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { GetCategoriesUseCase } from '../../../core/usecases/get-categories.usecase';
import { Category, CatalogService } from '../../../core/domain/entities/cms.model';

@Component({
  selector: 'app-services',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './services.component.html',
  styleUrl: './services.component.css'
})
export class ServicesComponent implements OnInit {
  categories = signal<Category[]>([]);
  allServices = signal<CatalogService[]>([]);
  selectedCategoryId = signal<string>('all');
  searchQuery = '';

  filteredCategories = computed(() => {
    const catId = this.selectedCategoryId();
    const q = this.searchQuery.trim().toLowerCase();
    let cats = this.categories();
    if (catId !== 'all') {
      cats = cats.filter(c => c.id === catId);
    }
    if (q) {
      cats = cats.filter(c =>
        c.name.toLowerCase().includes(q) ||
        this.getServicesForCategory(c.id).some(s =>
          s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q)
        )
      );
    }
    return cats;
  });

  constructor(
    private getCategoriesUseCase: GetCategoriesUseCase,
    private route: ActivatedRoute
  ) {}

  ngOnInit() {
    this.getCategoriesUseCase.execute().subscribe(cats => {
      this.categories.set(cats);
    });

    this.getCategoriesUseCase.getServices().subscribe(services => {
      this.allServices.set(services);
    });

    this.route.queryParams.subscribe(params => {
      if (params['category']) {
        this.selectedCategoryId.set(params['category']);
      }
      if (params['q']) {
        this.searchQuery = params['q'];
      }
    });
  }

  selectCategory(id: string) {
    this.selectedCategoryId.set(id);
    if (id !== 'all') {
      setTimeout(() => {
        const el = document.getElementById(`cat-${id}`);
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 100);
    }
  }

  getServicesForCategory(catId: string): CatalogService[] {
    const q = this.searchQuery.trim().toLowerCase();
    const services = this.allServices().filter(s => s.categoryId === catId);
    if (!q) return services;
    return services.filter(s =>
      s.name.toLowerCase().includes(q) ||
      s.description.toLowerCase().includes(q)
    );
  }

  getCategoryIconType(catName: string): string {
    const lower = (catName || '').toLowerCase();
    if (lower.includes('auto') || lower.includes('transport')) return 'automotive';
    if (lower.includes('beauty') || lower.includes('groom')) return 'beauty';
    if (lower.includes('care') || lower.includes('family')) return 'caregiving';
    if (lower.includes('clean') || lower.includes('sanit')) return 'cleaning';
    if (lower.includes('educat') || lower.includes('digit') || lower.includes('tutor')) return 'education';
    if (lower.includes('fash') || lower.includes('appar')) return 'fashion';
    if (lower.includes('food') || lower.includes('culin')) return 'food';
    if (lower.includes('health') || lower.includes('well')) return 'healthcare';
    if (lower.includes('home') || lower.includes('repair') || lower.includes('improv')) return 'home';
    if (lower.includes('lawn') || lower.includes('pest') || lower.includes('garden')) return 'lawn';
    if (lower.includes('pet')) return 'pet';
    return 'professional';
  }

  onSearch() {
    // Triggers computed signal
  }

  resetFilters() {
    this.selectedCategoryId.set('all');
    this.searchQuery = '';
  }
}
