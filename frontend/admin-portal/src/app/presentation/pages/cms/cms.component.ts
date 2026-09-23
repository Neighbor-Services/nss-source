import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CmsUseCase } from '../../../core/usecases/cms.usecase';
import { FAQ, Testimonial, HeroSection, SiteStat, AboutContent } from '../../../core/domain/entities/cms.model';

@Component({
  selector: 'app-cms',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './cms.component.html',
  styleUrl: './cms.component.css'
})
export class CmsComponent implements OnInit {
  activeTab = signal<'HERO' | 'FAQS' | 'TESTIMONIALS' | 'STATS_ABOUT'>('HERO');

  // Hero Section
  hero = signal<HeroSection>({
    headline: '',
    subheadline: '',
    cta_text: 'Get Started',
    cta_link: '#',
    image: '',
    is_active: true
  });

  // FAQs
  faqs = signal<FAQ[]>([]);
  faqCategoryFilter = signal<string>('ALL');
  isFaqModalOpen = signal(false);
  isFaqEditing = signal(false);
  currentFaq = signal<Partial<FAQ>>({
    question: '',
    answer: '',
    category: 'general',
    order: 1,
    is_active: true
  });

  // Testimonials
  testimonials = signal<Testimonial[]>([]);
  isTestimonialModalOpen = signal(false);
  isTestimonialEditing = signal(false);
  currentTestimonial = signal<Partial<Testimonial>>({
    name: '',
    role: 'Homeowner / Seeker',
    content: '',
    rating: 5,
    is_active: true
  });

  // Stats & About
  stats = signal<SiteStat[]>([]);
  isStatModalOpen = signal(false);
  isStatEditing = signal(false);
  currentStat = signal<Partial<SiteStat>>({
    label: '',
    value: '',
    order: 1,
    is_active: true
  });

  about = signal<AboutContent>({
    title: 'Our Mission',
    story_headline: 'Building Community Connections',
    story_text_1: '',
    story_text_2: '',
    mission_text: '',
    vision_text: '',
    year_founded: '2025',
    cities_covered: '50+',
    is_active: true
  });

  // States
  isLoading = signal(false);
  isSaving = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed FAQs
  filteredFaqs = computed(() => {
    if (this.faqCategoryFilter() === 'ALL') return this.faqs();
    return this.faqs().filter(f => f.category === this.faqCategoryFilter());
  });

  constructor(private cmsUC: CmsUseCase) {}

  ngOnInit(): void {
    this.loadAllCMS();
  }

  setTab(tab: 'HERO' | 'FAQS' | 'TESTIMONIALS' | 'STATS_ABOUT'): void {
    this.activeTab.set(tab);
  }

  loadAllCMS(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    // Load Hero
    this.cmsUC.getHero().subscribe({
      next: (h) => {
        if (h) this.hero.set(h);
      }
    });

    // Load FAQs
    this.cmsUC.listFAQs().subscribe({
      next: (res) => {
        this.faqs.set(res.results || []);
      }
    });

    // Load Testimonials
    this.cmsUC.listTestimonials().subscribe({
      next: (res) => {
        this.testimonials.set(res.results || []);
      }
    });

    // Load Stats
    this.cmsUC.listStats().subscribe({
      next: (res) => {
        this.stats.set(res.results || []);
      }
    });

    // Load About
    this.cmsUC.getAbout().subscribe({
      next: (a) => {
        if (a) this.about.set(a);
        this.isLoading.set(false);
      },
      error: () => {
        this.isLoading.set(false);
      }
    });
  }

  // ── HERO ACTIONS ───────────────────────────────────────────────────────────
  saveHero(): void {
    this.isSaving.set(true);
    this.cmsUC.updateHero(this.hero()).subscribe({
      next: (updated) => {
        this.isSaving.set(false);
        this.hero.set(updated);
        this.showSuccess('Hero section successfully updated!');
      },
      error: (err) => {
        this.isSaving.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update hero section.');
      }
    });
  }

  // ── FAQ ACTIONS ────────────────────────────────────────────────────────────
  openCreateFaqModal(): void {
    this.isFaqEditing.set(false);
    this.currentFaq.set({
      question: '',
      answer: '',
      category: 'general',
      order: (this.faqs().length + 1),
      is_active: true
    });
    this.isFaqModalOpen.set(true);
  }

  openEditFaqModal(faq: FAQ): void {
    this.isFaqEditing.set(true);
    this.currentFaq.set({ ...faq });
    this.isFaqModalOpen.set(true);
  }

  closeFaqModal(): void {
    this.isFaqModalOpen.set(false);
  }

