export interface AdminAppointment {
  id: string;
  seeker: string;
  provider: string;
  title: string;
  description: string;
  appointment_date?: string;
  status: 'SCHEDULED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED' | 'DISPUTED' | string;
  payment_mode: string;
  is_consultation: boolean;
  is_funded: boolean;
  total_price: number;
  created_at: string;
  updated_at?: string;
  seeker_details?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
    phone?: string;
  };
  provider_details?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
    phone?: string;
  };
  service_request_details?: {
    id: string;
    title: string;
    budget?: number;
  };
}
