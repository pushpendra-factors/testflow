import React, { useState } from 'react';
import { SVG } from '../factorsComponents';
import { DatePicker, Menu, Dropdown, Button } from 'antd';
import MomentTz from 'Components/MomentTz';
import { useSelector } from 'react-redux';
// import { TimeZoneOffsetValues } from 'Utils/constants';
import {
  getFirstDayOfLastWeek,
  getLastDayOfLastWeek,
  getFirstDayOfLastMonth,
  getLastDayOfLastMonth,
  getRangeByLabel,
} from './utils';

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
  comparison_supported = false,
  handleCompareWithClick,
}) => {
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [datePickerType, setdatePickerType] = useState('');
  const [dateString, setdateString] = useState(false);

  // const { active_project } = useSelector((state) => state.global); 
  
  // active_project.time_zone ? MomentTz.tz.setDefault(TimeZoneOffsetValues[active_project.time_zone]?.city): MomentTz.tz.setDefault('Asia/Kolkata');
  // console.log('MomentTz.tz.setDefault',TimeZoneOffsetValues[active_project.time_zone]?.city, MomentTz().format())

  const MomentTzKey = {
    day: 'days',
    week: 'weeks',
    month: 'months',
    quarter: 'quarters',
    year: 'years',
    hour: 'hours',
    minutes: 'minutes',
  };
  const dateData = {
    startDate: null,
    endDate: null,
    dateString: null,
    dateType: null,
  };
  const onChange = (startDate, dateString) => {
    setShowDatePicker(false);
    const dateType = datePickerType;
    const endDate = MomentTz(startDate).add(1, MomentTzKey[dateType]);

    const newDateData = {
      ...dateData,
      startDate,
      endDate,
      dateString,
      dateType,
    };

    if (datePickerType == 'month') {
      let startDateMonth = MomentTz(startDate).startOf('month');
      let endDateMonth = MomentTz(startDate).endOf('month');
      let newDateDataMonth = {
        ...dateData,
        startDate: startDateMonth,
        endDate: endDateMonth,
        dateType: datePickerType,
      };
      onSelect(newDateDataMonth);
      // setdateString('++Month');
    } else if (datePickerType === 'quarter') {
      if (endDate.isAfter(MomentTz())) {
        const endDateMonth = MomentTz();
        let newDateDataMonth = {
          ...dateData,
          startDate,
          endDate: endDateMonth,
          dateType: datePickerType,
        };

        onSelect(newDateDataMonth);
      } else {
        let newDateDataMonth = {
          ...dateData,
          startDate,
          endDate,
          dateType: datePickerType,
        };

        onSelect(newDateDataMonth);
      }
    } else {
      onSelect(newDateData);
    }
  };

  const onCustomChange = (startDate, dateString) => {
    const startDt = MomentTz(startDate[0]).startOf('day');
    let endDt = MomentTz(startDate[1]);
    if (endDt.isBefore(MomentTz().startOf('day'))) {
      endDt = endDt.endOf('day');
    } else {
      endDt = MomentTz();
    }

    let newDateData = {
      ...dateData,
      startDate: startDt,
      endDate: endDt,
      datePickerType,
      dateString,
    };
    setdateString(dateString);
    onSelect(newDateData);
    setShowDatePicker(false);
  };

  const returnPreSetDate = (type) => {
    setdatePickerType(null);
    const today = MomentTz();
    if (type == 'now') {
      let newDateData = {
        ...dateData,
        startDate: MomentTz().subtract(30, 'minutes'),
        endDate: today,
        dateType: type,
        dateString: 'Now',
      };
      setdateString('Now');
      onSelect(newDateData);
    }
    if (type == 'today') {
      let newDateData = {
        ...dateData,
        startDate: MomentTz(today).startOf('day'),
        endDate: today,
        dateType: type,
        dateString: 'Today',
      };
      setdateString('Today');
      onSelect(newDateData);
    }
    if (type == 'yesterday') {
      let newDateData = {
        ...dateData,
        startDate: MomentTz(today).subtract(1, 'days').startOf('day'),
        endDate: MomentTz(today).subtract(1, 'days').endOf('day'),
        dateType: type,
        dateString: 'Yesterday',
      };
      setdateString('Yesterday');
      onSelect(newDateData);
    }
    if (type == 'this_week') {
      const dateRng = getRangeByLabel('This Week');
      let startDate = dateRng.startDate;
      let endDate = dateRng.endDate;
      let newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'This Week',
      };
      setdateString('This Week');
      onSelect(newDateData);
    }
    if (type == 'last_week') {
      let startDate = MomentTz(getFirstDayOfLastWeek()).startOf('day').toDate();
      let endDate = MomentTz(getLastDayOfLastWeek()).endOf('day').toDate();
      let newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'Last Week',
      };
      setdateString('Last Week');
      onSelect(newDateData);
    }
    if (type == 'this_month') {
      const dateRng = getRangeByLabel('This Month');
      let startDate = dateRng.startDate;
      let endDate = dateRng.endDate;
      let newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'This Month',
      };
      setdateString('This Month');
      onSelect(newDateData);
    }
    if (type == 'last_month') {
      let startDate = MomentTz(getFirstDayOfLastMonth()).startOf('day').toDate();
      let endDate = MomentTz(getLastDayOfLastMonth()).endOf('day').toDate();
      let newDateData = {
        ...dateData,
        startDate,
        endDate,
        dateType: type,
        dateString: 'Last Month',
      };
      setdateString('Last Month');
      onSelect(newDateData);
    }
  };

  const showDatePickerFn = (type) => {
    setdatePickerType(type);
    setShowDatePicker(true);
  };

  const menu = (
    <Menu>
      {nowPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => returnPreSetDate('now')}>
            Now
          </a>
        </Menu.Item>
      )}

      {presetRange && (
        <>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('today')}>
              Today
            </a>
          </Menu.Item>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('yesterday')}>
              Yesterday
            </a>
          </Menu.Item>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('this_week')}>
              This Week
            </a>
          </Menu.Item>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('last_week')}>
              Last Week
            </a>
          </Menu.Item>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('this_month')}>
              This Month
            </a>
          </Menu.Item>
          <Menu.Item>
            <a target='_blank' onClick={() => returnPreSetDate('last_month')}>
              Last Month
            </a>
          </Menu.Item>
          <Menu.Divider />
        </>
      )}

      {weekPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => showDatePickerFn('week')}>
            Select Week
          </a>
        </Menu.Item>
      )}
      {monthPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => showDatePickerFn('month')}>
            Select Month
          </a>
        </Menu.Item>
      )}
      {quarterPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => showDatePickerFn('quarter')}>
            Select Quarter
          </a>
        </Menu.Item>
      )}
      {yearPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => showDatePickerFn('year')}>
            Select Year
          </a>
        </Menu.Item>
      )}
      {(weekPicker || monthPicker || quarterPicker || yearPicker) && (
        <Menu.Divider />
      )}

      {customPicker && (
        <Menu.Item>
          <a target='_blank' onClick={() => showDatePickerFn('custom')}>
            Select Custom Range
          </a>
        </Menu.Item>
      )}

      {comparison_supported && <Menu.Divider />}

      {comparison_supported && (
        <Menu.Item>
          <a target='_blank' onClick={handleCompareWithClick}>
            Compare with...
          </a>
        </Menu.Item>
      )}
    </Menu>
  );

  const displayRange = (range) => {
    if (dateString == 'Now') {
      // return MomentTz(range.startDate).format('MMM DD, YYYY hh:mma')
      return 'Now';
    }
    if (dateString == 'Today' || range.startDate == range.endDate) {
      return MomentTz(range.startDate).format('MMM DD, YYYY');
    } else {
      return (
        MomentTz(range.startDate).format('MMM DD, YYYY') +
        ' - ' +
        MomentTz(range.endDate).format('MMM DD, YYYY')
      );
    }
  };

  return (
    <div className={`fa-custom-datepicker`}>
      {
        <>
          <Dropdown
            overlayClassName={'fa-custom-datepicker--dropdown'}
            overlay={menu}
            placement={placement}
            trigger={!showDatePicker ? ['click'] : []}
          >
            <Button className={className} size={buttonSize ? buttonSize : null}>
              <SVG name={'calendar'} size={16} extraClass={'mr-1'} />
              {!showDatePicker && range ? displayRange(range) : null}
              {!showDatePicker && !range ? `Choose Date` : null}
              {showDatePicker && (
                <>
                  {datePickerType == 'custom' ? (
                    <RangePicker
                      format={'MMM DD YYYY'}
                      disabledDate={(d) => !d || d.isAfter(MomentTz())}
                      dropdownClassName={'fa-custom-datepicker--datepicker'}
                      size={'small'}
                      suffixIcon={null}
                      showToday={false}
                      bordered={false}
                      autoFocus={true}
                      allowClear={true}
                      open={true}
                      onOpenChange={()=>{
                        setShowDatePicker(false);
                      }}
                      onChange={onCustomChange}
                    />
                  ) : (
                    <DatePicker
                      picker={datePickerType}
                      disabledDate={(d) => !d || d.isAfter(MomentTz())}
                      dropdownClassName={'fa-custom-datepicker--datepicker'}
                      autoFocus={true}
                      open={true}
                      onOpenChange={()=>{
                        setShowDatePicker(false);
                      }}
                      size={'small'}
                      suffixIcon={null}
                      showToday={false}
                      bordered={false}
                      allowClear={true}
                      onChange={onChange}
                    />
                  )}
                </>
              )}
              {showDatePicker && (
                <span onClick={() => setShowDatePicker(false)}>
                  <SVG name={'Times'} size={16} extraClass={'mr-1'} />
                </span>
              )}
            </Button>
          </Dropdown>
        </>
      }
    </div>
  );
};

export default FaDatepicker;
