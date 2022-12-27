import React, { useState, useContext, useCallback } from 'react';
// import SavedGoals from './savedList';
import { Text, SVG, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { Button, Select, message, Checkbox, Tooltip, Modal, Form, Input } from 'antd';
import styles from './index.module.scss';
import { connect, useSelector } from 'react-redux';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import QueryBlock from './QueryBlock';
import { deleteGroupByForEvent, getUserProperties } from 'Reducers/coreQuery/middleware';
import { fetchSavedPathAnalysis, createPathPathAnalysisQuery } from 'Reducers/pathAnalysis';
import { useHistory } from 'react-router-dom';
import FaDatepicker from 'Components/FaDatepicker';
import { useEffect } from 'react';
import moment from 'moment';
import GLobalFilter from './GlobalFilter';
import { getGlobalFilters, getGlobalFiltersfromSavedState } from './utils'
import _ from 'lodash';

const QueryBuilder = ({
  queryOptions = {},
  queryType = "", 
  fetchSavedPathAnalysis,
  createPathPathAnalysisQuery, 
  activeProject,
  collapse,
  setCollapse,
  activeQuery,
}) => { 
  const [singleQueries, setSingleQueries] = useState([]);
  const [globalFilters, setGlobalFilters] = useState([]);
  const [multipleQueries, setMultipleQueries] = useState([]);
  const groupState = useSelector((state) => state.groups);
  const [excludeEvents, setExcludeEvents] = useState('true');
  const [pathCondition, setPathCondition] = useState('startswith');
  const [pathStepCount, setPathStepCount] = useState('4');
  const [repetativeStep, setRepetativeStep] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedDateRange, setSelectedDateRange] = useState({});

  const [loading, setLoading] = useState(false);
  const groupOpts = groupState?.data;

  const history = useHistory();

  const returnEventname = (arr) => {
    return arr.map((item) => item.label)
  } 

  const transformIndividualFilter = (Arr) =>{ 
    if(!_.isEmpty(Arr)){
      let query  = { 
        ...Arr[0],
        filters: _, // removing filters key
        filter: !_.isEmpty(Arr[0]?.filters) ? getGlobalFilters(Arr[0]?.filters) : null
      }
      return query 
    }
    else return null
  }
 
  const buildPathAnalysisQuery = (data) => {

    setLoading(true);
    let payload = {
      "title": data?.title,
      "event_type": pathCondition,
      "event": transformIndividualFilter(singleQueries),
      "steps": Number(pathStepCount),
      "include_events": (excludeEvents == 'false') ? multipleQueries : null,
      "exclude_events": (excludeEvents == 'true') ? multipleQueries : null,
      "starttimestamp": moment(selectedDateRange.startDate).unix(),
      "endtimestamp": moment(selectedDateRange.endDate).unix(),
      "avoid_repeated_events": repetativeStep,
      "filter": globalFilters ? getGlobalFilters(globalFilters) : null
    }; 
 
    createPathPathAnalysisQuery(activeProject?.id, payload).then(() => {
      fetchSavedPathAnalysis(activeProject?.id);
      setLoading(false)
      history.push('/path-analysis');
      message.success('Report saved!');
    }).catch((err) => {
      setLoading(false)
      console.log('path analysis err->', err);
      message.error(err.data.error);
    });

  }


  useEffect(() => {
    if (activeQuery) {
      console.log("activeQuery useEffect-->>", activeQuery)
      setRepetativeStep(activeQuery?.avoid_repeated_events)
      setPathCondition(`${activeQuery?.event_type}`)
      setPathStepCount(`${activeQuery?.steps}`)
      setExcludeEvents(`${activeQuery?.include_events ? 'true' : 'false'}`)

      setSingleQueries([activeQuery?.event]);
      setMultipleQueries(activeQuery?.include_events ? (activeQuery?.include_events ? activeQuery?.include_events : []) : (activeQuery?.exclude_events ? activeQuery?.exclude_events : []));

      let defaultDate = {
        startDate: moment.unix(activeQuery?.starttimestamp),
        endDate: moment.unix(activeQuery?.endtimestamp),
      } 

      setGlobalFilters(activeQuery?.filter ? getGlobalFiltersfromSavedState(activeQuery?.filter) : null)
      setSelectedDateRange(defaultDate)
    }
  }, [activeQuery])

  const enabledGroups = () => {
    let groups = [['Users', 'users']];
    groupOpts?.forEach((elem) => {
      const formatName = PropTextFormat(elem.name);
      groups.push([formatName, elem.name]);
    });
    return groups;
  };

  const singleEventChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...singleQueries];
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
    [singleQueries]
  );
  const multipleEventChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...multipleQueries];
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
      setMultipleQueries(queryupdated);
    },
    [multipleQueries]
  );



  const renderGroupSection = () => {
    try {
      return (
        <div className={'flex items-center mt-4'}>
          <Text type={'title'} level={7} weight={'thin'} extraClass={`m-0`} >Analyse</Text>
          <Select
            bordered={false}
            disabled={true}
            value={'users'}
            className={'fa-select-ghost--bold'}
            options={[
              {
                value: 'users',
                label: 'Users',
              }
            ]}
          />
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const queryList = (type) => {

    let isSingleEvent = (type == 'singleEvent') ? true : false
    let filterConfig = (isSingleEvent) ?
      {
        eventLimit: 1,
        extraActions: true
      }
      :
      {
        eventLimit: 5,
        extraActions: false
      }

    const blockList = [];
    let queryArr = isSingleEvent ? singleQueries : multipleQueries
    queryArr.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queryArr}
            eventChange={isSingleEvent ? singleEventChange : multipleEventChange}
            filterConfig={filterConfig}
          ></QueryBlock>
        </div>
      );
    });

    if (queryArr.length < filterConfig?.eventLimit) {
      blockList.push(
        <div key={'init'} className={styles.composer_body__query_block}>
          <QueryBlock
            queryType={queryType}
            index={queryArr.length + 1}
            queries={queryArr}
            eventChange={isSingleEvent ? singleEventChange : multipleEventChange}
            groupBy={queryOptions.groupBy}
            filterConfig={filterConfig}
          ></QueryBlock>
        </div>
      );
    }


    return blockList;
  };


  const renderPathCondition = () => {
    return (<div className={'mt-4'}>
      <Text type={'title'} level={7} weight={'bold'} extraClass={`m-0 mb-2`} >SHOW PATHS THAT</Text>
      <Select
        style={{
          width: 250,
        }}
        className={'fa-select'}
        options={[
          {
            value: 'startswith',
            label: 'Start with an event',
          },
          {
            value: 'endswith',
            label: 'Ends with an event',
          },
        ]}
        defaultValue={pathCondition}
        value={pathCondition}
        onChange={(data) => setPathCondition(data)}
      />
    </div>
    )
  }


  const renderSingleEvent = () => {
    try {
      return (
        <div className={`mt-4`} >
          <Text type={'title'} level={7} extraClass={`m-0`} >Select Event</Text>
          {queryList('singleEvent')}
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };


  const renderMultiEventList = () => {
    try {
      return (<>
        <div className={`mt-4`}>
          <div className={'flex items-center'}>
            <Text type={'title'} level={7} extraClass={`m-0`} >In this path, Show</Text>

            <Select
              bordered={false}
              className={'fa-select-ghost--highlight-text'}
              options={[
                {
                  value: 'true',
                  label: 'All events except',
                },
                {
                  value: 'false',
                  label: 'Only specific events',
                },
              ]}
              value={excludeEvents}
              onChange={(data) => setExcludeEvents(data)}
            />

          </div>

          {queryList()}
        </div>
      </>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const renderGlobalFilterBlock = () => {
    try {
      return (
        <div className={`mt-6 pt-5 pb-5 border-top--thin-2 border-bottom--thin-2`}>
          <Text type={'title'} level={7} weight={'bold'} extraClass={`m-0 mb-2`} >FILTER PATH BY</Text>
          <GLobalFilter
            filters={globalFilters}
            setGlobalFilters={setGlobalFilters}
            onFiltersLoad={[
              () => {
                getUserProperties(activeProject?.id);
              },
            ]}
          ></GLobalFilter>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };


  const pathStepSelection = () => {
    try {
      return (<>
        <div className={`mt-6`}>
          <Text type={'title'} level={7} weight={'bold'} extraClass={`m-0 mb-2`} >IN THIS PATH</Text>
          <div className={'flex items-center'}>
            <Text type={'title'} level={7} extraClass={`m-0`} >Show</Text>
            <Select
              bordered={false}
              className={'fa-select-ghost--highlight-text'}
              options={[
                {
                  value: '1',
                  label: '1 Step',
                },
                {
                  value: '2',
                  label: '2 Steps',
                },
                {
                  value: '3',
                  label: '3 Steps',
                },
                {
                  value: '4',
                  label: '4 Steps',
                },
                {
                  value: '5',
                  label: '5 Steps',
                },
              ]}
              value={pathStepCount}
              onChange={(data) => setPathStepCount(data)}
            />
          </div>
          <div className={'mt-2'}>
            <Checkbox defaultChecked={repetativeStep} onChange={(e) => setRepetativeStep(e.target.checked)}>Avoid repeated events</Checkbox>
            <Text type={'title'} level={8} extraClass={`m-0 ml-6`} >Restrict events to appear only once in this path</Text>
          </div>
        </div>
      </>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const SaveReportModal = () => {
    const [form] = Form.useForm();
    const onFinish = (values) => {
      buildPathAnalysisQuery(values).then(() => {
        setIsModalOpen(false)
      });
    };
    const onFinishFailed = (errorInfo) => {
      console.log('Failed:', errorInfo);
    };
    const onReset = () => { 
      setIsModalOpen(false)
      form.resetFields();
    };
    return (
      <Modal title="Save report"
        visible={isModalOpen}
        okText={'Save'}
        footer={null} 
        className='fa-modal--regular'
        afterClose={onReset}
        onCancel={onReset}
      >

        <Form
          name="basic"
          initialValues={{
            remember: true,
          }}
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          autoComplete="off"
          form={form}
        >
          <Form.Item

            name="title"
            rules={[
              {
                required: true,
                message: 'Please enter a name (or) title for the report',
              },
            ]}
          >
            <Input placeholder={'Name'} className={'fa-input'} size={'large'} />
          </Form.Item>

          <div className='mt-2'>
            <Form.Item name="description">
              <Input.TextArea className={'fa-input'} size={'large'} placeholder={'Description'} />
            </Form.Item>
          </div>

          <div className='mt-8 flex justify-end'> 

            <Form.Item>
            <Button size={'large'} htmlType="button" onClick={onReset}>
                                    Cancel
                                    </Button>
              <Button size={'large'}  type="primary" htmlType="submit" loading={loading} className={'ml-2'}> Save </Button>
            </Form.Item> 
          </div>

        </Form>
      </Modal>
    )
  }

  const setDateRange = (data) => {
    setSelectedDateRange(data)
  }
  const footer = () => {
    try {
      // if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
      //   return null;
      // }
      // if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
      //   return null;
      // } else {
       
      return (
        <div
          // className={ !collapse ? styles.composer_footer : styles.composer_footer_right }
          className='flex justify-between w-100 mt-6 pt-6 mb-6 border-top--thin-2'
        >
          {!collapse ? (
            <FaDatepicker
              customPicker
              presetRange
              monthPicker
              quarterPicker
              placement='topRight'
              buttonSize={'large'}
              range={{
                startDate: selectedDateRange.startDate,
                endDate: selectedDateRange.endDate,
              }}
              onSelect={setDateRange}
            />
          ) : (
            <Button
              className={`mr-2`}
              size={'large'}
              type={'default'}
              onClick={() => setCollapse(false)}
            >
              <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>
              Collapse all
            </Button>
          )}

          <Button type='primary' size='large' disabled={_.isEmpty(singleQueries)} loading={loading} onClick={() => setIsModalOpen(true)}> {`Save and Build`}</Button>

        </div>
      );
      // }
    } catch (err) {
      console.log(err);
    }
  }; 

  return <>
    <div className={'relative'}>


      {
        <div
          className={`query_card_cont mb-10 ${!collapse ? `query_card_open` : `query_card_close`
            }`}
          onClick={(e) => setCollapse(false)}
        >
          {renderGroupSection()}

          {renderPathCondition()}

          {renderSingleEvent('singleEvent')} {/* Single event selection with filters */}

          {renderMultiEventList()} {/* Multi event selection without filters */}

          {renderGlobalFilterBlock()}

          {pathStepSelection()}

          {SaveReportModal()}


          {footer()}


          <Button size="large" onClick={(e) => setCollapse(false)} className="query_card_expand">
            <SVG name="expand" size={20} />
            Expand
          </Button>
        </div>}





    </div>
  </>
}

const mapStateToProps = (state) => {
  return {
    // savedQuery: state.pathAnalysis.savedQuery, 
    activeProject: state.global.active_project,
  };
};


export default connect(mapStateToProps, { fetchSavedPathAnalysis, createPathPathAnalysisQuery })(QueryBuilder);
