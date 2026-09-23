export interface AdminReview {
  id: string;
  provider: string;
  reviewer: string;
  rating: number;
  comment: string;
  is_hidden: boolean;
  created_at: string;
  updated_at?: string;
  provider_details?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
  };
  reviewer_details?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
  };
}