  saveFaq(): void {
    const faq = this.currentFaq();
    if (!faq.question || !faq.answer) {
      this.errorMessage.set('Question and answer are required.');
      return;
    }

    this.isSaving.set(true);
    if (this.isFaqEditing() && faq.id) {
      this.cmsUC.updateFAQ(faq.id, faq).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isFaqModalOpen.set(false);
          this.showSuccess('FAQ updated successfully.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update FAQ.');
        }
      });
    } else {
      this.cmsUC.createFAQ(faq).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isFaqModalOpen.set(false);
          this.showSuccess('FAQ created successfully.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create FAQ.');
        }
      });
    }
  }

  deleteFaq(faq: FAQ): void {
    if (!confirm(`Delete FAQ "${faq.question}"?`)) return;
    this.cmsUC.deleteFAQ(faq.id).subscribe({
      next: () => {
        this.showSuccess('FAQ deleted.');
        this.loadAllCMS();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete FAQ.');
      }
    });
  }

  // ── TESTIMONIAL ACTIONS ────────────────────────────────────────────────────
  openCreateTestimonialModal(): void {
    this.isTestimonialEditing.set(false);
    this.currentTestimonial.set({
      name: '',
      role: 'Verified Neighbor',
      content: '',
      rating: 5,
      is_active: true
    });
    this.isTestimonialModalOpen.set(true);
  }

  openEditTestimonialModal(t: Testimonial): void {
    this.isTestimonialEditing.set(true);
    this.currentTestimonial.set({ ...t });
    this.isTestimonialModalOpen.set(true);
  }

  closeTestimonialModal(): void {
    this.isTestimonialModalOpen.set(false);
  }

  saveTestimonial(): void {
    const t = this.currentTestimonial();
    if (!t.name || !t.content) {
      this.errorMessage.set('Author name and testimonial story are required.');
      return;
    }

    this.isSaving.set(true);
    if (this.isTestimonialEditing() && t.id) {
      this.cmsUC.updateTestimonial(t.id, t).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isTestimonialModalOpen.set(false);
          this.showSuccess('Testimonial updated.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update testimonial.');
        }
      });
    } else {
      this.cmsUC.createTestimonial(t).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isTestimonialModalOpen.set(false);
          this.showSuccess('Testimonial added.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create testimonial.');
        }
      });
    }
  }

  deleteTestimonial(t: Testimonial): void {
    if (!confirm(`Delete testimonial from ${t.name}?`)) return;
    this.cmsUC.deleteTestimonial(t.id).subscribe({
      next: () => {
        this.showSuccess('Testimonial removed.');
        this.loadAllCMS();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete testimonial.');
      }
    });
  }

  // ── STATS & ABOUT ACTIONS ──────────────────────────────────────────────────
  openCreateStatModal(): void {
    this.isStatEditing.set(false);
    this.currentStat.set({
      label: '',
      value: '1,000+',
      order: (this.stats().length + 1),
      is_active: true
    });
    this.isStatModalOpen.set(true);
  }

  openEditStatModal(s: SiteStat): void {
    this.isStatEditing.set(true);
    this.currentStat.set({ ...s });
    this.isStatModalOpen.set(true);
  }

  closeStatModal(): void {
    this.isStatModalOpen.set(false);
  }

  saveStat(): void {
    const s = this.currentStat();
    if (!s.label || !s.value) {
      this.errorMessage.set('Label and value are required.');
      return;
    }

    this.isSaving.set(true);
    if (this.isStatEditing() && s.id) {
      this.cmsUC.updateStat(s.id, s).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isStatModalOpen.set(false);
          this.showSuccess('Stat metric updated.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to update stat.');
        }
      });
    } else {
      this.cmsUC.createStat(s).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.isStatModalOpen.set(false);
          this.showSuccess('Stat metric added.');
          this.loadAllCMS();
        },
        error: (err) => {
          this.isSaving.set(false);
          this.errorMessage.set(err?.error?.message || 'Failed to create stat.');
        }
      });
    }
  }

  deleteStat(s: SiteStat): void {
    if (!confirm(`Delete stat "${s.label}"?`)) return;
    this.cmsUC.deleteStat(s.id).subscribe({
      next: () => {
        this.showSuccess('Stat deleted.');
        this.loadAllCMS();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete stat.');
      }
    });
  }

  saveAbout(): void {
    this.isSaving.set(true);
    this.cmsUC.updateAbout(this.about()).subscribe({
      next: (updated) => {
        this.isSaving.set(false);
        this.about.set(updated);
        this.showSuccess('About page content saved!');
      },
      error: (err) => {
        this.isSaving.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update about content.');
      }
    });
  }

  private showSuccess(msg: string): void {
    this.successMessage.set(msg);
    setTimeout(() => this.successMessage.set(null), 3500);
  }
}
