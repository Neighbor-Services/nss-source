import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { LegalDocument } from '../domain/entities/cms.model';
import { LegalRepository } from '../repositories/legal.repository';

@Injectable({
  providedIn: 'root'
})
export class GetLegalDocumentUseCase {
  constructor(private legalRepo: LegalRepository) {}

  execute(slug: 'terms' | 'privacy'): Observable<LegalDocument> {
    return this.legalRepo.getLegalDocument(slug);
  }
}
