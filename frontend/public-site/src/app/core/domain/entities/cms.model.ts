export interface SiteSettings {
  siteName: string;
  supportEmail: string;
  supportPhone: string;
  headquarters: string;
  supportHours: string;
  emergencyPhone: string;
  officeAddress: string;
}

export interface SafetyStat {
  value: string;
  label: string;
  description?: string;
}

export interface HowItWorksStep {
  number: string;
  title: string;
  description: string;
  icon?: string;
}

export interface ProviderBenefit {
  title: string;
  description: string;
  icon?: string;
}

export interface Category {
  id: string;
  name: string;
  slug?: string;
  description: string;
  icon?: string;
  iconName?: string;
  serviceCount?: number;
  featured?: boolean;
}

export interface CatalogService {
  id: string;
  categoryId: string;
  categoryName?: string;
  name: string;
  description: string;
  suggestedMinPrice?: number;
  suggestedMaxPrice?: number;
  pricingType?: 'hourly' | 'fixed' | 'quote';
  popular?: boolean;
}

export interface Testimonial {
  id: string;
  quote: string;
  author: string;
  role: string;
  location: string;
  rating: number;
  avatarUrl?: string;
}

export interface FAQItem {
  id: string;
  question: string;
  answer: string;
  category?: 'general' | 'safety' | 'providers' | 'payment';
}

export interface CMSContent {
  siteSettings: SiteSettings;
  stats: SafetyStat[];
  steps: HowItWorksStep[];
  providerBenefits: ProviderBenefit[];
  testimonials: Testimonial[];
  faqs: FAQItem[];
  categories: Category[];
  services: CatalogService[];
}

export interface LegalDocument {
  title: string;
  effectiveDate: string;
  lastUpdated: string;
  sections: {
    heading: string;
    content: string[];
  }[];
}
