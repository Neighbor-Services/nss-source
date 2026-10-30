export interface RefundRequestItem {
  id: string;
  user_id: string;
  appointment_id?: string;
  payment_intent_id: string;
  amount: number;
  currency: string;
  reason: string;
  status: 'PENDING' | 'APPROVED' | 'PROCESSED' | 'REJECTED' | 'FAILED' | string;
  stripe_refund_id?: string;
  admin_id?: string;
  admin_notes?: string;
  created_at: string;
  updated_at?: string;
  user?: {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    phone?: string;
    profile?: {
      first_name?: string;
      last_name?: string;
      phone?: string;
      profile_picture?: string;
    };
  };
  appointment?: {
    id: string;
    title?: string;
    total_price?: number;
    status?: string;
    payment_mode?: string;
  };
}

