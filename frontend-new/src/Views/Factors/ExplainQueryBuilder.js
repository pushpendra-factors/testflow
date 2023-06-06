import React, { useEffect, useState } from 'react';
import { Modal, Form, Input, Button, Row, Col, Select, message } from 'antd';
import { SVG, Text } from 'factorsComponents';
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import {
  fetchEventNames,
  getUserProperties,
  getEventProperties
} from 'Reducers/coreQuery/middleware';
import {
  createExplainJob,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel,
  fetchFactorsModelMetadata
} from 'Reducers/factors';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import _ from 'lodash';
import EventFilterBy from './DrawerUtil/EventFilterBy';
// import MomentTz from 'Components/MomentTz';
import moment from 'moment-timezone';
import FaSelect from 'Components/FaSelect';
import ComposerBlock from '../../components/QueryCommons/ComposerBlock';
import EventTag from './FactorsInsightsNew/Components/EventTag';
import factorsai from 'factorsai';
import FaDatepicker from 'Components/FaDatepicker';
// import { operatorMap } from 'Utils/operatorMapping';

const symbolToTextConv = (symbol) =>{
  switch(symbol){
    case '=':
      return 'equals';
    case '!=':
      return 'notEqual';
    default:
      return 'equals'
  }
}


const title = (props) => {
  return (
    <div className={'flex justify-between items-center'}>
      <div className={'flex'}>
        <SVG name={'templates_cq'} size={24} />
        <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>
          New Goal
        </Text>
      </div>
      <div className={'flex justify-end items-center'}>
        <Button size={'large'} type='text' onClick={() => props.onClose()}>
          <SVG name='times'></SVG>
        </Button>
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
  const [selectedDateRange, setSelectedDateRange] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [loading, setLoading] = useState(false);


  const defaultStartDate = selectedDateRange ? selectedDateRange?.startDate : moment().subtract(1, 'days')
  const defaultEndDate = selectedDateRange ? selectedDateRange?.endDate : moment().subtract(1, 'days')

  const getFilters = (filters) => {
    const result = [];
    filters.forEach((filter) => {
      filter.values.forEach((value, index) => {
        result.push({
          en: filter.props[2],
          lop: !index ? 'AND' : 'OR',
          op: symbolToTextConv(filter.operator),
          pr: filter.props[0],
          ty: filter.props[1],
          va: value
        });
      });
    });
    return result;
  };

  const onChangeGroupSelect1 = (grp, value) => {
    setShowDropDown(false); 
    setEvent1(value[0]);
  };
  const removeFilter = (index) => {
    const fltrs = globalFilter.filter((v, i) => i !== index);
    setglobalFilter(fltrs);
  };
  const onChangeGroupSelect2 = (grp, value) => {
    setShowDropDown2(false);
    setEvent2(value[0]);
  };

  // const readableTimstamp = (unixTime) => {
  //   return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  // }

  // const factorsModels = !_.isEmpty(props.factors_models) && _.isArray(props.factors_models) ? props.factors_models.map((item) => { return [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`] }) : [];

  // const modelMetaDataFn = (projectID, modelID) => {
  //   props.fetchFactorsModelMetadata(projectID, modelID);
  //   if (props.factors_model_metadata) {
  //     setModelMetadata(props?.factors_model_metadata[0]?.events)
  //   }
  // }

  // useEffect(() => {
  //   let calcModelId = modelIDtoStringMap();
  //   if (calcModelId) {
  //     modelMetaDataFn(props?.activeProject?.id, calcModelId[0]?.mid);
  //   }
  // }, [selectedModel]);

  useEffect(() => {
    let goalInsights = props.goal_insights;
    if (goalInsights) {

      let defaultDate = {
        startDate: moment.unix(goalInsights?.sts),
        endDate: moment.unix(goalInsights?.ets)
      };
      setSelectedDateRange(defaultDate);

      if (goalInsights.type == 'singleevent') {
        if (goalInsights?.goal?.st_en == '') {
          setEvent1(goalInsights?.goal?.en_en);
        } else {
          setEvent1(goalInsights?.goal?.st_en);
          setEvent2(goalInsights?.goal?.en_en);
        }
      } else {
        setEvent1(goalInsights?.goal?.st_en);
        setEvent2(goalInsights?.goal?.en_en);
      }
      
      if(goalInsights?.goal?.rule?.in_en){
        setEventsToInc(goalInsights?.goal?.rule?.in_en) 
      }

    }
  }, []);

  const matchEventName = (item) => {
    let findItem =
      props?.eventPropNames?.[item] || props?.userPropNames?.[item];
    return findItem ? findItem : item;
  };

  useEffect(() => {
    //fetching model metada for 'events to include'

    // let modelID = props.factors_models ? props.factors_models[0]?.mid : 0;

    // if (props.factors_models) {
    //   if (props.factors_insight_model) {
    //     setSelectedModel(props.factors_insight_model);
    //   } else {
    //     setSelectedModel(factorsModels[0]);
    //   }
    // }
    if (props.activeProject && props.activeProject.id) {
      props.getUserProperties(props.activeProject.id, 'channel');
    }
    if (props.tracked_events) {
      const fromatterTrackedEvents = props.tracked_events.map((item) => {
        let displayName = matchEventName(item.name);
        return [displayName];
      });
      SetTrackedEventNames(fromatterTrackedEvents);
    }
  }, [
    props.activeProject,
    props.tracked_events,
    props.factors_models,
    props.goal_insights,
    props.factors_insight_model
  ]);

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

  const smoothScroll = (element) => {
    document.querySelector(element).scrollIntoView({
      behavior: 'smooth'
    });
  };

  // const modelIDtoStringMap = () => {
  //   if (_.isArray(props?.factors_models)) {
  //     return props?.factors_models?.filter((item) => {
  //       const generateStringArray = [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`];
  //       if (_.isEqual(selectedModel, generateStringArray)) {
  //         return item
  //       }
  //     });
  //   }
  //   else return []

  // }

  const delOption = (val, index) => {
    const filteredItems = eventsToInc.filter((item) => item !== val);
    setEventsToInc(filteredItems);
  };

  const valuesSelect = (val) => {
    let values = val.map((vl) => JSON.parse(vl)[0]);
    setEventsToInc(values);
    setShowEventsToIncDD(false);
  };

  const removeEvent1 = () => {
    setEvent1(null);
    setfiltersEvent1([]);
  };
  const removeEvent2 = () => {
    setEvent2(null);
    setfiltersEvent2([]);
  };

  const modelMetadataDDValue = modelMetadata?.map((item) => {
    return [matchEventName(item)];
  });
  const queryBuilderCollapse = !_.isEmpty(props?.goal_insights?.insights);

  const SaveReportModal = () => {
    const [form] = Form.useForm();

    const getInsights = (reportName) => {
      setInsightBtnLoading(true);
      // const calcModelId = modelIDtoStringMap();
      let projectID = props.activeProject.id; 
      let gprData = getFilters(filters);
      let event1pr = getFilters(filtersEvent1);
      let event2pr = getFilters(filtersEvent2);
      let payload = {
        rule: {
          st_en: {
            na: event2 ? (event2 ? event1 : null) : null,
            pr: event2 ? (event2pr ? event1pr : null) : null
          },
          en_en: {
            na: event2 ? (event2 ? event2 : event1) : event1,
            pr: event2 ? (event2pr ? event2pr : event1pr) : event1pr
          },
          gpr: gprData,
          in_en: eventsToInc
        },
        name: reportName,
        sts: moment(defaultStartDate).unix(),
        ets: moment(defaultEndDate).unix()
      }; 

      // creating explain job
      props
        .createExplainJob(projectID, payload)
        .then((data) => {
          setInsightBtnLoading(false);
          history.push('/explain');
          message.success(
            'Your report is saved. Find under the saved reports of Explain 2.0'
          );
        })
        .catch((err) => {
          console.log('createExplainJob catch', err);
          const ErrMsg = err?.data?.error
            ? err.data.error
            : `Oops! Something went wrong!`;
          message.error(ErrMsg);
          setInsightBtnLoading(false);
        });

      //Factors RUN_EXPLAIN tracking
      factorsai.track('RUN_EXPLAIN', {
        query_type: 'explain',
        project_name: props.activeProject.name,
        project_id: props.activeProject.id
      });
    };

    const onFinish = (values) => {
      getInsights(values?.title).then(() => {
        setIsModalOpen(false);
      });
    };
    const onFinishFailed = (errorInfo) => {
      console.log('Failed:', errorInfo);
    };
    const onReset = () => {
      setIsModalOpen(false);
      form.resetFields();
    };
    return (
      <Modal
        title='Save report'
        visible={isModalOpen}
        okText={'Save'}
        footer={null}
        className='fa-modal--regular'
        afterClose={onReset}
        onCancel={onReset}
      >
        <Form
          name='basic'
          initialValues={{
            remember: true
          }}
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          autoComplete='off'
          form={form}
        >
          <Form.Item
            name='title'
            rules={[
              {
                required: true,
                message: 'Please enter a name (or) title for the report'
              }
            ]}
          >
            <Input placeholder={'Name'} className={'fa-input'} size={'large'} />
          </Form.Item>

          <div className='mt-2'>
            <Form.Item name='description'>
              <Input.TextArea
                className={'fa-input'}
                size={'large'}
                placeholder={'Description'}
              />
            </Form.Item>
          </div>

          <div className='mt-8 flex justify-end'>
            <Form.Item>
              <Button size={'large'} htmlType='button' onClick={onReset}>
                Cancel
              </Button>
              <Button
                size={'large'}
                type='primary'
                htmlType='submit'
                loading={insightBtnLoading}
                className={'ml-2'}
              >
                {' '}
                Save{' '}
              </Button>
            </Form.Item>
          </div>
        </Form>
      </Modal>
    );
  };

  return (
    <div>
      <div className={`flex flex-col py-4 px-20 border--thin-2 relative `}>
        <div
          className={`explain-builder--content ${
            collapse ? 'explain-builder--collapsed' : ''
          }`}
        >
          <ComposerBlock
            blockTitle={'SELECT ANALYSIS WINDOW'}
            isOpen={showDateBlock}
            showIcon={true}
            onClick={() => setDateBlock(!showDateBlock)}
            extraClass={`no-padding-l`}
          >
            <div className={'relative mt-4'}>
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                quarterPicker
                placement='bottomRight'
                buttonSize={'large'}
                todayPicker={false}
                range={{
                  startDate: defaultStartDate,
                  endDate: defaultEndDate
                }}
                onSelect={setSelectedDateRange}
              />
            </div>
          </ComposerBlock>

          <ComposerBlock
            blockTitle={'EXPLAIN CONVERSIONS BETWEEN'}
            isOpen={eventsBlockOpen}
            showIcon={true}
            onClick={() => setEventsBlockOpen(!eventsBlockOpen)}
            extraClass={`no-padding-l`}
          >
            <Row gutter={[24, 4]}>
              <Col span={24}>
                <div className={'mt-4'}>
                  <div className={'flex flex-col'}>
                    <div
                      className={
                        'flex items-center justify-start query_block--actions '
                      }
                    >
                      <div className={'flex items-center'}>
                        {event1 && (
                          <>
                            <EventTag />
                            {/* <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>Users who perform</Text>  */}
                          </>
                        )}
                        <div className='relative' style={{ height: '42px' }}>
                          {!showDropDown && !event1 && (
                            <Button
                              onClick={() => setShowDropDown(true)}
                              type={'text'}
                              size={'large'}
                              icon={
                                <SVG
                                  name={'plus'}
                                  size={14}
                                  color={'grey'}
                                  extraClass={'mr-2'}
                                />
                              }
                            >
                              {eventCount === 2
                                ? 'Add First event'
                                : 'Add an event'}
                            </Button>
                          )}
                          {showDropDown && (
                            <>
                              <GroupSelect2
                                allowEmpty={true}
                                groupedProperties={
                                  TrackedEventNames
                                    ? [
                                        {
                                          label: 'Most Recent',
                                          icon: 'most_recent',
                                          values: TrackedEventNames
                                        }
                                      ]
                                    : null
                                }
                                placeholder='Select Events'
                                optionClick={(group, val) =>
                                  onChangeGroupSelect1(group, val)
                                }
                                onClickOutside={() => setShowDropDown(false)}
                              />
                            </>
                          )}

                          {event1 && !showDropDown && (
                            <Button
                              type={'link'}
                              size={'large'}
                              className={
                                'ml-2 fa-button--truncate fa-button--truncate-sm'
                              }
                              ellipsis
                              onClick={() => {
                                setShowDropDown(true);
                              }}
                            >
                              {event1}
                            </Button>
                          )}
                        </div>
                      </div>
                      {event1 && filtersEvent1.length < 1 && (
                        <Button
                          type={'text'}
                          onClick={() => setEventFilter1DD(true)}
                          className={'fa-btn--custom m-0'}
                        >
                          <SVG name={'filter'} extraClass={'m-0'} />
                        </Button>
                      )}
                      {event1 && (
                        <Button
                          type={'text'}
                          onClick={() => removeEvent1()}
                          className={'fa-btn--custom m-0'}
                        >
                          <SVG name={'delete'} extraClass={'m-0'} />
                        </Button>
                      )}
                    </div>
                    <EventFilterBy
                      event={event1}
                      setfiltersParent={setfiltersEvent1}
                      showEventFilterDD={showEventFilter1DD}
                      setEventFilterDD={setEventFilter1DD}
                    />
                  </div>
                </div>
              </Col>
            </Row>

            {(event1 || event2) && (
              <Row gutter={[24, 4]}>
                <Col span={24}>
                  <div className={'mt-4'}>
                    <div className={'flex flex-col'}>
                      <div
                        className={
                          'flex items-center justify-start query_block--actions'
                        }
                      >
                        <div className={'flex items-center'}>
                          {event2 && (
                            <>
                              <EventTag text={'B'} color={'yellow'} />
                              {/* <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>And then</Text> */}
                              {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                            </>
                          )}
                          <div className='relative' style={{ height: '42px' }}>
                            {!showDropDown2 && !event2 && (
                              <Button
                                onClick={() => setShowDropDown2(true)}
                                type={'text'}
                                size={'large'}
                                icon={
                                  <SVG
                                    name={'plus'}
                                    size={14}
                                    color={'grey'}
                                    extraClass={'mr-2'}
                                  />
                                }
                              >
                                Add next event
                              </Button>
                            )}
                            {showDropDown2 && (
                              <>
                                <GroupSelect2
                                  allowEmpty={true}
                                  groupedProperties={
                                    TrackedEventNames
                                      ? [
                                          {
                                            label: 'Most Recent',
                                            icon: 'most_recent',
                                            values: TrackedEventNames
                                          }
                                        ]
                                      : null
                                  }
                                  placeholder='Select Events'
                                  optionClick={(group, val) =>
                                    onChangeGroupSelect2(group, val)
                                  }
                                  onClickOutside={() => setShowDropDown2(false)}
                                />
                              </>
                            )}

                            {event2 && !showDropDown2 && (
                              <Button
                                type={'link'}
                                size={'large'}
                                className={
                                  'ml-2 fa-button--truncate fa-button--truncate-xs'
                                }
                                ellipsis
                                onClick={() => {
                                  setShowDropDown2(true);
                                }}
                              >
                                {event2}
                              </Button>
                            )}
                          </div>
                        </div>
                        {event2 && filtersEvent2.length < 1 && (
                          <Button
                            type={'text'}
                            onClick={() => setEventFilter2DD(true)}
                            className={'fa-btn--custom m-0'}
                          >
                            <SVG name={'filter'} extraClass={'m-0'} />
                          </Button>
                        )}
                        {event2 && (
                          <Button
                            type={'text'}
                            onClick={() => removeEvent2()}
                            className={'fa-btn--custom m-0'}
                          >
                            <SVG name={'delete'} extraClass={'m-0'} />
                          </Button>
                        )}
                      </div>
                      <EventFilterBy
                        event={event2}
                        setfiltersParent={setfiltersEvent2}
                        showEventFilterDD={showEventFilter2DD}
                        setEventFilterDD={setEventFilter2DD}
                      />
                      {/* {event2 && <EventFilterBy setfiltersParent={setfiltersEvent2} /> } */}
                    </div>
                  </div>
                </Col>
              </Row>
            )}
          </ComposerBlock>

          <ComposerBlock
            blockTitle={'EVENTS TO INCLUDE'}
            isOpen={eventsToIncBlock}
            showIcon={true}
            onClick={() => setEventsToIncBlock(!eventsToIncBlock)}
            extraClass={`no-padding-l`}
          >
            <div>
              <div className={`relative mt-2`}>
                {eventsToInc &&
                  eventsToInc?.map((item, index) => {
                    return (
                      <div key={index} className={`flex items-center mt-2`}>
                        <Button
                          type='text'
                          onClick={() => delOption(item, index)}
                          size={'small'}
                          className={`fa-btn--custom mr-1`}
                        >
                          <SVG name={'remove'} />
                        </Button>
                        <Button
                          className={`flex justify-start`}
                          type='link'
                          onClick={() => setShowEventsToIncDD(true)}
                        >
                          {item}
                        </Button>
                      </div>
                    );
                  })}

                <Button
                  className={` ml-2 mt-4`}
                  type={'text'}
                  onClick={() => setShowEventsToIncDD(true)}
                  icon={
                    <SVG
                      name={'plus'}
                      size={14}
                      color={'grey'}
                      extraClass={'mr-2'}
                    />
                  }
                >
                  {`Add Event`}
                </Button>

                {showEventsToIncDD && (
                  <div
                    style={{
                      top: '-32px',
                      position: 'absolute'
                    }}
                  >
                    <FaSelect
                      options={TrackedEventNames ? TrackedEventNames : []}
                      onClickOutside={() => setShowEventsToIncDD(false)}
                      applClick={(val) => valuesSelect(val)}
                      allowSearch={true}
                      multiSelect={true}
                      selectedOpts={eventsToInc}
                      extraClass={'top-0'}
                    />
                  </div>
                )}
              </div>
            </div>
          </ComposerBlock>
        </div>
        <div
          id={`explain-builder--footer`}
          className={`flex items-center pt-4 border-top--thin-2 justify-between`}
        >
          <div></div>

          {SaveReportModal()}

          <div className={`flex items-center justify-end`}>
            {(queryBuilderCollapse || collapse) && (
              <Button
                className={`mr-2`}
                size={'large'}
                type={'default'}
                onClick={() => setCollapse(!collapse)}
              >
                <SVG
                  name={collapse ? `Expand` : `arrowUp`}
                  size={20}
                  extraClass={`mr-1`}
                ></SVG>
                {`${collapse ? 'Expand' : 'Collapse all'}`}
              </Button>
            )}

            <Button
              type='primary'
              size={'large'}
              loading={insightBtnLoading}
              disabled={!event1}
              // onClick={() => getInsights(props.activeProject.id, eventCount === 2 ? true : false)}
              onClick={() => setIsModalOpen(true)}
            >
              {'Save and Build'}
            </Button>
          </div>
        </div>
      </div>
      {!queryBuilderCollapse && <div style={{ marginBottom: '200px' }} />}
    </div>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    userProperties: state.coreQuery.userProperties,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events,
    factors_model_metadata: state.factors.factors_model_metadata,
    factors_insight_model: state.factors.factors_insight_model,
    userPropNames: state.coreQuery?.userPropNames,
    eventPropNames: state.coreQuery?.eventPropNames
  };
};
export default connect(mapStateToProps, {
  fetchEventNames,
  createExplainJob,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel,
  getUserProperties,
  getEventProperties,
  fetchFactorsModelMetadata
})(CreateGoalDrawer);
