export interface ResolutionReport {
  role: 'seeker' | 'provider';
  issueType: 'quality' | 'no_show' | 'safety' | 'payment' | 'conduct' | 'other';
  bookingRef?: string;
  otherNeighbor?: string;
  description: string;
  expectedOutcome?: string;
}

export interface ResolutionResponse {
  status: 'success' | 'error';
  message: string;
  caseId?: string;
}
