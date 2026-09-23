import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CmsRepository } from '../../core/repositories/cms.repository';
import { FAQ, Testimonial, HeroSection, SiteStat, AboutContent } from '../../core/domain/entities/cms.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class CmsRepositoryImpl implements CmsRepository {
  private readonly faqsUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.cmsFaqs;
  private readonly testimonialsUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.cmsTestimonials;
  private readonly heroUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.cmsHero;
  private readonly statsUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.cmsStats;
  private readonly aboutUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.cmsAbout;

  constructor(private http: HttpClient) {}

  // FAQs
  listFAQs(params?: { category?: string; is_active?: boolean }): Observable<{ results: FAQ[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.category) httpParams = httpParams.set('category', params.category);
    if (params?.is_active !== undefined) httpParams = httpParams.set('is_active', params.is_active.toString());

    return this.http.get<{ results: FAQ[]; count: number }>(this.faqsUrl, { params: httpParams });
  }

  createFAQ(faq: Partial<FAQ>): Observable<FAQ> {
    return this.http.post<FAQ>(this.faqsUrl, faq);
  }

  updateFAQ(id: string, faq: Partial<FAQ>): Observable<FAQ> {
    return this.http.put<FAQ>(`${this.faqsUrl}/${id}`, faq);
  }

  deleteFAQ(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.faqsUrl}/${id}`);
  }

  // Testimonials
  listTestimonials(params?: { is_active?: boolean }): Observable<{ results: Testimonial[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.is_active !== undefined) httpParams = httpParams.set('is_active', params.is_active.toString());

    return this.http.get<{ results: Testimonial[]; count: number }>(this.testimonialsUrl, { params: httpParams });
  }

  createTestimonial(t: Partial<Testimonial>): Observable<Testimonial> {
    return this.http.post<Testimonial>(this.testimonialsUrl, t);
  }

  updateTestimonial(id: string, t: Partial<Testimonial>): Observable<Testimonial> {
    return this.http.put<Testimonial>(`${this.testimonialsUrl}/${id}`, t);
  }

  deleteTestimonial(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.testimonialsUrl}/${id}`);
  }

  // Hero Section
  getHero(): Observable<HeroSection> {
    return this.http.get<HeroSection>(this.heroUrl);
  }

  updateHero(hero: Partial<HeroSection>): Observable<HeroSection> {
    return this.http.put<HeroSection>(this.heroUrl, hero);
  }

  // Site Stats
  listStats(): Observable<{ results: SiteStat[]; count: number }> {
    return this.http.get<{ results: SiteStat[]; count: number }>(this.statsUrl);
  }

  createStat(stat: Partial<SiteStat>): Observable<SiteStat> {
    return this.http.post<SiteStat>(this.statsUrl, stat);
  }

  updateStat(id: string, stat: Partial<SiteStat>): Observable<SiteStat> {
    return this.http.put<SiteStat>(`${this.statsUrl}/${id}`, stat);
  }

  deleteStat(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.statsUrl}/${id}`);
  }

  // About Content
  getAbout(): Observable<AboutContent> {
    return this.http.get<AboutContent>(this.aboutUrl);
  }

  updateAbout(about: Partial<AboutContent>): Observable<AboutContent> {
    return this.http.put<AboutContent>(this.aboutUrl, about);
  }
}
