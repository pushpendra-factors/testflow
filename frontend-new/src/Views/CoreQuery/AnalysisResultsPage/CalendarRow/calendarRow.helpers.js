import MomentTz from '../../../../components/MomentTz';
import { SELECTABLE_OPTIONS_KEYS as COMPARISON_DATE_SELECTABLE_OPTIONS } from '../../../../components/ComparisonDatePicker/ComparisonDatePicker';
import { COMPARISON_DATE_RANGE_TYPE } from '../../../../components/ComparisonDatePicker/comparisonDatePicker.constants';

export const getCompareRange = ({
  selectedValue,
  durationObj,
  isPreset,
  customRangeType
}) => {
  if (isPreset) {
    if (selectedValue === COMPARISON_DATE_SELECTABLE_OPTIONS.PREVIOUS_DAY) {
      return {
        ...durationObj,
        startDate: parseInt(
          MomentTz(durationObj.from).subtract(1, 'day').format('x')
        ),
        endDate: parseInt(
          MomentTz(durationObj.to).subtract(1, 'day').format('x')
        ),
        selectedOption: selectedValue
      };
    }
    if (selectedValue === COMPARISON_DATE_SELECTABLE_OPTIONS.PREVIOUS_7_DAYS) {
      return {
        ...durationObj,
        startDate: parseInt(
          MomentTz(durationObj.from).subtract(7, 'days').format('x')
        ),
        endDate: parseInt(
          MomentTz(durationObj.to).subtract(7, 'days').format('x')
        ),
        selectedOption: selectedValue
      };
    }
    if (selectedValue === COMPARISON_DATE_SELECTABLE_OPTIONS.PREVIOUS_30_DAYS) {
      return {
        ...durationObj,
        startDate: parseInt(
          MomentTz(durationObj.from).subtract(30, 'days').format('x')
        ),
        endDate: parseInt(
          MomentTz(durationObj.to).subtract(30, 'days').format('x')
        ),
        selectedOption: selectedValue
      };
    }
    if (selectedValue === COMPARISON_DATE_SELECTABLE_OPTIONS.PREVIOUS_90_DAYS) {
      return {
        ...durationObj,
        startDate: parseInt(
          MomentTz(durationObj.from).subtract(90, 'days').format('x')
        ),
        endDate: parseInt(
          MomentTz(durationObj.to).subtract(90, 'days').format('x')
        ),
        selectedOption: selectedValue
      };
    }
    if (
      selectedValue === COMPARISON_DATE_SELECTABLE_OPTIONS.PREVIOUS_365_DAYS
    ) {
      return {
        ...durationObj,
        startDate: parseInt(
          MomentTz(durationObj.from).subtract(365, 'days').format('x')
        ),
        endDate: parseInt(
          MomentTz(durationObj.to).subtract(365, 'days').format('x')
        ),
        selectedOption: selectedValue
      };
    }
  }

  const fr = MomentTz(durationObj.from).startOf('day').utc().unix();
  const to = MomentTz(durationObj.to).endOf('day').utc().unix();
  const daysDiff = MomentTz(to * 1000).diff(fr * 1000, 'days');
  if (customRangeType === COMPARISON_DATE_RANGE_TYPE.START_DATE) {
    let endDate = MomentTz(selectedValue).add(daysDiff, 'days');
    if (MomentTz(endDate).isAfter(MomentTz())) {
      endDate = MomentTz();
    }
    return {
      ...durationObj,
      startDate: selectedValue,
      endDate,
      selectedOption: 'custom'
    };
  }
  if (customRangeType === COMPARISON_DATE_RANGE_TYPE.END_DATE) {
    return {
      ...durationObj,
      startDate: MomentTz(selectedValue).subtract(daysDiff, 'days'),
      endDate: selectedValue,
      selectedOption: 'custom'
    };
  }
};
