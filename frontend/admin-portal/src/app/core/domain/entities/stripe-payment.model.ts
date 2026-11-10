export interface StripePayment {
  id: string;
  payment_intent_id: string;
  charge_id?: string;
  customer_id?: string;
  user_id?: string;
  user_email?: string;
  user_name?: string;
  amount: number;
  amount_refunded: number;
  currency: string;
  status: 'succeeded' | 'requires_payment_method' | 'requires_action' | 'processing' | 'canceled' | 'refunded' | 'partially_refunded' | 'failed' | string;
  payment_method_type?: string;
  card_brand?: string;
  card_last4?: string;
  receipt_url?: string;
  description?: string;
  payment_type?: 'APPOINTMENT_ESCROW' | 'BACKGROUND_CHECK' | 'TIP' | 'SUBSCRIPTION' | 'DIRECT' | 'GENERAL' | string;
  appointment_id?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;

  user?: {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
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
    appointment_date?: string;
    seeker_id?: string;
    provider_id?: string;
    seeker?: {
      id: string;
      email: string;
      profile?: { first_name?: string; last_name?: string; phone?: string; profile_picture?: string };
    };
    provider?: {
      id: string;
      email: string;
      profile?: { first_name?: string; last_name?: string; phone?: string; profile_picture?: string };
    };
  };
}

export interface StripePaymentStats {
  total_volume: number;
  total_count: number;
  succeeded_count: number;
  refunded_count: number;
  total_refunded_amount: number;
  pending_count: number;
}

export interface StripePaymentDetail {
  payment: StripePayment;
  refund_requests: Array<{
    id: string;
    amount: number;
    currency: string;
    reason: string;
    status: string;
    stripe_refund_id?: string;
    admin_notes?: string;
    created_at: string;
    processed_at?: string;
  }>;
  stripe_dashboard_url: string;
  refundable_amount: number;
  is_refundable: boolean;
}

export interface StripePaymentListResponse {
  results: StripePayment[];
  count: number;
  page: number;
  page_size: number;
  stats?: StripePaymentStats;
}

export interface IssueRefundInput {
  amount?: number;
  reason?: string;
  notes?: string;
}

