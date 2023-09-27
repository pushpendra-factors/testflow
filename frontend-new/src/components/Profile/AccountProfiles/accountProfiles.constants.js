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

const INITIAL_ACCOUNT_STATE = ['All Accounts', 'All'];

export const INITIAL_FILTERS_STATE = {
  filters: [],
  eventsList: [],
  eventProp: 'any',
  account: INITIAL_ACCOUNT_STATE
};
