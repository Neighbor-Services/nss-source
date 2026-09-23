export interface ContactMessage {
  firstName: string;
  lastName?: string;
  email: string;
  inquiryType: 'general' | 'safety' | 'provider' | 'press' | 'partnership';
  message: string;
}

export interface ContactResponse {
  status: 'success' | 'error';
  message: string;
}
