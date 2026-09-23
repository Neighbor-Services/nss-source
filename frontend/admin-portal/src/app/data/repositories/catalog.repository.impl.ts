import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { CatalogRepository } from '../../core/repositories/catalog.repository';
import { CategoryItem, CatalogServiceItem } from '../../core/domain/entities/catalog.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class CatalogRepositoryImpl implements CatalogRepository {
  constructor(private http: HttpClient) {}

  listCategories(): Observable<CategoryItem[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.categories}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((c: any) => ({
          id: c.id?.toString() || '',
          name: c.name || '',
          slug: c.slug || '',
          description: c.description || '',
          icon: c.icon || c.image || 'star',
          isActive: c.is_active ?? true,
          order: c.order || 1
        }));
      })
    );
  }

  createCategory(data: Partial<CategoryItem>): Observable<CategoryItem> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.categories}`, {
      name: data.name,
      description: data.description || '',
      image: data.icon || ''
    }).pipe(
      map(res => ({
        id: res.id?.toString() || '',
        name: res.name || data.name || '',
        slug: res.slug || data.slug || '',
        description: res.description || data.description || '',
        icon: res.image || data.icon || 'star',
        isActive: res.is_active ?? true,
        order: res.order || 1
      }))
    );
  }

  updateCategory(id: string, data: Partial<CategoryItem>): Observable<CategoryItem> {
    return this.http.patch<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.categories}/${id}`, {
      name: data.name,
      description: data.description,
      image: data.icon
    }).pipe(
      map(res => ({
        id: res.id?.toString() || id,
        name: res.name || data.name || '',
        slug: res.slug || data.slug || '',
        description: res.description || data.description || '',
        icon: res.image || data.icon || 'star',
        isActive: res.is_active ?? true,
        order: res.order || 1
      }))
    );
  }

  deleteCategory(id: string): Observable<{ success: boolean }> {
    return this.http.delete<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.categories}/${id}`).pipe(
      map(() => ({ success: true }))
    );
  }

  listCatalogServices(categoryId?: string): Observable<CatalogServiceItem[]> {
    const url = categoryId
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.catalogServices}?category_id=${categoryId}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.catalogServices}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((s: any) => ({
          id: s.id?.toString() || '',
          categoryId: s.category?.toString() || s.category_id?.toString() || '',
          categoryName: s.category_name || s.category?.name || 'General',
          name: s.name || '',
          description: s.description || '',
          minPrice: s.min_price || s.suggested_min_price || 0,
          maxPrice: s.max_price || s.suggested_max_price || 0,
          pricingType: s.pricing_type || 'hourly',
          isPopular: s.is_popular ?? false,
          isActive: s.is_active ?? true
        }));
      })
    );
  }

  createCatalogService(data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.catalogServices}`, {
      category_id: data.categoryId,
      name: data.name,
      description: data.description || '',
      base_price: data.minPrice || 0,
      min_price: data.minPrice || 0,
      max_price: data.maxPrice || 0,
      pricing_type: data.pricingType || 'hourly',
      is_popular: data.isPopular || false
    }).pipe(
      map(res => ({
        id: res.id?.toString() || '',
        categoryId: res.category_id || data.categoryId || '',
        categoryName: res.category_name || data.categoryName || '',
        name: res.name || data.name || '',
        description: res.description || data.description || '',
        minPrice: res.base_price || res.min_price || data.minPrice || 0,
        maxPrice: res.max_price || data.maxPrice || 0,
        pricingType: res.pricing_type || data.pricingType || 'hourly',
        isPopular: res.is_popular ?? false,
        isActive: res.is_active ?? true
      }))
    );
  }

  updateCatalogService(id: string, data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem> {
    return this.http.patch<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.catalogServices}/${id}`, {
      category_id: data.categoryId,
      name: data.name,
      description: data.description,
      base_price: data.minPrice || 0,
      min_price: data.minPrice,
      max_price: data.maxPrice,
      pricing_type: data.pricingType,
      is_popular: data.isPopular
    }).pipe(
      map(res => ({
        id: res.id?.toString() || id,
        categoryId: res.category_id || data.categoryId || '',
        categoryName: res.category_name || data.categoryName || '',
        name: res.name || data.name || '',
        description: res.description || data.description || '',
        minPrice: res.base_price || res.min_price || data.minPrice || 0,
        maxPrice: res.max_price || data.maxPrice || 0,
        pricingType: res.pricing_type || data.pricingType || 'hourly',
        isPopular: res.is_popular ?? false,
        isActive: res.is_active ?? true
      }))
    );
  }

  deleteCatalogService(id: string): Observable<{ success: boolean }> {
    return this.http.delete<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.catalogServices}/${id}`).pipe(
      map(() => ({ success: true }))
    );
  }
}
