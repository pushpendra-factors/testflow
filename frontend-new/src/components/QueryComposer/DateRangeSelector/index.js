import React from 'react';
import styles from './index.module.scss';
import { DateRangePicker } from 'react-date-range';
import moment from 'moment';
import {
  DEFAULT_DATE_RANGE,
  DEFINED_DATE_RANGES
} from './utils';

const DateRangeSelector = ({
  ranges,
  setDates, staticRanges,
  inputRanges, minDate, maxDate, closeDatePicker, pickerVisible
}) => {
  const onChange = (dates) => {
    // console.log(staticRanges);
    // console.log(inputRanges);
    // console.log(dates);
    setDates(dates);
  };

  return (
        <div className={'fapp-date-picker'} style={{ display: 'block', marginTop: '10px' }}>
            <DateRangePicker ranges={ranges}
            onChange={onChange}
            staticRanges={ DEFINED_DATE_RANGES }
            inputRanges={[]}
            minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
            maxDate={moment(new Date()).subtract(1, 'days').endOf('day').toDate()}
            closeDatePicker={closeDatePicker} />
        </div>
  );
};

export default DateRangeSelector;
