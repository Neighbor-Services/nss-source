export interface ContactMessage {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  inquiry_type: string;
  message: string;
  created_at: string;
  is_resolved: boolean;
}

export interface ResolutionReport {
  id: string;
  role: string;
  issue_type: string;
  booking_ref?: string;
  other_neighbor?: string;
  date_of_service?: string;
  description: string;
  expected_outcome: string;
  created_at: string;
  is_reviewed: boolean;
}
