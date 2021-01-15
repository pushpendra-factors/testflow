import React, { useState, useEffect } from 'react';
import { Text, SVG } from 'factorsComponents';
import { DatePicker, Menu, Dropdown, Button } from 'antd';
import moment from 'moment';
const { RangePicker } = DatePicker;


const FaDatepicker = ({ placement,
    onSelect, customPicker, presetRange,
    weekPicker, monthPicker, quarterPicker, yearPicker,
    range

}) => {

    const [showDatePicker, setShowDatePicker] = useState(false);
    const [datePickerType, setdatePickerType] = useState('');
    const [dateString, setdateString] = useState(false);



    const momentKey = {
        day: 'days',
        week: 'weeks',
        month: 'months',
        quarter: 'quarters',
        year: 'years',
        hour: 'hours',
        minutes: 'minutes'

    }
    const dateData = {
        startDate: null,
        endDate: null,
        dateString: null,
        dateType: null
    }
    const onChange = (startDate, dateString) => {
        
        setShowDatePicker(false);
        const dateType = datePickerType;
        const endDate = moment(startDate).add(1, momentKey[dateType]);
     
        const newDateData = {
            ...dateData,
            startDate,
            endDate,
            dateString,
            dateType

        }
        console.log("inside onchange");

        if (datePickerType == 'month') {
            console.log("inside onchange month");
            let startDateMonth = moment(startDate).startOf('month');
            let endDateMonth = moment(startDate).endOf('month');
            let newDateDataMonth = {
                ...dateData,
                startDate: startDateMonth,
                endDate: endDateMonth
            }
            console.log("inside onchange-->>", newDateDataMonth);
            onSelect(newDateDataMonth);
            // setdateString('++Month'); 
        }
        else{ 
            onSelect(newDateData); 
        }  
    }

    const returnPreSetDate = (type) => {
        setdatePickerType(null)
        const today = moment();
        if (type == 'today') {
            let newDateData = {
                ...dateData,
                startDate: today,
                endDate: today
            }
            setdateString('Today');
            onSelect(newDateData);
        }
        if (type == 'this_week') {
            let startDate = moment().startOf('week');
            let endDate = moment();
            let newDateData = {
                ...dateData,
                startDate,
                endDate
            }
            setdateString('This Week');
            onSelect(newDateData);
        }
        if (type == 'this_month') {
            let startDate = moment().startOf('month');
            let endDate = moment();
            let newDateData = {
                ...dateData,
                startDate,
                endDate
            }
            setdateString('This Month');
            onSelect(newDateData);
        }
    }

    const showDatePickerFn = (type) => {
        setdatePickerType(type);
        setShowDatePicker(true);
    }

    const menu = (
        <Menu>
            {presetRange && <>
                <Menu.Item>
                    <a target="_blank" onClick={() => returnPreSetDate('today')}>
                        Today
                </a>
                </Menu.Item>
                <Menu.Item>
                    <a target="_blank" onClick={() => returnPreSetDate('this_week')}>
                        This Week
                </a>
                </Menu.Item>
                <Menu.Item>
                    <a target="_blank" onClick={() => returnPreSetDate('this_month')}>
                        This Month
                </a>
                </Menu.Item>
                <Menu.Divider />
            </>}

            {weekPicker &&
                <Menu.Item>
                    <a target="_blank" onClick={() => showDatePickerFn('week')}>
                        Select Week
                    </a>
                </Menu.Item>
            }
            {monthPicker &&
                <Menu.Item>
                    <a target="_blank" onClick={() => showDatePickerFn('month')}>
                        Select Month
                </a>
                </Menu.Item>
            }
            {quarterPicker &&
                <Menu.Item>
                    <a target="_blank" onClick={() => showDatePickerFn('quarter')}>
                        Select Quarter
                </a>
                </Menu.Item>
            }
            {yearPicker &&
                <Menu.Item>
                    <a target="_blank" onClick={() => showDatePickerFn('year')}>
                        Select Year
                </a>
                </Menu.Item>
            }
            {(weekPicker || monthPicker || quarterPicker || yearPicker) && <Menu.Divider />}

            {customPicker &&
                <Menu.Item>
                    <a target="_blank" onClick={() => showDatePickerFn('custom')}>
                        Select Custom Range
            </a>
                </Menu.Item>
            }

        </Menu>
    );


    const displayRange = (range) => {
        
        return moment(range.startDate).format('MMM DD, YYYY') + ' - ' +
            moment(range.endDate).format('MMM DD, YYYY');
    }

    return (
        <div className="fa-custom-datepicker">
            {<>
                <Dropdown overlayClassName={'fa-custom-datepicker--dropdown'} overlay={menu} placement={placement} trigger={!showDatePicker ? ['click'] : []} >

                    <Button size={'large'}><SVG name={'Calendar'} extraClass={'mr-1'} />
                        {!showDatePicker && range ? displayRange(range) : null}
                        {!showDatePicker && !range ? `Choose Date` : null}
                        {showDatePicker && <>
                            {datePickerType == 'custom' ? <RangePicker format={'MMM DD YYYY'} 
                            disabledDate={d => !d || d.isAfter(moment())} dropdownClassName={'fa-custom-datepicker--datepicker'} size={'small'} suffixIcon={null} showToday={false} bordered={false} autoFocus={true} allowClear={true} open={true} onChange={onChange} /> :
                                <DatePicker picker={datePickerType}
                                    disabledDate={d => !d || d.isAfter(moment())}
                                    dropdownClassName={'fa-custom-datepicker--datepicker'} autoFocus={true} open={true} size={'small'} suffixIcon={null} showToday={false} bordered={false} allowClear={true} onChange={onChange} />}
                        </>
                        }
                        {showDatePicker && <span onClick={() => setShowDatePicker(false)}>
                            <SVG name={'Times'} size={16} extraClass={'mr-1'} />
                        </span>}
                    </Button>
                </Dropdown>
            </>
            }

        </div>
    );
};

export default FaDatepicker;

