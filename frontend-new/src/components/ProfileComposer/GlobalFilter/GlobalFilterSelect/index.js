import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { Button, InputNumber, Tooltip, Select, DatePicker } from 'antd';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import FaDatepicker from 'Components/FaDatepicker';
import FaSelect from 'Components/FaSelect';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import moment from 'moment';
import { DEFAULT_OPERATOR_PROPS,dateTimeSelect } from 'Components/FaFilterSelect/utils';
import { DISPLAY_PROP } from '../../../../utils/constants';
import { toCapitalCase } from '../../../../utils/global';


const defaultOpProps = DEFAULT_OPERATOR_PROPS;

const { Option } = Select;

const rangePicker = ['=', '!='];
const customRangePicker = ['between', 'not between'];
const deltaPicker = ['in the previous', 'not in the previous'];
const currentPicker = ['in the current', 'not in the current'];
const datePicker = ['before', 'since'];

const GlobalFilterSelect = ({
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = [],
  setValuesByProps,
  applyFilter,
  filter,
  refValue,
}) => {
  const [propState, setPropState] = useState({
    icon: '',
    name: '',
    type: '',
  });

  const [operatorState, setOperatorState] = useState('=');
  const [valuesState, setValuesState] = useState(null);

  const [propSelectOpen, setPropSelectOpen] = useState(true);
  const [operSelectOpen, setOperSelectOpen] = useState(false);
  const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);
  const [grnSelectOpen, setGrnSelectOpen] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);

  const [updateState, updateStateApply] = useState(false);

  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const [dateOptionSelectOpen,setDateOptionSelectOpen]=useState(false);

  useEffect(() => {
    if (
      currentPicker.includes(
        isArray(operatorState) ? operatorState[0] : operatorState
      )
    ) {
      setCurrentFilt();
    }
    if (
      deltaPicker.includes(
        isArray(operatorState) ? operatorState[0] : operatorState
      )
    ) {
      setDeltaFilt();
    }
  }, [valuesState]);

  useEffect(() => {
    if (
      (filter && !valuesState) ||
      (filter && filter?.values !== valuesState)
    ) {
      const prop = filter.props;
      setPropState({ icon: prop[2], name: prop[0], type: prop[1] });
      setOperatorState(filter.operator);
      // Set values state
      setValues();
      setPropSelectOpen(false);
      setOperSelectOpen(false);
      setValuesSelectionOpen(false);
    }
  }, [filter]);

  useEffect(() => {
    if (updateState && valuesState && propState.type !== 'numerical') {
      emitFilter();
      updateStateApply(false);
    }
  }, [updateState]);

  const setValues = () => {
    let values;
    if (filter.props[1] === 'datetime') {
      const filterVals = isArray(filter.values)
        ? filter.values[0]
        : filter.values;
      const parsedValues = filterVals
        ? typeof filterVals === 'string'
          ? JSON.parse(filterVals)
          : filterVals
        : {};
      values = parseDateRangeFilter(
        parsedValues.fr,
        parsedValues.to,
        parsedValues
      );
    } else {
      values = filter.values;
    }
    if (
      currentPicker.includes(
        isArray(filter.operator) ? filter.operator[0] : filter.operator
      ) ||
      deltaPicker.includes(
        isArray(filter.operator) ? filter.operator[0] : filter.operator
      )
    ) {
      setValuesState(JSON.stringify(values));
    } else {
      setValuesState(values);
    }
  };

  const emitFilter = () => {
    if (propState && operatorState && valuesState) {
      applyFilter({
        props: [propState.name, propState.type, propState.icon],
        operator: operatorState,
        values: valuesState,
        ref: refValue,
      });
    }
  };

  const operatorSelect = (op) => {
    setOperatorState(op);
    setValuesState(null);
    setOperSelectOpen(false);
  };

  const renderDisplayName = (propState) => {
    let propertyName = '';
    if (
      propState.name &&
      (propState.icon === 'user' || propState.icon === 'user_g')
    ) {
      propertyName = userPropNames[propState.name]
        ? userPropNames[propState.name]
        : propState.name;
    }
    if (propState.name && propState.icon === 'event') {
      propertyName = eventPropNames[propState.name]
        ? eventPropNames[propState.name]
        : propState.name;
    }
    if (!propState.name) {
      propertyName = 'Select Property';
    }
    return propertyName;
  };

  const propSelect = (prop) => {
    setPropState({ icon: prop[3], name: prop[1], type: prop[2] });
    setPropSelectOpen(false);
    setOperatorState(prop[2] === 'datetime' ? 'between' : '=');
    setValuesState(null);
    setValuesByProps(prop);
  };

  const valuesSelect = (val) => {
    setValuesState(val.map((vl) => JSON.parse(vl)[0]));
    setValuesSelectionOpen(false);
    updateStateApply(true);
  };

  const onDateSelect = (rng) => {
    let startDate;
    let endDate;
    if (isArray(rng.startDate)) {
      startDate = rng.startDate[0].toDate().getTime();
      endDate = rng.startDate[1].toDate().getTime();
    } else {
      if (rng.startDate && rng.startDate._isAMomentObject) {
        startDate = rng.startDate.toDate().getTime();
      } else {
        startDate = rng.startDate.getTime();
      }

      if (rng.endDate && rng.endDate._isAMomentObject) {
        endDate = rng.endDate.toDate().getTime();
      } else {
        endDate = rng.endDate.getTime();
      }
    }

    const rangeValue = {
      fr: startDate,
      to: endDate,
      ovp: false,
    };

    setValuesState(JSON.stringify(rangeValue));
    updateStateApply(true);
  };
  const setNumericalValue = (ev) => {
    // onNumericalSelect(ev);

    setValuesState(String(ev).toString());
  };

  const parseDateRangeFilter = (fr, to, value) => {
    const fromVal = fr ? fr : new Date(MomentTz().startOf('day')).getTime();
    const toVal = to ? to : new Date(MomentTz()).getTime();
    return {
      from: fromVal,
      to: toVal,
      ovp: false,
      num: value['num'],
      gran: value['gran'],
    };
    // return (MomentTz(fromVal).format('MMM DD, YYYY') + ' - ' +
    //           MomentTz(toVal).format('MMM DD, YYYY'));
  };

  const renderGroupDisplayName = (propState) => {
    let propertyName = '';
    if (!propState.name) {
      propertyName = 'Select Property';
    } else {
      propertyName = propState.name;
    }
    return propertyName;
  };

  const renderPropSelect = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Tooltip title={renderGroupDisplayName(propState)}>
          <Button
            icon={
              propState && propState.icon ? (
                <SVG name={propState.icon} size={16} color={'purple'} />
              ) : null
            }
            className={`fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin`}
            type='link'
            onClick={() => setPropSelectOpen(!propSelectOpen)}
          >
            {renderDisplayName(propState)}
          </Button>
        </Tooltip>

        {propSelectOpen && (
          <div className={styles.filter__event_selector}>
            <GroupSelect2
              groupedProperties={propOpts}
              placeholder='Select Property'
              optionClick={(group, val) => propSelect([...val, group])}
              onClickOutside={() => setPropSelectOpen(false)}
            ></GroupSelect2>
          </div>
        )}
      </div>
    );
  };

  const renderOperatorSelector = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Button
          className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
          type='link'
          onClick={() => setOperSelectOpen(true)}
        >
          {operatorState ? operatorState : 'Select Operator'}
        </Button>

        {operSelectOpen && (
          <FaSelect
            options={operatorOpts[propState.type].map((op) => [op])}
            optionClick={(val) => operatorSelect(val)}
            onClickOutside={() => setOperSelectOpen(false)}
          ></FaSelect>
        )}
      </div>
    );
  };

  const setDeltaNumber = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues['num'] = val;
    setValuesState(JSON.stringify(parsedValues));
    updateStateApply(true);
  };

  const setDeltaGran = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues['gran'] = val;
    setValuesState(JSON.stringify(parsedValues));
    setGrnSelectOpen(false);
    setDateOptionSelectOpen(false);
    setDeltaFilt();
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

  const setCurrentGran = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues['gran'] = val;
    setValuesState(JSON.stringify(parsedValues));
    setGrnSelectOpen(false);
    setDateOptionSelectOpen(false);
    setCurrentFilt();
  };

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

  const onDatePickerSelect = (val) => {
    let dateT;
    let dateValue = {};
    const operatorSt = isArray(operatorState)
      ? operatorState[0]
      : operatorState;
    if (operatorSt === 'before') {
      dateT = MomentTz(val).startOf('day');
      dateValue['to'] = dateT.toDate().getTime();
    }

    if (operatorSt === 'since') {
      dateT = MomentTz(val).startOf('day');
      dateValue['fr'] = dateT.toDate().getTime();
    }

    setValuesState(JSON.stringify(dateValue));
    updateStateApply(true);
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
          className={'filter-buttons-margin filter-buttons-radius'}
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
          className={'filter-buttons-margin filter-buttons-radius'}
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
            placeholder={'number'}
            controls={false}
            className={'filter-buttons-radius date-input-number'}
          ></InputNumber>

          <Button
          className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
          type='link'
          onClick={() => setDateOptionSelectOpen(true)}
          >
          {parsedValues['gran'] ? dateTimeSelect.get(parsedValues['gran']) : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={[['Days'],['Weeks'],['Months'],['Quarters']]}
              optionClick={(val) => setDeltaGran(dateTimeSelect.get(val[0]))}
              onClickOutside={() => setDateOptionSelectOpen(false)}
            ></FaSelect>
          )}
      </div>
      );
    }

    if (currentPicker.includes(operator)) {
      selectorComponent = (
        <div className={`fa-filter-dateDeltaContainer`}>
          <Button
          className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
          type='link'
          onClick={() => setDateOptionSelectOpen(true)}
          >
          {parsedValues['gran'] ? toCapitalCase(parsedValues['gran']) : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={[['Week'],['Month'],['Quarter']]}
              optionClick={(val) => setCurrentGran(val[0].toLowerCase())}
              onClickOutside={() => setDateOptionSelectOpen(false)}
            ></FaSelect>
          )}
        </div>
      );
    }


    if (datePicker.includes(operator)) {
      selectorComponent = (
        <DatePicker
          // disabledDate={(d) => !d || d.isAfter(MomentTz())}
          autoFocus={false}
          className={`fa-date-picker`}
          open={showDatePicker}
          onOpenChange={() => {
            setShowDatePicker(!showDatePicker);
          }}
          value={
            operator === 'before'
              ? moment(parsedValues['to'])
              : moment(
                  parsedValues['from']
                    ? parsedValues['from']
                    : parsedValues['fr']
                )
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
    let selectionComponent = null;
    const values = [];

    selectionComponent = (
      <FaSelect
        multiSelect={true}
        options={
          valueOpts && valueOpts[propState.name]?.length
            ? valueOpts[propState.name].map((op) => [op])
            : []
        }
        applClick={(val) => valuesSelect(val)}
        onClickOutside={() => setValuesSelectionOpen(false)}
        selectedOpts={valuesState ? valuesState : []}
        allowSearch={true}
      ></FaSelect>
    );

    if (propState.type === 'datetime') {
      const parsedValues = valuesState
        ? typeof valuesState === 'string'
          ? JSON.parse(valuesState)
          : valuesState
        : {};
      const fromRange = parsedValues.fr ? parsedValues.fr : parsedValues.from;
      const dateRange = parseDateRangeFilter(
        fromRange,
        parsedValues.to,
        parsedValues
      );
      const rang = {
        startDate: dateRange.from,
        endDate: dateRange.to,
      };

      selectionComponent = selectDateTimeSelector(
        isArray(operatorState) ? operatorState[0] : operatorState,
        rang
      );
    }

    if (propState.type === 'numerical') {
      selectionComponent = (
        <InputNumber
          value={valuesState}
          onBlur={emitFilter}
          onChange={setNumericalValue}
        ></InputNumber>
      );
    }
    if (!operatorState || !propState?.name) return null;

    return (
      <div className={`${styles.filter__propContainer} w-7/12`}>
        {propState.type === 'categorical' ? (
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
                className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
                type='link'
              onClick={() => setValuesSelectionOpen(!valuesSelectionOpen)}
            >
              {valuesState && valuesState.length
                ? valuesState
                    .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                    .join(', ')
                : 'Select Values'}
            </Button>
          </Tooltip>
        ) : null}

        {valuesSelectionOpen || propState.type !== 'categorical'
          ? selectionComponent
          : null}
      </div>
    );
  };

  return (
    <div className={styles.filter}>
      {renderPropSelect()}

      {propState?.name ? renderOperatorSelector() : null}

      {operatorState ? renderValuesSelector() : null}
    </div>
  );
};

export default GlobalFilterSelect;
