import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { Button, InputNumber, Tooltip } from 'antd';
import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import FaDatepicker from '../../FaDatepicker';
import FaSelect from '../../FaSelect';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import { DISPLAY_PROP, OPERATORS } from 'Utils/constants';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';
import { DEFAULT_OP_PROPS } from 'Utils/operatorMapping';

const defaultOpProps = {
  categorical: DEFAULT_OP_PROPS['categorical'],
  numerical: DEFAULT_OP_PROPS['numerical'],
  datetime: [OPERATORS['equalTo']]
};

const AttrFilterSelect = ({
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = [],
  setValuesByProps,
  applyFilter,
  filter,
  refValue
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

  const [updateState, updateStateApply] = useState(false);

  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );

  useEffect(() => {
    if (filter) {
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
      const parsedValues = filter.values[0]
        ? typeof filter.values[0] === 'string'
          ? JSON.parse(filter.values)
          : filter.values
        : {};
      values = parseDateRangeFilter(parsedValues.fr, parsedValues.to);
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
        ref: refValue
      });
    }
  };

  const operatorSelect = (op) => {
    setOperatorState(op[0]);
    setValuesState(null);
    setOperSelectOpen(false);
  };

  const propSelect = (prop) => {
    setPropState({ icon: prop[2], name: prop[0], type: prop[1] });
    setPropSelectOpen(false);
    setOperatorState(OPERATORS['equalTo']);
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

    setValuesState(String(ev).toString());
  };

  const parseDateRangeFilter = (fr, to) => {
    const fromVal = fr ? fr : new Date(MomentTz().startOf('day')).getTime();
    const toVal = to ? to : new Date(MomentTz()).getTime();
    return {
      from: fromVal,
      to: toVal,
      ovp: false
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
            className={`fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin`}
            type='link'
            onClick={() => setPropSelectOpen(!propSelectOpen)}
          >
            {renderGroupDisplayName(propState)}
          </Button>
        </Tooltip>

        <div className={styles.filter__event_selector}>
          {propSelectOpen && (
            <div className={styles.filter__event_selector__btn}>
              <GroupSelect2
                groupedProperties={propOpts}
                placeholder='Select Property'
                optionClick={(group, val) => propSelect([...val, group])}
                onClickOutside={() => setPropSelectOpen(false)}
              ></GroupSelect2>
            </div>
          )}
        </div>
      </div>
    );
  };

  const renderOperatorSelector = () => {
    return (
      <div className={styles.filter__propContainer}>
        <Tooltip
          title='Select an equator to define your filter rules. '
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            className={`filter-buttons-radius filter-buttons-margin`}
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

  const renderValuesSelector = () => {
    let selectionComponent;
    const values = [];
    if (propState.type === 'categorical') {
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
    }

    if (propState.type === 'datetime') {
      const parsedValues = valuesState
        ? typeof valuesState === 'string'
          ? JSON.parse(valuesState)
          : valuesState
        : {};
      const fromRange = parsedValues.fr ? parsedValues.fr : parsedValues.from;
      const dateRange = parseDateRangeFilter(fromRange, parsedValues.to);
      const rang = {
        startDate: dateRange.from,
        endDate: dateRange.to
      };

      selectionComponent = (
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

    if (propState.type === 'numerical') {
      selectionComponent = (
        <InputNumber
          value={valuesState}
          onBlur={emitFilter}
          onChange={setNumericalValue}
        ></InputNumber>
      );
    }

    return (
      <div className={`${styles.filter__propContainer}`}>
        {propState.type === 'categorical' ? (
          <>
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

export default AttrFilterSelect;
