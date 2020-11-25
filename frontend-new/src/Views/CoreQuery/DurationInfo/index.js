import React, { useState, useEffect, useCallback } from 'react';
import { Popover, Button } from 'antd';
import DateRangeSelector from '../../../components/QueryComposer/DateRangeSelector';
import { getDateRange, readableDateRange } from '../../../components/QueryComposer/DateRangeSelector/utils';
import { SVG } from '../../../components/factorsComponents';

function DurationInfo({ durationObj, handleDurationChange }) {
  const [dateRangeVisible, setDateRangeVisibile] = useState(false);
  const [calendarLabel, setCalendarLabel] = useState('Pick Dates');

  const convertToDateRange = useCallback(() => {
    const range = getDateRange(durationObj);
    setCalendarLabel(readableDateRange(range[0]));
  }, [durationObj]);

  useEffect(() => {
    convertToDateRange();
  }, [durationObj.from, durationObj.to, convertToDateRange]);

  const setDateRange = (dates) => {
    handleDurationChange(dates);
    setDateRangeVisibile(false);
  };

  return (
    <Popover
      className="fa-event-popover"
      trigger="click"
      visible={dateRangeVisible}
      onVisibleChange={setDateRangeVisibile}
      content={
        <DateRangeSelector
          ranges={getDateRange()}
          pickerVisible={dateRangeVisible}
          setDates={setDateRange}
          closeDatePicker={() => setDateRangeVisibile(false)}
        />
      }
    >
      <Button><SVG name={'calendar'} extraClass={'mr-1'} /> {calendarLabel} </Button>
    </Popover>
  );
}

export default DurationInfo;
