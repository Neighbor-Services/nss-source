import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { CmsRepository } from '../../core/repositories/cms.repository';
import { CMSContent } from '../../core/domain/entities/cms.model';
import { API_CONFIG } from '../datasources/api.config';

@Injectable({
  providedIn: 'root'
})
export class CmsRepositoryImpl implements CmsRepository {
  constructor(private http: HttpClient) {}

  getCMSContent(): Observable<CMSContent> {
    return this.http.get<any>(`${API_CONFIG.baseUrl}${API_CONFIG.endpoints.cms}`).pipe(
      map(res => {
        const stats = (res?.stats || []).map((s: any) => ({
          value: s.value || s.stat_value || '',
          label: s.label || s.title || '',
          description: s.description || ''
        }));

        const steps = (res?.steps || []).map((st: any, i: number) => ({
          number: st.step_number ? String(st.step_number).padStart(2, '0') : String(i + 1).padStart(2, '0'),
          title: st.title || '',
          description: st.description || '',
          icon: st.icon || 'star'
        }));

        const providerBenefits = (res?.features || []).map((f: any) => ({
          title: f.title || '',
          description: f.description || '',
          icon: f.icon || 'check'
        }));

        const testimonials = (res?.testimonials || []).map((t: any) => ({
          id: t.id?.toString() || '',
          quote: t.content || t.quote || '',
          author: t.author_name || t.author || '',
          role: t.author_role || t.role || '',
          location: t.location || '',
          rating: t.rating || 5,
          avatarUrl: t.avatar_url || t.author_avatar || ''
        }));

        const faqs = (res?.faqs || []).map((f: any) => ({
          id: f.id?.toString() || '',
          question: f.question || '',
          answer: f.answer || '',
          category: f.category || 'general'
        }));

        const categories = (res?.categories || []).map((c: any) => ({
          id: c.id?.toString() || '',
          name: c.name || '',
          slug: c.slug || '',
          description: c.description || '',
          icon: c.image || c.icon || 'star',
          serviceCount: c.service_count || 0
        }));

        const services = (res?.catalog_services || []).map((s: any) => ({
          id: s.id?.toString() || '',
          categoryId: s.category?.toString() || s.category_id?.toString() || '',
          categoryName: s.category_name || '',
          name: s.name || '',
          description: s.description || '',
          suggestedMinPrice: s.min_price || s.suggested_min_price || 0,
          suggestedMaxPrice: s.max_price || s.suggested_max_price || 0,
          pricingType: s.pricing_type || 'hourly',
          popular: s.is_popular ?? false
        }));

        const siteSettings = {
          siteName: res?.settings?.site_name || 'Neighbor Service',
          supportEmail: res?.settings?.support_email || 'support@neighborservice.com',
          supportPhone: res?.settings?.support_phone || '+1 (800) 555-0199',
          headquarters: res?.settings?.headquarters || 'Austin, Texas',
          supportHours: res?.settings?.support_hours || 'Mon - Sun: 7:00 AM – 9:00 PM CST',
          emergencyPhone: res?.settings?.emergency_phone || '+1 (800) 555-0199',
          officeAddress: res?.settings?.office_address || '100 Congress Ave, Suite 2000, Austin, TX 78701'
        };

        return {
          siteSettings,
          stats,
          steps,
          providerBenefits,
          testimonials,
          faqs,
          categories,
          services
        };
      })
    );
  }
}
