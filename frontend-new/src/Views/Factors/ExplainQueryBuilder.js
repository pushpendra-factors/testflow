import React, { useState, useCallback, useMemo, useEffect } from 'react';
import { Modal, Form, Input, Button, Row, Col, Select, message } from 'antd';
import { SVG, Text } from 'factorsComponents';
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import {
  fetchEventNames,
  getUserPropertiesV2,
  getEventPropertiesV2,
  deleteGroupByForEvent
} from 'Reducers/coreQuery/middleware';
import {
  createExplainJob,
  createExplainJobv3,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel,
  fetchFactorsModelMetadata,
  fetchSavedExplainGoals,
  setActiveExplainQuery
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
import QueryBlock from '../PathAnalysis/PathAnalysisReport/QueryBuilder/QueryBlock';
import { formatBreakdownsForQuery, formatFiltersForQuery, processBreakdownsFromQuery, processFiltersFromQuery } from 'Views/CoreQuery/utils';

const symbolToTextConv = (symbol) => {
  switch (symbol) {
    case '=':
      return 'equals';
    case '!=':
      return 'notEqual';
    default:
      return 'equals';
  }
};

const INCLUDE_EVENTS = "includeEvents"

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

  const {activeQuery, setActiveExplainQuery} = props;

  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  moment.tz.setDefault(timeZone);
 
   
  
  const [insightBtnLoading, setInsightBtnLoading] = useState(false);
  const [collapse, setCollapse] = useState(false);
    

  const [filters, setfilters] = useState([]);
 

  const [eventsBlockOpen, setEventsBlockOpen] = useState(true);
  const [eventsToIncBlock, setEventsToIncBlock] = useState(true);
  const [showDateBlock, setDateBlock] = useState(true); 
   
 
  const [selectedDateRange, setSelectedDateRange] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
 



  const [queries, setSingleQueries] = useState([]);
  const [eventsToInclude, setEventsToInclude] = useState([]);

  const defaultStartDate = selectedDateRange
    ? selectedDateRange?.startDate
    : moment().subtract(1, 'days');
  const defaultEndDate = selectedDateRange
    ? selectedDateRange?.endDate
    : moment().subtract(1, 'days');

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


  const queryType = '';
  const queryOptions = {};


  const singleEventChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setSingleQueries(queryupdated);
    },
    [queries]
  );

  const eventsToIncludeChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...eventsToInclude];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setEventsToInclude(queryupdated);
    },
    [eventsToInclude]
  ); 

