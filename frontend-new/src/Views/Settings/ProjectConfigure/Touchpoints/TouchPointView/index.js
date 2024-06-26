import React, { useState, useEffect, useMemo } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import { Row, Col, Button, Radio } from 'antd';

import {
  getEventPropertiesV2,
  getEventPropertyValues
} from 'Reducers/coreQuery/middleware';

import FaFilterSelect from 'Components/FaFilterSelect';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';

import { toCapitalCase } from 'Utils/global';
import getGroupIcon from 'Utils/getGroupIcon';
import startCase from 'lodash/startCase';
import {
  convertGroupedPropertiesToUngrouped,
  setDisplayName
} from 'Utils/dataFormatter';
import FaSelect from '../../../../../components/GenericComponents/FaSelect';

import {
  formatFiltersForQuery,
  processFiltersFromQuery
} from '../../../../CoreQuery/utils';
import {
  RULE_TYPE_HS_CONTACT,
  RULE_TYPE_HS_CALLS,
  RULE_TYPE_HS_EMAILS,
  RULE_TYPE_HS_FORM_SUBMISSIONS,
  RULE_TYPE_HS_LISTS,
  RULE_TYPE_HS_MEETINGS,
  Extra_PROP_SHOW_OPTIONS,
  RULE_TYPE_SF_CONTACT,
  ruleTypesNameMappingForHS,
  RULE_TYPE_SF_CAMPAIGNS,
  RULE_TYPE_SF_TASKS,
  RULE_TYPE_SF_EVENTS,
  EVENTS_MAP,
  ruleTypesNameMappingForSF,
  DEFAULT_TIMESTAMPS,
  PROPERTY_MAP_OPTIONS
} from '../utils';
import { PropertySelect } from './PropertySelect';

