export const SELECTABLE_OPTIONS_KEYS = {
  PREVIOUS_DAY: 'previous_day',
  PREVIOUS_7_DAYS: 'previous_7_days',
  PREVIOUS_30_DAYS: 'previous_30_days',
  PREVIOUS_90_DAYS: 'previous_90_days',
  PREVIOUS_365_DAYS: 'previous_365_days'
};

export const SELECTABLE_OPTIONS = [
  {
    label: 'Previous Day',
    value: SELECTABLE_OPTIONS_KEYS.PREVIOUS_DAY
  },
  {
    label: 'Previous 7 Days',
    value: SELECTABLE_OPTIONS_KEYS.PREVIOUS_7_DAYS
  },
  {
    label: 'Previous 30 Days',
    value: SELECTABLE_OPTIONS_KEYS.PREVIOUS_30_DAYS
  },
  {
    label: 'Previous 90 Days',
    value: SELECTABLE_OPTIONS_KEYS.PREVIOUS_90_DAYS
  },
  {
    label: 'Previous 365 Days',
    value: SELECTABLE_OPTIONS_KEYS.PREVIOUS_365_DAYS
  }
];

export const COMPARISON_DATE_RANGE_TYPE = {
  START_DATE: 'start',
  END_DATE: 'end'
};
