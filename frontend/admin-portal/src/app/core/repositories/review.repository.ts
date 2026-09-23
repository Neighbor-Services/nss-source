import { Observable } from 'rxjs';
import { AdminReview } from '../domain/entities/review.model';

export abstract class ReviewRepository {
  abstract listReviews(params?: { rating?: number; isHidden?: boolean; page?: number; pageSize?: number }): Observable<{ results: AdminReview[]; count: number }>;
  abstract createReview(data: { provider: string; reviewer: string; rating: number; comment: string }): Observable<AdminReview>;
  abstract toggleReviewVisibility(id: string, isHidden: boolean): Observable<{ status: string; is_hidden: boolean }>;
  abstract deleteReview(id: string): Observable<{ status: string }>;
}
