import React, { useState, useCallback, useMemo } from 'react';
// import SavedGoals from './savedList';
import { Text, SVG } from 'factorsComponents';
import {
  Button,
  Select,
  message,
  Checkbox,
  Tooltip,
  Modal,
  Form,
  Input,
  Dropdown,
  Menu
} from 'antd';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import QueryBlock from './QueryBlock';
import {
  deleteGroupByForEvent,
  getGroupProperties,
  getUserProperties
} from 'Reducers/coreQuery/middleware';
import {
  fetchSavedPathAnalysis,
  createPathPathAnalysisQuery,
  fetchPathAnalysisInsights
} from 'Reducers/pathAnalysis';
import { useHistory } from 'react-router-dom';
import FaDatepicker from 'Components/FaDatepicker';
import FaSavedRangePicker from 'Components/FaSavedRangePicker';
import { useEffect } from 'react';
import moment from 'moment';
import { getGlobalFilters, getGlobalFiltersfromSavedState, getExpandBy, getExpandByFromState } from './utils';
import _ from 'lodash';
import FaSelect from 'Components/FaSelect';
import { fetchGroups } from 'Reducers/coreQuery/services';
import GlobalFilter from 'Components/GlobalFilter';
import EventFilter from './EventFilter';
import ExpandBy from './ExpandBy';

