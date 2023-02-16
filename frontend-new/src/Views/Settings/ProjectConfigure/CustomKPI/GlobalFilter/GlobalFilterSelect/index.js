import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, InputNumber, Tooltip, Select, DatePicker, Input } from 'antd';
// import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import GroupSelect2 from '../../GroupSelect2';
import FaDatepicker from 'Components/FaDatepicker';
import FaSelect from 'Components/FaSelect';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import moment from 'moment';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import { DISPLAY_PROP, OPERATORS } from 'Utils/constants';
import { TOOLTIP_CONSTANTS } from '../../../../../../constants/tooltips.constans';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

const { Option } = Select;

const rangePicker = [OPERATORS['equalTo'], OPERATORS['notEqualTo']];
const customRangePicker = [OPERATORS['between'], OPERATORS['notBetween']];
const deltaPicker = [OPERATORS['inThePrevious'], OPERATORS['notInThePrevious']];
const currentPicker = [OPERATORS['inTheCurrent'], OPERATORS['notInTheCurrent']];
const datePicker = [OPERATORS['before'], OPERATORS['since']];

const GlobalFilterSelect = ({
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = [],
  setValuesByProps,
  applyFilter,
  filter
}) => {
  const [propState, setPropState] = useState({
    icon: '',
    name: '',
    type: ''
  });

  const [operatorState, setOperatorState] = useState(OPERATORS['equalTo']);
  const [valuesState, setValuesState] = useState(null);

  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [operSelectOpen, setOperSelectOpen] = useState(false);
  const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);
  const [grnSelectOpen, setGrnSelectOpen] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [containButton, setContainButton] = useState(true);

  const [updateState, updateStateApply] = useState(false);
  const [eventFilterInfo, seteventFilterInfo] = useState(null);
  const { userPropNames, eventPropNames } = useSelector(
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
        if (filter.operator === OPERATORS['equalTo'])
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

  useEffect(() => {
    if (
      operatorState?.[0] === OPERATORS['isKnown'] ||
      operatorState?.[0] === OPERATORS['isUnknown']
    ) {
      valuesSelectSingle(['$none']);
    }
  }, [operatorState]);

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
        extra: eventFilterInfo ? eventFilterInfo : null
      });
    }
  };

  const operatorSelect = (op) => {
    setOperatorState(op);
    setValuesState(null);
    setOperSelectOpen(false);
  };

  const renderDisplayName = (propState) => {
    // propState?.name ? userPropNames[propState?.name] ? userPropNames[propState?.name] : propState?.name : 'Select Property'
    let propertyName = '';
    // if(propState.name && (propState.icon === 'user' || propState.icon === 'user_g')) {
    //   propertyName = userPropNames[propState.name]?  userPropNames[propState.name] : propState.name;
    // }
    // if(propState.name && propState.icon === 'event') {
    //   propertyName = eventPropNames[propState.name]?  eventPropNames[propState.name] : propState.name;
    // }

    propertyName = _.startCase(propState?.name);

    if (!propState.name) {
      propertyName = 'Select Property';
    }
    return propertyName;
  };

  const propSelect = (label, val, cat) => {
    let prop = [label, ...val];
    setPropState({ icon: prop[0], name: prop[1], type: prop[3], extra: val });
    setPropSelectOpen(false);
    setOperatorState(
      prop[3] === 'datetime' ? OPERATORS['between'] : OPERATORS['equalTo']
    );
    setValuesState(null);
    setValuesByProps([...val]);
    seteventFilterInfo(val);
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
    let propertyName = '';
    propertyName = _.startCase(propState?.name);

    if (!propState.name) {
      propertyName = 'Select Property';
    }
    return propertyName;
  };

  const renderPropSelect = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Tooltip
          title={renderGroupDisplayName(propState)}
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            icon={
              propState && propState.icon ? (
                <SVG name={propState.icon} size={16} color={'purple'} />
              ) : null
            }
            className={`fa-button--truncate fa-button--truncate-xs mr-2`}
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
              optionClick={(label, val, cat) => propSelect(label, val, cat)}
              onClickOutside={() => setPropSelectOpen(false)}
              hideTitle={true}
            ></GroupSelect2>
          </div>
        )}
      </div>
    );
  };

  const renderOperatorSelector = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Tooltip
          title='Select an equator to define your filter rules.'
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            className={`mr-2`}
            type='link'
            onClick={() => setOperSelectOpen(true)}
          >
            {operatorState ? operatorState : 'Select Operator'}
          </Button>
        </Tooltip>

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
        multiSelect={
          (isArray(operatorState) ? operatorState[0] : operatorState) ===
            '!=' ||
          (isArray(operatorState) ? operatorState[0] : operatorState) ===
            'does not contain'
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
        endDate: dateRange.to
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
              className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
              type='link'
              onClick={() => setContainButton(false)}
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
              className={`input-value filter-buttons-radius filter-buttons-margin`}
            ></Input>
          )}
        </div>
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
            color={TOOLTIP_CONSTANTS.DARK}
          >
            <Button
              className={`fa-button--truncate`}
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

export default GlobalFilterSelect;
