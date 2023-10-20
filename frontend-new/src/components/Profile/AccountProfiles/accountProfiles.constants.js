import { GROUP_NAME_DOMAINS } from "Components/GlobalFilter/FilterWrapper/utils";

export const eventMenuList = {
  any: {
    label: 'Any Event',
    key: 'any'
  },
  all: {
    label: 'All Events',
    key: 'all'
  },
  each: {
    label: 'Each Event',
    key: 'each'
  }
};

export const moreActionsMode = {
  DELETE: 'DELETE',
  RENAME: 'RENAME'
};

const INITIAL_ACCOUNT_STATE = ['All Accounts', GROUP_NAME_DOMAINS];

export const INITIAL_FILTERS_STATE = {
  filters: [],
  eventsList: [],
  eventProp: 'any',
  account: INITIAL_ACCOUNT_STATE
};
