export interface SubscriptionItem {
  id: string;
  userId: string;
  userName: string;
  userEmail: string;
  tier: string;
  interval: string;
  isActive: boolean;
  stripeSubscriptionId?: string;
  currentPeriodEnd?: string;
  createdAt: string;
}

export interface SubscriptionPlan {
  id?: string;
  name: string;
  tier: 'BASIC' | 'SILVER' | 'GOLD' | 'PLATINUM' | string;
  interval: 'month' | 'year' | string;
  description: string;
  price: number;
  currency: string;
  features?: string[];
  appleProductId?: string;
  googleProductId?: string;
  maxCatalogServices?: number;
  isActive?: boolean;
  displayOrder: number;
  created_at?: string;
}
