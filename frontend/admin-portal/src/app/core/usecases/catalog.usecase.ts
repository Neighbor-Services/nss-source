import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CatalogRepository } from '../repositories/catalog.repository';
import { CategoryItem, CatalogServiceItem } from '../domain/entities/catalog.model';

@Injectable({
  providedIn: 'root'
})
export class CatalogUseCase {
  constructor(private catalogRepo: CatalogRepository) {}

  listCategories(): Observable<CategoryItem[]> {
    return this.catalogRepo.listCategories();
  }

  createCategory(data: Partial<CategoryItem>): Observable<CategoryItem> {
    return this.catalogRepo.createCategory(data);
  }

  updateCategory(id: string, data: Partial<CategoryItem>): Observable<CategoryItem> {
    return this.catalogRepo.updateCategory(id, data);
  }

  deleteCategory(id: string): Observable<{ success: boolean }> {
    return this.catalogRepo.deleteCategory(id);
  }

  bulkDeleteCategories(ids: string[]): Observable<{ success: boolean; deleted_count?: number }> {
    return this.catalogRepo.bulkDeleteCategories(ids);
  }

  listCatalogServices(categoryId?: string): Observable<CatalogServiceItem[]> {
    return this.catalogRepo.listCatalogServices(categoryId);
  }

  createCatalogService(data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem> {
    return this.catalogRepo.createCatalogService(data);
  }

  updateCatalogService(id: string, data: Partial<CatalogServiceItem>): Observable<CatalogServiceItem> {
    return this.catalogRepo.updateCatalogService(id, data);
  }

  deleteCatalogService(id: string): Observable<{ success: boolean }> {
    return this.catalogRepo.deleteCatalogService(id);
  }

  bulkDeleteCatalogServices(ids: string[]): Observable<{ success: boolean; deleted_count?: number }> {
    return this.catalogRepo.bulkDeleteCatalogServices(ids);
  }

  importCatalogBatch(items: any[]): Observable<{ imported_count: number; message: string }> {
    return this.catalogRepo.importCatalogBatch(items);
  }

  exportCatalogBatch(): Observable<{ items: any[]; count: number }> {
    return this.catalogRepo.exportCatalogBatch();
  }

  getAISynonyms(): Observable<{ synonyms: Record<string, string[]>; count: number }> {
    return this.catalogRepo.getAISynonyms();
  }

  addAISynonym(key: string, synonyms: string[]): Observable<{ status: string }> {
    return this.catalogRepo.addAISynonym(key, synonyms);
  }

  deleteAISynonym(key: string): Observable<{ status: string }> {
    return this.catalogRepo.deleteAISynonym(key);
  }
}

