import moment from 'moment';
import React, { useState } from 'react';
import ComparisonDatePicker from '../ComparisonDatePicker';
import { COMPARISON_DATE_RANGE_TYPE } from '../comparisonDatePicker.constants';

export default {
  title: 'Components/ComparisonDatePicker',
  component: ComparisonDatePicker
};

export const DefaultComparisonDatePicker = () => {
  const [value, setValue] = useState(null);

  const handleChange = (selectedOption) => {
    const { value, isPreset, customRangeType } = selectedOption;
    if (isPreset) {
      setValue({
        optionValue: value
      });
      return;
    }
    // this makes an assumption that main selected duration has 7 days difference
    if (customRangeType === COMPARISON_DATE_RANGE_TYPE.START_DATE) {
      setValue({
        startDate: value,
        endDate: moment(value).add(7, 'days')
      });
    }
    if (customRangeType === COMPARISON_DATE_RANGE_TYPE.END_DATE) {
      setValue({
        endDate: value,
        startDate: moment(value).subtract(7, 'days')
      });
    }
  };

  return (
    <ComparisonDatePicker
      value={value}
      onChange={handleChange}
      comparisonLabel="Compare"
      buttonSize
      placement="bottomLeft"
      onRemoveClick={setValue.bind(null)}
    />
  );
};
