import { OPERATORS } from 'Utils/constants';
import { DEFAULT_OP_PROPS } from 'Utils/operatorMapping';

export const DEFAULT_OPERATOR_PROPS = {
  categorical: DEFAULT_OP_PROPS['categorical'],
  numerical: DEFAULT_OP_PROPS['numerical'],
  datetime: DEFAULT_OP_PROPS['datetime']
};

export const dateTimeSelect = new Map([
  ['Days', 'days'],
  ['Weeks', 'week'],
  ['Months', 'month'],
  ['Quarters', 'quarter'],
  ['days', 'Days'],
  ['week', 'Weeks'],
  ['month', 'Months'],
  ['quarter', 'Quarters']
]);
