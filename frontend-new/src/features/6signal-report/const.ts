import { Filters } from './types';

export const CHANNEL_KEY = 'Channel';
export const CAMPAIGN_KEY = 'Campaign';
export const SESSION_SPENT_TIME = 'Time_Spent';
export const PAGE_COUNT_KEY = 'Page_Count';
export const PAGE_URL_KEY = 'Page_Seen';
export const COUNTRY_KEY = 'Country';
export const COMPANY_KEY = 'Company';
export const KEY_LABELS = {
  [COMPANY_KEY]: 'Company Name',
  [COUNTRY_KEY]: 'Country',
  [PAGE_URL_KEY]: 'Initial Page seen',
  [CAMPAIGN_KEY]: 'Campaign',
  [SESSION_SPENT_TIME]: 'Time spent',
  [PAGE_COUNT_KEY]: 'Pages viewed',
  [CHANNEL_KEY]: 'Channel'
};

export const CHANNEL_QUICK_FILTERS: Filters[] = [
  {
    id: 'all',
    label: 'All'
  },
  {
    id: 'Paid Search',
    label: 'Paid Search'
  },
  {
    id: 'Paid Social',
    label: 'Paid Social'
  },
  {
    id: 'Organic Search',
    label: 'Organic'
  },
  {
    id: 'Direct',
    label: 'Direct'
  }
];

export const SHARE_QUERY_PARAMS = {
  queryId: 'queryId',
  projectId: 'pId',
  routeVersion: 'version'
};
