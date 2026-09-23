import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Category, CatalogService } from '../domain/entities/cms.model';
import { CatalogRepository } from '../repositories/catalog.repository';

@Injectable({
  providedIn: 'root'
})
export class GetCategoriesUseCase {
  constructor(private catalogRepo: CatalogRepository) {}

  execute(): Observable<Category[]> {
    return this.catalogRepo.getCategories();
  }

  getServices(categoryId?: string): Observable<CatalogService[]> {
    return this.catalogRepo.getCatalogServices(categoryId);
  }
}
