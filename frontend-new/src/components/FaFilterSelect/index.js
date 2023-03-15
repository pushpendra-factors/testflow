import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { Button, Input, InputNumber, Tooltip, DatePicker, Select } from 'antd';
import GroupSelect2 from '../QueryComposer/GroupSelect2';
import FaDatepicker from '../FaDatepicker';
import FaSelect from '../FaSelect';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import {
  DEFAULT_OPERATOR_PROPS,
  dateTimeSelect
} from 'Components/FaFilterSelect/utils';
import moment from 'moment';
import { DISPLAY_PROP, OPERATORS } from '../../utils/constants';
import { toCapitalCase } from '../../utils/global';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

const { Option } = Select;

const rangePicker = [OPERATORS['equalTo'], OPERATORS['notEqualTo']];
const customRangePicker = [OPERATORS['between'], OPERATORS['notBetween']];
const deltaPicker = [OPERATORS['inThePrevious'], OPERATORS['notInThePrevious']];
const currentPicker = [OPERATORS['inTheCurrent'], OPERATORS['notInTheCurrent']];
const datePicker = [OPERATORS['before'], OPERATORS['since']];

const FAFilterSelect = ({
  displayMode,
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = [],
  setValuesByProps,
  applyFilter,
  filter,
  disabled = false,
  refValue,
  caller,
  propsDDPos,
  propsDDHeight,
  operatorDDPos,
  operatorDDHeight,
  valuesDDPos,
  valuesDDHeight
}) => {
  const [propState, setPropState] = useState({
    icon: '',
    name: '',
    type: ''
  });

  const [operatorState, setOperatorState] = useState(OPERATORS['equalTo']);
  const [valuesState, setValuesState] = useState(null);

  const [propSelectOpen, setPropSelectOpen] = useState(true);
  const [operSelectOpen, setOperSelectOpen] = useState(false);
  const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);
  const [grnSelectOpen, setGrnSelectOpen] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [dateOptionSelectOpen, setDateOptionSelectOpen] = useState(false);
  const [containButton, setContainButton] = useState(true);

  const [updateState, updateStateApply] = useState(false);

  const { userPropNames, eventPropNames, groupPropNames } = useSelector(
    (state) => state.coreQuery
  );

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
      operatorState === OPERATORS['isKnown'] ||
      operatorState === OPERATORS['isUnknown'] ||
      operatorState?.[0] === OPERATORS['isKnown'] ||
      operatorState?.[0] === OPERATORS['isUnknown']
    ) {
      valuesSelectSingle(['$none']);
    }
  }, [operatorState]);

  useEffect(() => {
    if (
      (filter && !valuesState) ||
      (filter && filter?.values !== valuesState)
    ) {
      const prop = filter.props;
      setPropState({ icon: prop[2], name: prop[0], type: prop[1] });
      if (
        (filter.operator === OPERATORS['equalTo'] ||
          filter.operator === OPERATORS['notEqualTo'] ||
          filter.operator?.[0] === OPERATORS['equalTo'] ||
          filter.operator?.[0] === OPERATORS['notEqualTo']) &&
        filter.values?.[0] === '$none'
      ) {
        if (
          filter.operator === OPERATORS['equalTo'] ||
          filter.operator?.[0] === OPERATORS['equalTo']
        )
          setOperatorState(OPERATORS['isUnknown']);
        else setOperatorState(OPERATORS['isKnown']);
      } else {
        setOperatorState(filter.operator);
      }
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
        ref: refValue
      });
    }
  };

  const operatorSelect = (op) => {
    setOperatorState(op);
    setValuesState(null);
    setOperSelectOpen(false);
  };

  const propSelect = (prop) => {
    setPropState({ icon: prop[3], name: prop[1], type: prop[2] });
    setPropSelectOpen(false);
    setOperatorState(
      prop[2] === 'datetime' ? OPERATORS['between'] : OPERATORS['equalTo']
    );
    setValuesState(null);
    setValuesByProps(prop);
    setValuesSelectionOpen(true);
  };

  const valuesSelect = (val) => {
    setValuesState(val.map((vl) => JSON.parse(vl)[0]));
    setValuesSelectionOpen(false);
    updateStateApply(true);
  };

  const valuesSelectSingle = (val) => {
    setValuesState(val);
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
      ovp: false
    };

    setValuesState(JSON.stringify(rangeValue));
    updateStateApply(true);
  };
  const setNumericalValue = (ev) => {
    // onNumericalSelect(ev);

    setValuesState(String(ev.target.value).toString());
  };

  const parseDateRangeFilter = (fr, to, value) => {
    const fromVal = fr ? fr : new Date(MomentTz().startOf('day')).getTime();
    const toVal = to ? to : new Date(MomentTz()).getTime();
    return {
      from: fromVal,
      to: toVal,
      ovp: false,
      num: value['num'],
      gran: value['gran']
    };
    // return (MomentTz(fromVal).format('MMM DD, YYYY') + ' - ' +
    //           MomentTz(toVal).format('MMM DD, YYYY'));
  };

  const renderGroupDisplayName = (propState) => {
    // propState?.name ? userPropNames[propState?.name] ? userPropNames[propState?.name] : propState?.name : 'Select Property'
    let propertyName = propState?.name;
    if (propState.name && propState.icon === 'group') {
      propertyName = groupPropNames[propState.name]
        ? groupPropNames[propState.name]
        : propState.name;
    }
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

  const renderPropSelect = () => {
    return (
      <div
        className={`${styles.filter__propContainer} ${
          disabled ? `fa-truncate-150` : ''
        }`}
      >
        <Tooltip
          title={renderGroupDisplayName(propState)}
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            disabled={disabled}
            icon={
              propState && propState.icon ? (
                <SVG
                  name={propState.icon}
                  size={16}
                  color={displayMode ? 'grey' : 'purple'}
                />
              ) : null
            }
            className={`fa-button--truncate fa-button--truncate-xs ${
              displayMode ? 'static-button' : ''
            }  btn-left-round filter-buttons-margin`}
            type={displayMode ? 'default' : 'link'}
            onClick={() =>
              displayMode ? null : setPropSelectOpen(!propSelectOpen)
            }
          >
            {renderGroupDisplayName(propState)}
          </Button>
        </Tooltip>
        {propSelectOpen && (
          <div className={styles.filter__event_selector}>
            <GroupSelect2
              groupedProperties={propOpts}
              placeholder='Select Property'
              optionClick={(group, val) => propSelect([...val, group])}
              onClickOutside={() => setPropSelectOpen(false)}
              placement={propsDDPos}
              height={propsDDHeight}
            />
          </div>
        )}
      </div>
    );
  };

  const renderOperatorSelector = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Tooltip
          title='Select an equator to define your filter rules. '
          color={TOOLTIP_CONSTANTS.DARK}
          trigger={displayMode ? [] : 'hover'}
        >
          <Button
            disabled={disabled}
            className={`fa-button--truncate ${
              displayMode ? 'static-button' : ''
            } filter-buttons-radius filter-buttons-margin`}
            type={displayMode ? 'default' : 'link'}
            onClick={() => (displayMode ? null : setOperSelectOpen(true))}
          >
            {operatorState ? operatorState : 'Select Operator'}
          </Button>
        </Tooltip>

        {operSelectOpen && (
          <FaSelect
            options={operatorOpts[propState.type].map((op) => [op])}
            optionClick={(val) => operatorSelect(val)}
            onClickOutside={() => setOperSelectOpen(false)}
            placement={operatorDDPos}
          />
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
    if (operatorSt === OPERATORS['before']) {
      dateT = MomentTz(val).startOf('day');
      dateValue['to'] = dateT.toDate().getTime();
    }

    if (operatorSt === OPERATORS['since']) {
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
          disabled={disabled}
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
          disabled={disabled}
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
            disabled={disabled}
            placeholder={'number'}
            controls={false}
            className={'filter-buttons-radius date-input-number'}
          ></InputNumber>

          <Button
            disabled={disabled}
            trigger={displayMode ? [] : 'hover'}
            className={`fa-button--truncate ${
              displayMode ? 'static-button' : ''
            } filter-buttons-radius filter-buttons-margin`}
            type={displayMode ? 'default' : 'link'}
            onClick={() => (displayMode ? null : setDateOptionSelectOpen(true))}
          >
            {parsedValues['gran']
              ? dateTimeSelect.get(parsedValues['gran'])
              : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={[['Days'], ['Weeks'], ['Months'], ['Quarters']]}
              optionClick={(val) => setDeltaGran(dateTimeSelect.get(val[0]))}
              onClickOutside={() => setDateOptionSelectOpen(false)}
              placement={valuesDDPos}
            />
          )}
        </div>
      );
    }

    if (currentPicker.includes(operator)) {
      selectorComponent = (
        <div className={`fa-filter-dateDeltaContainer`}>
          <Button
            disabled={disabled}
            trigger={displayMode ? [] : 'hover'}
            className={`fa-button--truncate ${
              displayMode ? 'static-button' : ''
            } filter-buttons-radius filter-buttons-margin`}
            type={displayMode ? 'default' : 'link'}
            onClick={() => (displayMode ? null : setDateOptionSelectOpen(true))}
          >
            {parsedValues['gran']
              ? toCapitalCase(parsedValues['gran'])
              : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={[['Week'], ['Month'], ['Quarter']]}
              optionClick={(val) => setCurrentGran(val[0].toLowerCase())}
              onClickOutside={() => setDateOptionSelectOpen(false)}
              placement={valuesDDPos}
            />
          )}
        </div>
      );
    }

    if (datePicker.includes(operator)) {
      selectorComponent = (
        <DatePicker
          disabled={disabled}
          // disabledDate={(d) => !d || d.isAfter(MomentTz())}
          autoFocus={false}
          className={`fa-date-picker`}
          open={showDatePicker}
          onOpenChange={() => {
            setShowDatePicker(!showDatePicker);
          }}
          value={
            operator === OPERATORS['before']
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
    let selectionComponent;
    const values = [];
    if (propState.type === 'categorical') {
      selectionComponent = (
        <FaSelect
          multiSelect={
            (isArray(operatorState) ? operatorState[0] : operatorState) ===
              OPERATORS['notEqualTo'] ||
            (isArray(operatorState) ? operatorState[0] : operatorState) ===
              OPERATORS['doesNotContain']
              ? false
              : true
          }
          options={
            valueOpts && valueOpts[propState.name]?.length
              ? valueOpts[propState.name].map((op) => [op])
              : []
          }
          applClick={(val) => valuesSelect(val)}
          optionClick={(val) => valuesSelectSingle(val)}
          onClickOutside={() => setValuesSelectionOpen(false)}
          selectedOpts={valuesState ? valuesState : []}
          allowSearch={true}
          placement={valuesDDPos}
        />
      );
    }

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
        num: dateRange.num,
        gran: dateRange.gran
      };

      selectionComponent = selectDateTimeSelector(
        isArray(operatorState) ? operatorState[0] : operatorState,
        rang
      );
    }

    if (propState.type === 'numerical') {
      selectionComponent = (
        <div>
          {containButton && (
            <Button
              disabled={disabled}
              trigger={displayMode ? [] : 'hover'}
              className={`fa-button--truncate ${
                displayMode ? 'static-button' : ''
              } filter-buttons-radius filter-buttons-margin`}
              type={displayMode ? 'default' : 'link'}
              onClick={() => (displayMode ? null : setContainButton(false))}
            >
              {valuesState ? valuesState : 'Enter Value'}
            </Button>
          )}
          {!containButton && (
            <Input
              type='number'
              value={valuesState}
              placeholder={'Enter Value'}
              autoFocus={true}
              onBlur={() => {
                emitFilter();
                setContainButton(true);
              }}
              onPressEnter={() => {
                emitFilter();
                setContainButton(true);
              }}
              onChange={setNumericalValue}
              disabled={disabled}
              className={`input-value filter-buttons-radius filter-buttons-margin`}
            ></Input>
          )}
        </div>
      );
    }

    return (
      <div
        className={`${styles.filter__propContainer}${
          disabled ? `fa-truncate-150` : ''
        }`}
      >
        {propState.type === 'categorical' ? (
          <>
            <Tooltip
              mouseLeaveDelay={0}
              title={
                valuesState && valuesState.length
                  ? valuesState
                      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                      .join(', ')
                  : null
              }
              color={TOOLTIP_CONSTANTS.DARK}
            >
              <Button
                className={`fa-button--truncate ${
                  caller === 'profiles' ? 'fa-button--truncate-sm' : ''
                }  ${
                  displayMode
                    ? 'btn-right-round static-button'
                    : 'filter-buttons-radius'
                } filter-buttons-margin`}
                type={displayMode ? 'default' : 'link'}
                disabled={disabled}
                onClick={() =>
                  displayMode
                    ? null
                    : setValuesSelectionOpen(!valuesSelectionOpen)
                }
              >
                {valuesState && valuesState.length
                  ? valuesState
                      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                      .join(', ')
                  : 'Select Values'}
              </Button>
            </Tooltip>
            {valuesSelectionOpen && selectionComponent}
          </>
        ) : null}

        {propState.type !== 'categorical' ? selectionComponent : null}
      </div>
    );
  };

  return (
    <div className={styles.filter}>
      {renderPropSelect()}

      {propState?.name ? renderOperatorSelector() : null}

      {operatorState &&
      operatorState !== OPERATORS['isKnown'] &&
      operatorState !== OPERATORS['isUnknown'] &&
      operatorState?.[0] !== OPERATORS['isKnown'] &&
      operatorState?.[0] !== OPERATORS['isUnknown']
        ? renderValuesSelector()
        : null}
    </div>
  );
};

export default FAFilterSelect;
