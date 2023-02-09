import { OPERATORS } from './constants';

export const operatorMap = {
  [OPERATORS['equalTo']]: 'equals',
  [OPERATORS['notEqualTo']]: 'notEqual',
  [OPERATORS['contain']]: 'contains',
  [OPERATORS['doesNotContain']]: 'notContains',
  [OPERATORS['lesserThan']]: 'lesserThan',
  [OPERATORS['lesserThanOrEqual']]: 'lesserThanOrEqual',
  [OPERATORS['greaterThan']]: 'greaterThan',
  [OPERATORS['greaterThanOrEqual']]: 'greaterThanOrEqual',
  [OPERATORS['between']]: 'between',
  [OPERATORS['notBetween']]: 'notInBetween',
  [OPERATORS['inThePrevious']]: 'inLast',
  [OPERATORS['notInThePrevious']]: 'notInLast',
  [OPERATORS['inTheCurrent']]: 'inCurrent',
  [OPERATORS['notInTheCurrent']]: 'notInCurrent',
  [OPERATORS['before']]: 'before',
  [OPERATORS['since']]: 'since',
  [OPERATORS['isKnown']]: 'notEqual',
  [OPERATORS['isUnknown']]: 'equals'
};

export const reverseOperatorMap = {
  equals: OPERATORS['equalTo'],
  notEqual: OPERATORS['notEqualTo'],
  contains: OPERATORS['contain'],
  notContains: OPERATORS['doesNotContain'],
  lesserThan: OPERATORS['lesserThan'],
  lesserThanOrEqual: OPERATORS['lesserThanOrEqual'],
  greaterThan: OPERATORS['greaterThan'],
  greaterThanOrEqual: OPERATORS['greaterThanOrEqual']
};

export const reverseDateOperatorMap = {
  equals: OPERATORS['equalTo'],
  notEqual: OPERATORS['notEqualTo'],
  between: OPERATORS['between'],
  notInBetween: OPERATORS['notBetween'],
  inLast: OPERATORS['inThePrevious'],
  notInLast: OPERATORS['notInThePrevious'],
  inCurrent: OPERATORS['inTheCurrent'],
  notInCurrent: OPERATORS['notInTheCurrent'],
  before: OPERATORS['before'],
  since: OPERATORS['since']
};

export const DEFAULT_OP_PROPS = {
  categorical: [
    OPERATORS['equalTo'],
    OPERATORS['notEqualTo'],
    OPERATORS['contain'],
    OPERATORS['doesNotContain'],
    OPERATORS['isKnown'],
    OPERATORS['isUnknown']
  ],
  numerical: [
    OPERATORS['equalTo'],
    OPERATORS['notEqualTo'],
    OPERATORS['lesserThan'],
    OPERATORS['lesserThanOrEqual'],
    OPERATORS['greaterThan'],
    OPERATORS['greaterThanOrEqual']
  ],
  datetime: [
    OPERATORS['between'],
    OPERATORS['notBetween'],
    OPERATORS['inTheCurrent'],
    OPERATORS['notInTheCurrent'],
    OPERATORS['inThePrevious'],
    OPERATORS['notInThePrevious'],
    OPERATORS['before'],
    OPERATORS['since']
  ]
};
