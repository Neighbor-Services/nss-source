import { Observable } from 'rxjs';
import { CategoryItem, CatalogServiceItem } from '../domain/entities/catalog.model';

export abstract class CatalogRepository {
  abstract listCategories(): Observable<CategoryItem[]>;
  abstract createCategory(data: Partial<CategoryItem>): Observable<CategoryItem>;
  abstract updateCategory(id: string, data: Partial<CategoryItem>): Observable<CategoryItem>;
  abstract deleteCategory(id: string): Observable<{ success: boolean }>;
  abstract bulkDeleteCategories(ids: string[]): Observable<{ success: boolean; deleted_count?: number }>;

  abstract listCatalogServices(categoryId?: string): Observable<CatalogServiceItem[]>;
  abstract createCatalogService(data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem>;
  abstract updateCatalogService(id: string, data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem>;
  abstract deleteCatalogService(id: string): Observable<{ success: boolean }>;
  abstract bulkDeleteCatalogServices(ids: string[]): Observable<{ success: boolean; deleted_count?: number }>;

  abstract importCatalogBatch(items: any[]): Observable<{ imported_count: number; message: string }>;
  abstract exportCatalogBatch(): Observable<{ items: any[]; count: number }>;
  abstract getAISynonyms(): Observable<{ synonyms: Record<string, string[]>; count: number }>;
  abstract addAISynonym(key: string, synonyms: string[]): Observable<{ status: string }>;
  abstract deleteAISynonym(key: string): Observable<{ status: string }>;
}

