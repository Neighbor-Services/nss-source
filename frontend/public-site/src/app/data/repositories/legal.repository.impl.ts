import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { LegalRepository } from '../../core/repositories/legal.repository';
import { LegalDocument } from '../../core/domain/entities/cms.model';
import { API_CONFIG } from '../datasources/api.config';

@Injectable({
  providedIn: 'root'
})
export class LegalRepositoryImpl implements LegalRepository {
  constructor(private http: HttpClient) {}

  getLegalDocument(slug: 'terms' | 'privacy'): Observable<LegalDocument> {
    return this.http.get<any>(`${API_CONFIG.baseUrl}${API_CONFIG.endpoints.legal}`).pipe(
      map(res => {
        const doc = res?.[slug] || res?.documents?.[slug];
        if (doc) {
          return {
            title: doc.title || (slug === 'terms' ? 'Terms & Conditions' : 'Privacy Policy'),
            effectiveDate: doc.effective_date || '2026-09-01',
            lastUpdated: doc.last_updated || '2026-09-19',
            sections: doc.sections || []
          };
        }
        return {
          title: slug === 'terms' ? 'Terms & Conditions of Service' : 'Privacy Policy',
          effectiveDate: 'September 1, 2026',
          lastUpdated: 'September 19, 2026',
          sections: [
            {
              heading: '1. Introduction & Agreement',
              content: [
                'Welcome to Neighbor Service. By using our application and verified marketplace, you agree to these legal terms.'
              ]
            }
          ]
        };
      })
    );
  }
}
