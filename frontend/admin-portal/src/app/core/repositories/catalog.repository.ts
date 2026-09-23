import { Observable } from 'rxjs';
import { CategoryItem, CatalogServiceItem } from '../domain/entities/catalog.model';

export abstract class CatalogRepository {
  abstract listCategories(): Observable<CategoryItem[]>;
  abstract createCategory(data: Partial<CategoryItem>): Observable<CategoryItem>;
  abstract updateCategory(id: string, data: Partial<CategoryItem>): Observable<CategoryItem>;
  abstract deleteCategory(id: string): Observable<{ success: boolean }>;

  abstract listCatalogServices(categoryId?: string): Observable<CatalogServiceItem[]>;
  abstract createCatalogService(data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem>;
  abstract updateCatalogService(id: string, data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem>;
  abstract deleteCatalogService(id: string): Observable<{ success: boolean }>;
}
