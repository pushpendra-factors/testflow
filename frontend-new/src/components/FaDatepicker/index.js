import React, { useState } from 'react';
import { DatePicker, Menu, Dropdown, Button, Tooltip } from 'antd';
import { isEqual } from 'lodash';
import MomentTz from 'Components/MomentTz';
// import { TimeZoneOffsetValues } from 'Utils/constants';
import { getRangeByLabel } from './utils';
import { SVG } from '../factorsComponents';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';

const { RangePicker } = DatePicker;

const FaDatepicker = ({
  placement,
  onSelect,
  customPicker,
  presetRange,
  weekPicker,
  monthPicker,
  quarterPicker,
  yearPicker,
  range,
  buttonSize,
  nowPicker,
  className,
  comparison_supported: comparisonSupported = false,
  handleCompareWithClick,
  disabled = false,
  todayPicker = true
}) => {
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [datePickerType, setDatePickerType] = useState('');
  const [dateString, setDateString] = useState(false);
  const [quarterDateStr, setQuarterDateStr] = useState('');

  const MomentTzKey = {
    day: 'days',
    week: 'weeks',
    month: 'months',
    quarter: 'quarters',
    year: 'years',
    hour: 'hours',
    minutes: 'minutes'
  };
  const dateData = {
    startDate: null,
    endDate: null,
    dateString: null,
    dateType: null
  };
  const onChange = (startDate, dateString) => {
    setShowDatePicker(false);
    const dateType = datePickerType;
    let endDate = MomentTz(startDate)
      .startOf('day')
      .add(1, MomentTzKey[dateType]);

    const newDateData = {
      ...dateData,
      startDate,
      endDate,
      dateString,
      dateType
    };

    if (datePickerType === 'month') {
      const startDateMonth = MomentTz(startDate).startOf('month');
      const endDateMonth = MomentTz(startDate).endOf('month');
      const newDateDataMonth = {
        ...dateData,
        startDate: startDateMonth,
        endDate: endDateMonth,
        dateType: datePickerType
      };
      onSelect(newDateDataMonth);
      setDateString('++Month');
    } else if (datePickerType === 'quarter') {
      if (endDate.isAfter(MomentTz())) {
        const endDateMonth = MomentTz();
        const newDateDataMonth = {
          ...dateData,
          startDate,
          endDate: endDateMonth,
          dateType: datePickerType
        };
        setQuarterDateStr(`${endDate.year()}, Q${startDate.quarter()}`);
        onSelect(newDateDataMonth);
      } else {
        endDate = MomentTz(startDate).endOf('Q');
        const newDateDataMonth = {
          ...dateData,
          startDate,
          endDate,
          dateType: datePickerType
        };
        setQuarterDateStr(`${endDate.year()}, Q${startDate.quarter()}`);
        onSelect(newDateDataMonth);
      }
    } else {
      onSelect(newDateData);
    }
  };

  const ifOnlyCustomPicker = () => {
    if (
      customPicker &&
      !presetRange &&
      !weekPicker &&
      !monthPicker &&
      !quarterPicker &&
      !yearPicker &&
      !nowPicker
    ) {
      return true;
    }
    return false;
  };

  const onCustomChange = (startDate, dateString) => {
    const startDt = MomentTz(startDate[0]).startOf('day');
    const endDt = MomentTz(startDate[1]).endOf('day');
    const newDateData = {
      ...dateData,
      startDate: startDt,
      endDate: endDt,
      datePickerType,
      dateString
    };
    setDateString(dateString);
    onSelect(newDateData);
    setShowDatePicker(false);
  };

  const returnPreSetDate = (type) => {
    setDatePickerType(null);
    const today = MomentTz();
    if (type === 'now') {
      const newDateData = {
        ...dateData,
        startDate: MomentTz().subtract(30, 'minutes'),
        endDate: today,
        dateType: type,
        dateString: 'Now'
      };
      setDateString('Now');
      onSelect(newDateData);
    }
    if (type === 'today') {
      const newDateData = {
        ...dateData,
        startDate: MomentTz(today).startOf('day'),
        endDate: today,
        dateType: type,
        dateString: 'Today'
      };
      setDateString('Today');
      onSelect(newDateData);
    }
    if (type === 'yesterday') {
      const newDateData = {
        ...dateData,
        startDate: MomentTz(today).subtract(1, 'days').startOf('day'),
        endDate: MomentTz(today).subtract(1, 'days').endOf('day'),
        dateType: type,
        dateString: 'Yesterday'
      };
      setDateString('Yesterday');
      onSelect(newDateData);
    }
    if (type === 'this_week') {
      const dateRng = getRangeByLabel('This Week');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'This Week'
      };
      setDateString('This Week');
      onSelect(newDateData);
    }
    if (type === 'last_week') {
      const startDate = MomentTz(today).subtract(1, 'weeks').startOf('week');
      const endDate = MomentTz(today).subtract(1, 'weeks').endOf('week');
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'Last Week'
      };
      setDateString('Last Week');
      onSelect(newDateData);
    }
    if (type === 'this_month') {
      const dateRng = getRangeByLabel('This Month');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'This Month'
      };
      setDateString('This Month');
      onSelect(newDateData);
    }
    if (type === 'last_month') {
      const startDate = MomentTz().subtract(1, 'months').startOf('month');
      const endDate = MomentTz().subtract(1, 'months').endOf('month');
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'Last Month'
      };
      setDateString('Last Month');
      onSelect(newDateData);
    }
    if (type === 'this_quarter') {
      const dateRng = getRangeByLabel('This Quarter');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: dateRng
      };
      setDateString('This Quarter');
      onSelect(newDateData);
      setQuarterDateStr(dateRng.dateStr);
    }
    if (type === 'last_quarter') {
      const dateRng = getRangeByLabel('Last Quarter');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: dateRng
      };
      setDateString('Last Quarter');
      onSelect(newDateData);
      setQuarterDateStr(dateRng.dateStr);
    }
    if (type === 'last_7days') {
      const dateRng = getRangeByLabel('Last 7 Days');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: dateRng
      };
      setDateString('Last 7 Days');
      onSelect(newDateData);
      setQuarterDateStr(dateRng.dateStr);
    }
    if (type === 'last_14days') {
      const dateRng = getRangeByLabel('Last 14 Days');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: dateRng
      };
      setDateString('Last 14 Days');
      onSelect(newDateData);
      setQuarterDateStr(dateRng.dateStr);
    }
    if (type === 'last_28days') {
      const dateRng = getRangeByLabel('Last 28 Days');
      const { startDate } = dateRng;
      const { endDate } = dateRng;
      const newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: dateRng
      };
      setDateString('Last 28 Days');
      onSelect(newDateData);
      setQuarterDateStr(dateRng.dateStr);
    }
  };

  const showDatePickerFn = (type) => {
    setDatePickerType(type);
    setShowDatePicker(true);
  };

  const menu = (
    <Menu>
      {nowPicker && (
        <Menu.Item key="now">
          <a target="_blank" onClick={() => returnPreSetDate('now')}>
            Now
          </a>
        </Menu.Item>
      )}

      {presetRange && (
        <>
        {todayPicker &&
          <Menu.Item key="today">
            <a target="_blank" onClick={() => returnPreSetDate('today')}>
              Today
            </a>
          </Menu.Item>
          }
          <Menu.Item key="yesterday">
            <a target="_blank" onClick={() => returnPreSetDate('yesterday')}>
              Yesterday
            </a>
          </Menu.Item>
          <Menu.Item key="this_week">
            <a target="_blank" onClick={() => returnPreSetDate('this_week')}>
              This Week
            </a>
          </Menu.Item>
          <Menu.Item key="last_week">
            <a target="_blank" onClick={() => returnPreSetDate('last_week')}>
              Last Week
            </a>
          </Menu.Item> 
          <Menu.Item key="this_week">
            <a target="_blank" onClick={() => returnPreSetDate('last_7days')}>
              Last 7 days
            </a>
          </Menu.Item>
          <Menu.Item key="this_week">
            <a target="_blank" onClick={() => returnPreSetDate('last_14days')}>
              Last 14 days
            </a>
          </Menu.Item>
          <Menu.Item key="this_week">
            <a target="_blank" onClick={() => returnPreSetDate('last_28days')}>
              Last 28 days
            </a>
          </Menu.Item> 
          <Menu.Item key="this_month">
            <a target="_blank" onClick={() => returnPreSetDate('this_month')}>
              This Month
            </a>
          </Menu.Item>
          <Menu.Item key="last_month">
            <a target="_blank" onClick={() => returnPreSetDate('last_month')}>
              Last Month
            </a>
          </Menu.Item>
          <Menu.Item key="this_quarter">
            <a target="_blank" onClick={() => returnPreSetDate('this_quarter')}>
              This Quarter
            </a>
          </Menu.Item>
          <Menu.Item key="last_quarter">
            <a target="_blank" onClick={() => returnPreSetDate('last_quarter')}>
              Last Quarter
            </a>
          </Menu.Item>
          <Menu.Divider />
        </>
      )}

      {weekPicker && (
        <Menu.Item key="week">
          <a target="_blank" onClick={() => showDatePickerFn('week')}>
            Select Week
          </a>
        </Menu.Item>
      )}
      {monthPicker && (
        <Menu.Item key="month">
          <a target="_blank" onClick={() => showDatePickerFn('month')}>
            Select Month
          </a>
        </Menu.Item>
      )}
      {quarterPicker && (
        <Menu.Item key="quarter">
          <a target="_blank" onClick={() => showDatePickerFn('quarter')}>
            Select Quarter
          </a>
        </Menu.Item>
      )}
      {yearPicker && (
        <Menu.Item key="year">
          <a target="_blank" onClick={() => showDatePickerFn('year')}>
            Select Year
          </a>
        </Menu.Item>
      )}
      {(weekPicker || monthPicker || quarterPicker || yearPicker) && (
        <Menu.Divider />
      )}

      {customPicker && (
        <Menu.Item key="custom">
          <a target="_blank" onClick={() => showDatePickerFn('custom')}>
            Select Custom Range
          </a>
        </Menu.Item>
      )}

      {comparisonSupported && <Menu.Divider />}

      {/* {comparisonSupported && (
        <Menu.Item key="compare">
          <a target="_blank" onClick={handleCompareWithClick}>
            Compare with...
          </a>
        </Menu.Item>
      )} */}
    </Menu>
  );

  const displayRange = (range) => {
    if (dateString === 'Now') {
      // return MomentTz(range.startDate).format('MMM DD, YYYY hh:mma')
      return 'Now';
    }
    if (
      dateString === 'This Quarter' ||
      dateString === 'Last Quarter' ||
      datePickerType === 'quarter'
    ) {
      return quarterDateStr;
    }
    if (dateString === 'Today' || isEqual(range.startDate === range.endDate)) {
      return MomentTz(range.startDate).format('MMM DD, YYYY');
    }
    return `${MomentTz(range.startDate).format('MMM DD, YYYY')} - ${MomentTz(
      range.endDate
    ).format('MMM DD, YYYY')}`;
  };

  const renderCustomPicker = () => (
    <div className="fa-custom-datepicker">
      {
        <Button
          disabled={disabled}
          className={className}
          size={buttonSize || null}
          onClick={() => setShowDatePicker(true)}
        >
          <SVG name="calendar" size={16} extraClass="mr-1" />
          {!showDatePicker && range ? displayRange(range) : null}
          {!showDatePicker && !range ? 'Choose Date' : null}
          {showDatePicker && (
            <RangePicker
              format="MMM DD YYYY"
              // disabledDate={(d) => !d || d.isAfter(MomentTz())}
              dropdownClassName="fa-custom-datepicker--datepicker"
              size="small"
              suffixIcon={null}
              showToday={false}
              bordered={false}
              autoFocus
              allowClear
              open
              onOpenChange={() => {
                setShowDatePicker(false);
              }}
              onChange={onCustomChange}
            />
          )}
          {showDatePicker && (
            <span onClick={() => setShowDatePicker(false)}>
              <SVG name="Times" size={16} extraClass="mr-1" />
            </span>
          )}
        </Button>
      }
    </div>
  );

  const renderFaDatePicker = () => (
    <div className="fa-custom-datepicker">
      <Dropdown
        disabled={disabled}
        overlayClassName="fa-custom-datepicker--dropdown"
        overlay={menu}
        placement={placement}
        trigger={!showDatePicker ? ['click'] : []}
      >
        <Button
          disabled={disabled}
          className={className}
          size={buttonSize || null}
        >
          <SVG name="calendar" size={16} extraClass="mr-1" />
          {!showDatePicker && range ? displayRange(range) : null}
          {!showDatePicker && !range ? 'Choose Date' : null}
          {showDatePicker && (
            <>
              {datePickerType === 'custom' ? (
                <RangePicker
                  disabled={disabled}
                  format="MMM DD YYYY"
                  disabledDate={(d) => !d || (todayPicker ?  d.isAfter(MomentTz()) : d.isAfter(MomentTz().subtract(1, 'days')) )}
                  dropdownClassName="fa-custom-datepicker--datepicker"
                  size="small"
                  suffixIcon={null}
                  showToday={false}
                  bordered={false}
                  autoFocus
                  allowClear
                  open
                  onOpenChange={(open) => {
                    if (open) {
                      setShowDatePicker(false);
                    }
                  }}
                  onChange={onCustomChange}
                />
              ) : (
                <DatePicker
                  picker={datePickerType}
                  disabledDate={(d) => !d || d.isAfter(MomentTz())}
                  dropdownClassName="fa-custom-datepicker--datepicker"
                  autoFocus
                  open
                  onOpenChange={(open) => {
                    if (open) {
                      setShowDatePicker(false);
                    }
                  }}
                  size="small"
                  suffixIcon={null}
                  showToday={false}
                  bordered={false}
                  allowClear
                  onChange={onChange}
                />
              )}
            </>
          )}
          {showDatePicker && (
            <span onClick={() => setShowDatePicker(false)}>
              <SVG name="Times" size={16} extraClass="mr-1" />
            </span>
          )}
        </Button>
      </Dropdown>
    </div>
  );

  return ifOnlyCustomPicker() ? renderCustomPicker() : renderFaDatePicker();
};

export default FaDatepicker;
