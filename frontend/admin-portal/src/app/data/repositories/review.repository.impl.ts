import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ReviewRepository } from '../../core/repositories/review.repository';
import { AdminReview } from '../../core/domain/entities/review.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class ReviewRepositoryImpl implements ReviewRepository {
  private readonly baseUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.reviews;

  constructor(private http: HttpClient) {}

  listReviews(params?: { rating?: number; isHidden?: boolean; page?: number; pageSize?: number }): Observable<{ results: AdminReview[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.rating) httpParams = httpParams.set('rating', params.rating.toString());
    if (params?.isHidden !== undefined) httpParams = httpParams.set('is_hidden', params.isHidden.toString());
    if (params?.page) httpParams = httpParams.set('page', params.page.toString());
    if (params?.pageSize) httpParams = httpParams.set('page_size', params.pageSize.toString());

    return this.http.get<{ results: AdminReview[]; count: number }>(this.baseUrl, { params: httpParams });
  }

  createReview(data: { provider: string; reviewer: string; rating: number; comment: string }): Observable<AdminReview> {
    return this.http.post<AdminReview>(this.baseUrl, {
      provider: data.provider,
      reviewer: data.reviewer,
      rating: data.rating,
      comment: data.comment
    });
  }

  toggleReviewVisibility(id: string, isHidden: boolean): Observable<{ status: string; is_hidden: boolean }> {
    return this.http.post<{ status: string; is_hidden: boolean }>(`${this.baseUrl}/${id}/toggle-hide`, { is_hidden: isHidden });
  }

  deleteReview(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.baseUrl}/${id}`);
  }
}