const queryList = (type) => {

  let filterConfig = {
        eventLimit: 2,
        extraActions: false //disabling individual level filter/breakdown
      }
  if(type === INCLUDE_EVENTS){
    filterConfig = {
      eventLimit: 10,
      extraActions: false //disabling individual level filter/breakdown
    }
  }

  const blockList = [];
  let queryArr =  type===INCLUDE_EVENTS? eventsToInclude : queries; 
  
  queryArr?.forEach((event, index) => {
    blockList.push(
      <div key={index} 
      className={'flex'}
      >

      <EventTag text={index==0? "A" : "B"} color={index==0? "blue" : "yellow"} />
        
        <QueryBlock
          index={index + 1}
          queryType={queryType}
          event={event}
          queries={queryArr}
          eventChange={type===INCLUDE_EVENTS? eventsToIncludeChange : singleEventChange}
          filterConfig={filterConfig}
        ></QueryBlock>
      </div>
    );
  });

  if (queryArr?.length < filterConfig?.eventLimit) {
    blockList.push(
      <div key={'init'} 
      // className={styles.composer_body__query_block}
      >
        <QueryBlock
          queryType={queryType}
          index={queryArr.length + 1}
          queries={queryArr}
          eventChange={type===INCLUDE_EVENTS? eventsToIncludeChange : singleEventChange}
          groupBy={queryOptions.groupBy}
          filterConfig={filterConfig}
        ></QueryBlock>
      </div>
    );
  }

  return blockList;
}; 

  useEffect(() => {
 
    let goalInsights = activeQuery;
    if (goalInsights) {
      let defaultDate = {
        startDate: moment.unix(goalInsights?.sts),
        endDate: moment.unix(goalInsights?.ets)
      };
      setSelectedDateRange(defaultDate);

      let queryList = []
      if(!_.isEmpty(goalInsights?.query?.st_en) && !_.isEmpty(goalInsights?.query?.st_en?.label)){
        queryList.push(goalInsights?.query?.st_en)
        queryList.push(goalInsights?.query?.en_en)
      }
      else{
        queryList.push(goalInsights?.query?.en_en)
      }

      setEventsToInclude(goalInsights?.query?.in_en)

      let finalQueryList = queryList.map((item)=>{
        if(item?.filter){
          return {
            ...item,
            filters: processFiltersFromQuery(item.filter)
          }
        }
        else return item
      }) 

      setSingleQueries(finalQueryList) 
    }
    return () =>{
      setActiveExplainQuery(null)
    }
  }, [activeQuery]);

  const matchEventName = (item) => {
    let findItem =
      props?.eventPropNames?.[item] || props?.userPropNames?.[item];
    return findItem ? findItem : item;
  };

  const queryBuilderCollapse = !_.isEmpty(props?.goal_insights?.insights);

  const SaveReportModal = () => {
    const [form] = Form.useForm();

    const transformIndividualFilter = (Arr, index) => {
      if (!_.isEmpty(Arr)) {
        let query = {
          ...Arr[index],
          filter: !_.isEmpty(Arr[index]?.filters)
            ? formatFiltersForQuery(Arr[index]?.filters)
            : null
        };
        if(query?.hasOwnProperty('filters')){
          delete query.filters  
        }
        return query;
      } else return null;
    };

    const startAndEndEvents = (eventArr) =>{
      let events = {};
      if (!_.isEmpty(eventArr)){
          if(eventArr.length == 1){
            events = {
              st_en:null,
              en_en:transformIndividualFilter(eventArr,0), 
            }
          }
          else{
            events = { 
              st_en:transformIndividualFilter(eventArr,0),
              en_en:transformIndividualFilter(eventArr,1),
            }
          }

      }
      return events
    }


    const getInsights = (reportName) => {
      // setInsightBtnLoading(true);
      let projectID = props.activeProject.id;
      let gprData = getFilters(filters);  
      let payload = {
        rule: {
          ...startAndEndEvents(queries),
          gpr: gprData,
          in_en: eventsToInclude
        },
        name: reportName,
        sts: moment(defaultStartDate).unix(),
        ets: moment(defaultEndDate).unix()
      }; 

      // creating explain job
      props
        .createExplainJobv3(projectID, payload)
        .then((data) => {
          setInsightBtnLoading(false);
          props.fetchSavedExplainGoals(projectID);
          history.push('/explain');
          message.success('Report saved!');
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
                disabledDateRange={{
                  startDate: moment().subtract(3, 'months'),
                  endDate: moment().subtract(4, 'days')
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
                    <div className={'flex flex-col justify-start'}>
                      {queryList()} 
                    </div> 
                    </div> 
                  </div>
                </div>
              </Col>
            </Row>

            
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

                    {queryList(INCLUDE_EVENTS)}

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
              disabled={queries.length == 0 }
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
    userPropertiesV2: state.coreQuery.userPropertiesV2,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    activeQuery: state.factors.activeQuery,
    tracked_events: state.factors.tracked_events,
    factors_model_metadata: state.factors.factors_model_metadata,
    factors_insight_model: state.factors.factors_insight_model,
    userPropNames: state.coreQuery?.userPropNames,
    eventPropNames: state.coreQuery?.eventPropNames, 
  };
};
export default connect(mapStateToProps, {
  fetchEventNames,
  createExplainJob,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel,
  getUserPropertiesV2,
  getEventPropertiesV2,
  fetchFactorsModelMetadata,
  fetchSavedExplainGoals,
  createExplainJobv3,
  setActiveExplainQuery
})(CreateGoalDrawer);
