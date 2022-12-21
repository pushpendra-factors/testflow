import React, { useState, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import {
  Row,
  Col,
  Select,
  Menu,
  Dropdown,
  Button,
  Form,
  Table,
  Input,
  notification
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined } from '@ant-design/icons';
import {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI
} from 'Reducers/kpi';
import {
  getUserProperties,
  deleteGroupByForEvent,
  fetchEventNames,
  getEventProperties
} from 'Reducers/coreQuery/middleware';
import _ from 'lodash';
import GLobalFilter from './GLobalFilter';
import { formatFilterDate } from '../../../../utils/dataFormatter';
import styles from './index.module.scss';
import {
  reverseOperatorMap,
  reverseDateOperatorMap,
  convertDateTimeObjectValuesToMilliSeconds,
  getKPIQuery,
  DefaultDateRangeFormat
} from './utils';
import { FILTER_TYPES } from '../../../CoreQuery/constants';
import QueryBlock from './QueryBlock';
import {
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE
} from '../../../../utils/constants';
import EventFilter from './EventFilter/GlobalFilter';

const { Option } = Select;

function CustomKPI({
  activeProject,
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  customKPIConfig,
  savedCustomKPI,
  addNewCustomKPI,
  eventPropNames,
  userPropNames,
  removeCustomKPI,
  currentAgent,
  fetchKPIConfigWithoutDerivedKPI,
  fetchEventNames,
  eventNames,
  eventNameOptions,
  getEventProperties,
  eventProperties
}) {
  const [showForm, setShowForm] = useState(false);
  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [selKPICategory, setKPICategory] = useState(false);
  const [selKPIType, setKPIType] = useState('default');
  const [KPIPropertyDetails, setKPIPropertyDetails] = useState({});
  const [filterDDValues, setFilterDDValues] = useState();
  const [filterValues, setFilterValues] = useState([]);
  const [KPIFn, setKPIFn] = useState(false);
  const [viewMode, KPIviewMode] = useState(false);
  const [viewKPIDetails, setKPIDetails] = useState(false);

  const [selEventName, setEventName] = useState(false);
  const [EventPropertyDetails, setEventPropertyDetails] = useState({});
  const [EventfilterDDValues, setEventFilterDDValues] = useState();
  const [EventfilterValues, setEventFilterValues] = useState([]);
  const [EventFn, setEventFn] = useState(false);

  const [form] = Form.useForm();

  // const [queryOptions, setQueryOptions] = useState({});

  // KPI SELECTION
  const [queryType, setQueryType] = useState('kpi');
  const [queries, setQueries] = useState([]);
  const [selectedMainCategory, setSelectedMainCategory] = useState(false);
  const [KPIConfigProps, setKPIConfigProps] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });

  const { groupBy } = useSelector((state) => state.coreQuery);

  const matchEventName = (item) => {
    const findItem = eventPropNames?.[item] || userPropNames?.[item];
    return findItem || item;
  };

  const menu = (item) => (
    <Menu>
      <Menu.Item
        key='0'
        onClick={() => {
          KPIviewMode(true);
          setKPIDetails(item);
        }}
      >
        <a>View</a>
      </Menu.Item>
      <Menu.Item
        key='1'
        onClick={() => {
          deleteKPI(item);
        }}
      >
        <a>Remove</a>
      </Menu.Item>
    </Menu>
  );

  const alphabetIndex = 'ABCDEFGHIJK';

  const columns = [
    {
      title: 'KPI Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Text type='title' level={7} truncate charLimit={25}>
          {text}
        </Text>
      )
      // width: 100,
    },
    {
      title: 'Description',
      dataIndex: 'desc',
      key: 'desc',
      render: (text) => (
        <Text type='title' level={7} truncate charLimit={25}>
          {text}
        </Text>
      )
      // width: 200,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (item) => (
        <Text type='title' level={7} truncate charLimit={35}>
          {item}
        </Text>
      ),
      width: 'auto'
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      width: 75,
      render: (obj) => (
        <Dropdown overlay={() => menu(obj)} trigger={['click']}>
          <Button
            type='text'
            icon={
              <MoreOutlined
                rotate={90}
                style={{ color: 'gray', fontSize: '18px' }}
              />
            }
          />
        </Dropdown>
      )
    }
  ];
  const onChange = () => {
    seterrorInfo(null);
  };

  const setGlobalFiltersOption = (filters) => {
    const opts = { ...queryOptions };
    opts.globalFilters = filters;
    setFilterValues(opts);
  };

  const setEventGlobalFiltersOption = (filters) => {
    const opts = { ...queryOptions };
    opts.globalFilters = filters;
    setEventFilterValues(opts);
  };

  const operatorMap = {
    '=': 'equals',
    '!=': 'notEqual',
    contains: 'contains',
    'does not contain': 'notContains',
    '<': 'lesserThan',
    '<=': 'lesserThanOrEqual',
    '>': 'greaterThan',
    '>=': 'greaterThanOrEqual',
    between: 'between',
    'not between': 'notInBetween',
    'in the previous': 'inLast',
    'not in the previous': 'notInLast',
    'in the current': 'inCurrent',
    'not in the current': 'notInCurrent',
    before: 'before',
    since: 'since'
  };

  const getEventsWithPropertiesKPI = (filters, category = null) => {
    const filterProps = [];
    // adding fil?.extra ? fil?.extra[*] check as a hotfix for timestamp filters
    filters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : 'event',
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : 'event'
              : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values
        });
      }
    });
    return filterProps;
  };

  const queryChange = (newEvent, index, changeType = 'add', flag = null) => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
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
    setQueries(queryupdated);
  };

  useEffect(() => {
    setSelectedMainCategory(queries[0]);
  }, [queries]);

  const handleEventChange = (...props) => {
    queryChange(...props);
  };

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={handleEventChange}
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
        <div key='init' className={styles.composer_body__query_block}>
          <QueryBlock
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={handleEventChange}
            groupBy={queryOptions.groupBy}
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    }

    return blockList;
  };

  const onReset = () => {
    form.resetFields();
    setShowForm(false);
    setFilterValues([]);
    setKPICategory(false);
    setKPIType('default');
    setQueries([]);
    setKPIPropertyDetails({});
    setKPIFn(false);
    setEventPropertyDetails({});
    setEventFn(false);
    setEventFilterValues([]);
    setEventName(false);
  };

  const onFinish = (data) => {
    let payload;
    if (selKPIType === 'default') {
      payload = {
        name: data?.name,
        description: data?.description,
        type_of_query: 1,
        obj_ty: data.kpi_category,
        transformations: {
          agFn: data.kpi_function,
          agPr: KPIPropertyDetails?.name,
          agPrTy: KPIPropertyDetails?.data_type,
          fil: filterValues?.globalFilters
            ? getEventsWithPropertiesKPI(filterValues?.globalFilters)
            : [],
          daFie: data.kpi_dateField
        }
      };
    } else if (selKPIType === 'derived_kpi') {
      const KPIquery = getKPIQuery(
        queries,
        queryOptions.date_range,
        groupBy,
        queryOptions,
        data?.for
      );

      payload = {
        name: data?.name,
        description: data?.description,
        type_of_query: 2,
        transformations: {
          ...KPIquery
        }
      };
    } else {
      payload = {
        name: data?.name,
        description: data?.description,
        type_of_query: 3,
        obj_ty: 'event_based',
        transformations: {
          agFn: data?.event_function,
          agPr: EventPropertyDetails?.name,
          agPrTy: EventPropertyDetails?.data_type,
          fil: EventfilterValues?.globalFilters
            ? getEventsWithPropertiesKPI(EventfilterValues?.globalFilters)
            : [],
          daFie: '',
          evNm: data?.event,
          en: 'events_occurrence'
        }
      };
    }

    setLoading(true);
    addNewCustomKPI(activeProject.id, payload)
      .then(() => {
        setLoading(false);
        fetchSavedCustomKPI(activeProject.id);
        notification.success({
          message: 'KPI Saved',
          description:
            'New KPI is created and saved successfully. You can start using it across the product shortly.'
        });
        onReset();
      })
      .catch((err) => {
        setLoading(false);
        notification.error({
          message: 'Error',
          description: err?.data?.error
        });
        console.log('addNewCustomKPI error->', err);
      });
  };

  const deleteKPI = (item) => {
    removeCustomKPI(activeProject.id, item?.id)
      .then(() => {
        fetchSavedCustomKPI(activeProject.id);
        notification.success({
          message: 'KPI Removed',
          description: 'Custom KPI is removed successfully.'
        });
      })
      .catch((err) => {
        notification.error({
          message: 'Error',
          description: err?.data?.error
        });
        console.log('addNewCustomKPI error->', err);
      });
  };

  useEffect(() => {
    // if (!customKPIConfig) {
    fetchCustomKPIConfig(activeProject.id);
    // }
    // if (!savedCustomKPI) {
    fetchSavedCustomKPI(activeProject.id);
    // }
    fetchKPIConfigWithoutDerivedKPI(activeProject.id);
    fetchEventNames(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    const DDCategory = customKPIConfig?.result?.find((category) => {
      if (category.obj_ty === selKPICategory) {
        return category;
      }
    });

    const DDvalues = DDCategory?.properties?.map((item) => [
      item.display_name,
      item.name,
      item.data_type,
      item.entity
    ]);

    setFilterDDValues(DDvalues);
  }, [selKPICategory, customKPIConfig]);

  useEffect(() => {
    let DDCategory;
    for (const key of Object.keys(eventProperties)) {
      if (key === selEventName) {
        DDCategory = eventProperties[key];
      }
    }

    const DDvalues = DDCategory?.map((item) => [
      item[0],
      item[1],
      item[2],
      'event'
    ]);

    setEventFilterDDValues(DDvalues);
  }, [selEventName, eventProperties]);

  useEffect(() => {
    if (selEventName || viewKPIDetails?.transformations?.evNm) {
      getEventProperties(
        activeProject.id,
        selEventName || viewKPIDetails?.transformations?.evNm
      );
    }
  }, [selEventName, viewKPIDetails?.transformations?.evNm]);

  const onKPICategoryChange = (value) => {
    setKPICategory(value);
  };

  const onEventNameChange = (value) => {
    setEventName(value);
  };

  const onKPITypeChange = (value) => {
    setKPIType(value);
  };

  useEffect(() => {
    if (savedCustomKPI) {
      const savedArr = [];
      savedCustomKPI?.map((item, index) => {
        savedArr.push({
          key: index,
          name: item.name,
          desc: item.description,
          type:
            item.type_of_query === 1
              ? 'Default'
              : item.type_of_query === 2
              ? 'Derived'
              : 'Event Based',
          actions: item
        });
      });
      setTableData(savedArr);
    }
  }, [savedCustomKPI]);

  const getStateFromFilters = (rawFilters) => {
    const filters = [];

    rawFilters.forEach((pr) => {
      if (pr.lOp === 'AND') {
        const val = pr.prDaTy === FILTER_TYPES.CATEGORICAL ? [pr.va] : pr.va;

        const DNa = matchEventName(pr.prNa);

        filters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, 'filter'],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy]
        });
      } else if (pr.prDaTy === FILTER_TYPES.CATEGORICAL) {
        filters[filters.length - 1].values.push(pr.va);
      }
    });
    return filters;
  };

  const excludeEventsFromList = [
    'Contact Created',
    'Contact Updated',
    'Lead Created',
    'Lead Updated',
    'Account Created',
    'Account Updated',
    'Company Created',
    'Company Updated',
    'Deal Created',
    'Deal Updated',
    'Opportunity Created',
    'Opportunity Updated'
  ];

  const whiteListedAccounts = [
    'junaid@factors.ai',
    'solutions@factors.ai',
    'parveenr@factors.ai',
    'sonali@factors.ai'
  ];

  function renderEventBasedKPIForm() {
    return (
      <div>
        <Row className='mt-6'>
          <Col span={24}>
            <div className='border-top--thin-2 pt-3 mt-3' />
          </Col>
        </Row>
        <Row className='m-0'>
          <Col span={18}>
            {/* <div className={'border-top--thin-2 pt-3 mt-3'} /> */}
            <Text type='title' level={7} extraClass='m-0'>
              Select Event
            </Text>
            <Form.Item
              name='event'
              className='m-0'
              rules={[
                {
                  required: true,
                  message: 'Please select Event'
                }
              ]}
            >
              <Select
                className='fa-select w-full'
                size='large'
                onChange={(value) => onEventNameChange(value)}
                placeholder='Select Event'
                showSearch
                filterOption={(input, option) =>
                  option.children.toLowerCase().indexOf(input.toLowerCase()) >=
                  0
                }
              >
                {Object.entries(eventNames)
                  .filter((entry) => {
                    const key = entry[0];
                    const value = entry[1];

                    // Check if the key or value matches one of the values to be removed
                    return (
                      !excludeEventsFromList.includes(key) &&
                      !excludeEventsFromList.includes(value)
                    );
                  })
                  .map((item) => (
                    <Option key={item[0]} value={item[0]}>
                      {item[1]}
                    </Option>
                  ))}
              </Select>
            </Form.Item>
          </Col>
        </Row>

        {selEventName && (
          <Row className='mt-8'>
            <Col span={18}>
              <Text type='title' level={7} extraClass='m-0'>
                Function
              </Text>
              <Form.Item
                name='event_function'
                className='m-0'
                rules={[
                  {
                    required: true,
                    message: 'Please select a Function'
                  }
                ]}
              >
                <Select
                  className='fa-select w-full'
                  size='large'
                  placeholder='Function'
                  onChange={(value) => {
                    setEventFn(value);
                  }}
                  showSearch
                  filterOption={(input, option) =>
                    option.children
                      .toLowerCase()
                      .indexOf(input.toLowerCase()) >= 0
                  }
                >
                  {customKPIConfig?.result?.map((item) => {
                    if (item.type_of_query === 3) {
                      return item?.agFn.map((it) => {
                        return (
                          <Option key={it} value={it}>
                            {_.startCase(it)}
                          </Option>
                        );
                      });
                    }
                  })}
                </Select>
              </Form.Item>
            </Col>
          </Row>
        )}

        {EventFn &&
          EventFn !== 'unique' &&
          EventFn !== 'count' &&
          EventfilterDDValues && (
            <Row className='mt-8'>
              <Col span={18}>
                <Text type='title' level={7} extraClass='m-0'>
                  Select Property
                </Text>
                <Form.Item
                  name='event_property'
                  className='m-0'
                  rules={[
                    {
                      required: true,
                      message: 'Please select a property'
                    }
                  ]}
                >
                  <Select
                    className='fa-select w-full'
                    size='large'
                    disabled={!selEventName}
                    onChange={(value, details) => {
                      setEventPropertyDetails(details);
                    }}
                    placeholder='Select Property'
                    showSearch
                    filterOption={(input, option) =>
                      option.children
                        .toLowerCase()
                        .indexOf(input.toLowerCase()) >= 0
                    }
                  >
                    {Object.keys(eventProperties)?.map((category) => {
                      if (category === selEventName) {
                        return eventProperties[category]?.map((item) => {
                          if (item[2] === 'numerical') {
                            return (
                              <Option
                                key={item[0]}
                                value={item[0]}
                                name={item[1]}
                                data_type={item[2]}
                                en={'event'}
                              >
                                {item[0]}
                              </Option>
                            );
                          }
                        });
                      }
                    })}
                  </Select>
                </Form.Item>
              </Col>
            </Row>
          )}

        {EventfilterDDValues && (
          <Row className='mt-8'>
            <Col span={18}>
              <div className='border-top--thin-2 border-bottom--thin-2 pt-5 pb-5'>
                <Text type='title' level={7} weight='bold' extraClass='m-0'>
                  FILTER BY
                </Text>
                <EventFilter
                  filters={EventfilterValues?.globalFilters}
                  setGlobalFilters={setEventGlobalFiltersOption}
                  selEventName={selEventName}
                  eventProperties={eventProperties}
                />
              </div>
            </Col>
          </Row>
        )}
      </div>
    );
  }

  function renderEventBasedKPIView() {
    return (
      <div>
        <Row>
          <Col span={18}>
            <Text type='title' level={7} extraClass='m-0 mt-6'>
              Event
            </Text>
            <Input
              disabled
              size='large'
              value={eventNames[viewKPIDetails?.transformations?.evNm]}
              className='fa-input w-full'
              placeholder='Display Name'
            />
          </Col>
        </Row>
        <Row>
          <Col span={18}>
            <Text type='title' level={7} extraClass='m-0 mt-6'>
              Function
            </Text>
            <Input
              disabled
              size='large'
              value={_.startCase(viewKPIDetails?.transformations?.agFn)}
              className='fa-input w-full'
              placeholder='Display Name'
            />
          </Col>
        </Row>
        {!_.isEmpty(viewKPIDetails?.transformations?.agPr) && (
          <Row>
            <Col span={18}>
              <Text type='title' level={7} extraClass='m-0 mt-6'>
                Property
              </Text>
              <Input
                disabled
                size='large'
                value={eventPropNames[viewKPIDetails?.transformations?.agPr]}
                className='fa-input w-full'
                placeholder='Display Name'
              />
            </Col>
          </Row>
        )}
        {!_.isEmpty(viewKPIDetails?.transformations?.fil) && (
          <Row>
            <Col span={18}>
              <Text type='title' level={7} extraClass='m-0 mt-6'>
                Filter
              </Text>
              {/* {getGlobalFilters(viewKPIDetails?.transformations?.fil)} */}
              <GLobalFilter
                filters={getStateFromFilters(
                  viewKPIDetails?.transformations?.fil
                )}
                setGlobalFilters={setEventGlobalFiltersOption}
                selKPICategory={selEventName}
                DDKPIValues={EventfilterDDValues}
                delFilter={false}
                viewMode
              />
            </Col>
          </Row>
        )}
      </div>
    );
  }

  return (
    <div className='fa-container mt-32 mb-12 min-h-screen'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className='mb-10 pl-4'>
            {!showForm && !viewMode && (
              <>
                <Row>
                  <Col span={12}>
                    <Text type='title' level={3} weight='bold' extraClass='m-0'>
                      Custom KPIs
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className='flex justify-end'>
                      <Button size='large' onClick={() => setShowForm(true)}>
                        <SVG name='plus' extraClass='mr-2' size={16} />
                        Add New
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className='mt-4'>
                  <Col span={24}>
                    <div className='mt-6'>
                      <Text
                        type='title'
                        level={7}
                        color='grey-2'
                        extraClass='m-0'
                      >
                        Have a specific KPI that you measure based on the values
                        of a CRM object's fields? Say no more — it's easy to set
                        this up, so you can measure them over time.
                      </Text>
                      <Text
                        type='title'
                        level={7}
                        color='grey-2'
                        extraClass='m-0 mt-2'
                      >
                        All it takes is filtering for the CRM objects, adding
                        your custom conditions over it, and you should be good
                        to go!
                      </Text>

                      <Table
                        className='fa-table--basic mt-8'
                        columns={columns}
                        dataSource={tableData}
                        pagination={false}
                        loading={tableLoading}
                        tableLayout='fixed'
                      />
                    </div>
                  </Col>
                </Row>
              </>
            )}
            {showForm && !viewMode && (
              <Form
                form={form}
                onFinish={onFinish}
                className='w-full'
                onChange={onChange}
                loading
              >
                <Row>
                  <Col span={12}>
                    <Text type='title' level={3} weight='bold' extraClass='m-0'>
                      New Custom KPI
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className='flex justify-end'>
                      <Button
                        size='large'
                        disabled={loading}
                        onClick={() => {
                          onReset();
                        }}
                      >
                        Cancel
                      </Button>
                      <Button
                        size='large'
                        disabled={loading}
                        loading={loading}
                        className='ml-2'
                        type='primary'
                        htmlType='submit'
                      >
                        Save
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className='mt-8'>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0'>
                      KPI Name
                    </Text>
                    <Form.Item
                      name='name'
                      rules={[
                        { required: true, message: 'Please enter KPI name' }
                      ]}
                    >
                      <Input
                        disabled={loading}
                        size='large'
                        className='fa-input w-full'
                        placeholder='Display Name'
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Row className='mt-8'>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0'>
                      Description
                    </Text>
                    <Form.Item
                      name='description'
                      rules={[
                        {
                          required: true,
                          message: 'Please enter description'
                        }
                      ]}
                    >
                      <Input
                        disabled={loading}
                        size='large'
                        className='fa-input w-full'
                        placeholder='Description'
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Row className='mt-8'>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0'>
                      KPI Type
                    </Text>
                    <Form.Item name='kpi_type' className='m-0'>
                      <Select
                        className='fa-select w-full'
                        size='large'
                        onChange={(value) => onKPITypeChange(value)}
                        placeholder='KPI Type'
                        defaultValue='default'
                      >
                        <Option value='default'>Default</Option>
                        <Option value='derived_kpi'>Derived KPI</Option>
                        {whiteListedAccounts.includes(currentAgent?.email) && (
                          <Option value='event_based'>Event Based</Option>
                        )}
                      </Select>
                    </Form.Item>
                  </Col>
                </Row>

                {selKPIType === 'default' ? (
                  <div>
                    <Row className='mt-6'>
                      <Col span={24}>
                        <div className='border-top--thin-2 pt-3 mt-3' />
                      </Col>
                    </Row>
                    <Row className='m-0'>
                      <Col span={18}>
                        {/* <div className={'border-top--thin-2 pt-3 mt-3'} /> */}
                        <Text type='title' level={7} extraClass='m-0'>
                          Category
                        </Text>
                        <Form.Item
                          name='kpi_category'
                          className='m-0'
                          rules={[
                            {
                              required: true,
                              message: 'Please select KPI Category'
                            }
                          ]}
                        >
                          <Select
                            className='fa-select w-full'
                            size='large'
                            onChange={(value) => onKPICategoryChange(value)}
                            placeholder='KPI Category'
                            showSearch
                            filterOption={(input, option) =>
                              option.children
                                .toLowerCase()
                                .indexOf(input.toLowerCase()) >= 0
                            }
                          >
                            {customKPIConfig?.result?.map((item) => (
                              <Option key={item.obj_ty} value={item.obj_ty}>
                                {_.startCase(item.obj_ty)}
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                    </Row>

                    {selKPICategory && (
                      <Row className='mt-8'>
                        <Col span={18}>
                          <Text type='title' level={7} extraClass='m-0'>
                            Select Function
                          </Text>
                          <Form.Item
                            name='kpi_function'
                            className='m-0'
                            rules={[
                              {
                                required: true,
                                message: 'Please select a Function'
                              }
                            ]}
                          >
                            <Select
                              className='fa-select w-full'
                              size='large'
                              placeholder='Function'
                              onChange={(value) => {
                                setKPIFn(value);
                              }}
                              showSearch
                              filterOption={(input, option) =>
                                option.children
                                  .toLowerCase()
                                  .indexOf(input.toLowerCase()) >= 0
                              }
                            >
                              {customKPIConfig?.result?.map((item) => {
                                if (item.obj_ty === selKPICategory) {
                                  return item?.agFn.map((it) => {
                                    return (
                                      <Option key={it} value={it}>
                                        {_.startCase(it)}
                                      </Option>
                                    );
                                  });
                                }
                              })}
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    )}

                    {KPIFn && KPIFn != 'unique' && filterDDValues && (
                      <Row className='mt-8'>
                        <Col span={18}>
                          <Text type='title' level={7} extraClass='m-0'>
                            Select Property
                          </Text>
                          <Form.Item
                            name='kpi_property'
                            className='m-0'
                            rules={[
                              {
                                required: true,
                                message: 'Please select a property'
                              }
                            ]}
                          >
                            <Select
                              className='fa-select w-full'
                              size='large'
                              disabled={!selKPICategory}
                              onChange={(value, details) => {
                                setKPIPropertyDetails(details);
                              }}
                              placeholder='Select Property'
                              showSearch
                              filterOption={(input, option) =>
                                option.children
                                  .toLowerCase()
                                  .indexOf(input.toLowerCase()) >= 0
                              }
                            >
                              {customKPIConfig?.result?.map((category) => {
                                if (category.obj_ty == selKPICategory) {
                                  return category?.properties?.map((item) => {
                                    if (item.data_type == 'numerical') {
                                      return (
                                        <Option
                                          key={item.name}
                                          value={item.display_name}
                                          name={item.name}
                                          data_type={item.data_type}
                                          entity={item.entity}
                                        >
                                          {_.startCase(item.display_name)}
                                        </Option>
                                      );
                                    }
                                  });
                                }
                              })}
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    )}

                    {filterDDValues && (
                      <>
                        <Row className='mt-8'>
                          <Col span={18}>
                            <div className='border-top--thin-2 border-bottom--thin-2 pt-5 pb-5'>
                              {/* <Collapse defaultActiveKey={['1']} ghost expandIconPosition={'right'}>
                                        <Panel header={<Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>FILTER BY</Text>} key="1">
                                         */}

                              <Text
                                type='title'
                                level={7}
                                weight='bold'
                                extraClass='m-0'
                              >
                                FILTER BY
                              </Text>
                              <GLobalFilter
                                filters={filterValues?.globalFilters}
                                onFiltersLoad={[
                                  () => {
                                    getUserProperties(activeProject.id, null);
                                  }
                                ]}
                                setGlobalFilters={setGlobalFiltersOption}
                                selKPICategory={selKPICategory}
                                DDKPIValues={filterDDValues}
                              />
                              {/* </Panel>
                                    </Collapse> */}
                            </div>
                          </Col>
                        </Row>

                        <Row className='mt-8'>
                          <Col span={18}>
                            <Text type='title' level={7} extraClass='m-0'>
                              Set time to
                            </Text>
                            <Form.Item
                              name='kpi_dateField'
                              className='m-0'
                              rules={[
                                {
                                  required: true,
                                  message: 'Please select a date field'
                                }
                              ]}
                            >
                              <Select
                                className='fa-select w-full'
                                size='large'
                                disabled={!selKPICategory}
                                placeholder='Date field'
                                showSearch
                                filterOption={(input, option) =>
                                  option.children
                                    .toLowerCase()
                                    .indexOf(input.toLowerCase()) >= 0
                                }
                              >
                                {customKPIConfig?.result?.map((category) => {
                                  if (category.obj_ty === selKPICategory) {
                                    return category?.properties?.map((item) => {
                                      if (item.data_type === 'datetime')
                                        return (
                                          <Option
                                            key={item.name}
                                            value={item.name}
                                            name={item.name}
                                            data_type={item.data_type}
                                            entity={item.entity}
                                          >
                                            {_.startCase(item.display_name)}
                                          </Option>
                                        );
                                    });
                                  }
                                })}
                              </Select>
                            </Form.Item>
                          </Col>
                        </Row>
                      </>
                    )}
                  </div>
                ) : selKPIType === 'derived_kpi' ? (
                  <>
                    <Row className='mt-6'>
                      <Col span={24}>
                        <div className='border-top--thin-2 pt-3 mt-3' />
                        <Text type='title' level={6} extraClass='m-0'>
                          Select KPIs and Formula
                        </Text>
                      </Col>
                    </Row>
                    <div className='mt-4 border rounded-lg'>
                      <Row className='m-0 ml-4 my-2'>
                        <Col span={18}>
                          <Form.Item name='query_type' className='m-0'>
                            {queryList()}
                          </Form.Item>
                        </Col>
                      </Row>
                      <Row className='m-0'>
                        <Col span={24}>
                          <div className='border-top--thin-2 pt-3 mt-3' />
                        </Col>
                      </Row>
                      <Row className='m-0 ml-4 my-3'>
                        <Col>
                          <Text
                            type='title'
                            level={7}
                            color='grey'
                            extraClass='m-0 pt-2 mr-3'
                          >
                            Formula:
                          </Text>
                        </Col>
                        <Col span={14}>
                          <Form.Item
                            name='for'
                            rules={[
                              {
                                required: true,
                                message: 'Please enter formula'
                              }
                            ]}
                          >
                            <Input
                              // disabled={loading}
                              size='large'
                              // className={'fa-input w-full'}
                              placeholder='please type the formula. Eg A/B, A+B, A-B, A*B'
                              bordered={false}
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                    </div>
                  </>
                ) : (
                  [renderEventBasedKPIForm()]
                )}
              </Form>
            )}

            {viewMode && (
              <>
                <Row>
                  <Col span={12}>
                    <Text type='title' level={3} weight='bold' extraClass='m-0'>
                      View Custom KPI
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className='flex justify-end'>
                      <Button
                        size='large'
                        disabled={loading}
                        onClick={() => {
                          KPIviewMode(false);
                        }}
                      >
                        Back
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className='mt-8'>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0'>
                      KPI Name
                    </Text>
                    <Input
                      disabled
                      size='large'
                      value={viewKPIDetails?.name}
                      className='fa-input w-full'
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                <Row>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0 mt-6'>
                      Description
                    </Text>
                    <Input
                      disabled
                      size='large'
                      value={viewKPIDetails?.description}
                      className='fa-input w-full'
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                <Row>
                  <Col span={18}>
                    <Text type='title' level={7} extraClass='m-0 mt-6'>
                      KPI Type
                    </Text>
                    <Input
                      disabled
                      size='large'
                      value={
                        viewKPIDetails?.type_of_query === 1
                          ? 'Default'
                          : viewKPIDetails?.type_of_query === 2
                          ? 'Derived'
                          : 'Event Based'
                      }
                      className='fa-input w-full'
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                {viewKPIDetails?.type_of_query === 1 ? (
                  <div>
                    <Row>
                      <Col span={18}>
                        <Text type='title' level={7} extraClass='m-0 mt-6'>
                          Category
                        </Text>
                        <Input
                          disabled
                          size='large'
                          value={_.startCase(viewKPIDetails?.obj_ty)}
                          className='fa-input w-full'
                          placeholder='Display Name'
                        />
                      </Col>
                    </Row>
                    <Row>
                      <Col span={18}>
                        <Text type='title' level={7} extraClass='m-0 mt-6'>
                          Function
                        </Text>
                        <Input
                          disabled
                          size='large'
                          value={_.startCase(
                            viewKPIDetails?.transformations?.agFn
                          )}
                          className='fa-input w-full'
                          placeholder='Display Name'
                        />
                      </Col>
                    </Row>
                    {!_.isEmpty(viewKPIDetails?.transformations?.agPr) && (
                      <Row>
                        <Col span={18}>
                          <Text type='title' level={7} extraClass='m-0 mt-6'>
                            Property
                          </Text>
                          <Input
                            disabled
                            size='large'
                            value={matchEventName(
                              viewKPIDetails?.transformations?.agPr
                            )}
                            className='fa-input w-full'
                            placeholder='Display Name'
                          />
                        </Col>
                      </Row>
                    )}
                    {!_.isEmpty(viewKPIDetails?.transformations?.fil) && (
                      <Row>
                        <Col span={18}>
                          <Text type='title' level={7} extraClass='m-0 mt-6'>
                            Filter
                          </Text>
                          {/* {getGlobalFilters(viewKPIDetails?.transformations?.fil)} */}
                          <GLobalFilter
                            filters={getStateFromFilters(
                              viewKPIDetails?.transformations?.fil
                            )}
                            onFiltersLoad={[
                              () => {
                                getUserProperties(activeProject.id, null);
                              }
                            ]}
                            setGlobalFilters={setGlobalFiltersOption}
                            selKPICategory={selKPICategory}
                            DDKPIValues={filterDDValues}
                            delFilter={false}
                            viewMode
                          />
                        </Col>
                      </Row>
                    )}
                    <Row>
                      <Col span={18}>
                        <Text type='title' level={7} extraClass='m-0 mt-6'>
                          Set time to
                        </Text>
                        <Input
                          disabled
                          size='large'
                          value={matchEventName(
                            viewKPIDetails?.transformations?.daFie
                          )}
                          className='fa-input w-full'
                          placeholder='Display Name'
                        />
                      </Col>
                    </Row>
                  </div>
                ) : viewKPIDetails?.type_of_query === 2 ? (
                  <>
                    <Row className='mt-6'>
                      <Col span={24}>
                        <div className='border-top--thin-2 pt-3 mt-3' />
                        <Text type='title' level={6} extraClass='m-0'>
                          KPIs and Formula
                        </Text>
                      </Col>
                    </Row>
                    <div className='mt-4 border rounded-lg'>
                      {viewKPIDetails?.transformations?.qG.map(
                        (item, index) => (
                          <div className='py-2'>
                            <Row className='m-0 mt-1 ml-4'>
                              <Col>
                                <div className='flex items-center fa--query_block_section borderless no-padding mt-1'>
                                  <div className='fa--query_block--add-event active flex justify-center items-center mr-2'>
                                    <Text
                                      disabled
                                      type='title'
                                      level={7}
                                      weight='bold'
                                      color='white'
                                      extraClass='m-0'
                                    >
                                      {alphabetIndex[index]}
                                    </Text>
                                  </div>
                                </div>
                              </Col>
                              <Col>
                                <Button className='mr-2' type='link' disabled>
                                  {_.startCase(item?.me[0])}
                                </Button>
                              </Col>
                              <Col>
                                {item?.pgUrl && (
                                  <div>
                                    <span className='mr-2'>from</span>
                                    <Button
                                      className='mr-2'
                                      type='link'
                                      disabled
                                    >
                                      {item?.pgUrl}
                                    </Button>
                                  </div>
                                )}
                              </Col>
                            </Row>
                            {item?.fil?.length > 0 && (
                              <Row className='mt-2 ml-4'>
                                <Col span={18}>
                                  <Text
                                    type='title'
                                    level={7}
                                    color='grey'
                                    extraClass='m-0 ml-1 my-1'
                                  >
                                    Filters
                                  </Text>
                                  {getStateFromFilters(item.fil).map(
                                    (filter, index) => (
                                      <div key={index} className='mt-1'>
                                        <Button
                                          className='mr-2'
                                          type='link'
                                          disabled
                                        >
                                          {filter.extra[0]}
                                        </Button>
                                        <Button
                                          className='mr-2'
                                          type='link'
                                          disabled
                                        >
                                          {filter.operator}
                                        </Button>
                                        <Button
                                          className='mr-2'
                                          type='link'
                                          disabled
                                        >
                                          {filter.values[0]}
                                        </Button>
                                      </div>
                                    )
                                  )}
                                </Col>
                              </Row>
                            )}
                          </div>
                        )
                      )}
                      <Row className='m-0'>
                        <Col span={24}>
                          <div className='border-top--thin-2 pt-3 mt-3' />
                        </Col>
                      </Row>
                      <Row className='m-0 ml-4 my-3'>
                        <Col>
                          <Text
                            type='title'
                            level={7}
                            color='grey'
                            extraClass='m-0 pt-2 mr-3'
                          >
                            Formula:
                          </Text>
                        </Col>
                        <Col span={14}>
                          <Input
                            disabled
                            size='large'
                            value={viewKPIDetails?.transformations?.for}
                            // className={'fa-input w-full'}
                            placeholder='Type your formula.  Eg A/B, A+B, A-B, A*B'
                            bordered={false}
                          />
                        </Col>
                      </Row>
                    </div>
                  </>
                ) : (
                  [renderEventBasedKPIView()]
                )}
              </>
            )}
          </div>
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  customKPIConfig: state.kpi?.custom_kpi_config,
  savedCustomKPI: state.kpi?.saved_custom_kpi,
  userPropNames: state.coreQuery?.userPropNames,
  eventPropNames: state.coreQuery?.eventPropNames,
  currentAgent: state.agent.agent_details,
  eventNames: state.coreQuery?.eventNames,
  eventNameOptions: state.coreQuery.eventOptions,
  eventProperties: state.coreQuery.eventProperties
});

export default connect(mapStateToProps, {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI,
  fetchEventNames,
  getEventProperties
})(CustomKPI);
