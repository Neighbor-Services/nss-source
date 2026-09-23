import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { ReviewUseCase } from '../../../core/usecases/review.usecase';
import { AdminReview } from '../../../core/domain/entities/review.model';

@Component({
  selector: 'app-reviews',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './reviews.component.html',
  styleUrl: './reviews.component.css'
})
export class ReviewsComponent implements OnInit {
  reviews = signal<AdminReview[]>([]);
  totalCount = signal(0);
  
  // Filters
  selectedRating = signal<number>(0);
  selectedVisibility = signal<'ALL' | 'VISIBLE' | 'HIDDEN'>('ALL');
  
  // Pagination
  currentPage = signal(1);
  pageSize = signal(15);
  
  // Async states
  isLoading = signal(false);
  isActioning = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalReviews = computed(() => this.totalCount() || this.reviews().length);
  averageRating = computed(() => {
    const list = this.reviews();
    if (!list.length) return 5.0;
    const sum = list.reduce((acc, r) => acc + (r.rating || 0), 0);
    return (sum / list.length);
  });
  hiddenCount = computed(() => this.reviews().filter(r => r.is_hidden).length);
  oneStarCount = computed(() => this.reviews().filter(r => r.rating === 1).length);
  totalPages = computed(() => Math.max(1, Math.ceil(this.totalCount() / this.pageSize())));

  constructor(private reviewUC: ReviewUseCase) {}

  ngOnInit(): void {
    this.loadReviews();
  }

  loadReviews(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    let isHidden: boolean | undefined = undefined;
    if (this.selectedVisibility() === 'VISIBLE') isHidden = false;
    if (this.selectedVisibility() === 'HIDDEN') isHidden = true;

    this.reviewUC.listReviews({
      rating: this.selectedRating() > 0 ? this.selectedRating() : undefined,
      isHidden: isHidden,
      page: this.currentPage(),
      pageSize: this.pageSize()
    }).subscribe({
      next: (res) => {
        this.reviews.set(res.results || []);
        this.totalCount.set(res.count || 0);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load customer reviews.');
      }
    });
  }

  setRatingFilter(stars: number): void {
    this.selectedRating.set(stars);
    this.currentPage.set(1);
    this.loadReviews();
  }

  setVisibilityFilter(v: 'ALL' | 'VISIBLE' | 'HIDDEN'): void {
    this.selectedVisibility.set(v);
    this.currentPage.set(1);
    this.loadReviews();
  }

  toggleHide(review: AdminReview): void {
    const newHidden = !review.is_hidden;
    this.isActioning.set(true);

    this.reviewUC.toggleReviewVisibility(review.id, newHidden).subscribe({
      next: () => {
        this.isActioning.set(false);
        review.is_hidden = newHidden;
        this.successMessage.set(`Review visibility updated to ${newHidden ? 'HIDDEN (Moderated)' : 'PUBLIC'}.`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update review visibility.');
      }
    });
  }

  deleteReview(review: AdminReview): void {
    if (!confirm('Are you sure you want to permanently delete this review? This action is irreversible.')) return;

    this.isActioning.set(true);
    this.reviewUC.deleteReview(review.id).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set('Review permanently removed from database.');
        this.loadReviews();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to delete review.');
      }
    });
  }

  // Add Review Modal
  showCreateModal = signal(false);
  newReview = {
    provider: '',
    reviewer: '',
    rating: 5,
    comment: ''
  };

  openCreateModal(): void {
    this.newReview = {
      provider: '',
      reviewer: '',
      rating: 5,
      comment: ''
    };
    this.showCreateModal.set(true);
  }

  closeCreateModal(): void {
    this.showCreateModal.set(false);
  }

  submitReview(): void {
    if (!this.newReview.provider || !this.newReview.reviewer || !this.newReview.comment) {
      this.errorMessage.set('Provider ID, Reviewer ID, and Comment are required.');
      return;
    }

    this.isActioning.set(true);
    this.errorMessage.set(null);

    this.reviewUC.createReview({
      provider: this.newReview.provider.trim(),
      reviewer: this.newReview.reviewer.trim(),
      rating: Number(this.newReview.rating) || 5,
      comment: this.newReview.comment.trim()
    }).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.showCreateModal.set(false);
        this.successMessage.set('Review submitted successfully.');
        this.loadReviews();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || err?.error?.error || 'Failed to submit review.');
      }
    });
  }

  changePage(newPage: number): void {
    if (newPage >= 1 && newPage <= this.totalPages()) {
      this.currentPage.set(newPage);
      this.loadReviews();
    }
  }
}
