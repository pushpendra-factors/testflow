import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';

export const eventMenuList = {
  any: {
    label: 'Any Event',
    key: 'any'
  },
  all: {
    label: 'All Events',
    key: 'all'
  }
};

export const eventTimelineMenuList = {
  7: {
    label: 'Last 7 days',
    key: '7'
  },
  14: {
    label: 'Last 14 days',
    key: '14'
  },
  30: {
    label: 'Last 30 days',
    key: '30'
  },
  60: {
    label: 'Last 60 days',
    key: '60'
  },
  90: {
    label: 'Last 90 days',
    key: '90'
  }
};

export const moreActionsMode = {
  DELETE: 'DELETE',
  RENAME: 'RENAME'
};

const INITIAL_ACCOUNT_STATE = ['All Accounts', GROUP_NAME_DOMAINS];

const EVENT_TIMELINE_DEFAULT_VALUE = '7';

export const INITIAL_FILTERS_STATE = {
  filters: [],
  eventsList: [],
  eventProp: 'any',
  account: INITIAL_ACCOUNT_STATE,
  secondaryFilters: [],
  eventTimeline: EVENT_TIMELINE_DEFAULT_VALUE
};

export const INITIAL_USER_PROFILES_FILTERS_STATE = {
  filters: [],
  eventsList: [],
  eventProp: 'any',
  account: ['All People', 'users'],
  secondaryFilters: [],
  eventTimeline: EVENT_TIMELINE_DEFAULT_VALUE
};
