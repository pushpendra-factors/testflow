import React, { useState, useEffect } from 'react';
import { DateRangePicker } from 'react-date-range';
import styles from './index.module.scss';
import MomentTz from 'Components/MomentTz';
import {
  DEFINED_DATE_RANGES,
  DEFAULT_DATE_RANGE
} from './utils';
import { Button } from 'antd';

const DateRangeSelector = ({
  ranges,
  setDates, closeDatePicker
}) => {
  const [selectedRngState, setSelectedRngState] = useState(ranges);
  const [selectedDate, setSelectedDate] = useState({});

  useEffect(() => setDateRange(), [selectedDate]);
  useEffect(() => setSelectedRngState(ranges), [ranges]);

  const onChange = (range) => {
    setSelectedDate(range);
  };

  const setDateRange = () => {
    if (!selectedDate || !selectedDate.selected) return;

    const ranges = [{ ...DEFAULT_DATE_RANGE }];
    ranges[0].startDate = MomentTz(selectedDate.selected.startDate).toDate();
    ranges[0].endDate = MomentTz(selectedDate.selected.endDate).toDate();

    setSelectedRngState(ranges);
  };

  const applyRange = () => {
    setDates(selectedDate);
  };

  return (
        <div className={'fapp-date-picker'}>
            <DateRangePicker ranges={selectedRngState}
            onChange={onChange}
            staticRanges={ DEFINED_DATE_RANGES }
            inputRanges={[]}
            minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
            maxDate={MomentTz(new Date()).subtract(1, 'days').endOf('day').toDate()}
            closeDatePicker={closeDatePicker} />
            <div className={styles.dt_actions}>
              <Button className={'mr-2'} onClick={closeDatePicker}>Cancel</Button>
              <Button type="primary" onClick={applyRange}>Apply</Button>
            </div>
        </div>
  );
};

export default DateRangeSelector;
