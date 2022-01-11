import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents'; 
import { Button, Input, InputNumber, Tooltip, DatePicker, Select } from 'antd';
import GroupSelect2 from '../GroupSelect2';
import FaDatepicker from 'Components/FaDatepicker';
import FaSelect from '../FaSelect';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import {DEFAULT_OPERATOR_PROPS} from 'Components/FaFilterSelect/utils';
import moment from 'moment';
import _ from 'lodash'; 
import { DISPLAY_PROP } from '../../../utils/constants';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;
 
const {Option} = Select;

const FAFilterSelect = ({
    propOpts = [],
    operatorOpts=defaultOpProps,
    valueOpts=[],
    setValuesByProps,
    applyFilter,
    filter
},
) => {

    const [propState, setPropState] = useState({
        icon: '',
        name: '',
        type: '',
        extra: ''
    });

    const rangePicker = ['=', '!='];
    const customRangePicker = ['between', 'not between'];
    const deltaPicker = ['in the previous', 'not in the previous'];
    const currentPicker = ['in the current', 'not in the current'];
    const datePicker = ['before', 'since'];


    const [operatorState, setOperatorState] = useState("="); 
    const [valuesState, setValuesState] = useState(null);

    const [propSelectOpen, setPropSelectOpen] = useState(false);
    const [operSelectOpen, setOperSelectOpen] = useState(false);
    const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);
    const [grnSelectOpen, setGrnSelectOpen] = useState(false);
    const [showDatePicker ,setShowDatePicker] = useState(false);

    const [updateState, updateStateApply] = useState(false);
    const [eventFilterInfo, seteventFilterInfo] = useState(null);

    const {userPropNames, eventPropNames} = useSelector((state) => state.coreQuery)

    useEffect(() => {
        if (filter) {  
            const prop = filter.props; 
            setPropState({ icon: prop[2], name: prop[0], type: prop[1]});
            setOperatorState(filter.operator);
            // Set values state
            setValues();
            setPropSelectOpen(false);
            setOperSelectOpen(false);
            setValuesSelectionOpen(false);

        }

    }, [filter])

    useEffect(() => {
        if(updateState && valuesState && propState.type !== 'numerical') {
            emitFilter();
            updateStateApply(false);
        }
    }, [updateState])


    const setValues = () => {
        let values;
        if (filter.props[1] === 'datetime') { 
            const parsedValues = (filter.values[0] ? (typeof filter.values[0] === 'string')? JSON.parse(filter.values) : filter.values : {});
            values = parseDateRangeFilter(parsedValues.fr, parsedValues.to, parsedValues)
        } else {
            values = filter.values;
        }
        setValuesState(values);
    }


    const emitFilter = () => {
        if(propState && operatorState && valuesState) {
            applyFilter({
                props: [propState.name, propState.type, propState.icon],
                operator: operatorState,
                values: valuesState,
                extra: eventFilterInfo ? eventFilterInfo : null
            })
        }
    }

    const operatorSelect = (op) => {
        setOperatorState(op);
        setValuesState(null);
        setOperSelectOpen(false);
    }

    const propSelect = (label, val, cat) => { 
        let prop = [label, ...val];
        setPropState({ icon: prop[0], name: prop[1], type: prop[3], extra: val});
        setPropSelectOpen(false);
        setOperatorState(prop[3] === 'datetime' ? 'between' : "=");
        setValuesState(null);
        setValuesByProps([...val]);
        seteventFilterInfo(val)
    }

    const valuesSelect = (val) => {
        setValuesState(val.map(vl => JSON.parse(vl)[0]));
        setValuesSelectionOpen(false);
        updateStateApply(true);
    }

    const onDateSelect = (rng) => {
        let startDate;
        let endDate;
        if(isArray(rng.startDate)) {
            startDate = rng.startDate[0].toDate().getTime();
            endDate = rng.startDate[1].toDate().getTime();
        } else {
            if(rng.startDate && rng.startDate._isAMomentObject){
                startDate = rng.startDate.toDate().getTime();
            } else {
                startDate = rng.startDate.getTime();
            }
    
            if(rng.endDate && rng.endDate._isAMomentObject){
                endDate = rng.endDate.toDate().getTime();
            } else {
                endDate = rng.endDate.getTime();
            }
        }
        
        const rangeValue = {
            "fr": startDate,
            "to": endDate,
            "ovp": false
        }

        setValuesState(JSON.stringify(rangeValue));
        updateStateApply(true);
    }
    const setNumericalValue = (ev) => {
        // onNumericalSelect(ev);

        setValuesState(String(ev).toString());
    }

    const parseDateRangeFilter = (fr, to, value) => {
        const fromVal = fr ? fr : new Date(MomentTz().startOf('day')).getTime();
        const toVal = to ? to : new Date(MomentTz()).getTime();
        return {
            from: fromVal,
            to: toVal,
            ovp: false,
            num: value["num"],
            gran: value["gran"]
        }
        // return (MomentTz(fromVal).format('MMM DD, YYYY') + ' - ' +
        //           MomentTz(toVal).format('MMM DD, YYYY'));
    }

    const renderGroupDisplayName = (propState) => { 
        // propState?.name ? userPropNames[propState?.name] ? userPropNames[propState?.name] : propState?.name : 'Select Property'
        let propertyName = '';
        // if(propState.name && propState.icon === 'user') {
        //   propertyName = userPropNames[propState.name]?  userPropNames[propState.name] : propState.name;
        // }
        // if(propState.name && propState.icon === 'event') {
        //   propertyName = eventPropNames[propState.name]?  eventPropNames[propState.name] : propState.name;
        // }
        
        propertyName = _.startCase(propState?.name); 

        if(!propState.name) {
          propertyName = 'Select Property';
        }
        return propertyName;
      }

    const renderPropSelect = () => {
        return (<div className={styles.filter__propContainer}>

            <Tooltip title={renderGroupDisplayName(propState)}>
                <Button
                    // icon={propState && propState.icon ? <SVG name={propState.icon} size={16} color={'purple'} /> : null}
                    className={`fa-button--truncate-xs`}
                    type="link"
                    onClick={() => setPropSelectOpen(!propSelectOpen)}> {renderGroupDisplayName(propState)}
                </Button>
            </Tooltip>
            <div className={styles.filter__event_selector}>
                {propSelectOpen && (
                    <div className={styles.filter__event_selector__btn}>
                        <GroupSelect2
                            groupedProperties={propOpts}
                            placeholder="Select Property"
                            optionClick={(label, val, cat) => propSelect(label, val, cat)}
                            onClickOutside={() => setPropSelectOpen(false)}
                            hideTitle={true}
                            isFilterDD={true}
                        />
                    </div>
                )
                }
            </div>

        </div>)
    } 
    const renderOperatorSelector = () => {
        return (<div className={styles.filter__propContainer}>

            <Button
                className={`fa-button--truncate ml-2`}
                type="link"
                onClick={() => setOperSelectOpen(true)}> {operatorState ? operatorState : 'Select Operator'}
            </Button>

            {operSelectOpen &&
                <FaSelect
                    options={operatorOpts[propState.type].map(op => [op])}
                    optionClick={(val) => operatorSelect(val)}
                    onClickOutside={() => setOperSelectOpen(false)}
                >
                </FaSelect>
            }

        </div>);
    }

    const setDeltaNumber = (val = 1) => {
        const parsedValues = (valuesState ? (typeof valuesState === 'string')? JSON.parse(valuesState) : valuesState : {});
        parsedValues['num'] = val;
        parsedValues['gran'] = 'days';
        setValuesState(JSON.stringify(parsedValues));
        updateStateApply(true);
    }

    const setDeltaGran = (val) => {
        const parsedValues = (valuesState ? (typeof valuesState === 'string')? JSON.parse(valuesState) : valuesState : {});
        parsedValues['gran'] = val;
        setValuesState(JSON.stringify(parsedValues));
        setGrnSelectOpen(false);
        setDeltaFilt();
    }

    const setCurrentFilt = () => {
        const parsedValues = valuesState
          ? typeof valuesState === 'string'
            ? JSON.parse(valuesState)
            : valuesState
          : {};
        if (parsedValues['gran']) {
          updateStateApply(true);
        }
      };

    const setDeltaFilt = () => {
        const parsedValues = valuesState
          ? typeof valuesState === 'string'
            ? JSON.parse(valuesState)
            : valuesState
          : {};
        if (parsedValues['num'] && parsedValues['gran']) {
          updateStateApply(true);
        }
      };

    const onDatePickerSelect = (val) => {
        let dateT;
        let dateValue = {};
        const operatorSt = isArray(operatorState)? operatorState[0] : operatorState;
        if(operatorSt === 'before') {
            dateT = MomentTz(val).startOf('day');
            dateValue["to"] = dateT.toDate().getTime();
        }

        if(operatorSt === 'since') {
            dateT = MomentTz(val).startOf('day');
            dateValue["fr"] = dateT.toDate().getTime();
        }

        setValuesState(JSON.stringify(dateValue));
        updateStateApply(true);
    }

    const setCurrentGran = (val) => {
        const parsedValues = valuesState
          ? typeof valuesState === 'string'
            ? JSON.parse(valuesState)
            : valuesState
          : {};
        parsedValues['gran'] = val;
        setValuesState(JSON.stringify(parsedValues));
        setGrnSelectOpen(false);
        setCurrentFilt();
      };


    const selectDateTimeSelector = (operator, rang, parsedVals) => {
        let selectorComponent = null;
    
        const parsedValues = valuesState
          ? typeof valuesState === 'string'
            ? JSON.parse(valuesState)
            : valuesState
          : {};
    
        if (rangePicker.includes(operator)) {
          selectorComponent = (
            <FaDatepicker
              customPicker
              presetRange
              monthPicker
              placement='topRight'
              range={rang}
              onSelect={(rng) => onDateSelect(rng)}
            />
          );
        }
    
        if (customRangePicker.includes(operator)) {
          selectorComponent = (
            <FaDatepicker
              customPicker
              placement='topRight'
              range={rang}
              onSelect={(rng) => onDateSelect(rng)}
            />
          );
        }
    
        if (deltaPicker.includes(operator)) {
          selectorComponent = (
            <div className={`fa-filter-dateDeltaContainer`}>
              <InputNumber
                value={parsedValues['num']}
                min={1}
                max={999}
                onChange={setDeltaNumber}
              ></InputNumber>
    
              <Select
                defaultValue=''
                value={parsedValues['gran']}
                className={'fa-select--ghost'}
                onChange={setDeltaGran}
              >
                <Option value='' disabled>
                  <i>Select:</i>
                </Option>
                <Option value='days'>Days</Option>
                <Option value='week'>Weeks</Option>
                <Option value='month'>Months</Option>
                <Option value='quarter'>Quarters</Option>
              </Select>
            </div>
          );
        }
    
        if (currentPicker.includes(operator)) {
          selectorComponent = (
            <div className={`fa-filter-dateDeltaContainer`}>
              <Select
                defaultValue=''
                value={parsedValues['gran']}
                className={'fa-select--ghost'}
                onChange={setCurrentGran}
              >
                <Option value='' disabled>
                  <i>Select:</i>
                </Option>
                <Option value='week'>Week</Option>
                <Option value='month'>Month</Option>
                <Option value='quarter'>Quarter</Option>
              </Select>
            </div>
          );
        }
    
        if (datePicker.includes(operator)) {
          selectorComponent = (
            <DatePicker
              disabledDate={(d) => !d || d.isAfter(MomentTz())}
              autoFocus={false}
              className={`fa-date-picker`}
              open={showDatePicker}
              onOpenChange={() => {
                setShowDatePicker(!showDatePicker);
              }}
              value={
                operator === 'before'
                  ? moment(parsedValues['to'])
                  : moment(parsedValues['from'] ? parsedValues['from'] : parsedValues['fr'])
              }
              size={'small'}
              suffixIcon={null}
              showToday={false}
              bordered={true}
              allowClear={true}
              onChange={onDatePickerSelect}
            />
          );
        }
    
        return selectorComponent;
      };

    const renderValuesSelector = () => {
        let selectionComponent;
        const values = [];
        if (propState.type === 'categorical') {
            selectionComponent = (<FaSelect
                multiSelect={true}
                options={valueOpts && valueOpts[propState.name]?.length ? valueOpts[propState.name].map(op => [op]) : []}
                applClick={(val) => valuesSelect(val)}
                onClickOutside={() => setValuesSelectionOpen(false)}
                selectedOpts={valuesState ? valuesState : []}
                allowSearch={true}
            >
            </FaSelect>);
        }

        if (propState.type === 'datetime') {
            const parsedValues = (valuesState ? (typeof valuesState === 'string')? JSON.parse(valuesState) : valuesState : {});
            const fromRange = parsedValues.fr? parsedValues.fr : parsedValues.from;
            const dateRange = parseDateRangeFilter(fromRange, parsedValues.to, parsedValues);
            const rang = {
                startDate: dateRange.from,
                endDate: dateRange.to,
                num: dateRange.num,
                gran: dateRange.gran
            }

            selectionComponent = selectDateTimeSelector(isArray(operatorState)? operatorState[0] : operatorState, rang);
        }

        if (propState.type === 'numerical') {
            selectionComponent = (<InputNumber value={valuesState} onBlur={emitFilter} onChange={setNumericalValue}></InputNumber>);
        }

        return (
          <div className={`${styles.filter__propContainer} ml-4`}>
            {propState.type === 'categorical' ? (
              <>
                {' '}
                <Tooltip
                  title={
                    valuesState && valuesState.length
                      ? valuesState
                          .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                          .join(', ')
                      : null
                  }
                >
                  <Button
                    className={`fa-button--truncate fa-button--truncate-lg`}
                    type='link'
                    onClick={() => setValuesSelectionOpen(!valuesSelectionOpen)}
                  >
                    {' '}
                    {valuesState && valuesState.length
                      ? valuesState
                          .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                          .join(', ')
                      : 'Select Values'}
                  </Button>{' '}
                </Tooltip>{' '}
                {valuesSelectionOpen && selectionComponent}{' '}
              </>
            ) : null}

            {propState.type !== 'categorical' ? selectionComponent : null}
          </div>
        );

    }

    return (<div className={styles.filter}>
        {renderPropSelect()}

        {propState?.name ? renderOperatorSelector() : null}

        {operatorState? renderValuesSelector() : null}

    </div>);

}

export default FAFilterSelect;