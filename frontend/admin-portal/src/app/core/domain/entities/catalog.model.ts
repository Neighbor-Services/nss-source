export interface CategoryItem {
  id: string;
  name: string;
  slug: string;
  description: string;
  icon: string;
  isActive: boolean;
  order: number;
}

export interface CatalogServiceItem {
  id: string;
  categoryId: string;
  categoryName?: string;
  name: string;
  description: string;
  minPrice: number;
  maxPrice: number;
  pricingType: 'hourly' | 'fixed' | 'quote';
  isPopular: boolean;
  isActive: boolean;
  specialties?: string[];
  defaultServiceLocation?: string;
}

export interface CatalogImportItem {
  category_name: string;
  category_description?: string;
  service_name: string;
  service_description?: string;
  default_service_location?: string;
  specialties?: string[];
}

export interface CatalogExportItem {
  category_id: string;
  category_name: string;
  service_id: string;
  service_name: string;
  service_description: string;
  default_service_location: string;
  specialties: string[];
}

export interface AISynonymsResponse {
  synonyms: Record<string, string[]>;
  count: number;
}

