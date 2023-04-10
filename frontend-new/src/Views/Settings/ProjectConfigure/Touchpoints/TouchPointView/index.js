import React, { useState, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import { Row, Col, Button, Radio, Input, Select, Tooltip } from 'antd';

import { getEventProperties } from 'Reducers/coreQuery/middleware';

import FaFilterSelect from 'Components/FaFilterSelect';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';

import { fetchEventPropertyValues } from 'Reducers/coreQuery/services';
import FaSelect from '../../../../../components/FaSelect';

import {
  getFiltersWithoutOrProperty,
  getStateFromFilters
} from '../../../../../Views/CoreQuery/utils';
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
  reverseRuleTypesNameMappingForHS,
  RULE_TYPE_SF_CAMPAIGNS,
  RULE_TYPE_SF_TASKS,
  RULE_TYPE_SF_EVENTS,
  EVENTS_MAP,
  ruleTypesNameMappingForSF,
  reverseRuleTypesNameMappingForSF,
  DEFAULT_TIMESTAMPS
} from '../utils';
import { toCapitalCase } from 'Utils/global';
import styles from './index.module.scss';
import logger from 'Utils/logger';

const TouchpointView = ({
  activeProject,
  tchType = '2',
  getEventProperties,
  eventProperties,
  userProperties,
  rule,
  onCancel,
  onSave
}) => {
  const { eventPropNames } = useSelector((state) => state.coreQuery);

  const [dropDownValues, setDropDownValues] = useState({});
  const [filterDD, setFilterDD] = useState(false);

  const [tchRuleType, setTchRuleType] = useState(
    rule
      ? rule.rule_type
      : tchType === '2'
      ? RULE_TYPE_HS_CONTACT
      : RULE_TYPE_SF_CONTACT
  );

  const [timestampRef, setTimestampRefState] = useState(
    tchType === '2'
      ? DEFAULT_TIMESTAMPS[RULE_TYPE_HS_CONTACT]
      : DEFAULT_TIMESTAMPS[RULE_TYPE_SF_CONTACT]
  );
  //touch_point_time_ref
  const [touchPointPropRef, setTouchPointPropRef] = useState(
    tchType === '2'
      ? DEFAULT_TIMESTAMPS[RULE_TYPE_HS_CONTACT]
      : DEFAULT_TIMESTAMPS[RULE_TYPE_SF_CAMPAIGNS]
  );
  const [timestampPropertyRef, setTimestampPropRef] = useState(false);
  const [dateTypeDD, setDateTypeDD] = useState(false);
  const [dateTypeProps, setDateTypeProps] = useState([]);
  //filters
  const [newFilterStates, setNewFilterStates] = useState([]);

  const [extraPropBtn, setExtraPropBtn] = useState(false);
  const [initialRender, setInitialRender] = useState(true);

  const [propertyValArray, setPropertyValArray] = useState(null);

  //property map
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

  const [ruleSelectorOpen, setRuleSelectorOpen] = useState(false);
  const [typeSelectorOpen, setTypeSelectorOpen] = useState(false);
  const [sourceSelectorOpen, setSourceSelectorOpen] = useState(false);
  const [campaignSelectorOpen, setCampaignSelectorOpen] = useState(false);
  const [channelSelectorOpen, setChannelSelectorOpen] = useState(false);
  const [extraPropSelectorOpen, setExtraPropSelectorOpen] = useState(false);

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
    getEventProperties(activeProject.id, eventToCall);
    if (!rule || !initialRender) reInitialise();
  }, [tchRuleType]);
  useEffect(() => {
    if (rule) {
      const filterState = getStateFromFilters(rule.filters);
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
    //Gets the extra Properties Filtered and return the defined properties.
    const propMap = { ...properties };
    const extraProps = {};
    const propKeys = Object.keys(propertyMap);
    Object.keys(propMap).forEach((key) => {
      if (key !== '$type' && propMap[key].va?.[0] !== '$') {
        propMap[key].va = reversePropertyNameMap(propMap[key].va);
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
        return await fetchEventPropertyValues(
          activeProject.id,
          eventToCall,
          propToCall
        );
      })
    );
    setPropertyValArray(filterData);
  };

  const setPropData = (propToCall, data) => {
    const ddValues = Object.assign({}, dropDownValues);
    ddValues[propToCall] = [...data, '$none'];
    setDropDownValues(ddValues);
  };

  const setValuesByProps = (props) => {
    const eventToCall = getEventToCall();
    const propToCall = props.length > 3 ? props[1] : props[0];
    if (dropDownValues[propToCall]?.length >= 1) {
      return null;
    }
    fetchEventPropertyValues(activeProject.id, eventToCall, propToCall)
      .then((res) => {
        setPropData(propToCall, res.data);
      })
      .catch((err) => {
        const ddValues = Object.assign({}, dropDownValues);
        ddValues[propToCall] = ['$none'];
        setDropDownValues(ddValues);
      });
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

  useEffect(() => {
    const eventToCall = getEventToCall();
    const tchUserProps = [];
    const filterDD = Object.assign({}, filterDropDownOptions);
    const propState = [];
    const eventProps = [];
    const startsWith = getStartsWith();
    if (tchType === '2') {
      eventProperties[eventToCall]
        ? eventProperties[eventToCall].forEach((prop) => {
            if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
              eventProps.push(prop);
            }
          })
        : (() => {})();
      tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS &&
        userProperties.forEach((prop) => {
          if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
            tchUserProps.push(prop);
          }
        });
    } else if (tchType === '3') {
      eventProperties[eventToCall]
        ? eventProperties[eventToCall].forEach((prop) => {
            if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
              eventProps.push(prop);
            }
          })
        : (() => {})();
      userProperties.forEach((prop) => {
        if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
          tchUserProps.push(prop);
        }
      });
    }

    filterDropDownOptions.props.forEach((k) => {
      propState.push({
        label: k.label,
        icon: 'event',
        values: [...eventProps, ...tchUserProps]
      });
    });

    const dateTypepoperties = [];
    eventProps.forEach((prop) => {
      if (prop[2] === 'datetime') {
        dateTypepoperties.push(prop);
      }
    });
    tchUserProps.forEach((prop) => {
      if (prop[2] === 'datetime') {
        dateTypepoperties.push(prop);
      }
    });
    setDateTypeProps(dateTypepoperties);
    filterDD.props = propState;
    setFiltDD(filterDD);
  }, [eventProperties, timestampRef]);

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
            <div className={`relative flex`}>
              <FaFilterSelect
                filter={filter}
                propOpts={filterDropDownOptions.props}
                operatorOpts={filterDropDownOptions.operator}
                valueOpts={dropDownValues}
                applyFilter={(filt) => applyFilter(filt, index)}
                setValuesByProps={setValuesByProps}
              ></FaFilterSelect>
            </div>
            <Button
              type='text'
              className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
              onClick={() => closeFilter(index)}
              size={'small'}
            >
              <SVG name={'remove'} />
            </Button>
          </div>
        );
      });
    }

    filterRows.push(
      <div className={`mt-2 flex items-center relative `}>
        {filterDD ? (
          <>
            <div className={`relative flex`}>
              <FaFilterSelect
                propOpts={filterDropDownOptions.props}
                operatorOpts={filterDropDownOptions.operator}
                valueOpts={dropDownValues}
                applyFilter={(filt) => applyFilter(filt, -1)}
                setValuesByProps={setValuesByProps}
              ></FaFilterSelect>
            </div>
            <Button
              type='text'
              className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
              onClick={() => setFilterDD(false)}
              size={'small'}
            >
              <SVG name={'remove'} />
            </Button>
          </>
        ) : (
          <Button
            size={'large'}
            type={'text'}
            onClick={() => setFilterDD(true)}
          >
            <SVG name={'plus'} extraClass={'mr-1'} />
            {'Add Filter'}
          </Button>
        )}
      </div>
    );

    return filterRows;
  };

  const renderFilterBlock = () => {
    return (
      <Row className={`mt-4`}>
        <Col span={6} className={`justify-items-start`}>
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>
            Add a Touchpoint Rule<sup>*</sup>
          </Text>
        </Col>

        <Col span={14}>{renderFilters().map((component) => component)}</Col>
      </Row>
    );
  };

  const setTimestampRefSF = (val) => {
    const timeStVal = val?.target?.value;
    setTimestampRefState(timeStVal);
    setTimestampPropRef(
      RULE_TYPE_HS_CONTACT && timeStVal === '' ? true : false
    );
    setTouchPointPropRef(timeStVal);
  };

  const setTimestampProp = (val) => {
    setTouchPointPropRef(val[1]);
    setDateTypeDD(false);
  };

  const validateInputs = () => {
    let isReady = true;
    const propKeys = Object.keys(propertyMap);
    for (let i = 0; i < propKeys.length; i++) {
      propertyMap[propKeys[i]]['va'] ? (() => {})() : (isReady = false);
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
          <Radio value={`LAST_MODIFIED_TIME_REF`}>
            Factors Last modified time
          </Radio>
          <Radio value={``}>Select a property</Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_HS_EMAILS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`$hubspot_engagement_timestamp`}>Email Timestamp</Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_HS_FORM_SUBMISSIONS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`$timestamp`}>Form submission timestamp</Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_HS_MEETINGS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`$hubspot_engagement_timestamp`}>
            Meeting Done Timestamp
          </Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_HS_CALLS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`$hubspot_engagement_timestamp`}>Call timestamp</Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_HS_LISTS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`$hubspot_contact_list_timestamp`}>
            Added to the List timestamp
          </Radio>
          <Radio value={`$hubspot_contact_list_list_create_timestamp`}>
            List create timestamp
          </Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_SF_CONTACT) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={`campaign_member_created_date`}>
            Campaign Created Date
          </Radio>
          <Radio value={`campaign_member_first_responded_date`}>
            Campaign First Responded Date
          </Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_SF_TASKS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={'$salesforce_task_lastmodifieddate'}>
            Task Modified Date
          </Radio>
          <Radio value={'$salesforce_task_createddate'}>
            Task Created Date
          </Radio>
          <Radio value={'$salesforce_task_completeddatetime'}>
            Task Completed Date
          </Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_SF_EVENTS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={'$salesforce_event_lastmodifieddate'}>
            Event Modified Date
          </Radio>
          <Radio value={'$salesforce_event_createddate'}>
            Event Created Date
          </Radio>
        </Radio.Group>
      );
    } else if (tchRuleType === RULE_TYPE_SF_CAMPAIGNS) {
      return (
        <Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
          <Radio value={'$sf_campaign_member_created'}>
            Campaign Created Date
          </Radio>
          <Radio value={'$sf_campaign_member_updated'}>
            Campaign First Responded Date
          </Radio>
        </Radio.Group>
      );
    }
  };

  const renderTimestampSelector = () => {
    return (
      <div className={`mt-8`}>
        <Row className={`mt-2`}>
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>
            Touchpoint Timestamp<sup>*</sup>
          </Text>
        </Row>
        <Row className={`mt-4`}>{renderTimestampRenderOption()}</Row>
        <Row className={`mt-2`}>
          {tchRuleType === RULE_TYPE_HS_CONTACT && timestampPropertyRef && (
            <div className={`relative`}>
              <Button type='link' onClick={() => setDateTypeDD(!dateTypeDD)}>
                {touchPointPropRef
                  ? eventPropNames[touchPointPropRef]
                    ? eventPropNames[touchPointPropRef]
                    : touchPointPropRef
                  : 'Select Date type property'}
              </Button>
              {dateTypeDD && (
                <FaSelect
                  optionClick={(opt) => setTimestampProp(opt)}
                  onClickOutside={() => setDateTypeDD(false)}
                  options={dateTypeProps}
                ></FaSelect>
              )}
            </div>
          )}
        </Row>
      </div>
    );
  };

  const setPropType = (val) => {
    const propMap = Object.assign({}, propertyMap);
    propMap['$type']['va'] = val[0].toLowerCase();
    setPropertyMap(propMap);
    setTypeSelectorOpen(false);
  };

  const setPropSource = (val) => {
    let propMap = Object.assign({}, propertyMap);
    propMap['$source']['va'] = reversePropertyNameMap(val[0]);
    if (val[0].length !== 0 && isSearchedValue(val[0]))
      propMap['$source']['ty'] = 'Constant';
    else propMap['$source']['ty'] = 'Property';
    setPropertyMap(propMap);
    setSourceSelectorOpen(false);
  };

  const setPropCampaign = (val) => {
    let propMap = Object.assign({}, propertyMap);
    propMap['$campaign']['va'] = reversePropertyNameMap(val[0]);
    if (val[0].length !== 0 && isSearchedValue(val[0]))
      propMap['$campaign']['ty'] = 'Constant';
    else propMap['$campaign']['ty'] = 'Property';

    setPropertyMap(propMap);
    setCampaignSelectorOpen(false);
  };

  const setPropChannel = (val) => {
    let propMap = Object.assign({}, propertyMap);
    propMap['$channel']['va'] = reversePropertyNameMap(val[0]);
    if (val[0].length !== 0 && isSearchedValue(val[0]))
      propMap['$channel']['ty'] = 'Constant';
    else propMap['$channel']['ty'] = 'Property';

    setPropertyMap(propMap);
    setChannelSelectorOpen(false);
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

  const renderEventPropertyCampOptions = () => {
    const eventToCall = getEventToCall();
    const propertiesMp = [];
    const startsWith = getStartsWith();
    if (tchType === '2') {
      eventProperties[eventToCall]?.forEach((prop) => {
        if (startsWith?.length ? prop[1]?.startsWith(startsWith) : true) {
          propertiesMp.push([prop[0]]);
        }
      });
      tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS &&
        userProperties.forEach((prop) => {
          if (prop[1]?.startsWith(startsWith)) {
            propertiesMp.push([prop[0]]);
          }
        });
    } else if (tchType === '3') {
      eventProperties[eventToCall]?.forEach((prop) => {
        if (prop[1]?.startsWith(startsWith)) {
          propertiesMp.push([prop[0]]);
        }
      });
      userProperties.forEach((prop) => {
        if (prop[1]?.startsWith(startsWith)) {
          propertiesMp.push([prop[0]]);
        }
      });
    }
    return propertiesMp;
  };

  const propertyNameMap = (val) => {
    const eventToCall = getEventToCall();
    if (eventProperties[eventToCall] === undefined) return '';
    const index = eventProperties[eventToCall]
      ?.map((prop) => prop[1])
      .indexOf(val);
    if (index === -1) return val;
    const name = eventProperties[eventToCall]?.[index]?.[0];
    return name === undefined ? '' : name;
  };
  const reversePropertyNameMap = (val) => {
    const eventToCall = getEventToCall();
    if (eventProperties[eventToCall] === undefined) return '';
    const index = eventProperties[eventToCall]
      ?.map((prop) => prop[0])
      .indexOf(val);
    if (index === -1) return val;
    const name = eventProperties[eventToCall]?.[index]?.[1];
    return name === undefined ? '' : name;
  };

  //To check if the value is new value entered by user.
  const isSearchedValue = (val) => {
    const eventToCall = getEventToCall();
    const index1 = eventProperties[eventToCall]
      ?.map((prop) => prop[0])
      .indexOf(val);
    const index2 = eventProperties[eventToCall]
      ?.map((prop) => prop[1])
      .indexOf(val);
    if (index1 === -1 && index2 === -1) return true;
    return false;
  };
  const renderTypePropertyOptions = () => {
    const options = [];
    if (tchRuleType !== RULE_TYPE_HS_FORM_SUBMISSIONS) options.push(['Tactic']);
    if (tchRuleType !== RULE_TYPE_HS_EMAILS) options.push(['Offer']);
    return options;
  };
  const renderPropertyMap = () => {
    return (
      <div className={`border-top--thin pt-5 mt-8 `}>
        <Row>
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>
            Map the properties<sup>*</sup>
          </Text>
        </Row>

        <Row className={`mt-10`}>
          <Col span={7}>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              Type
            </Text>
          </Col>
          <div
            className={`flex flex-col relative items-center ${styles.dropdown}`}
          >
            <Tooltip
              title={
                propertyMap['$type']['va'] === ''
                  ? 'Select Type Property'
                  : toCapitalCase(propertyMap['$type']['va'])
              }
            >
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                onClick={() => setTypeSelectorOpen(true)}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {propertyMap['$type']['va'] === ''
                    ? 'Select Type'
                    : toCapitalCase(propertyMap['$type']['va'])}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
            </Tooltip>
            {typeSelectorOpen && (
              <FaSelect
                options={renderTypePropertyOptions()}
                optionClick={(val) => setPropType(val)}
                onClickOutside={() => setTypeSelectorOpen(false)}
                extraClass={`${styles.dropdownSelect}`}
              ></FaSelect>
            )}
          </div>
        </Row>

        <Row className={`mt-4`}>
          <Col span={7}>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              Source
            </Text>
          </Col>
          <div
            className={`flex flex-col relative items-center ${styles.dropdown}`}
          >
            <Tooltip
              title={
                propertyMap['$source']['va'] === ''
                  ? 'Select Source Property'
                  : propertyNameMap(propertyMap['$source']['va'])
              }
            >
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                onClick={() => setSourceSelectorOpen(true)}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {propertyMap['$source']['va'] === ''
                    ? 'Select Source Property'
                    : propertyNameMap(propertyMap['$source']['va'])}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
            </Tooltip>
            {sourceSelectorOpen && (
              <FaSelect
                allowSearch
                options={renderEventPropertyCampOptions()}
                optionClick={(val) => setPropSource(val)}
                onClickOutside={() => setSourceSelectorOpen(false)}
                extraClass={`${styles.dropdownSelect}`}
              ></FaSelect>
            )}
          </div>
        </Row>

        <Row className={`mt-4`}>
          <Col span={7}>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              Campaign
            </Text>
          </Col>

          <div
            className={`flex flex-col relative items-center ${styles.dropdown}`}
          >
            <Tooltip
              title={
                propertyMap['$campaign']['va'] === ''
                  ? 'Select Campaign Property'
                  : propertyNameMap(propertyMap['$campaign']['va'])
              }
            >
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                onClick={() => setCampaignSelectorOpen(true)}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {propertyMap['$campaign']['va'] === ''
                    ? 'Select Campaign Property'
                    : propertyNameMap(propertyMap['$campaign']['va'])}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
            </Tooltip>
            {campaignSelectorOpen && (
              <FaSelect
                allowSearch
                options={renderEventPropertyCampOptions()}
                optionClick={(val) => setPropCampaign(val)}
                onClickOutside={() => setCampaignSelectorOpen(false)}
                extraClass={`${styles.dropdownSelect}`}
              ></FaSelect>
            )}
          </div>
        </Row>

        <Row className={`mt-4`}>
          <Col span={7}>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              Channel
            </Text>
          </Col>

          <div
            className={`flex flex-col relative items-center ${styles.dropdown}`}
          >
            <Tooltip
              title={
                propertyMap['$channel']['va'] === ''
                  ? 'Select Channel Property'
                  : propertyNameMap(propertyMap['$channel']['va'])
              }
            >
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                onClick={() => setChannelSelectorOpen(true)}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {propertyMap['$channel']['va'] === ''
                    ? 'Select Channel Property'
                    : propertyNameMap(propertyMap['$channel']['va'])}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
            </Tooltip>
            {channelSelectorOpen && (
              <FaSelect
                allowSearch
                options={renderEventPropertyCampOptions()}
                optionClick={(val) => setPropChannel(val)}
                onClickOutside={() => setChannelSelectorOpen(false)}
                extraClass={`${styles.dropdownSelect}`}
              ></FaSelect>
            )}
          </div>
        </Row>
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
      //parse and set filterstate
      filters: getFiltersWithoutOrProperty(newFilterStates),
      // set propMap
      properties_map: propMap,
      touch_point_time_ref: touchPointPropRef
    };
    touchPointObj['rule_type'] = tchRuleType;
    onSave(touchPointObj);
  };

  const renderFooterActions = () => {
    return (
      <div>
        <Row className={`mt-20 relative justify-start`}>
          <Text
            level={7}
            type={'title'}
            extraClass={'m-0 italic'}
            weight={'thin'}
          >
            <sup>*</sup> All these fields are mandatory
          </Text>
        </Row>
        <Row className={`border-top--thin mt-4 relative justify-start`}>
          <Col className={`mt-6`} span={10}>
            <Button size={'large'} onClick={() => onCancel()}>
              Cancel
            </Button>
            <Button
              disabled={!validateRuleInfo() || !validateInputs()}
              size={'large'}
              type='primary'
              className={'ml-2'}
              htmlType='submit'
              onClick={onSaveToucPoint}
            >
              Save
            </Button>
          </Col>
        </Row>
      </div>
    );
  };

  //Rule Type Selection
  const ruleTypeSelect = (val) => {
    setTchRuleType(
      tchType === '2'
        ? ruleTypesNameMappingForHS[val]
        : ruleTypesNameMappingForSF[val]
    );
    setRuleSelectorOpen(false);
  };
  const renderTchRuleTypeOptions = () => {
    let ruleTypes = Object.keys(ruleTypesNameMappingForHS).map((type) => [
      type
    ]);
    if (tchType === '3') {
      ruleTypes = Object.keys(ruleTypesNameMappingForSF).map((type) => [type]);
    }
    return (
      <div className={`flex flex-col relative items-center ${styles.dropdown}`}>
        <Tooltip title='Change in Hubspot Contact Field Value'>
          <Button
            className={`${styles.dropdownbtn}`}
            type='text'
            onClick={() => setRuleSelectorOpen(true)}
          >
            <div className={styles.dropdownbtntext + '  text-sm'}>
              {tchType === '2'
                ? reverseRuleTypesNameMappingForHS[tchRuleType]
                : reverseRuleTypesNameMappingForSF[tchRuleType]}{' '}
            </div>
            <div className={styles.dropdownbtnicon}>
              <SVG name='caretDown' size={18} />
            </div>
          </Button>
        </Tooltip>
        {ruleSelectorOpen && (
          <FaSelect
            options={ruleTypes}
            optionClick={(val) => ruleTypeSelect(val[0])}
            onClickOutside={() => setRuleSelectorOpen(false)}
            extraClass={`${styles.dropdownSelect}`}
          ></FaSelect>
        )}
      </div>
    );
  };
  const renderTchRuleType = () => {
    return (
      <div className={`mt-8`}>
        <Row className={`mt-2`}>
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
            Create a touchpoint using<sup>*</sup>
          </Text>
        </Row>
        <Row className={`mt-4`}>{renderTchRuleTypeOptions()}</Row>
      </div>
    );
  };

  //Extra Property only in Form Submission
  const setExtraMapByProp = (extraProp) => {
    const extraMap = { ...extraPropMap };
    extraMap[`$` + extraProp] = {
      ty: 'Property',
      va: ''
    };
    setExtraPropMap(extraMap);
  };
  const setExtraPropVal = (val, key) => {
    let propMap = Object.assign({}, extraPropMap);
    propMap['$' + key]['va'] = reversePropertyNameMap(val[0]);
    if (val[0].length !== 0 && isSearchedValue(val[0]))
      propMap['$' + key]['ty'] = 'Constant';
    else propMap['$' + key]['ty'] = 'Property';
    setExtraPropMap(propMap);
    setExtraPropSelectorOpen(false);
  };
  const renderAddExtraPropBtn = () => {
    return (
      <div className={`mr-2 items-center relative`}>
        <Button
          type='link'
          icon={<SVG name={'plus'} color={'grey'} />}
          onClick={() => setExtraPropBtn(!extraPropBtn)}
        >
          Add touchpoint property
        </Button>

        {extraPropBtn && (
          <FaSelect
            options={Extra_PROP_SHOW_OPTIONS}
            optionClick={(op) => {
              setExtraMapByProp(op[2]);
              setExtraPropBtn(false);
            }}
            onClickOutside={() => setExtraPropBtn(false)}
          ></FaSelect>
        )}
      </div>
    );
  };
  const renderExtraPropMap = () => {
    const extraMapRows = [];
    Extra_PROP_SHOW_OPTIONS.forEach((key, index) => {
      if (!Object.keys(extraPropMap).includes(`$` + key[2])) return null;
      extraMapRows.push(
        <Row key={index} className={`mt-4`}>
          <Col span={7}>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              {key[0]}
            </Text>
          </Col>
          <div
            className={`flex flex-col relative items-center ${styles.dropdown}`}
          >
            <Tooltip title='Select Property'>
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                onClick={() => setExtraPropSelectorOpen(true)}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {extraPropMap['$' + key[2]]['va'] === ''
                    ? 'Select Property'
                    : propertyNameMap(extraPropMap[`$` + key[2]]['va'])}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
            </Tooltip>
            {extraPropSelectorOpen && (
              <FaSelect
                allowSearch
                options={renderEventPropertyCampOptions()}
                optionClick={(val) => setExtraPropVal(val, key[2])}
                onClickOutside={() => setExtraPropSelectorOpen(false)}
                extraClass={`${styles.dropdownSelect}`}
              ></FaSelect>
            )}
          </div>
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
  eventProperties: state.coreQuery.eventProperties,
  userProperties: state.coreQuery.userProperties
});
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(TouchpointView);
