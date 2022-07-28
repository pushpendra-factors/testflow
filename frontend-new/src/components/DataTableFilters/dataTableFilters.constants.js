export const EQUALITY_OPERATOR_KEYS = {
  EQUAL: 'equal',
  NOT_EQUAL: 'not-equal',
  CONTAINS: 'contains',
  DOES_NOT_CONTAIN: 'does-not-contain'
};

export const EQUALITY_OPERATOR_MENU = [
  {
    title: '=',
    key: EQUALITY_OPERATOR_KEYS.EQUAL
  },
  {
    title: '!=',
    key: EQUALITY_OPERATOR_KEYS.NOT_EQUAL
  },
  {
    title: 'contains',
    key: EQUALITY_OPERATOR_KEYS.CONTAINS
  },
  {
    title: 'does not contain',
    key: EQUALITY_OPERATOR_KEYS.DOES_NOT_CONTAIN
  }
];

export const CATEGORY_COMBINATION_OPERATOR_KEYS = {
  AND: 'AND',
  OR: 'OR'
};

export const CATEGORY_COMBINATION_OPERATOR_MENU = [
  {
    title: 'AND',
    key: CATEGORY_COMBINATION_OPERATOR_KEYS.AND
  },
  {
    title: 'OR',
    key: CATEGORY_COMBINATION_OPERATOR_KEYS.OR
  }
];

export const DEFAULT_CATEGORY_VALUE = {
  values: [],
  equalityOperator: EQUALITY_OPERATOR_KEYS.EQUAL
};

export const TEST_FILTER_OPTIONS = [
  {
    title: 'Campaign Name',
    key: 'Campaign_Name',
    options: ['option1', 'option2', 'option3', 'option4', 'option5', 'option6']
  },
  {
    title: 'AdGroup Name',
    key: 'AdGroup_Name',
    options: []
  },
  {
    title: 'Keyword Match Type',
    key: 'Keyword_Match_Type',
    options: []
  },
  {
    title: 'Keyword',
    key: 'Keyword',
    options: []
  }
];
