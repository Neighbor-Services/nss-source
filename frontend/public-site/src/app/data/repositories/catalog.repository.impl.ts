import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { CatalogRepository } from '../../core/repositories/catalog.repository';
import { Category, CatalogService } from '../../core/domain/entities/cms.model';
import { API_CONFIG } from '../datasources/api.config';

@Injectable({
  providedIn: 'root'
})
export class CatalogRepositoryImpl implements CatalogRepository {
  constructor(private http: HttpClient) {}

  getCategories(): Observable<Category[]> {
    return this.http.get<any>(`${API_CONFIG.baseUrl}${API_CONFIG.endpoints.categories}`).pipe(
      map(res => {
        const list = Array.isArray(res) ? res : (res?.results || []);
        return list.map((item: any) => ({
          id: item.id?.toString() || item.uuid || '',
          name: item.name || '',
          slug: item.slug || '',
          description: item.description || '',
          icon: item.image || item.icon || 'star',
          serviceCount: item.service_count || 0
        }));
      })
    );
  }

  getCatalogServices(categoryId?: string): Observable<CatalogService[]> {
    const url = categoryId
      ? `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.catalogServices}?category_id=${categoryId}`
      : `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.catalogServices}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const list = Array.isArray(res) ? res : (res?.results || []);
        return list.map((item: any) => {
          let catId = '';
          if (item.category_id) {
            catId = item.category_id.toString();
          } else if (item.category && typeof item.category === 'object') {
            catId = (item.category.id || item.category.uuid || '').toString();
          } else if (item.category) {
            catId = item.category.toString();
          }

          let catName = 'General';
          if (item.category_name) {
            catName = item.category_name;
          } else if (item.category && typeof item.category === 'object' && item.category.name) {
            catName = item.category.name;
          }

          return {
            id: (item.id || item.uuid || '').toString(),
            categoryId: catId,
            categoryName: catName,
            name: item.name || '',
            description: item.description || '',
            suggestedMinPrice: item.min_price || item.suggested_min_price || item.base_price || 0,
            suggestedMaxPrice: item.max_price || item.suggested_max_price || (item.base_price ? item.base_price * 1.5 : 0),
            pricingType: item.pricing_type || 'hourly',
            popular: item.is_popular ?? false
          };
        });
      })
    );
  }
}
