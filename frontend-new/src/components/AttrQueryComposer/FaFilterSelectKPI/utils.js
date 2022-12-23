

export const  DEFAULT_OPERATOR_PROPS = {
    "categorical": [
      '=',
      '!=',
      'contains',
      'does not contain'
    ],
    "numerical": [
      '=',
      '!=',
      '<',
      '<=',
      '>',
      '>='
    ],
    "datetime": [
      'between',
      'not between',
      "in the last",
      "not in the last",
      "before",
      "since"
    ]
};

export const dateTimeSelect = new Map(
  [['Days','days'],['Weeks','week'],['Months','month'],['Quarters','quarter'],['days','Days'],['week','Weeks'],['month','Months'],['quarter','Quarters']]
);