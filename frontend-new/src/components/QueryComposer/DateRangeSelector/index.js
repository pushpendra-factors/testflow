import React from 'react';
import styles from './index.module.scss';

import { DatePicker } from 'antd';

const { RangePicker } = DatePicker;

const DateRangeSelector = ({ setDates, pickerVisible }) => {
  const onChange = (dates) => {
    setDates(dates);
  };

  return (
        <div className={styles.dr_container}>
            <RangePicker open={pickerVisible} onChange={onChange}></RangePicker>
        </div>
  );
};

export default DateRangeSelector;
