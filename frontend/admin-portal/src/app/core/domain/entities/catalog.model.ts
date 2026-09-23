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
}
