export interface FAQ {
  id: string;
  question: string;
  answer: string;
  order: number;
  category: string;
  is_active: boolean;
}

export interface Testimonial {
  id: string;
  name: string;
  role?: string;
  content: string;
  image?: string;
  rating: number;
  is_active: boolean;
  created_at?: string;
}

export interface HeroSection {
  id?: string;
  headline: string;
  subheadline: string;
  cta_text: string;
  cta_link: string;
  image?: string;
  is_active: boolean;
  updated_at?: string;
}

export interface SiteStat {
  id: string;
  label: string;
  value: string;
  order: number;
  is_active: boolean;
}

export interface AboutContent {
  id?: string;
  title: string;
  story_headline: string;
  story_text_1: string;
  story_text_2?: string;
  mission_text: string;
  vision_text: string;
  year_founded: string;
  cities_covered: string;
  is_active: boolean;
  updated_at?: string;
}
