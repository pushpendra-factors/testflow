export interface DataMap {
  [key: string]: number;
}

export interface TopPages {
  page_url: string;
  views: number;
  users_count: number;
  total_time: number;
  avg_page_scroll_percent: number;
}

export interface TopUsers {
  name: string;
  views: number;
  users_count: number;
  active_time: number;
  pages_count: number;
}

export type Overview = {
  temperature: number;
  engagement: string;
  users_count: number;
  time_active: number;
  scores_list: DataMap;
  top_pages: TopPages[];
  top_users: TopUsers[];
};

export interface AccountOverviewProps {
  overview: Overview;
  loading: boolean;
}

export interface CustomStyles {
  '--bg-color': string;
}
