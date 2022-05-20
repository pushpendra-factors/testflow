import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select, message
} from 'antd';
import { SVG, Text } from 'factorsComponents';
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import { fetchEventNames, getUserProperties, getEventProperties } from 'Reducers/coreQuery/middleware';
import { fetchGoalInsights, fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel, fetchFactorsModelMetadata } from 'Reducers/factors';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import _ from 'lodash';
import FilterBlock from '../../components/QueryComposer/FilterBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
import EventFilterBy from './DrawerUtil/EventFilterBy';
// import MomentTz from 'Components/MomentTz';
import moment from 'moment-timezone';
import FaSelect from 'Components/FaSelect';
import ComposerBlock from '../../components/QueryCommons/ComposerBlock';
import EventTag from './FactorsInsightsNew/Components/EventTag'
import factorsai from 'factorsai';


const title = (props) => {
  return (
    <div className={'flex justify-between items-center'}>
      <div className={'flex'}>
        <SVG name={'templates_cq'} size={24} />
        <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>New Goal</Text>
      </div>
      <div className={'flex justify-end items-center'}>
        <Button size={'large'} type="text" onClick={() => props.onClose()}><SVG name="times"></SVG></Button>
      </div>
    </div>
  );
};

const CreateGoalDrawer = (props) => {
  const history = useHistory();
  const { Option } = Select;

  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  moment.tz.setDefault(timeZone);


  const [TrackedEventNames, SetTrackedEventNames] = useState([]);

  const [EventNames, SetEventNames] = useState([]);
  const [eventCount, SetEventCount] = useState(2);

  const [showDropDown, setShowDropDown] = useState(false);
  const [event1, setEvent1] = useState(null);

  const [showDropDown2, setShowDropDown2] = useState(false);
  const [event2, setEvent2] = useState(null);

  const [showFtDropDown, setshowFtDropDown] = useState(false);
  const [globalFilter, setglobalFilter] = useState([]);
  const [filterLoader, setfilterLoader] = useState(false);

  const [showDateTime, setShowDateTime] = useState(false);
  const [selectedModel, setSelectedModel] = useState(null);
  const [insightBtnLoading, setInsightBtnLoading] = useState(false);
  const [collapse, setCollapse] = useState(false);

  const [filtersEvent1, setfiltersEvent1] = useState([]);
  const [showEventFilter1DD, setEventFilter1DD] = useState(false);
  const [filtersEvent2, setfiltersEvent2] = useState([]);
  const [showEventFilter2DD, setEventFilter2DD] = useState(false);

  const [filters, setfilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({
    user: []
  });
  const [filterDD, setFilterDD] = useState(false);

  const [eventsBlockOpen, setEventsBlockOpen] = useState(true);
  const [eventsToIncBlock, setEventsToIncBlock] = useState(true);
  const [showDateBlock, setDateBlock] = useState(true);

  const [eventsToInc, setEventsToInc] = useState([]);
  const [showEventsToIncDD, setShowEventsToIncDD] = useState(false);

  const [modelMetadata, setModelMetadata] = useState([]);

  const operatorMap = {
    "=": "equals",
    "!=": "notEqual",
    contains: "contains",
    "does not contain": "notContains",
    "<": "lesserThan",
    "<=": "lesserThanOrEqual",
    ">": "greaterThan",
    ">=": "greaterThanOrEqual",
  };

  const getFilters = (filters) => {
    const result = [];
    filters.forEach((filter) => {
      filter.values.forEach((value, index) => {
        result.push({
          en: filter.props[2],
          lop: !index ? "AND" : "OR",
          op: operatorMap[filter.operator],
          pr: filter.props[0],
          ty: filter.props[1],
          va: value,
        });
      });
    });
    return result;
  };

  const onChangeGroupSelect1 = (grp, value) => {
    setShowDropDown(false);
    // console.log("value-->",value);
    // console.log("event value-->",value);
    setEvent1(value[0]);
  }
  // const onChangeFilter = (grp, value) => {
  //   setshowFtDropDown(false)
  //   console.log('onChangeFilter',value)
  //   setfilterLoader(true);
  //   fetchUserPropertyValues(props.activeProject.id,value[0]).then(res => {
  //     setglobalFilter([ ...globalFilter , 
  //       {
  //         "key": value[0],
  //         "vl": res.data
  //       }
  //   ]); 
  //   console.log('fetchUserPropertyValues',res);
  //   setfilterLoader(false);
  //   });
  //   // setEvent1(value[0]); 
  // }
  const removeFilter = (index) => {
    const fltrs = globalFilter.filter((v, i) => i !== index);
    setglobalFilter(fltrs);
  }
  const onChangeGroupSelect2 = (grp, value) => {

    setShowDropDown2(false);
    setEvent2(value[0]);
  }
  const onChangeDateTime = (grp, value) => {
    setShowDateTime(false);
    setSelectedModel(value);
  }

  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  }
  const factorsModels = !_.isEmpty(props.factors_models) && _.isArray(props.factors_models) ? props.factors_models.map((item) => { return [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`] }) : [];

  // useEffect(()=>{ 
  //   console.log('event-->>',event1);
  //   if(event1){
  //     props.getEventProperties(props.activeProject.id,event1)
  //   }
  // },[event1]); 
  const modelMetaDataFn = (projectID, modelID) => {
    props.fetchFactorsModelMetadata(projectID, modelID);
    if (props.factors_model_metadata) {
      setModelMetadata(props?.factors_model_metadata[0]?.events)
    }
  }

  useEffect(() => {
    let calcModelId = modelIDtoStringMap();
    if (calcModelId) {
      modelMetaDataFn(props?.activeProject?.id, calcModelId[0]?.mid);
    }
  }, [selectedModel]);

  useEffect(() => {
    let goalInsights = props.goal_insights
    if (goalInsights) {
      console.log("coming from saved insights", goalInsights);
      if (goalInsights.type == "singleevent") {
        setEvent1(goalInsights?.goal?.en_en)
      }
      else {
        setEvent1(goalInsights?.goal?.st_en)
        setEvent2(goalInsights?.goal?.en_en)
      }
    }
  }, []);

  const matchEventName = (item) => {
    let findItem = props?.eventPropNames?.[item] || props?.userPropNames?.[item]
    return findItem ? findItem : item
  }

  useEffect(() => {
    // if(!props.GlobalEventNames || !factorsModels){
    //   const getData = async () => {
    //     await props.fetchEventNames(props.activeProject.id);
    //     await props.fetchFactorsModels(props.activeProject.id);
    //   };
    //   getData();    
    // }
    // if(props.GlobalEventNames){ 
    //   SetEventNames(props.GlobalEventNames);

    // }

    //fetching model metada for 'events to include'
    let modelID = props.factors_models ? props.factors_models[0]?.mid : 0;
    modelMetaDataFn(props.activeProject.id, modelID);

    if (props.factors_models) {
      if (props.factors_insight_model) {
        setSelectedModel(props.factors_insight_model);
      } else {
        setSelectedModel(factorsModels[0]);
      }
    }
    if (props.activeProject && props.activeProject.id) {
      props.getUserProperties(props.activeProject.id, 'channel')
    }
    if (props.tracked_events) {
      const fromatterTrackedEvents = props.tracked_events.map((item) => {
        let displayName = matchEventName(item.name)
        return [displayName]
      });
      SetTrackedEventNames(fromatterTrackedEvents);
    }
  }, [props.activeProject, props.tracked_events, props.factors_models, props.goal_insights, props.factors_insight_model])

  useEffect(() => {
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.user = props.userProperties;
    let catAndNumericalProps = [];

    props.userProperties.map((item) => {
      if (item[1] == 'categorical' || item[1] == 'numerical') {
        catAndNumericalProps.push(item);
      }
    });
    assignFilterProps.user = catAndNumericalProps;
    setFilterProperties(assignFilterProps);

  }, [props.userProperties]);


  const factorsDataFormatNew = {
    "name": "",
    "rule": {
      "st_en": {
        "na": "",
        "pr": [
          // {
          //     "en": "user",
          //     "lop": "",
          //     "op": "equals",
          //     "pr": "$user_id12",
          //     "ty": "categorical",
          //     "va": "1"
          // }
        ]
      },
      "en_en": {
        "na": "",
        "pr": []
      },
      "gpr": [],
      "vs": true
    }
  };


  // const factorsDataFormat = {
  //   name: "",
  //   rule: {
  //       st_en: "",
  //       en_en: "",
  //       vs: true,
  //       rule: {
  //           ft: []
  //       }
  //   }
  // };

  const smoothScroll = (element) => {
    document.querySelector(element).scrollIntoView({
      behavior: 'smooth',
    });
  }

  const modelIDtoStringMap = () => {
    if (_.isArray(props?.factors_models)) {
      return props?.factors_models?.filter((item) => {
        const generateStringArray = [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`];
        if (_.isEqual(selectedModel, generateStringArray)) {
          return item
        }
      });
    }
    else return []

  }


  const getInsights = (projectID, isJourney = false) => {
    setInsightBtnLoading(true);
    const calcModelId = modelIDtoStringMap();

    // console.log("calcModelId",calcModelId[0].mid);
    let gprData = getFilters(filters);
    let event1pr = getFilters(filtersEvent1);
    let event2pr = getFilters(filtersEvent2);
    let factorsData = {
      ...factorsDataFormatNew,
      rule: {
        ...factorsDataFormatNew.rule,
        st_en: {
          na: event2 ? (event2 ? event1 : null) : null,
          pr: event2 ? (event2pr ? event1pr : null) : null,
        },
        en_en: {
          na: event2 ? (event2 ? event2 : event1) : event1,
          pr: event2 ? (event2pr ? event2pr : event1pr) : event1pr,
        },
        gpr: gprData,
        in_en: eventsToInc
      }
    }

    props.fetchGoalInsights(projectID, isJourney, factorsData, calcModelId[0]?.mid).then((data) => {
      props.saveGoalInsightRules(factorsData);
      props.saveGoalInsightModel(selectedModel);
      setInsightBtnLoading(false);
      // history.push('/explain/insights');
      setCollapse(true);
      setTimeout(() => {
        smoothScroll('#explain-builder--footer');
      }, 200);

    }).catch((err) => {
      console.log("fetchGoalInsights catch", err);
      const ErrMsg = err?.data?.error ? err.data.error : `Oops! Something went wrong!`;
      message.error(ErrMsg);
      setInsightBtnLoading(false);
    });

    //Factors RUN_EXPLAIN tracking
    factorsai.track('RUN_EXPLAIN', { 'query_type': 'explain', project_name: props.activeProject.name, project_id: props.activeProject.id });

  }


  const valuesSelect = (val) => {
    let values = val.map((vl) => JSON.parse(vl)[0]);
    setEventsToInc(values);
    setShowEventsToIncDD(false);
  };

  const modelMetadataDDValue = modelMetadata?.map((item) => { return [matchEventName(item)] })
  const queryBuilderCollapse = !_.isEmpty(props?.goal_insights?.insights)

  return (
    <div
    >
      {/* <Drawer
        title={title(props)}
        placement="left"
        closable={false}
        visible={props.visible}
        onClose={props.onClose}
        getContainer={false}
        width={'650px'}
        className={'fa-drawer'}
      > */}

      <div className={`flex flex-col py-4 px-8 border--thin-2 relative `}>
        <div className={`explain-builder--content ${collapse ? 'explain-builder--collapsed' : ''}`}>


          <ComposerBlock blockTitle={'SELECT ANALYSIS WINDOW'} isOpen={showDateBlock}
            showIcon={true} onClick={() => setDateBlock(!showDateBlock)}
            extraClass={`no-padding-l`}
          >
            <div className={'relative mr-2'}>
              {<Button size={'large'} onClick={() => setShowDateTime(true)} className='mt-2 border--thin-2 '><SVG name={'calendar'} extraClass={'mr-1'} />{selectedModel ? selectedModel : 'Select Date Range'} </Button>}
              {showDateTime &&
                <GroupSelect2
                  groupedProperties={factorsModels ? [
                    {
                      label: 'Most Recent',
                      icon: 'most_recent',
                      values: factorsModels
                    }
                  ] : null}
                  placeholder="Select Date Range "
                  optionClick={(group, val) => onChangeDateTime(group, val)}
                  onClickOutside={() => setShowDateTime(false)}
                />
              }
            </div>

          </ComposerBlock>

          <ComposerBlock blockTitle={'EXPLAIN CONVERSIONS BETWEEN'} isOpen={eventsBlockOpen}
            showIcon={true} onClick={() => setEventsBlockOpen(!eventsBlockOpen)}
            extraClass={`no-padding-l`}
          >

            <Row gutter={[24, 4]}>
              <Col span={24}>
                <div className={'mt-4'}>

                  <div className={'flex flex-col'}>

                    <div className={'flex items-center justify-start query_block--actions '}>
                      <div className={'flex items-center'}>
                        {event1 && <>
                          <EventTag />
                          {/* <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>Users who perform</Text>  */}
                        </>}
                        <div className='relative' style={{ height: '42px' }}>
                          {!showDropDown && !event1 && <Button onClick={() => setShowDropDown(true)} type={'text'} size={'large'} icon={<SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'} />}>{eventCount === 2 ? 'Add First event' : 'Add an event'}</Button>}
                          {showDropDown && <>
                            <GroupSelect2
                              allowEmpty={true}
                              groupedProperties={TrackedEventNames ? [
                                {
                                  label: 'Most Recent',
                                  icon: 'most_recent',
                                  values: TrackedEventNames
                                }
                              ] : null}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect1(group, val)}
                              onClickOutside={() => setShowDropDown(false)}
                            />
                          </>
                          }

                          {event1 && !showDropDown && <Button type={'link'} size={'large'} className={'ml-2 fa-button--truncate fa-button--truncate-sm'} ellipsis onClick={() => {
                            setShowDropDown(true);
                          }} >{event1}</Button>
                          }
                        </div>
                      </div>
                      {event1 && <Button type={'text'} onClick={() => setEventFilter1DD(true)} className={'fa-btn--custom m-0'}><SVG name={'filter'} extraClass={'m-0'} /></Button>}
                      {event1 && <Button type={'text'} onClick={() => setEvent1(null)} className={'fa-btn--custom m-0'}><SVG name={'delete'} extraClass={'m-0'} /></Button>}
                    </div>
                    <EventFilterBy event={event1} setfiltersParent={setfiltersEvent1} showEventFilterDD={showEventFilter1DD} setEventFilterDD={setEventFilter1DD} />
                  </div>
                </div>
              </Col>
            </Row>

            {(event1 || event2) &&
              <Row gutter={[24, 4]}>
                <Col span={24}>
                  <div className={'mt-4'}>

                    <div className={'flex flex-col'}>

                      <div className={'flex items-center justify-start query_block--actions'}>
                        <div className={'flex items-center'}>

                          {event2 && <>
                            <EventTag text={'B'} color={'yellow'} />
                            {/* <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>And then</Text> */}
                            {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                          </>}
                          <div className='relative' style={{ height: '42px' }}>
                            {!showDropDown2 && !event2 && <Button onClick={() => setShowDropDown2(true)} type={'text'} size={'large'} icon={<SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'} />}>Add next event</Button>}
                            {showDropDown2 && <>

                              <GroupSelect2
                                allowEmpty={true}
                                groupedProperties={TrackedEventNames ? [
                                  {
                                    label: 'Most Recent',
                                    icon: 'most_recent',
                                    values: TrackedEventNames
                                  }
                                ] : null}
                                placeholder="Select Events"
                                optionClick={(group, val) => onChangeGroupSelect2(group, val)}
                                onClickOutside={() => setShowDropDown2(false)}
                              />
                            </>
                            }

                            {event2 && !showDropDown2 && <Button type={'link'} size={'large'} className={'ml-2 fa-button--truncate fa-button--truncate-xs'} ellipsis onClick={() => {
                              setShowDropDown2(true);
                            }} >{event2}</Button>
                            }
                          </div>
                        </div>
                        {event2 && <Button type={'text'} onClick={() => setEventFilter2DD(true)} className={'fa-btn--custom m-0'}><SVG name={'filter'} extraClass={'m-0'} /></Button>}
                        {event2 && <Button type={'text'} onClick={() => setEvent2(null)} className={'fa-btn--custom m-0'}><SVG name={'delete'} extraClass={'m-0'} /></Button>}
                      </div>
                      <EventFilterBy event={event2} setfiltersParent={setfiltersEvent2} showEventFilterDD={showEventFilter2DD} setEventFilterDD={setEventFilter2DD} />
                      {/* {event2 && <EventFilterBy setfiltersParent={setfiltersEvent2} /> } */}
                    </div>
                  </div>
                </Col>
              </Row>
            }


          </ComposerBlock>

          <ComposerBlock blockTitle={'EVENTS TO INCLUDE'} isOpen={eventsToIncBlock}
            showIcon={true} onClick={() => setEventsToIncBlock(!eventsToIncBlock)}
            extraClass={`no-padding-l`}
          >

            <div>


              {eventsToInc && eventsToInc?.map((item) => {
                return <Button
                  className={`ml-2 mt-4 flex justify-start`}
                  type='link'
                  onClick={() => setShowEventsToIncDD(true)}
                >
                  {item}
                </Button>
              })}

              <Button
                className={` ml-2 mt-4`}
                type={'text'}
                onClick={() => setShowEventsToIncDD(true)}
                icon={<SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'} />}
              >
                {`Add Event`}
              </Button>

              {showEventsToIncDD && (
                <FaSelect
                  options={modelMetadata ? modelMetadataDDValue : []}
                  onClickOutside={() => setShowEventsToIncDD(false)}
                  applClick={(val) => valuesSelect(val)}
                  allowSearch={true}
                  multiSelect={true}
                  selectedOpts={eventsToInc}

                />
              )}
            </div>

          </ComposerBlock>

        </div>
        <div id={`explain-builder--footer`} className={`flex items-center pt-4 border-top--thin-2 justify-end`} >

          {(queryBuilderCollapse || collapse) && <Button className={`mr-2`} size={'large'} type={'default'} onClick={() => setCollapse(!collapse)} > <SVG name={collapse ? `Expand` : `arrowUp`} size={20} extraClass={`mr-1`}></SVG>{`${collapse ? 'Expand' : 'Collapse all'}`} </Button>}
          <Button type="primary" size={'large'} loading={insightBtnLoading} disabled={!(event1 && selectedModel)} onClick={() => getInsights(props.activeProject.id, eventCount === 2 ? true : false)}>Find Insights</Button>

        </div>


      </div>

      {/* </Drawer> */}
      {!queryBuilderCollapse && <div style={{ marginBottom: '200px' }} />}
    </div>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    userProperties: state.coreQuery.userProperties,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    userProperties: state.coreQuery.userProperties,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events,
    factors_model_metadata: state.factors.factors_model_metadata,
    factors_insight_model: state.factors.factors_insight_model,
    userPropNames: state.coreQuery?.userPropNames,
    eventPropNames: state.coreQuery?.eventPropNames,
  };
};
export default connect(mapStateToProps, {
  fetchEventNames, fetchGoalInsights,
  fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel, getUserProperties, fetchUserPropertyValues, getEventProperties, fetchFactorsModelMetadata
})(CreateGoalDrawer);
