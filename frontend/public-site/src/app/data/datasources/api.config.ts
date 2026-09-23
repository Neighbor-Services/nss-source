function getApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    if ((window as any).__API_BASE_URL__) {
      return (window as any).__API_BASE_URL__;
    }
    const hostname = window.location.hostname;
    if (hostname.includes('staging')) {
      return 'https://staging-api.neighborservice.com/api/v1';
    }
    if (hostname.includes('neighborservice.com')) {
      return 'https://api.neighborservice.com/api/v1';
    }
  }
  return 'http://localhost:8000/api/v1';
}

export const API_CONFIG = {
  get baseUrl(): string {
    return getApiBaseUrl();
  },
  endpoints: {
    cms: '/public/cms/',
    contact: '/public/contact/',
    resolution: '/public/resolution/',
    categories: '/services/categories/',
    catalogServices: '/services/catalog-services/',
    legal: '/accounts/legal/'
  }
};
