import { Observable } from 'rxjs';
import { Category, CatalogService } from '../domain/entities/cms.model';

export abstract class CatalogRepository {
  abstract getCategories(): Observable<Category[]>;
  abstract getCatalogServices(categoryId?: string): Observable<CatalogService[]>;
}
