import { Observable } from 'rxjs';
import { LegalDocument } from '../domain/entities/legal.model';

export abstract class LegalRepository {
  abstract listDocuments(params?: { type?: string; is_active?: boolean }): Observable<{ results: LegalDocument[]; count: number }>;
  abstract getDocument(id: string): Observable<LegalDocument>;
  abstract createDocument(doc: Partial<LegalDocument>): Observable<LegalDocument>;
  abstract updateDocument(id: string, doc: Partial<LegalDocument>): Observable<LegalDocument>;
  abstract deleteDocument(id: string): Observable<{ status: string }>;
}
