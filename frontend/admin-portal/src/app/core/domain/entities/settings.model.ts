export interface PlatformSettings {
  backgroundCheckPaymentMode: 'PLATFORM_PAYS' | 'PROVIDER_PAYS';
  backgroundCheckFee: number;
  broadcastRadiusKm: number;
  matchRadiusKm: number;
}
