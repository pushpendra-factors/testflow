import React, { useState, useEffect, useMemo } from 'react';
import { useSelector, connect } from 'react-redux';
import {
  Button,
  Input,
  InputNumber,
  Tooltip,
  DatePicker,
  Upload,
  Row,
  Col,
  message
} from 'antd';
import MomentTz from 'Components/MomentTz';
import { isArray } from 'lodash';
import {
  DEFAULT_OPERATOR_PROPS,
  checkIfValueSelectorCanRender,
  convertOptionsToGroupSelectFormat,
  dateTimeSelect
} from 'Components/FaFilterSelect/utils';
import moment from 'moment';
import AppModal from 'Components/AppModal';
import { UploadOutlined } from '@ant-design/icons';
import { uploadList } from 'Reducers/global';
import FaSelect from 'Components/GenericComponents/FaSelect';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { selectedOptionsMapper } from 'Components/GenericComponents/FaSelect/utils';
import truncateURL from 'Utils/truncateURL';
import { processProperties, PropTextFormat } from 'Utils/dataFormatter';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import { toCapitalCase } from '../../utils/global';
import { DISPLAY_PROP, OPERATORS } from '../../utils/constants';
import FaDatepicker from '../FaDatepicker';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
// import truncateURL from 'Utils/truncateURL';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;
const rangePicker = [OPERATORS.equalTo, OPERATORS.notEqualTo];
const customRangePicker = [OPERATORS.between, OPERATORS.notBetween];
const deltaPicker = [OPERATORS.inThePrevious, OPERATORS.notInThePrevious];
const currentPicker = [OPERATORS.inTheCurrent, OPERATORS.notInTheCurrent];
const datePicker = [OPERATORS.before, OPERATORS.since];

