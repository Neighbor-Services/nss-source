import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CmsRepository } from '../repositories/cms.repository';
import { FAQ, Testimonial, HeroSection, SiteStat, AboutContent } from '../domain/entities/cms.model';

@Injectable({
  providedIn: 'root'
})
export class CmsUseCase {
  constructor(private cmsRepo: CmsRepository) {}

  // FAQs
  listFAQs(params?: { category?: string; is_active?: boolean }): Observable<{ results: FAQ[]; count: number }> {
    return this.cmsRepo.listFAQs(params);
  }

  createFAQ(faq: Partial<FAQ>): Observable<FAQ> {
    return this.cmsRepo.createFAQ(faq);
  }

  updateFAQ(id: string, faq: Partial<FAQ>): Observable<FAQ> {
    return this.cmsRepo.updateFAQ(id, faq);
  }

  deleteFAQ(id: string): Observable<{ status: string }> {
    return this.cmsRepo.deleteFAQ(id);
  }

  // Testimonials
  listTestimonials(params?: { is_active?: boolean }): Observable<{ results: Testimonial[]; count: number }> {
    return this.cmsRepo.listTestimonials(params);
  }

  createTestimonial(t: Partial<Testimonial>): Observable<Testimonial> {
    return this.cmsRepo.createTestimonial(t);
  }

  updateTestimonial(id: string, t: Partial<Testimonial>): Observable<Testimonial> {
    return this.cmsRepo.updateTestimonial(id, t);
  }

  deleteTestimonial(id: string): Observable<{ status: string }> {
    return this.cmsRepo.deleteTestimonial(id);
  }

  // Hero Section
  getHero(): Observable<HeroSection> {
    return this.cmsRepo.getHero();
  }

  updateHero(hero: Partial<HeroSection>): Observable<HeroSection> {
    return this.cmsRepo.updateHero(hero);
  }

  // Site Stats
  listStats(): Observable<{ results: SiteStat[]; count: number }> {
    return this.cmsRepo.listStats();
  }

  createStat(stat: Partial<SiteStat>): Observable<SiteStat> {
    return this.cmsRepo.createStat(stat);
  }

  updateStat(id: string, stat: Partial<SiteStat>): Observable<SiteStat> {
    return this.cmsRepo.updateStat(id, stat);
  }

  deleteStat(id: string): Observable<{ status: string }> {
    return this.cmsRepo.deleteStat(id);
  }

  // About Content
  getAbout(): Observable<AboutContent> {
    return this.cmsRepo.getAbout();
  }

  updateAbout(about: Partial<AboutContent>): Observable<AboutContent> {
    return this.cmsRepo.updateAbout(about);
  }
}
