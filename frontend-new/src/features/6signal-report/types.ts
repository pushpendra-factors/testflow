export interface APIResponse {
  status: number;
  ok: boolean;
  data?: any;
}

export type WeekStartEnd = {
  from: number;
  to: number;
  formattedRange: string;
  formattedRangeOption: string;
  isSaved: boolean;
};

export interface Query {
  six_signal_query_group: SixSignalQueryGroup[];
}

export interface ResultGroup {
  headers: string[];
  rows: string[][];
  query: null;
}
export interface SixSignalQueryGroup {
  fr: number;
  to: number;
  tz: string;
}
export interface ReportApiResponseData {
  result_group: ResultGroup[];
  query: Query;
  cache_meta?: CacheMeta;
  is_shareable: boolean;
}

export interface ReportIndex {
  1: ReportApiResponseData;
}

export interface ReportApiResponse extends APIResponse {
  data: ReportIndex;
}
export interface CacheMeta {
  from: number;
  last_computed_at: number;
  preset: string;
  refreshed_at: number;
  timezone: string;
  to: number;
}

export interface QuickFilterProps {
  filters: Filters[];
  onFilterChange: (id: string) => void;
  selectedFilter?: string;
}

export interface Filters {
  id: string;
  label: string;
}

export interface ReportTableProps {
  data?: ResultGroup | null;
  selectedCampaigns: string[];
  selectedChannel: string;
  isSixSignalActivated: boolean;
  dataSelected: string;
}

export interface StringObject {
  [key: string]: string;
}

export interface ShareApiData {
  project_id: number;
  query_id: string;
  route_version: string;
}
export interface ShareApiResponse extends APIResponse {
  data: ShareApiData;
}

export interface ShareData extends ShareApiData {
  dateSelected: string;
  publicUrl: string;
  from?: number;
  to?: number;
  timezone?: string;
  domain?: string;
  projectId: string;
}

export interface SavedReportDatesApiResponse extends APIResponse {
  data: string[];
}
