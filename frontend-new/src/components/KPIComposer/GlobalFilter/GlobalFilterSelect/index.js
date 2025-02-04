import React, { useState, useEffect, useMemo } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { Button, InputNumber, Tooltip, Select, DatePicker, Input } from 'antd';
import FaDatepicker from 'Components/FaDatepicker';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import moment from 'moment';
import _ from 'lodash';
import {
  DEFAULT_OPERATOR_PROPS,
  dateTimeSelect
} from 'Components/FaFilterSelect/utils';
import { toCapitalCase } from '../../../../utils/global';

import { TOOLTIP_CONSTANTS } from '../../../../constants/tooltips.constans';
import { DISPLAY_PROP, OPERATORS } from 'Utils/constants';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import FaSelect from 'Components/GenericComponents/FaSelect';
import { selectedOptionsMapper } from 'Components/GenericComponents/FaSelect/utils';
import { processProperties } from 'Utils/dataFormatter';
import truncateURL from 'Utils/truncateURL';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

const { Option } = Select;

const GlobalFilterSelect = ({
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = {},
  valueOptsLoading,
  setValuesByProps,
  applyFilter,
  filter,
  refValue,
  viewMode = false,
  maxAllowedSelection
}) => {
  const rangePicker = [OPERATORS['equalTo'], OPERATORS['notEqualTo']];
  const customRangePicker = [OPERATORS['between'], OPERATORS['notBetween']];
  const deltaPicker = [
    OPERATORS['inThePrevious'],
    OPERATORS['notInThePrevious']
  ];
  const currentPicker = [
    OPERATORS['inTheCurrent'],
    OPERATORS['notInTheCurrent']
  ];
  const datePicker = [OPERATORS['before'], OPERATORS['since']];

  const [propState, setPropState] = useState({
    icon: '',
    name: '',
    type: ''
  });

  const valueDisplayNames = useMemo(() => {
    return valueOpts?.[propState?.extra?.[1]]
      ? valueOpts?.[propState?.extra?.[1]]
      : DISPLAY_PROP;
  }, [valueOpts, propState.extra]);

  const [operatorState, setOperatorState] = useState(OPERATORS['between']);
  const [valuesState, setValuesState] = useState(null);

  const [propSelectOpen, setPropSelectOpen] = useState(true);
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
  const { projectDomainsList } = useSelector((state) => state.global);

  const [dateOptionSelectOpen, setDateOptionSelectOpen] = useState(false);

  useEffect(() => {
    if (filter) {
      const prop = filter.props;
      setPropState({
        icon: prop[2],
        name: prop[0],
        type: prop[1],
        extra: filter?.extra
      });
      if (
        (filter.operator === OPERATORS['equalTo'] ||
          filter.operator === OPERATORS['notEqualTo']) &&
        filter.values?.length === 1 &&
        filter.values?.[0] === '$none'
      ) {
        if (filter.operator === OPERATORS['equalTo'])
          setOperatorState(OPERATORS['isUnknown']);
        else setOperatorState(OPERATORS['isKnown']);
      } else {
        setOperatorState(filter.operator);
      }
      seteventFilterInfo(filter?.extra);
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
      operatorState === OPERATORS['isKnown'] ||
      operatorState === OPERATORS['isUnknown']
    ) {
      valuesSelectSingle(['$none']);
    }
  }, [operatorState]);

  const setValues = () => {
    let values;
    if (filter.props[1] === 'datetime') {
      const parsedValues = filter.values[0]
        ? typeof filter.values[0] === 'string'
          ? JSON.parse(filter.values)
          : filter.values
        : {};
      values = parseDateRangeFilter(
        parsedValues.fr,
        parsedValues.to,
        parsedValues
      );
    } else {
      values = filter.values;
    }
    setValuesState(values);
  };

  const emitFilter = () => {
    if (propState && operatorState && valuesState) {
      applyFilter({
        props: [propState.name, propState.type, propState.icon],
        operator: operatorState,
        values: valuesState,
        ref: refValue,
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

    propertyName = eventPropNames[propState?.name]
      ? eventPropNames[propState?.name]
      : propState?.name;
    // propertyName = _.startCase(propState?.name);

    if (!propState.name) {
      propertyName = 'Select Property';
    }
    return propertyName;
  };

  const propSelect = (option, group) => {
    const valueType = option.extraProps.valueType;
    const valuecategory = option.extraProps.queryType;
    const valueArray = [option.label, option.value, valueType, valuecategory];
    setPropState({
      icon: option.extraProps?.propertyType || group.iconName,
      name: option.label,
      type: valueType,
      extra: valueArray
    });
    setPropSelectOpen(false);
    setOperatorState(
      valueType === 'datetime' ? OPERATORS['between'] : OPERATORS['equalTo']
    );
    setValuesState(null);
    setValuesByProps([...valueArray]);
    seteventFilterInfo(valueArray);
    setValuesSelectionOpen(true);
  };

  const valuesSelect = (val) => {
    setValuesState(val);
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

  const getGroupLabel = (grp) => {
    if (grp === 'event') return 'Event Properties';
    if (grp === 'user') return 'User Properties';
    if (!grp || !grp.length) return 'Properties';
    return grp;
  };

  const renderGroupDisplayName = (propState) => {
    let propertyName = '';
    if (!propState.name) {
      propertyName = 'Select Property';
    } else {
      propertyName = _.startCase(propState.name);
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
            // icon={propState && propState.icon ? <SVG name={propState.icon} size={16} color={'purple'} /> : null}
            className={`fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin`}
            type='link'
            onClick={() => setPropSelectOpen(!propSelectOpen)}
            disabled={viewMode}
          >
            {renderDisplayName(propState)}
          </Button>
        </Tooltip>

        {propSelectOpen && (
          <div className={styles.filter__event_selector}>
            <GroupSelect
              options={propOpts?.map((groupOpt) => {
                return {
                  iconName: groupOpt?.icon,
                  label: getGroupLabel(groupOpt?.label),
                  values: processProperties(
                    groupOpt?.values,
                    groupOpt?.propertyType
                  )
                };
              })}
              onClickOutside={() => setPropSelectOpen(false)}
              placeholder='Select Property'
              allowSearch={true}
              optionClickCallback={propSelect}
              allowSearchTextSelection={false}
              extraClass={`${styles.filter__event_selector__select}`}
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
          title='Select an equator to define your filter rules.'
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
            type='link'
            onClick={() => setOperSelectOpen(true)}
            disabled={viewMode}
          >
            {operatorState ? operatorState : 'Select Operator'}
          </Button>
        </Tooltip>

        {operSelectOpen && (
          <FaSelect
            options={operatorOpts[propState.type]
              .filter(
                (op) =>
                  op !== OPERATORS['inList'] && op !== OPERATORS['notInList']
              )
              .map((op) => {
                return { value: op, label: op };
              })}
            optionClickCallback={(option) => operatorSelect(option.value)}
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
    const operatorSt = operatorState;
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
          disabled={viewMode}
          onSelect={(rng) => onDateSelect(rng)}
          className={'filter-buttons-margin filter-buttons-radius'}
        />
      );
    }

    if (customRangePicker.includes(operator)) {
      selectorComponent = (
        <FaDatepicker
          disabled={viewMode}
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
            disabled={viewMode}
            onChange={setDeltaNumber}
            placeholder={'number'}
            controls={false}
            className={'filter-buttons-radius date-input-number'}
          ></InputNumber>

          <Button
            className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
            type='link'
            disabled={viewMode}
            onClick={() => setDateOptionSelectOpen(true)}
          >
            {parsedValues['gran']
              ? dateTimeSelect.get(parsedValues['gran'])
              : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={['Days', 'Weeks', 'Months', 'Quarters'].map((option) => {
                return {
                  value: option,
                  label: option
                };
              })}
              optionClickCallback={(option) => {
                setDeltaGran(dateTimeSelect.get(option.value));
              }}
              onClickOutside={() => setDateOptionSelectOpen(false)}
            />
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
            disabled={viewMode}
            onClick={() => setDateOptionSelectOpen(true)}
          >
            {parsedValues['gran']
              ? toCapitalCase(parsedValues['gran'])
              : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={['Week', 'Month', 'Quarter'].map((option) => {
                return {
                  value: option,
                  label: option
                };
              })}
              optionClickCallback={(option) => {
                setCurrentGran(option.value.toLowerCase());
              }}
              onClickOutside={() => setDateOptionSelectOpen(false)}
            />
          )}
        </div>
      );
    }

    if (datePicker.includes(operator)) {
      selectorComponent = (
        <DatePicker
          // disabledDate={(d) => !d || d.isAfter(MomentTz())}
          disabled={viewMode}
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

  const formatCsvUploadValue = (value) => {
    if (
      operatorState === OPERATORS['inList'] ||
      operatorState?.[0] === OPERATORS['inList'] ||
      operatorState === OPERATORS['notInList'] ||
      operatorState?.[0] === OPERATORS['notInList']
    ) {
      const vl = value.split('_');
      let data = '';
      if (vl.length > 1) {
        data += vl[1];
        for (let i = 2; i < vl.length - 1; i++) {
          data = data + '_' + vl?.[i];
        }
        data = data + '.' + vl[vl.length - 1];
      } else {
        data = value;
      }
      return data;
    }

    return value;
  };

  const renderValuesSelector = () => {
    let selectionComponent = null;
    const values = [];

    if (propState.type === 'categorical') {
      const variant =
        operatorState === OPERATORS['notEqualTo'] ||
        operatorState === OPERATORS['doesNotContain']
          ? 'Single'
          : 'Multi';
      let valueOptions = valueOpts?.[propState?.extra?.[1]]
        ? Object.entries(valueOpts[propState?.extra?.[1]]).map((val) => {
            return {
              value: val[0],
              label: val[1]
            };
          })
        : [];
      valueOptions = selectedOptionsMapper(valueOptions, valuesState);

      if (variant === 'Single') {
        selectionComponent = (
          <FaSelect
            variant={'Single'}
            options={valueOptions}
            optionClickCallback={(option) => {
              valuesSelect([option.value]);
            }}
            onClickOutside={() => setValuesSelectionOpen(false)}
            allowSearch={true}
            loadingState={valueOptsLoading}
          />
        );
      } else {
        selectionComponent = (
          <FaSelect
            variant={'Multi'}
            options={valueOptions}
            applyClickCallback={(updatedOptions, selectedOptions) => {
              valuesSelect(selectedOptions);
            }}
            onClickOutside={() => setValuesSelectionOpen(false)}
            allowSearch={true}
            maxAllowedSelection={maxAllowedSelection}
            loadingState={valueOptsLoading}
          />
        );
      }
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
        endDate: dateRange.to
      };

      selectionComponent = selectDateTimeSelector(operatorState, rang);
    }

    if (propState.type === 'numerical') {
      selectionComponent = (
        <div>
          {containButton && (
            <Button
              className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
              type='link'
              onClick={() => setContainButton(false)}
              disabled={viewMode}
            >
              {valuesState ? valuesState : 'Enter Value'}
            </Button>
          )}
          {!containButton && (
            <Input
              type='number'
              disabled={viewMode}
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
                    .map((vl) =>
                      valueDisplayNames[vl]
                        ? valueDisplayNames[vl]
                        : formatCsvUploadValue(vl)
                    )
                    .join(', ')
                : null
            }
            color={TOOLTIP_CONSTANTS.DARK}
          >
            <Button
              className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
              type={viewMode ? 'default' : 'link'}
              onClick={() => setValuesSelectionOpen(!valuesSelectionOpen)}
              // disabled={viewMode}
              style={{ color: `${viewMode && '#00000040'}` }}
            >
              {valuesState && valuesState.length
                ? valuesState
                    .map((vl) =>
                      truncateURL(
                        valueDisplayNames[vl]
                          ? valueDisplayNames[vl]
                          : formatCsvUploadValue(vl),
                        projectDomainsList
                      )
                    )
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
      operatorState !== OPERATORS['isUnknown']
        ? renderValuesSelector()
        : null}
    </div>
  );
};

export default GlobalFilterSelect;