function FaFilterSelect({
  viewMode,
  propOpts = [],
  operatorOpts = defaultOpProps,
  valueOpts = {},
  valueOptsLoading,
  setValuesByProps,
  applyFilter,
  filter,
  disabled = false,
  refValue,
  caller,
  dropdownPlacement,
  dropdownMaxHeight,
  uploadList,
  showInList = false,
  minEntriesPerGroup,
  maxAllowedSelection,
  extraProps = {}
}) {
  const [propState, setPropState] = useState({
    groupName: '',
    icon: '',
    name: '',
    type: ''
  });

  const [operatorState, setOperatorState] = useState(OPERATORS.equalTo);
  const [valuesState, setValuesState] = useState(null);
  const [propSelectOpen, setPropSelectOpen] = useState(true);
  const [operSelectOpen, setOperSelectOpen] = useState(false);
  const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);
  const [grnSelectOpen, setGrnSelectOpen] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [dateOptionSelectOpen, setDateOptionSelectOpen] = useState(false);
  const [containButton, setContainButton] = useState(true);
  const [updateState, updateStateApply] = useState(false);
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [uploadFileName, setUploadFileName] = useState('');
  const [uploadFileByteArray, setUploadFileByteArray] = useState([]);
  const [loading, setLoading] = useState(false);

  const { userPropNames, eventPropNames, groupPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const activeProject = useSelector((state) => state.global.active_project);
  const { projectDomainsList } = useSelector((state) => state.global);

  const valueDisplayNames = useMemo(
    () =>
      valueOpts?.[propState?.name] ? valueOpts[propState.name] : DISPLAY_PROP,
    [valueOpts, propState.name]
  );

  useEffect(() => {
    if (currentPicker.includes(operatorState)) {
      setCurrentFilt();
    }
    if (deltaPicker.includes(operatorState)) {
      setDeltaFilt();
    }
  }, [valuesState]);

  // useEffect(() => {
  //   if (
  //     operatorState === OPERATORS['isKnown'] ||
  //     operatorState === OPERATORS['isUnknown']
  //   ) {
  //     valuesSelect(['$none']);
  //   }
  // }, [operatorState]);

  useEffect(() => {
    if (
      (filter && !valuesState) ||
      (filter && filter?.values !== valuesState)
    ) {
      const prop =
        filter.props.length === 3 ? ['', ...filter.props] : filter.props;
      setPropState({
        groupName: prop[0],
        icon: prop[3],
        name: prop[1],
        type: prop[2]
      });
      if (
        (filter.operator === OPERATORS.equalTo ||
          filter.operator === OPERATORS.notEqualTo) &&
        filter.values?.length === 1 &&
        filter.values?.[0] === '$none'
      ) {
        if (filter.operator === OPERATORS.equalTo)
          setOperatorState(OPERATORS.isUnknown);
        else setOperatorState(OPERATORS.isKnown);
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
    if (filter.props[2] === 'datetime') {
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
        props: [
          propState.groupName,
          propState.name,
          propState.type,
          propState.icon
        ],
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
    // instead of useEffect we are just putting it here
    if (op === OPERATORS.isKnown || op === OPERATORS.isUnknown) {
      valuesSelect(['$none']);
    }
  };

  const propSelect = (option, group) => {
    setPropState({
      groupName: option.extraProps?.groupName,
      icon: option.extraProps?.propertyType || group.iconName,
      name: option.value,
      type: option.extraProps.valueType
    });
    setPropSelectOpen(false);
    setOperatorState(
      option.extraProps.valueType === 'datetime'
        ? OPERATORS.between
        : OPERATORS.equalTo
    );
    setValuesState(null);
    setValuesByProps([
      option.extraProps?.groupName,
      option.value,
      option.extraProps.valueType,
      option.extraProps?.propertyType || group.iconName
    ]);
    setValuesSelectionOpen(true);
  };

  const valuesSelect = (values) => {
    setValuesState(values);
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

  const setTextValue = (event) => {
    setValuesState(event.target.value);
  };

  const parseDateRangeFilter = (fr, to, value) => {
    const fromVal = fr || new Date(MomentTz().startOf('day')).getTime();
    const toVal = to || new Date(MomentTz()).getTime();
    return {
      from: fromVal,
      to: toVal,
      ovp: false,
      num: value.num,
      gran: value.gran
    };
    // return (MomentTz(fromVal).format('MMM DD, YYYY') + ' - ' +
    //           MomentTz(toVal).format('MMM DD, YYYY'));
  };

  const renderGroupDisplayName = (propState) => {
    const mergedPropNames = {
      ...groupPropNames,
      ...userPropNames,
      ...eventPropNames,
      ...extraProps?.displayNames
    };
    let propertyName = propState?.name;
    propertyName = mergedPropNames[propState?.name]
      ? mergedPropNames[propState?.name]
      : PropTextFormat(propState?.name);
    if (!propState.name) {
      propertyName = 'Select Property';
    }
    return propertyName;
  };

  const getIcon = (propState) => {
    const { name, icon } = propState || {};
    if (!name) return null;
    const iconName = 'user';
    return (
      <SVG name={iconName} size={16} color={viewMode ? 'grey' : 'purple'} />
    );
  };

  const getGroupLabel = (grp) => {
    if (grp === 'event') return 'Event Properties';
    if (grp === 'user') return 'User Properties';
    if (!grp || !grp.length) return 'Properties';
    return grp;
  };

  const renderPropSelect = () => (
    <div
      className={`${styles.filter__propContainer} ${
        disabled ? `fa-truncate-150` : ''
      }`}
    >
      <Tooltip
        zIndex={99999}
        title={renderGroupDisplayName(propState)}
        color={TOOLTIP_CONSTANTS.DARK}
      >
        <Button
          disabled={disabled}
          icon={getIcon(propState)}
          className={`fa-button--truncate fa-button--truncate-xs ${
            viewMode ? 'static-button' : ''
          }  btn-left-round filter-buttons-margin`}
          type={viewMode ? 'default' : 'link'}
          onClick={() => (viewMode ? null : setPropSelectOpen(!propSelectOpen))}
        >
          {renderGroupDisplayName(propState)}
        </Button>
      </Tooltip>
      {propSelectOpen && (
        <div className={styles.filter__event_selector}>
          <GroupSelect
            options={convertOptionsToGroupSelectFormat(propOpts)}
            placement={dropdownPlacement}
            onClickOutside={() => setPropSelectOpen(false)}
            placeholder='Select Property'
            allowSearch
            optionClickCallback={propSelect}
            allowSearchTextSelection={false}
            extraClass={`${styles.filter__event_selector__select}`}
          />
        </div>
      )}
    </div>
  );

  const renderOperatorSelector = () => (
    <div className={styles.filter__propContainer}>
      <Tooltip
        zIndex={99999}
        title='Select an equator to define your filter rules. '
        color={TOOLTIP_CONSTANTS.DARK}
        trigger={viewMode ? [] : 'hover'}
      >
        <Button
          disabled={disabled}
          className={`fa-button--truncate ${
            viewMode ? 'static-button' : ''
          } filter-buttons-radius filter-buttons-margin`}
          type={viewMode ? 'default' : 'link'}
          onClick={() => (viewMode ? null : setOperSelectOpen(true))}
        >
          {operatorState || 'Select Operator'}
        </Button>
      </Tooltip>

      {operSelectOpen && (
        <FaSelect
          options={operatorOpts[propState.type]
            .filter(
              (op) =>
                // Only include the operator if showInList is true or it's not 'inList'
                showInList ||
                (op !== OPERATORS.inList && op !== OPERATORS.notInList)
            )
            .map((op) => ({ value: op, label: op }))}
          optionClickCallback={(option) => operatorSelect(option.value)}
          onClickOutside={() => setOperSelectOpen(false)}
          placement={dropdownPlacement}
        />
      )}
    </div>
  );

  const setDeltaNumber = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues.num = val;
    setValuesState(JSON.stringify(parsedValues));
    updateStateApply(true);
  };

  const setDeltaGran = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues.gran = val;
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
    if (parsedValues.num && parsedValues.gran) {
      updateStateApply(true);
    }
  };

  const setCurrentGran = (val) => {
    const parsedValues = valuesState
      ? typeof valuesState === 'string'
        ? JSON.parse(valuesState)
        : valuesState
      : {};
    parsedValues.gran = val;
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
    if (parsedValues.gran) {
      updateStateApply(true);
    }
  };

  const onDatePickerSelect = (val) => {
    let dateT;
    const dateValue = {};
    const operatorSt = operatorState;
    if (operatorSt === OPERATORS.before) {
      dateT = MomentTz(val).startOf('day');
      dateValue.to = dateT.toDate().getTime();
    }

    if (operatorSt === OPERATORS.since) {
      dateT = MomentTz(val).startOf('day');
      dateValue.fr = dateT.toDate().getTime();
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
          className='filter-buttons-margin filter-buttons-radius'
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
          className='filter-buttons-margin filter-buttons-radius'
        />
      );
    }

    if (deltaPicker.includes(operator)) {
      selectorComponent = (
        <div className='fa-filter-dateDeltaContainer'>
          <InputNumber
            value={parsedValues.num}
            min={1}
            max={999}
            onChange={setDeltaNumber}
            disabled={disabled}
            placeholder='number'
            controls={false}
            className='filter-buttons-radius date-input-number'
          />

          <Button
            disabled={disabled}
            trigger={viewMode ? [] : 'hover'}
            className={`fa-button--truncate ${
              viewMode ? 'static-button' : ''
            } filter-buttons-radius filter-buttons-margin`}
            type={viewMode ? 'default' : 'link'}
            onClick={() => (viewMode ? null : setDateOptionSelectOpen(true))}
          >
            {parsedValues.gran
              ? dateTimeSelect.get(parsedValues.gran)
              : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={['Days', 'Weeks', 'Months', 'Quarters'].map(
                (option) => ({
                  value: option,
                  label: option
                })
              )}
              optionClickCallback={(option) => {
                setDeltaGran(dateTimeSelect.get(option.value));
              }}
              onClickOutside={() => setDateOptionSelectOpen(false)}
              placement={dropdownPlacement}
            />
          )}
        </div>
      );
    }

    if (currentPicker.includes(operator)) {
      selectorComponent = (
        <div className='fa-filter-dateDeltaContainer'>
          <Button
            disabled={disabled}
            trigger={viewMode ? [] : 'hover'}
            className={`fa-button--truncate ${
              viewMode ? 'static-button' : ''
            } filter-buttons-radius filter-buttons-margin`}
            type={viewMode ? 'default' : 'link'}
            onClick={() => (viewMode ? null : setDateOptionSelectOpen(true))}
          >
            {parsedValues.gran ? toCapitalCase(parsedValues.gran) : 'Select'}
          </Button>

          {dateOptionSelectOpen && (
            <FaSelect
              options={['Week', 'Month', 'Quarter'].map((option) => ({
                value: option,
                label: option
              }))}
              optionClickCallback={(option) => {
                setCurrentGran(option.value.toLowerCase());
              }}
              onClickOutside={() => setDateOptionSelectOpen(false)}
              placement={dropdownPlacement}
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
          className='fa-date-picker'
          open={showDatePicker}
          onOpenChange={() => {
            setShowDatePicker(!showDatePicker);
          }}
          value={
            operator === OPERATORS.before
              ? moment(parsedValues.to)
              : moment(parsedValues.from ? parsedValues.from : parsedValues.fr)
          }
          size='small'
          suffixIcon={null}
          showToday={false}
          bordered
          allowClear
          onChange={onDatePickerSelect}
        />
      );
    }

    return selectorComponent;
  };
  const renderValuesSelector = () => {
    let selectionComponent;
    if (propState.type === 'categorical') {
      const variant =
        operatorState === OPERATORS.notEqualTo ||
        operatorState === OPERATORS.doesNotContain
          ? 'Single'
          : operatorState === OPERATORS.contain
            ? 'Multi'
            : 'Multi';
      let valueOptions = valueOpts?.[propState?.name]
        ? Object.entries(valueOpts[propState.name]).map((val) => ({
            value: val[0],
            label: val[1]
          }))
        : [];

      valueOptions = selectedOptionsMapper(valueOptions, valuesState);
      if (variant === 'Single') {
        selectionComponent = (
          <FaSelect
            variant='Single'
            options={valueOptions}
            optionClickCallback={(option) => {
              valuesSelect([option.value]);
            }}
            onClickOutside={() => setValuesSelectionOpen(false)}
            allowSearch
            loadingState={valueOptsLoading}
          />
        );
      } else if (variant === 'Input') {
        selectionComponent = (
          <div>
            {containButton && (
              <Button
                disabled={disabled}
                trigger={viewMode ? [] : 'hover'}
                className={`fa-button--truncate ${
                  viewMode ? 'static-button' : ''
                } filter-buttons-radius filter-buttons-margin`}
                type={viewMode ? 'default' : 'link'}
                onClick={() => (viewMode ? null : setContainButton(false))}
              >
                {valuesState || 'Enter Value'}
              </Button>
            )}
            {!containButton && (
              <Input
                type='text'
                value={valuesState}
                placeholder='Enter Value'
                autoFocus
                onBlur={() => {
                  emitFilter();
                  setContainButton(true);
                }}
                onPressEnter={() => {
                  emitFilter();
                  setContainButton(true);
                }}
                onChange={setTextValue}
                disabled={disabled}
                className='input-value filter-buttons-radius filter-buttons-margin'
              />
            )}
          </div>
        );
      } else {
        selectionComponent = (
          <FaSelect
            variant='Multi'
            options={valueOptions}
            applyClickCallback={(updatedOptions, selectedOptions) => {
              valuesSelect(selectedOptions);
            }}
            onClickOutside={() => setValuesSelectionOpen(false)}
            allowSearch
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
        endDate: dateRange.to,
        num: dateRange.num,
        gran: dateRange.gran
      };

      selectionComponent = selectDateTimeSelector(operatorState, rang);
    }

    if (propState.type === 'numerical') {
      selectionComponent = (
        <div>
          {containButton && (
            <Button
              disabled={disabled}
              trigger={viewMode ? [] : 'hover'}
              className={`fa-button--truncate ${
                viewMode ? 'static-button' : ''
              } filter-buttons-radius filter-buttons-margin`}
              type={viewMode ? 'default' : 'link'}
              onClick={() => (viewMode ? null : setContainButton(false))}
            >
              {valuesState || 'Enter Value'}
            </Button>
          )}
          {!containButton && (
            <Input
              type='number'
              value={valuesState}
              placeholder='Enter Value'
              autoFocus
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
              className='input-value filter-buttons-radius filter-buttons-margin'
            />
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
            {/* {operatorState === OPERATORS['contain'] ? (
              selectionComponent
            ) :  */}
            <>
              <Tooltip
                zIndex={99999}
                mouseLeaveDelay={0}
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
                  className={`fa-button--truncate ${
                    caller === 'profiles' ? 'fa-button--truncate-sm' : ''
                  }  ${
                    viewMode
                      ? 'btn-right-round static-button'
                      : 'filter-buttons-radius'
                  } filter-buttons-margin`}
                  type={viewMode ? 'default' : 'link'}
                  disabled={disabled}
                  onClick={() =>
                    viewMode
                      ? null
                      : setValuesSelectionOpen(!valuesSelectionOpen)
                  }
                >
                  {valuesState && valuesState.length
                    ? valuesState
                        .map((vl) =>
                          valueDisplayNames[vl]
                            ? truncateURL(
                                valueDisplayNames[vl],
                                projectDomainsList
                              )
                            : formatCsvUploadValue(vl)
                        )
                        .join(', ')
                    : 'Select Values'}
                </Button>
              </Tooltip>
              {valuesSelectionOpen && selectionComponent}
            </>
          </>
        ) : null}

        {propState.type !== 'categorical' ? selectionComponent : null}
      </div>
    );
  };

  const handleChange = (info) => {
    const reader = new FileReader();
    const fileByteArray = [];
    reader.readAsArrayBuffer(info?.file?.originFileObj);
    reader.onloadend = function (evt) {
      if (evt.target.readyState === FileReader.DONE) {
        const arrayBuffer = evt.target.result;
        const array = new Uint8Array(arrayBuffer);
        for (let i = 0; i < array.length; i++) {
          fileByteArray.push(array[i]);
        }
      }
    };

    setUploadFileName(info?.file?.name);
    setUploadFileByteArray(fileByteArray);
  };

  const handleCancel = () => {
    setUploadModalOpen(false);
    setUploadFileName('');
    setUploadFileByteArray([]);
  };

  const handleOk = () => {
    setLoading(true);

    uploadList(activeProject?.id, {
      file_name: uploadFileName,
      payload: uploadFileByteArray
    })
      .then((res) => {
        valuesSelectSingle([res?.data?.file_reference]);
        handleCancel();
        setLoading(false);
      })
      .catch((err) => {
        setLoading(false);
        message.error(err?.data?.error);
      });
  };

  const formatCsvUploadValue = (value) => {
    const vl = value.split('_');
    let data = '';
    if (vl.length > 1) {
      data += vl[1];
      for (let i = 2; i < vl.length - 1; i++) {
        data = `${data}_${vl?.[i]}`;
      }
      data = `${data}.${vl[vl.length - 1]}`;
    } else {
      data = value;
    }
    return data;
  };

  const renderCsvUpload = () => {
    let selectionComponent;

    if (propState.type === 'categorical') {
      selectionComponent = (
        <AppModal
          visible={uploadModalOpen}
          width={500}
          closable={false}
          title={null}
          footer={null}
        >
          <Text type='title' level={6} weight='bold' extraClcalass='m-0'>
            Upload a CSV with a single column
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0 -mt-2'>
            We’ll only look at the first column as your reference list of data
          </Text>
          <div className='border rounded mt-2 flex justify-center '>
            <Upload
              showUploadList={false}
              onChange={handleChange}
              accept='.csv'
              maxCount={1}
              className='text-center'
            >
              <div className='p-8'>
                {uploadFileName ? (
                  <Button className='inline'>
                    {uploadFileName}
                    <SVG extraClass='ml-1' name='close' color='grey' />
                  </Button>
                ) : (
                  <Button icon={<UploadOutlined />}>Upload CSV</Button>
                )}
              </div>
            </Upload>
          </div>
          <Row className='mt-4'>
            <Col span={24}>
              <div className='flex justify-end'>
                <Button
                  size='large'
                  className='mr-2'
                  onClick={() => handleCancel()}
                >
                  Cancel
                </Button>
                <Button
                  size='large'
                  className='ml-2'
                  type='primary'
                  onClick={() => handleOk()}
                  disabled={!uploadFileName}
                  loading={loading}
                >
                  Done
                </Button>
              </div>
            </Col>
          </Row>
        </AppModal>
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
              zIndex={99999}
              mouseLeaveDelay={0}
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
                className={`fa-button--truncate ${
                  caller === 'profiles' ? 'fa-button--truncate-sm' : ''
                }  ${
                  viewMode
                    ? 'btn-right-round static-button'
                    : 'filter-buttons-radius'
                } filter-buttons-margin`}
                type={viewMode ? 'default' : 'link'}
                disabled={disabled}
                onClick={() =>
                  viewMode ? null : setUploadModalOpen(!uploadModalOpen)
                }
              >
                {valuesState && valuesState.length
                  ? valuesState
                      .map((vl) =>
                        valueDisplayNames[vl]
                          ? truncateURL(
                              valueDisplayNames[vl],
                              projectDomainsList
                            )
                          : formatCsvUploadValue(vl)
                      )
                      .join(', ')
                  : 'Upload list'}
              </Button>
            </Tooltip>
            {uploadModalOpen && selectionComponent}
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

      {checkIfValueSelectorCanRender(operatorState)
        ? renderValuesSelector()
        : operatorState === OPERATORS.inList ||
            operatorState === OPERATORS.notInList
          ? renderCsvUpload()
          : null}
    </div>
  );
}

export default connect(null, { uploadList })(FaFilterSelect);
