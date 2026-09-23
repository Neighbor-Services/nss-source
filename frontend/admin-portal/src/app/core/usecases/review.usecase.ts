import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ReviewRepository } from '../repositories/review.repository';
import { AdminReview } from '../domain/entities/review.model';

@Injectable({
  providedIn: 'root'
})
export class ReviewUseCase {
  constructor(private reviewRepo: ReviewRepository) {}

  listReviews(params?: { rating?: number; isHidden?: boolean; page?: number; pageSize?: number }): Observable<{ results: AdminReview[]; count: number }> {
    return this.reviewRepo.listReviews(params);
  }

  createReview(data: { provider: string; reviewer: string; rating: number; comment: string }): Observable<AdminReview> {
    return this.reviewRepo.createReview(data);
  }

  toggleReviewVisibility(id: string, isHidden: boolean): Observable<{ status: string; is_hidden: boolean }> {
    return this.reviewRepo.toggleReviewVisibility(id, isHidden);
  }

  deleteReview(id: string): Observable<{ status: string }> {
    return this.reviewRepo.deleteReview(id);
  }
}
