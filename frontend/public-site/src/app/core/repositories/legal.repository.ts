import { Observable } from 'rxjs';
import { LegalDocument } from '../domain/entities/cms.model';

export abstract class LegalRepository {
  abstract getLegalDocument(slug: 'terms' | 'privacy'): Observable<LegalDocument>;
}
