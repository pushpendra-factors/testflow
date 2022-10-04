export const EQUALITY_OPERATOR_KEYS = {
  EQUAL: 'equal',
  NOT_EQUAL: 'not-equal',
  CONTAINS: 'contains',
  DOES_NOT_CONTAIN: 'does-not-contain',
  GREATER_THAN_OR_EQUAL_TO: 'greater-than-or-equal-to',
  LESS_THAN_OR_EQUAL_TO: 'less-than-or-equal-to'
};

export const EQUALITY_OPERATOR_MENU = [
  {
    title: '=',
    key: EQUALITY_OPERATOR_KEYS.EQUAL,
    valueType: ['numerical', 'categorical', 'percentage']
  },
  {
    title: '!=',
    key: EQUALITY_OPERATOR_KEYS.NOT_EQUAL,
    valueType: ['numerical', 'categorical', 'percentage']
  },
  {
    title: 'contains',
    key: EQUALITY_OPERATOR_KEYS.CONTAINS,
    valueType: ['categorical']
  },
  {
    title: 'does not contain',
    key: EQUALITY_OPERATOR_KEYS.DOES_NOT_CONTAIN,
    valueType: ['categorical']
  },
  {
    title: '>=',
    key: EQUALITY_OPERATOR_KEYS.GREATER_THAN_OR_EQUAL_TO,
    valueType: ['numerical', 'percentage']
  },
  {
    title: '<=',
    key: EQUALITY_OPERATOR_KEYS.LESS_THAN_OR_EQUAL_TO,
    valueType: ['numerical', 'percentage']
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
  },
  {
    title: 'Impressions',
    key: 'Impressions',
    options: [],
    valueType: 'numerical'
  },
  {
    title: 'Clicks',
    key: 'Clicks',
    options: [],
    valueType: 'numerical'
  },
  {
    title: 'CTR (%)',
    key: 'CTR (%)',
    options: [],
    valueType: 'percentage'
  }
];