const QueryBuilder = ({
  queryOptions = {},
  queryType = '',
  fetchSavedPathAnalysis,
  createPathPathAnalysisQuery,
  fetchGroups,
  activeProject,
  activeQuery,
  groupOpts,
  getGroupProperties,
  getUserProperties,
  eventOptions,
  savedQuery,
  fetchPathAnalysisInsights
}) => {
  const [singleQueries, setSingleQueries] = useState([]);
  const [globalFilters, setGlobalFilters] = useState([]);
  const [multipleQueries, setMultipleQueries] = useState([]);
  const [excludeEvents, setExcludeEvents] = useState('true');
  const [pathCondition, setPathCondition] = useState('startswith');
  const [pathStepCount, setPathStepCount] = useState('4');
  const [repetativeStep, setRepetativeStep] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedDateRange, setSelectedDateRange] = useState({});
  const [isGroupDDVisible, setGroupDDVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [groupCategory, setGroupCategory] = useState('users');
  const [collapse, setCollapse] = useState(false);
  const [showCollapseBtn, setCollapseBtn] = useState(false);
  const [considerEventArr, setConsiderEventArr] = useState([]);
  const [savedRanges, setSavedRanges] = useState(null);

  const history = useHistory();

  useEffect(() => {
    fetchGroups(activeProject.id);
  }, [activeProject?.id]);

  // useEffect(() => {
  //   console.log("exclude events triggered", excludeEvents)
  //   if(excludeEvents != activeQuery?.exclude_events){
  //     setConsiderEventArr([]);
  //     setMultipleQueries([]);
  //   }
  // }, [excludeEvents]);

  useEffect(() => {
    if (groupCategory === 'users') {
      getUserProperties(activeProject?.id);
    } else {
      getGroupProperties(activeProject?.id, groupCategory);
    }
  }, [activeProject?.id, groupCategory]);

  const groupsList = useMemo(() => {
    let groups = [['Users', 'users']];
    Object.entries(groupOpts || {}).forEach(([group_name, display_name]) => {
      groups.push([display_name, group_name]);
    });
    return groups;
  }, [groupOpts]);

  const returnEventname = (arr) => {
    return arr.map((item) => item.label);
  };

  const transformIndividualFilter = (Arr) => {
    if (!_.isEmpty(Arr)) {
      let query = {
        ...Arr[0],
        filters: _, // removing filters key
        filter: !_.isEmpty(Arr[0]?.filters)
          ? getGlobalFilters(Arr[0]?.filters)
          : null
      };
      return query;
    } else return null;
  };

  const separateEventsArr = () => {
    let Arr = considerEventArr.filter(item => item.type == "eventOnly")
    let payload = Arr.map(item => {
      return {
        alias: item.value?.split(",")[1],
        label: item.value?.split(",")[0],
        group: item.value?.split(",")[2],
        filter: item?.filter ? getGlobalFilters(item?.filter) : null,
        expand_property: item?.expand_property ? getExpandBy(item?.expand_property) : null
      }
    })
    return payload
  }
  const separateGroupArr = () => {
    let Arr = considerEventArr.filter(item => item.type == "eventType")
    let payload = Arr.map(item => {
      return {
        alias: "",
        label: item.value,
        filter: item?.filter ? getGlobalFilters(item?.filter) : null,
        expand_property: item?.expand_property ? getExpandBy(item?.expand_property) : null
      }
    })
    return payload
  }


  const buildPathAnalysisQuery = (data) => {
    setLoading(true);
    let payload = {
      "query": {
        title: data?.title,
        event_type: pathCondition,
        group: groupCategory,
        event: transformIndividualFilter(singleQueries),
        steps: Number(pathStepCount),
        include_events: excludeEvents == 'false' ? separateEventsArr() : null,
        exclude_events: excludeEvents == 'true' ? multipleQueries : null,
        include_group: separateGroupArr(),
        starttimestamp: moment(selectedDateRange.startDate).unix(),
        endtimestamp: moment(selectedDateRange.endDate).unix(),
        avoid_repeated_events: repetativeStep,
        filter: globalFilters ? getGlobalFilters(globalFilters) : null
      },
      referenceid: activeQuery ? activeQuery?.referenceid : ''
    };
    createPathPathAnalysisQuery(activeProject?.id, payload)
      .then(() => {
        fetchSavedPathAnalysis(activeProject?.id);
        setLoading(false);
        history.push('/path-analysis');
        message.success('Report saved!');
      })
      .catch((err) => {
        setLoading(false);
        console.log('path analysis err->', err);
        message.error(err.data.error);
      });
  };

  useEffect(() => {
    let activeQueryItem = activeQuery?.query;
    if (activeQueryItem) {


      setCollapseBtn(true);
      setCollapse(true);
      setRepetativeStep(activeQueryItem?.avoid_repeated_events);
      setPathCondition(`${activeQueryItem?.event_type}`);
      setPathStepCount(`${activeQueryItem?.steps}`);
      setExcludeEvents(`${activeQueryItem?.include_events ? 'false' : 'true'}`);


      let includeGroupFromState = activeQueryItem?.include_group ? activeQueryItem?.include_group?.map((item) => {
        return {
          ...item,
          value: item?.label,
          type: "eventType",
          filter: item?.filter ? getGlobalFiltersfromSavedState(item?.filter) : null,
          expand_property: item?.expand_property ? getExpandByFromState(item?.expand_property) : null
        }
      }) : []
      let includeEventsFromState = activeQueryItem?.include_events ? activeQueryItem?.include_events?.map((item) => {
        return {
          ...item,
          type: "eventOnly",
          value: `${item?.label},${item?.alias},${item?.group},`,
          filter: item?.filter ? getGlobalFiltersfromSavedState(item?.filter) : null,
          expand_property: item?.expand_property ? getExpandByFromState(item?.expand_property) : null
        }
      }) : []
      let considerEventsFromState = [...includeGroupFromState, ...includeEventsFromState];

      setConsiderEventArr(considerEventsFromState)
      let eventFromState = {
        ...activeQueryItem?.event,
        //adding filters key for filters to work
        filters: activeQueryItem?.event?.filter
          ? getGlobalFiltersfromSavedState(activeQueryItem?.event?.filter)
          : null
      };
      setSingleQueries([eventFromState]);
      setMultipleQueries(
        activeQueryItem?.include_events
          ? activeQueryItem?.include_events
            ? activeQueryItem?.include_events
            : []
          : activeQueryItem?.exclude_events
            ? activeQueryItem?.exclude_events
            : []
      );

      let defaultDate = {
        startDate: moment.unix(activeQueryItem?.starttimestamp),
        endDate: moment.unix(activeQueryItem?.endtimestamp)
      };

      setGlobalFilters(
        activeQueryItem?.filter
          ? getGlobalFiltersfromSavedState(activeQueryItem?.filter)
          : null
      );
      setSelectedDateRange(defaultDate);

      let referenceid = activeQuery?.referenceid;
      let savedReferences = savedQuery[referenceid];
      if (savedReferences?.length > 1) {
        let savedRangesArr = savedReferences?.map((item) => {
          return {
            startDate: item?.query?.starttimestamp, endDate: item?.query?.endtimestamp, status: item?.status, id: item?.id
          }
        })
        setSavedRanges(savedRangesArr)
      }
    }
  }, [activeQuery]);

  const triggerDropDown = () => {
    setGroupDDVisible(true);
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

  const onGroupSelect = (val) => {
    setGroupCategory(val);
    setGroupDDVisible(false);
  };

  const selectGroup = () => {
    return (
      <div className={`${styles.groupsection_dropdown}`}>
        {isGroupDDVisible ? (
          <FaSelect
            extraClass={`${styles.groupsection_dropdown_menu}`}
            options={groupsList}
            onClickOutside={() => setGroupDDVisible(false)}
            optionClick={(val) => onGroupSelect(val[1])}
          ></FaSelect>
        ) : null}
      </div>
    );
  };

  const renderGroupSection = () => {
    try {
      return (
        <div className={`flex items-center pt-6`}>
          <Text
            type={'title'}
            level={6}
            weight={'normal'}
            extraClass={`m-0 mr-3`}
          >
            Analyse
          </Text>
          <div className={`${styles.groupsection}`}>
            <Tooltip title='Attribute at a User, Deal, or Opportunity level'>
              <Button
                className={`${styles.groupsection_button}`}
                type='text'
                onClick={triggerDropDown}
              >
                <div className={`flex items-center`}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={`m-0 mr-1`}
                  >
                    {
                      groupsList?.find(
                        ([_, groupName]) => groupName === groupCategory
                      )?.[0]
                    }
                  </Text>
                  <SVG name='caretDown' />
                </div>
              </Button>
            </Tooltip>
            {selectGroup()}
          </div>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const queryList = (type) => {
    let isSingleEvent = type === 'singleEvent' ? true : false;
    let filterConfig = isSingleEvent
      ? {
        eventLimit: 1,
        extraActions: true
      }
      : {
        eventLimit: 15,
        extraActions: false
      };

    const blockList = [];
    let queryArr = isSingleEvent ? singleQueries : multipleQueries;
    queryArr.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queryArr}
            eventChange={
              isSingleEvent ? singleEventChange : multipleEventChange
            }
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
            eventChange={
              isSingleEvent ? singleEventChange : multipleEventChange
            }
            groupBy={queryOptions.groupBy}
            filterConfig={filterConfig}
          ></QueryBlock>
        </div>
      );
    }

    return blockList;
  };

  const renderPathCondition = () => {
    return (
      <div className={'mt-4'}>
        <Text type={'title'} level={7} weight={'bold'} extraClass={`m-0 mb-2`}>
          SHOW PATHS THAT
        </Text>
        <Select
          style={{
            width: 250
          }}
          className={'fa-select'}
          options={[
            {
              value: 'startswith',
              label: 'Start with an event'
            },
            {
              value: 'endswith',
              label: 'Ends with an event'
            }
          ]}
          defaultValue={pathCondition}
          value={pathCondition}
          onChange={(data) => setPathCondition(data)}
        />
      </div>
    );
  };

  const renderSingleEvent = () => {
    try {
      return (
        <div className={`mt-4`}>
          <Text type={'title'} level={7} extraClass={`m-0`}>
            Select Event
          </Text>
          {queryList('singleEvent')}
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const CustomPath = ({ item, index }) => {

    const [eventType, setEventType] = useState('eventOnly');
    const [eventLevelFilter, setEventLevelFilter] = useState([]);
    const [filterDD, setFilterDD] = useState(false);
    const [expandByDD, setExpandByDD] = useState([false]);
    const [eventLevelExpandBy, setEventLevelExpandBy] = useState([]);


    useEffect(() => {
      //reset filter when event type changes
      setEventLevelFilter([]);
      setEventLevelExpandBy([]);
    }, [eventType]);

    const setValOnChange = (val) => {
      let mainArr = considerEventArr;
      mainArr[index] = {
        ...mainArr[index],
        type: mainArr[index]?.type ? mainArr[index]?.type : 'eventOnly',
        value: val
      }
      setConsiderEventArr(mainArr)
    }
    const setOnTypeChange = (val) => {
      setEventType(val);
      let mainArr = considerEventArr;
      mainArr[index] = {
        ...mainArr[index],
        type: val,
      }
      setConsiderEventArr(mainArr)
    }
    const removeItem = (indexVal) => {
      let newArr = considerEventArr?.filter((item, index) => index != indexVal)
      setConsiderEventArr(newArr);
    }

    const setFilterChange = (val) => {
      setEventLevelFilter(val)
      let mainArr = considerEventArr;
      mainArr[index] = {
        ...mainArr[index],
        filter: val
      }
      setConsiderEventArr(mainArr)
    }

    const setExpandByChange = (val) => {
      setEventLevelExpandBy(val)
      let mainArr = considerEventArr;
      mainArr[index] = {
        ...mainArr[index],
        expand_property: val
      }
      setConsiderEventArr(mainArr)
    }

    const DDMenuItems = (eventType) => {
      return (
        <Menu>
          {eventType != 'eventOnly' &&
            <Menu.Item key='0' disabled={eventLevelFilter?.length >= 1} onClick={() => setFilterDD(true)}>
              <a disabled={eventLevelFilter?.length >= 1}> Add Filter</a>
            </Menu.Item>}
          <Menu.Item key='1' disabled={eventLevelExpandBy?.length >= 1} onClick={() => setExpandByDD([true])}>
            <a disabled={eventLevelExpandBy?.length >= 1}>Expand by property</a>
          </Menu.Item>
        </Menu>
      );
    };

    const DDEventTypes = [
      {
        value: 'Page Views',
        label: 'Page Views',
      },
      {
        value: 'CRM Events',
        label: 'CRM Events',
      },
      {
        value: 'Button Clicks',
        label: 'Button Clicks',
      },
      {
        value: 'Sessions',
        label: 'Sessions',
      },
    ]


    let DDValues = eventOptions.map((item) => {
      let optionsVal = item?.values.map((val) => {
        return { label: val[0], value: `${val[1]},${val[0]},${item?.label}` }
      })
      return {
        label: item?.label,
        options: optionsVal
      }
    });

    let typeCheck = considerEventArr[index]?.type ? considerEventArr[index]?.type : 'eventOnly';
    return (
      <div className='flex flex-col'>
        <div className='flex items-center mt-2'>
          <div className=''>
            <Select
              style={{ minWidth: 185 }}
              className={'fa-select'}
              defaultActiveFirstOption={true}
              defaultValue={typeCheck}
              onChange={(val) => setOnTypeChange(val)}
              options={[
                { value: 'eventOnly', label: 'If the event equals' },
                { value: 'eventType', label: 'If the event is of type' },
              ]}
            />
          </div>
          <div className='ml-2'>
            {typeCheck == 'eventOnly' ?
              <Select
                showSearch
                style={{ minWidth: 285 }}
                className={'fa-select'}
                options={DDValues}
                placeholder={'Select event'}
                allowClear={true}
                // defaultActiveFirstOption={true}
                onChange={(val) => setValOnChange(val)}
                defaultValue={considerEventArr[index]?.value?.split(",")[1]}
              /> :
              <Select
                style={{ minWidth: 245 }}
                options={DDEventTypes}
                className={'fa-select'}
                placeholder={'Select type'}
                // defaultActiveFirstOption={true}
                onChange={(val) => setValOnChange(val)}
                defaultValue={considerEventArr[index]?.value}
              />
            }
          </div>
          <div className='ml-2'>
            <Dropdown overlay={() => DDMenuItems(eventType)} trigger={['click']} >
              <Button type={'text'} icon={<SVG name={'Threedot'} size={16} color='gray' />} />
            </Dropdown>

            <Button type={'text'} icon={<SVG
              name={'Trash'}
              size={16}
              color='gray'
            />}
              onClick={() => removeItem(index)} />
          </div>



        </div>
        <EventFilter
          filters={item?.filter ? item?.filter : eventLevelFilter}
          setGlobalFilters={(val) => setFilterChange(val)}
          groupName={groupCategory}
          filterDD={filterDD}
          setFilterDD={setFilterDD}
        />

        <ExpandBy isDDVisible={expandByDD} setDDVisible={setExpandByDD}
          eventLevelExpandBy={item?.expand_property ? item?.expand_property : eventLevelExpandBy} setEventLevelExpandBy={setExpandByChange}
        />
      </div>
    )
  }


  const addEventFn = () => {
    try {



      return (
        <div className={`mt-4`}>
          {considerEventArr?.map((item, index) => {
            return (
              <CustomPath item={item} index={index} />
            )
          })}

          <Button type={'text'} className={considerEventArr?.length > 0 ? 'mt-4' : ''}
            onClick={() => setConsiderEventArr([...considerEventArr, {}])
            }
            icon={<SVG name='plus' />}
          >Add New</Button>

        </div>
      );
    } catch (err) {
      console.log("error caught-->>>", err);
    }
  };

  const renderMultiEventList = () => {
    try {
      return (
        <>
          <div className={`mt-4`}>
            <div className={'flex items-center'}>
              <Text type={'title'} level={7} extraClass={`m-0`}>
                In this path, Show
              </Text>

              <Select
                bordered={false}
                className={'fa-select-ghost--highlight-text'}
                options={[
                  {
                    value: 'true',
                    label: 'All events except'
                  },
                  {
                    value: 'false',
                    label: 'Only specific events'
                  }
                ]}
                value={excludeEvents}
                onChange={(data) => setExcludeEvents(data)}
              />
            </div>

            {excludeEvents == 'true' ? queryList() : addEventFn()}
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
        <div
          className={`mt-6 pt-5 pb-5 border-top--thin-2 border-bottom--thin-2`}
        >
          <Text
            type={'title'}
            level={7}
            weight={'bold'}
            extraClass={`m-0 mb-2`}
          >
            FILTER PATH BY
          </Text>
          <GlobalFilter
            filters={globalFilters}
            setGlobalFilters={setGlobalFilters}
            groupName={groupCategory}
          />
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const pathStepSelection = () => {
    try {
      return (
        <>
          <div className={`mt-6`}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              extraClass={`m-0 mb-2`}
            >
              IN THIS PATH
            </Text>
            <div className={'flex items-center'}>
              <Text type={'title'} level={7} extraClass={`m-0`}>
                Show
              </Text>
              <Select
                bordered={false}
                className={'fa-select-ghost--highlight-text'}
                options={[
                  {
                    value: '1',
                    label: '1 Step'
                  },
                  {
                    value: '2',
                    label: '2 Steps'
                  },
                  {
                    value: '3',
                    label: '3 Steps'
                  },
                  {
                    value: '4',
                    label: '4 Steps'
                  },
                  {
                    value: '5',
                    label: '5 Steps'
                  }
                ]}
                value={pathStepCount}
                onChange={(data) => setPathStepCount(data)}
              />
            </div>
            <div className={'mt-2'}>
              <Checkbox
                checked={repetativeStep}
                onChange={(e) => setRepetativeStep(e.target.checked)}
              >
                Avoid repeated events
              </Checkbox>
              <Text type={'title'} level={8} extraClass={`m-0 ml-6`}>
                Restrict events to appear only once in this path
              </Text>
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
                loading={loading}
                className={'ml-2'}
              >
                Save
              </Button>
            </Form.Item>
          </div>
        </Form>
      </Modal>
    );
  };

  const setDateRange = (data) => {
    setSelectedDateRange(data);
  };
  const footer = () => {
    try {
      // if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
      //   return null;
      // }
      // if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
      //   return null;
      // } else {

      const smoothScroll = (element) => {
        document.querySelector(element).scrollIntoView({
          behavior: 'smooth'
        });
      };

      const onSelectSavedRangeFn = (data) => {
        setLoading(true);
        fetchPathAnalysisInsights(activeProject?.id, data?.id).then(() => {
          smoothScroll('#fa-report-container');
          setLoading(false);
        }).catch((err) => {
          setLoading(false);
          console.log('path analysis saved range err->', err);
          message.error(err.data.error);
        });
      }

      return (
        <div
          // className={ !collapse ? styles.composer_footer : styles.composer_footer_right }
          className='flex justify-between w-100 mt-6 pt-6 mb-6 border-top--thin-2'
        >

          <FaSavedRangePicker
            todayPicker={false}
            customPicker
            presetRange
            monthPicker
            quarterPicker
            placement='topLeft'
            buttonSize={'large'}
            range={{
              startDate: selectedDateRange.startDate,
              endDate: selectedDateRange.endDate
            }}
            savedRanges={savedRanges}
            onSelect={setDateRange}
            onSelectSavedRange={onSelectSavedRangeFn}
          />

          {/* <FaDatepicker
            todayPicker={false}
            customPicker
            presetRange
            monthPicker
            quarterPicker
            placement='topRight'
            buttonSize={'large'}
            range={{
              startDate: selectedDateRange.startDate,
              endDate: selectedDateRange.endDate
            }}
            onSelect={setDateRange}
          /> */}

          <div className='flex justify-end items-center'>
            {showCollapseBtn && (
              <Button
                className={`mr-2`}
                size={'large'}
                type={'default'}
                onClick={() => setCollapse(true)}
              >
                <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>
                Collapse all
              </Button>
            )}
            <Button
              type='primary'
              size='large'
              disabled={_.isEmpty(singleQueries)}
              loading={loading}
              onClick={() => setIsModalOpen(true)}
            >
              {`Save and Build`}
            </Button>
          </div>
        </div>
      );
      // }
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <>
      <div className={'relative px-20'}>
        {
          <div
            className={`query_card_cont mb-10 ${!collapse ? `query_card_open` : `query_card_close`
              }`}
          >
            {renderGroupSection()}
            {renderPathCondition()}
            {renderSingleEvent('singleEvent')}
            {/* Single event selection with filters */}
            {renderMultiEventList()}
            {/* Multi event selection without filters */}
            {renderGlobalFilterBlock()}
            {pathStepSelection()}
            {SaveReportModal()}
            {footer()}
            <Button
              size='large'
              onClick={(e) => setCollapse(false)}
              className='query_card_expand'
            >
              <SVG name='expand' size={20} />
              Expand
            </Button>
          </div>
        }
      </div>
    </>
  );
};

const mapStateToProps = (state) => {
  return {
    savedQuery: state.pathAnalysis?.savedQuery,
    activeQuery: state.pathAnalysis?.activeQuery,
    activeProject: state.global.active_project,
    eventOptions: state.coreQuery.eventOptions,
    groupOpts: state.groups.data
  };
};

export default connect(mapStateToProps, {
  fetchSavedPathAnalysis,
  createPathPathAnalysisQuery,
  fetchGroups,
  getGroupProperties,
  getUserProperties,
  fetchPathAnalysisInsights
})(QueryBuilder);
