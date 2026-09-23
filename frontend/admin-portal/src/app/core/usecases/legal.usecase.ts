import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { LegalRepository } from '../repositories/legal.repository';
import { LegalDocument } from '../domain/entities/legal.model';

@Injectable({
  providedIn: 'root'
})
export class LegalUseCase {
  constructor(private legalRepo: LegalRepository) {}

  listDocuments(params?: { type?: string; is_active?: boolean }): Observable<{ results: LegalDocument[]; count: number }> {
    return this.legalRepo.listDocuments(params);
  }

  getDocument(id: string): Observable<LegalDocument> {
    return this.legalRepo.getDocument(id);
  }

  createDocument(doc: Partial<LegalDocument>): Observable<LegalDocument> {
    return this.legalRepo.createDocument(doc);
  }

  updateDocument(id: string, doc: Partial<LegalDocument>): Observable<LegalDocument> {
    return this.legalRepo.updateDocument(id, doc);
  }

  deleteDocument(id: string): Observable<{ status: string }> {
    return this.legalRepo.deleteDocument(id);
  }
}
