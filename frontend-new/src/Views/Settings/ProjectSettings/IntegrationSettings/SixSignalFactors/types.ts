export type FeatureModes = 'view' | 'edit' | 'configure';
export type EnrichTypes = 'include' | 'exclude';
export type EnrichPageUrlType = 'contains' | 'equals';

export interface EnrichPageData {
  value: string;
  type: EnrichPageUrlType;
}

export interface EnrichCountryData {
  value: string;
  type: 'equals';
}

export interface SixSignalConfigType {
  api_limit?: number;
  country_include?: EnrichCountryData[];
  country_exclude?: EnrichCountryData[];
  pages_include?: EnrichPageData[];
  pages_exclude?: EnrichPageData[];
}
