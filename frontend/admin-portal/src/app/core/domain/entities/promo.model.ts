export interface AdminPromoCode {
  id: string;
  code: string;
  discount_type: 'PERCENTAGE' | 'FIXED' | string;
  discount_value: number;
  min_spend: number;
  max_discount: number;
  max_uses: number;
  uses_count: number;
  is_active: boolean;
  expires_at?: string;
  created_at: string;
  updated_at?: string;
}
