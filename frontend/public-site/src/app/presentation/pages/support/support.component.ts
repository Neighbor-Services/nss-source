import { Component, OnInit, AfterViewInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { GetCMSContentUseCase } from '../../../core/usecases/get-cms-content.usecase';
import { FAQItem } from '../../../core/domain/entities/cms.model';

interface FaqGroup {
  category: string;
  label: string;
  value: string;
  items: FAQItem[];
}

@Component({
  selector: 'app-support',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './support.component.html',
  styleUrl: './support.component.css'
})
export class SupportComponent implements OnInit, AfterViewInit {
  faqs = signal<FAQItem[]>([]);
  activeCat = 'all';
  searchQuery = '';
  openFaqId = signal<string | null>(null);

  private getCategoryLabel(cat: string): string {
    const map: Record<string, string> = {
      general: 'General Questions',
      safety: 'Trust & Safety',
      providers: 'For Providers',
      payment: 'Payments & Pricing'
    };
    return map[cat] || cat.charAt(0).toUpperCase() + cat.slice(1);
  }

  faqCategories = computed<{ label: string; value: string }[]>(() => {
    const cats = new Map<string, string>();
    this.faqs().forEach(f => {
      if (f.category) cats.set(f.category, this.getCategoryLabel(f.category));
    });
    return Array.from(cats.entries()).map(([value, label]) => ({ value, label }));
  });

  visibleGroups = computed<FaqGroup[]>(() => {
    const all = this.faqs();
    if (all.length === 0) return [];

    const q = this.searchQuery.trim().toLowerCase();

    // Group by category
    const groupMap = new Map<string, FaqGroup>();
    all.forEach(faq => {
      const cat = faq.category || 'general';
      const label = this.getCategoryLabel(cat);
      if (!groupMap.has(cat)) {
        groupMap.set(cat, { category: cat, label, value: cat, items: [] });
      }

      // Category filter
      if (this.activeCat !== 'all' && cat !== this.activeCat) return;

      // Search filter
      if (q) {
        const matches =
          faq.question.toLowerCase().includes(q) ||
          faq.answer.toLowerCase().includes(q);
        if (!matches) return;
      }

      groupMap.get(cat)!.items.push(faq);
    });

    return Array.from(groupMap.values()).filter(g => g.items.length > 0);
  });

  constructor(
    private getCMSContentUC: GetCMSContentUseCase,
    private route: ActivatedRoute
  ) {}

  ngOnInit() {
    this.getCMSContentUC.execute().subscribe(cms => {
      this.faqs.set(cms.faqs || []);
    });

    this.route.queryParams.subscribe(params => {
      if (params['tab']) this.activeCat = params['tab'];
    });
  }

  ngAfterViewInit() {
    const io = new IntersectionObserver(entries => {
      entries.forEach(e => {
        if (e.isIntersecting) { e.target.classList.add('in'); io.unobserve(e.target); }
      });
    }, { threshold: 0.14 });
    document.querySelectorAll('.rv').forEach(el => io.observe(el));
  }

  setCat(cat: string) {
    this.activeCat = cat;
    this.openFaqId.set(null);
  }

  toggleFaq(id: string) {
    this.openFaqId.update(curr => curr === id ? null : id);
  }

  onSearch() {
    this.openFaqId.set(null);
  }

  clearSearch() {
    this.searchQuery = '';
    this.openFaqId.set(null);
  }
}