const TouchpointView = ({
  activeProject,
  tchType = 'hubspot',
  getEventPropertiesV2,
  eventPropertiesV2,
  eventUserPropertiesV2,
  rule,
  onCancel,
  onSave,
  getEventPropertyValues,
  propertyValuesMap
}) => {
  const { eventPropNames } = useSelector((state) => state.coreQuery);

  const [dropDownValues, setDropDownValues] = useState({});
  const [filterDD, setFilterDD] = useState(false);

  const [tchRuleType, setTchRuleType] = useState(
    rule
      ? rule.rule_type
      : tchType === 'hubspot'
        ? RULE_TYPE_HS_CONTACT
        : RULE_TYPE_SF_CONTACT
  );

  const [timestampRef, setTimestampRefState] = useState(
    tchType === 'hubspot'
      ? DEFAULT_TIMESTAMPS[RULE_TYPE_HS_CONTACT]
      : DEFAULT_TIMESTAMPS[RULE_TYPE_SF_CONTACT]
  );
  // touch_point_time_ref
  const [touchPointPropRef, setTouchPointPropRef] = useState(
    tchType === 'hubspot'
      ? DEFAULT_TIMESTAMPS[RULE_TYPE_HS_CONTACT]
      : DEFAULT_TIMESTAMPS[RULE_TYPE_SF_CAMPAIGNS]
  );
  const [timestampPropertyRef, setTimestampPropRef] = useState(false);
  const [dateTypeDD, setDateTypeDD] = useState(false);
  const [dateTypeProps, setDateTypeProps] = useState([]);
  // filters
  const [newFilterStates, setNewFilterStates] = useState([]);

  const [extraPropBtn, setExtraPropBtn] = useState(false);
  const [initialRender, setInitialRender] = useState(true);

  const [propertyValArray, setPropertyValArray] = useState(null);

  // property map
  const [propertyMap, setPropertyMap] = useState({
    $campaign: {
      ty: 'Property',
      va: ''
    },
    $channel: {
      ty: 'Property',
      va: ''
    },
    $source: {
      ty: 'Property',
      va: ''
    },
    $type: {
      ty: 'Property',
      va: ''
    }
  });

  const [extraPropMap, setExtraPropMap] = useState({});

  const [filterDropDownOptions, setFiltDD] = useState({
    props: [
      {
        label: '',
        icon: 'event'
      }
    ],
    operator: DEFAULT_OPERATOR_PROPS
  });

  const reInitialise = () => {
    setDefaultTimeStampValue();
    setNewFilterStates([]);
    setExtraPropBtn(false);
    setPropertyMap({
      $campaign: {
        ty: 'Property',
        va: ''
      },
      $channel: {
        ty: 'Property',
        va: ''
      },
      $source: {
        ty: 'Property',
        va: ''
      },
      $type: {
        ty: 'Property',
        va: ''
      }
    });
    setExtraPropMap({});
  };
  useEffect(() => {
    const eventToCall = getEventToCall();
    getEventPropertiesV2(activeProject.id, eventToCall);
    if (!rule || !initialRender) reInitialise();
  }, [tchRuleType]);
  useEffect(() => {
    if (rule) {
      const filterState = processFiltersFromQuery(rule.filters);
      chainEventPropertyValues(filterState);
      setNewFilterStates(filterState);
      setPropertyMap(reversePropertyMap(rule.properties_map));
      if (rule.touch_point_time_ref === 'LAST_MODIFIED_TIME_REF') {
        setTimestampRefState('LAST_MODIFIED_TIME_REF');
        setTimestampPropRef(false);
        setTouchPointPropRef('LAST_MODIFIED_TIME_REF');
      } else {
        setTimestampRefState(rule.touch_point_time_ref);
        setTouchPointPropRef(rule.touch_point_time_ref);
        setTimestampPropRef(true);
        setDateTypeDD(false);
      }
      setInitialRender(false);
    }
  }, [rule]);

  useEffect(() => {
    if (propertyValArray) {
      propertyValArray.then((res) => {
        newFilterStates.forEach((filt, index) => {
          const prop = filt.props;
          const propToCall = prop.length > 3 ? prop[1] : prop[0];
          setPropData(propToCall, res[index]?.data);
        });
      });
    }
  }, [propertyValArray]);

  const reversePropertyMap = (properties) => {
    // Gets the extra Properties Filtered and return the defined properties.
    const propMap = { ...properties };
    const extraProps = {};
    const propKeys = Object.keys(propertyMap);
    Object.keys(propMap).forEach((key) => {
      if (key !== '$type' && propMap[key].va?.[0] !== '$') {
        propMap[key].va = getReversePropertyName(propMap[key]?.va);
      }
      if (!propKeys.includes(key)) {
        extraProps[key] = propMap[key];
        delete propMap[key];
      }
    });
    setExtraPropMap(extraProps);
    return propMap;
  };

  const chainEventPropertyValues = (filters) => {
    const filterData = Promise.all(
      filters.map(async (filt) => {
        const eventToCall = getEventToCall();
        const prop = filt.props;
        const propToCall = prop.length > 3 ? prop[1] : prop[0];
        return await getEventPropertyValues(
          activeProject.id,
          eventToCall,
          propToCall
        );
      })
    );
    setPropertyValArray(filterData);
  };

  const setPropData = (propToCall, data) => {
    const ddValues = { ...dropDownValues };
    ddValues[propToCall] = [...data, '$none'];
    setDropDownValues(ddValues);
  };

  const setValuesByProps = (props) => {
    const eventToCall = getEventToCall();
    const propToCall = props.length > 3 ? props[1] : props[0];
    if (dropDownValues[propToCall]?.length >= 1) {
      return null;
    }
    getEventPropertyValues(activeProject.id, eventToCall, propToCall);
  };

  const getEventToCall = () => {
    if (
      tchRuleType === RULE_TYPE_SF_CONTACT ||
      tchRuleType === RULE_TYPE_SF_CAMPAIGNS
    ) {
      return timestampRef === 'campaign_member_created_date'
        ? EVENTS_MAP[RULE_TYPE_SF_CAMPAIGNS][0]
        : EVENTS_MAP[RULE_TYPE_SF_CAMPAIGNS][1];
    }
    return EVENTS_MAP[tchRuleType];
  };

  const getPropertiesFiltered = (propertiesObj, startsWith) => {
    const propertiesFiltered = {};
    if (propertiesObj) {
      for (const key in propertiesObj) {
        if (propertiesObj.hasOwnProperty(key)) {
          const filteredProperties =
            propertiesObj[key].filter((item) =>
              item?.[1]?.startsWith(startsWith)
            ) || [];
          if (filteredProperties?.length > 0) {
            propertiesFiltered[key] = filteredProperties;
          }
        }
      }
    }
    return propertiesFiltered;
  };

  useEffect(() => {
    const eventToCall = getEventToCall();
    let tchUserProps = {};
    const filterDD = { ...filterDropDownOptions };
    const propState = [];
    let eventProps = {};
    const propsArray = [];
    const startsWith = getStartsWith();
    if (tchType === 'hubspot') {
      eventProps = getPropertiesFiltered(
        eventPropertiesV2[eventToCall],
        startsWith
      );
      if (tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS) {
        tchUserProps = getPropertiesFiltered(eventUserPropertiesV2, startsWith);
      }
    } else if (tchType === 'salesforce') {
      eventProps = getPropertiesFiltered(
        eventPropertiesV2[eventToCall],
        startsWith
      );
      tchUserProps = getPropertiesFiltered(eventUserPropertiesV2, startsWith);
    }
    const filterOptsObj = {};
    if (eventProps) {
      Object.keys(eventProps)?.forEach((groupkey) => {
        if (!filterOptsObj[groupkey]) {
          filterOptsObj[groupkey] = {
            label: startCase(groupkey),
            icon: getGroupIcon(groupkey),
            propertyType: 'event',
            values: eventProps?.[groupkey]
          };
        } else {
          eventProps?.[groupkey]?.forEach((optionArray) =>
            filterOptsObj[groupkey].values.push(optionArray)
          );
        }
        // Convert to Array.
        eventProps[groupkey].forEach((userPropArray) => {
          propsArray.push(userPropArray);
        });
      });
    }
    if (tchUserProps) {
      Object.keys(tchUserProps)?.forEach((groupkey) => {
        if (!filterOptsObj[groupkey]) {
          filterOptsObj[groupkey] = {
            label: startCase(groupkey),
            icon: getGroupIcon(groupkey),
            propertyType: 'event',
            values: tchUserProps?.[groupkey]
          };
        } else {
          tchUserProps?.[groupkey]?.forEach((optionArray) =>
            filterOptsObj[groupkey].values.push(optionArray)
          );
        }
        // Convert to Array.
        tchUserProps[groupkey].forEach((userPropArray) => {
          propsArray.push(userPropArray);
        });
      });
    }

    const dateTypepoperties = [];
    propsArray.forEach((prop) => {
      if (prop[2] === 'datetime') {
        dateTypepoperties.push({ value: prop[1], label: prop[0] });
      }
    });
    setDateTypeProps(dateTypepoperties);

    filterDD.props = Object.values(filterOptsObj);
    setFiltDD(filterDD);
  }, [eventPropertiesV2, timestampRef]);

  const applyFilter = (fil, index) => {
    const filtState = [...newFilterStates];
    if (index && index < 0) {
      filtState.push(fil);
    } else {
      filtState[index] = fil;
    }
    setNewFilterStates(filtState);
    setFilterDD(false);
  };

  const closeFilter = (index) => {
    const filtrs = [...newFilterStates];
    filtrs.splice(index, 1);
    setNewFilterStates(filtrs);
  };

  const renderFilters = () => {
    const filterRows = [];
    if (newFilterStates) {
      newFilterStates.forEach((filter, index) => {
        filterRows.push(
          <div className={`mt-2 flex items-center relative `}>
            <div className='relative flex'>
              <FaFilterSelect
                filter={filter}
                propOpts={filterDropDownOptions.props}
                operatorOpts={filterDropDownOptions.operator}
                valueOpts={propertyValuesMap.data}
                applyFilter={(filt) => applyFilter(filt, index)}
                setValuesByProps={setValuesByProps}
                valueOptsLoading={propertyValuesMap.loading}
              />
            </div>
            <Button
              type='text'
              className='fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button'
              onClick={() => closeFilter(index)}
              size='small'
            >
              <SVG name='remove' />
            </Button>
          </div>
        );
      });
    }

    filterRows.push(
      <div className={`mt-2 flex items-center relative `}>
        {filterDD ? (
          <>
            <div className='relative flex'>
              <FaFilterSelect
                propOpts={filterDropDownOptions.props}
                operatorOpts={filterDropDownOptions.operator}
                valueOpts={propertyValuesMap.data}
                applyFilter={(filt) => applyFilter(filt, -1)}
                setValuesByProps={setValuesByProps}
                valueOptsLoading={propertyValuesMap.loading}
              />
            </div>
            <Button
              type='text'
              className='fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button'
              onClick={() => setFilterDD(false)}
              size='small'
            >
              <SVG name='remove' />
            </Button>
          </>
        ) : (
          <Button size='large' type='text' onClick={() => setFilterDD(true)}>
            <SVG name='plus' extraClass='mr-1' />
            Add Filter
          </Button>
        )}
      </div>
    );

    return filterRows;
  };

  const renderFilterBlock = () => (
    <Row className='mt-4'>
      <Col span={6} className='justify-items-start'>
        <Text level={7} type='title' extraClass='m-0' weight='bold'>
          Add a Touchpoint Rule<sup>*</sup>
        </Text>
      </Col>

      <Col span={14}>{renderFilters().map((component) => component)}</Col>
    </Row>
  );

  const setTimestampRefSF = (val) => {
    const timeStVal = val?.target?.value;
    setTimestampRefState(timeStVal);
    setTimestampPropRef(!!(RULE_TYPE_HS_CONTACT && timeStVal === ''));
    setTouchPointPropRef(timeStVal);
  };

  const setTimestampProp = ({ value, label }) => {
    setTouchPointPropRef(value);
    setDateTypeDD(false);
  };

  const validateInputs = () => {
    let isReady = true;
    const propKeys = Object.keys(propertyMap);
    for (let i = 0; i < propKeys.length; i++) {
      propertyMap[propKeys[i]].va ? (() => {})() : (isReady = false);
      if (!isReady) {
        break;
      }
    }
    return isReady;
  };

  const validateRuleInfo = () => {
    if (newFilterStates.length < 1 || !touchPointPropRef) {
      return false;
    }
    return true;
  };

  const setDefaultTimeStampValue = () => {
    const val = DEFAULT_TIMESTAMPS[tchRuleType];
    setTimestampRefState(val);
    setTimestampPropRef(false);
    setTouchPointPropRef(val);
  };

  const renderTimestampRenderOption = () => {
    if (tchRuleType === RULE_TYPE_HS_CONTACT) {
      return (
        <Radio.Group
          onChange={setTimestampRefSF}
          value={timestampRef === 'LAST_MODIFIED_TIME_REF' ? timestampRef : ''}
        >
          <Radio value='LAST_MODIFIED_TIME_REF'>
            Factors Last modified time
          </Radio>
          <Radio value=''>Select a property</Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_HS_EMAILS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$hubspot_engagement_timestamp'>Email Timestamp</Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_HS_FORM_SUBMISSIONS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$timestamp'>Form submission timestamp</Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_HS_MEETINGS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$hubspot_engagement_timestamp'>
            Meeting Done Timestamp
          </Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_HS_CALLS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$hubspot_engagement_timestamp'>Call timestamp</Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_HS_LISTS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$hubspot_contact_list_timestamp'>
            Added to the List timestamp
          </Radio>
          <Radio value='$hubspot_contact_list_list_create_timestamp'>
            List create timestamp
          </Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_SF_CONTACT) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='campaign_member_created_date'>
            Campaign Created Date
          </Radio>
          <Radio value='campaign_member_first_responded_date'>
            Campaign First Responded Date
          </Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_SF_TASKS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$salesforce_task_lastmodifieddate'>
            Task Modified Date
          </Radio>
          <Radio value='$salesforce_task_createddate'>Task Created Date</Radio>
          <Radio value='$salesforce_task_completeddatetime'>
            Task Completed Date
          </Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_SF_EVENTS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$salesforce_event_lastmodifieddate'>
            Event Modified Date
          </Radio>
          <Radio value='$salesforce_event_createddate'>
            Event Created Date
          </Radio>
        </Radio.Group>
      );
    }
    if (tchRuleType === RULE_TYPE_SF_CAMPAIGNS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value='$sf_campaign_member_created'>
            Campaign Created Date
          </Radio>
          <Radio value='$sf_campaign_member_updated'>
            Campaign First Responded Date
          </Radio>
        </Radio.Group>
      );
    }
  };

  const renderTimestampSelector = () => (
    <div className='mt-8'>
      <Row className='mt-2'>
        <Text level={7} type='title' extraClass='m-0' weight='bold'>
          Touchpoint Timestamp<sup>*</sup>
        </Text>
      </Row>
      <Row className='mt-4'>{renderTimestampRenderOption()}</Row>
      <Row className='mt-2'>
        {tchRuleType === RULE_TYPE_HS_CONTACT && timestampPropertyRef && (
          <div className='relative'>
            <Button type='link' onClick={() => setDateTypeDD(!dateTypeDD)}>
              {touchPointPropRef
                ? eventPropNames[touchPointPropRef]
                  ? eventPropNames[touchPointPropRef]
                  : touchPointPropRef
                : 'Select Date type property'}
            </Button>
            {dateTypeDD && (
              <FaSelect
                optionClickCallback={setTimestampProp}
                onClickOutside={() => setDateTypeDD(false)}
                options={dateTypeProps}
              />
            )}
          </div>
        )}
      </Row>
    </div>
  );

  const setPropType = ({ value, label }) => {
    const propMap = { ...propertyMap };
    propMap.$type.va = value;
    setPropertyMap(propMap);
    // setTypeSelectorOpen(false);
  };

  const setPropVal = ({ value, label }, key) => {
    const propMap = { ...propertyMap };
    propMap[key].va = value;
    if (value.length !== 0 && isSearchedValue(value))
      propMap[key].ty = 'Constant';
    else propMap[key].ty = 'Property';
    setPropertyMap(propMap);
  };

  const getStartsWith = () => {
    switch (tchRuleType) {
      case RULE_TYPE_HS_EMAILS:
        return '$hubspot_engagement';
      case RULE_TYPE_HS_MEETINGS:
        return '$hubspot_engagement';
      case RULE_TYPE_HS_CALLS:
        return '$hubspot_engagement';
      case RULE_TYPE_HS_CONTACT:
        return '$hubspot_contact';
      case RULE_TYPE_HS_FORM_SUBMISSIONS:
        return '';
      case RULE_TYPE_HS_LISTS:
        return '$hubspot_contact_list';
      case RULE_TYPE_SF_EVENTS:
        return '$salesforce_event';
      case RULE_TYPE_SF_TASKS:
        return '$salesforce_task';
      default:
        return '';
    }
  };

  const eventPropertiesModified = useMemo(() => {
    const eventToCall = getEventToCall();

    const eventProps = [];
    if (eventPropertiesV2?.[eventToCall]) {
      convertGroupedPropertiesToUngrouped(
        eventPropertiesV2?.[eventToCall],
        eventProps
      );
    }
    return eventProps;
  }, [eventPropertiesV2, tchRuleType]);

  const eventUserPropertiesModified = useMemo(() => {
    const userPropertiesModified = [];
    if (eventUserPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        eventUserPropertiesV2,
        userPropertiesModified
      );
    }
    return userPropertiesModified;
  }, [eventUserPropertiesV2]);

  const renderEventPropertyCampOptions = () => {
    const eventToCall = getEventToCall();
    const propertiesMp = [];
    const startsWith = getStartsWith();

    if (tchType === 'hubspot') {
      eventPropertiesModified?.forEach((prop) => {
        if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
          propertiesMp.push({ value: prop[1], label: prop[0] });
        }
      });
      tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS &&
        eventUserPropertiesModified.forEach((prop) => {
          if (prop[1]?.startsWith(startsWith)) {
            propertiesMp.push({ value: prop[1], label: prop[0] });
          }
        });
    } else if (tchType === 'salesforce') {
      eventPropertiesModified?.forEach((prop) => {
        if (prop[1]?.startsWith(startsWith)) {
          propertiesMp.push({ value: prop[1], label: prop[0] });
        }
      });
      eventUserPropertiesModified.forEach((prop) => {
        if (prop[1]?.startsWith(startsWith)) {
          propertiesMp.push({ value: prop[1], label: prop[0] });
        }
      });
    }
    return propertiesMp;
  };

  const { propertyNameMap, reversePropertyNameMap } = useMemo(() => {
    const eventToCall = getEventToCall();
    const propertyNameMap = {};
    const reversePropertyNameMap = {};
    if (eventToCall && eventPropertiesV2?.[eventToCall]) {
      Object.keys(eventPropertiesV2[eventToCall]).forEach((groupKey) => {
        eventPropertiesV2[eventToCall][groupKey]?.forEach((optArray) => {
          reversePropertyNameMap[optArray[0]] = optArray[1];
          propertyNameMap[optArray[1]] = optArray[0];
        });
      });
    }
    return { propertyNameMap, reversePropertyNameMap };
  }, [eventPropertiesV2, tchRuleType]);

  const getPropertyName = (val) => {
    if (propertyNameMap.hasOwnProperty(val)) return propertyNameMap[val];
    return val;
  };
  const getReversePropertyName = (val) => {
    if (reversePropertyNameMap.hasOwnProperty(val))
      return reversePropertyNameMap[val];
    return val;
  };
  // To check if the value is new value entered by user. Returns True For New Value
  const isSearchedValue = (val) => {
    const eventToCall = getEventToCall();
    return (
      !propertyNameMap.hasOwnProperty(val) &&
      !reversePropertyNameMap.hasOwnProperty(val)
    );
  };

  const renderTypePropertyOptions = () => {
    const options = [];
    if (tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS)
      options.push({ value: 'tactic', label: 'Tactic' });
    if (tchRuleType !== RULE_TYPE_HS_EMAILS)
      options.push({ value: 'offer', label: 'Offer' });
    return options;
  };
  const renderPropertyMap = () => {
    const propertyMapRows = [];
    PROPERTY_MAP_OPTIONS.forEach((property, index) => {
      const propTitle = property[0];
      const propKey = property[1];
      if (propTitle === 'Type') {
        propertyMapRows.push(
          <Row key={index} className='mt-10'>
            <Col span={7}>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                {propTitle}
              </Text>
            </Col>
            <PropertySelect
              title={
                propertyMap.$type.va === ''
                  ? 'Select Type Property'
                  : toCapitalCase(propertyMap.$type.va)
              }
              setPropValue={setPropType}
              renderOptions={renderTypePropertyOptions}
              allowSearch={false}
            />
          </Row>
        );
      } else {
        propertyMapRows.push(
          <Row key={index} className='mt-4'>
            <Col span={7}>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                {propTitle}
              </Text>
            </Col>
            <PropertySelect
              title={
                !propertyMap[propKey].va || propertyMap[propKey].va === ''
                  ? `Select ${toCapitalCase(propKey?.slice(1))} Property`
                  : setDisplayName(eventPropNames, propertyMap[propKey].va)
              }
              setPropValue={(option) => setPropVal(option, propKey)}
              renderOptions={renderEventPropertyCampOptions}
              allowSearch
            />
          </Row>
        );
      }
    });
    return (
      <div className={`border-top--thin pt-5 mt-8 `}>
        <Row>
          <Text level={7} type='title' extraClass='m-0' weight='bold'>
            Map the properties<sup>*</sup>
          </Text>
        </Row>

        {propertyMapRows}
      </div>
    );
  };

  const onSaveToucPoint = () => {
    // Prep settings obj;
    let propMap = { ...propertyMap };
    if (Object.keys(extraPropMap).length) {
      propMap = Object.assign(propMap, extraPropMap);
    }

    const touchPointObj = {
      // parse and set filterstate
      filters: formatFiltersForQuery(newFilterStates),
      // set propMap
      properties_map: propMap,
      touch_point_time_ref: touchPointPropRef
    };
    touchPointObj.rule_type = tchRuleType;
    onSave(touchPointObj);
  };

  const renderFooterActions = () => (
    <div>
      <Row className='mt-20 relative justify-start'>
        <Text level={7} type='title' extraClass='m-0 italic' weight='thin'>
          <sup>*</sup> All these fields are mandatory
        </Text>
      </Row>
      <Row className='border-top--thin mt-4 relative justify-start'>
        <Col className='mt-6' span={10}>
          <Button size='large' onClick={() => onCancel()}>
            Cancel
          </Button>
          <Button
            disabled={!validateRuleInfo() || !validateInputs()}
            size='large'
            type='primary'
            className='ml-2'
            htmlType='submit'
            onClick={onSaveToucPoint}
          >
            Save
          </Button>
        </Col>
      </Row>
    </div>
  );

  // Rule Type Selection
  const ruleTypeSelect = (option) => {
    setTchRuleType(option.value);
  };
  const renderTchRuleTypeOptions = () => {
    if (tchType === 'salesforce') {
      return Object.entries(ruleTypesNameMappingForSF).map((option) => ({
        value: option[0],
        label: option[1]
      }));
    }
    return Object.entries(ruleTypesNameMappingForHS).map((option) => ({
      value: option[0],
      label: option[1]
    }));
  };
  const renderTchRuleType = () => (
    <div className='mt-8'>
      <Row className='mt-2'>
        <Text level={7} type='title' extraClass='m-0' weight='thin'>
          Create a touchpoint using<sup>*</sup>
        </Text>
      </Row>
      <Row className='mt-4'>
        <PropertySelect
          title={
            tchType === 'hubspot'
              ? ruleTypesNameMappingForHS[tchRuleType]
              : ruleTypesNameMappingForSF[tchRuleType]
          }
          setPropValue={ruleTypeSelect}
          renderOptions={renderTchRuleTypeOptions}
          allowSearch={false}
        />
      </Row>
    </div>
  );

  // Extra Property only in Form Submission
  const setExtraMapByProp = (extraProp) => {
    const extraMap = { ...extraPropMap };
    extraMap[`$${extraProp}`] = {
      ty: 'Property',
      va: ''
    };
    setExtraPropMap(extraMap);
  };
  const setExtraPropVal = ({ value, label }, key) => {
    const propMap = { ...extraPropMap };
    propMap[key].va = value;
    if (value.length !== 0 && isSearchedValue(value))
      propMap[key].ty = 'Constant';
    else propMap[key].ty = 'Property';
    setExtraPropMap(propMap);
  };
  const renderAddExtraPropBtn = () => (
    <div className='mr-2 items-center relative'>
      <Button
        type='link'
        icon={<SVG name='plus' color='grey' />}
        onClick={() => setExtraPropBtn(!extraPropBtn)}
      >
        Add touchpoint property
      </Button>

      {extraPropBtn && (
        <FaSelect
          options={Extra_PROP_SHOW_OPTIONS.map((op) => ({
            value: op[2],
            label: op[0]
          }))}
          optionClickCallback={(option) => {
            setExtraMapByProp(option.value);
            setExtraPropBtn(false);
          }}
          onClickOutside={() => setExtraPropBtn(false)}
        />
      )}
    </div>
  );
  const renderExtraPropMap = () => {
    const extraMapRows = [];
    Extra_PROP_SHOW_OPTIONS.forEach((key, index) => {
      const propKey = `$${key[2]}`;
      if (!Object.keys(extraPropMap).includes(propKey)) return null;
      extraMapRows.push(
        <Row key={index} className='mt-4'>
          <Col span={7}>
            <Text level={7} type='title' extraClass='m-0' weight='thin'>
              {key[0]}
            </Text>
          </Col>
          <PropertySelect
            title={
              !extraPropMap[propKey].va || extraPropMap[propKey].va === ''
                ? 'Select Property'
                : setDisplayName(eventPropNames, extraPropMap[propKey].va)
            }
            setPropValue={(val) => setExtraPropVal(val, propKey)}
            renderOptions={renderEventPropertyCampOptions}
            allowSearch
          />
        </Row>
      );
    });

    return (
      <div className={`pt-5 mt-8 `}>
        {extraMapRows}
        <Row>{renderAddExtraPropBtn()}</Row>
      </div>
    );
  };

  return (
    <div>
      {renderTchRuleType()}

      {renderTimestampSelector()}

      {renderFilterBlock()}

      {renderPropertyMap()}

      {tchRuleType === RULE_TYPE_HS_FORM_SUBMISSIONS && renderExtraPropMap()}

      {renderFooterActions()}
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  propertyValuesMap: state.coreQuery.propertyValuesMap
});
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventPropertiesV2,
      getEventPropertyValues
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(TouchpointView);
