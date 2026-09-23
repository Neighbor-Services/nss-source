export interface LegalDocument {
  id: string;
  document_type: 'TERMS' | 'PRIVACY' | 'SAFETY' | 'COMMUNITY' | 'PAYMENT_TERMS' | 'PROMO_TERMS' | string;
  title: string;
  slug: string;
  content: string;
  version: string;
  is_active: boolean;
  published_at?: string | null;
  created_at?: string;
  updated_at?: string;
}
