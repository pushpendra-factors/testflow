import { Filters } from './types';

export const CHANNEL_KEY = 'channel';
export const CAMPAIGN_KEY = 'campaign';
export const SESSION_SPENT_TIME = 'time_spent';
export const PAGE_COUNT_KEY = 'page_count';
export const PAGE_URL_KEY = 'page_seen';
export const COUNTRY_KEY = 'country';
export const COMPANY_KEY = 'company';
export const INDUSTRY_KEY = 'industry';
export const EMP_RANGE_KEY = 'emp_range';
export const REVENUE_RANGE_KEY = 'revenue_range';
export const DOMAIN_KEY = 'domain';
export const KEY_LABELS = {
  [COMPANY_KEY]: 'Company Name',
  [COUNTRY_KEY]: 'Country',
  [PAGE_URL_KEY]: 'Initial Page seen',
  [CAMPAIGN_KEY]: 'Campaign',
  [SESSION_SPENT_TIME]: 'Time spent',
  [PAGE_COUNT_KEY]: 'Pages viewed',
  [CHANNEL_KEY]: 'Channel',
  [INDUSTRY_KEY]: 'Industry',
  [EMP_RANGE_KEY]: 'Employee Range',
  [REVENUE_RANGE_KEY]: 'Revenue Range',
  [DOMAIN_KEY]: 'Domain'
};

export const DEFAULT_COLUMNS = [
  COMPANY_KEY,
  CHANNEL_KEY,
  COUNTRY_KEY,
  INDUSTRY_KEY,
  REVENUE_RANGE_KEY,
  EMP_RANGE_KEY,
  SESSION_SPENT_TIME
];

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
