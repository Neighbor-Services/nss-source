import { Observable } from 'rxjs';
import { FAQ, Testimonial, HeroSection, SiteStat, AboutContent } from '../domain/entities/cms.model';

export abstract class CmsRepository {
  // FAQs
  abstract listFAQs(params?: { category?: string; is_active?: boolean }): Observable<{ results: FAQ[]; count: number }>;
  abstract createFAQ(faq: Partial<FAQ>): Observable<FAQ>;
  abstract updateFAQ(id: string, faq: Partial<FAQ>): Observable<FAQ>;
  abstract deleteFAQ(id: string): Observable<{ status: string }>;

  // Testimonials
  abstract listTestimonials(params?: { is_active?: boolean }): Observable<{ results: Testimonial[]; count: number }>;
  abstract createTestimonial(t: Partial<Testimonial>): Observable<Testimonial>;
  abstract updateTestimonial(id: string, t: Partial<Testimonial>): Observable<Testimonial>;
  abstract deleteTestimonial(id: string): Observable<{ status: string }>;

  // Hero Section
  abstract getHero(): Observable<HeroSection>;
  abstract updateHero(hero: Partial<HeroSection>): Observable<HeroSection>;

  // Site Stats
  abstract listStats(): Observable<{ results: SiteStat[]; count: number }>;
  abstract createStat(stat: Partial<SiteStat>): Observable<SiteStat>;
  abstract updateStat(id: string, stat: Partial<SiteStat>): Observable<SiteStat>;
  abstract deleteStat(id: string): Observable<{ status: string }>;

  // About Content
  abstract getAbout(): Observable<AboutContent>;
  abstract updateAbout(about: Partial<AboutContent>): Observable<AboutContent>;
}
