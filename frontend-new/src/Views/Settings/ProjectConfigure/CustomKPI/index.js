import React, { useState, useEffect } from 'react';
import { connect, useSelector, useDispatch } from 'react-redux';
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
  message,
  Collapse,
  notification,
} from 'antd';
import styles from './index.module.scss';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined } from '@ant-design/icons';
import {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI
} from 'Reducers/kpi';
import GLobalFilter from './GLobalFilter';
import { getUserProperties } from 'Reducers/coreQuery/middleware';
import { formatFilterDate } from '../../../../utils/dataFormatter';
import _ from 'lodash';
import {
  reverseOperatorMap,
  reverseDateOperatorMap,
  convertDateTimeObjectValuesToMilliSeconds,
  getKPIQuery,
  DefaultDateRangeFormat,
} from './utils';
import { FILTER_TYPES } from '../../../CoreQuery/constants';
import QueryBlock from './QueryBlock';
import { deleteGroupByForEvent } from '../../../../reducers/coreQuery/middleware';
import { INITIAL_SESSION_ANALYTICS_SEQ, QUERY_OPTIONS_DEFAULT_VALUE } from '../../../../utils/constants';

const { Panel } = Collapse;
const { Option, OptGroup } = Select; 
 

const CustomKPI = ({
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
  fetchKPIConfigWithoutDerivedKPI
}) => {
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
      date_range: { ...DefaultDateRangeFormat },
    });

    const { groupBy } = useSelector((state) => state.coreQuery);

 

const matchEventName = (item) => { 
  let findItem = eventPropNames?.[item] || userPropNames?.[item]
  return findItem ? findItem : item
}


  const menu = (item) => {
    return (
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
  };




  const columns = [
    {
      title: 'KPI Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Text type={'title'} level={7} truncate={true} charLimit={25}>
          {text}
        </Text>
      ),
      // width: 100,
    },
    {
      title: 'Description',
      dataIndex: 'desc',
      key: 'desc',
      render: (text) => (
        <Text type={'title'} level={7} truncate={true} charLimit={25}>
          {text}
        </Text>
      ),
      // width: 200,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (item) => (
        <Text type={'title'} level={7} truncate={true} charLimit={35}>{item}</Text>
      ),
      width: 'auto',
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
      ),
    },
  ];
  const onChange = () => {
    seterrorInfo(null);
  };

  const setGlobalFiltersOption = (filters) => {
    const opts = Object.assign({}, queryOptions);
    opts.globalFilters = filters;
    setFilterValues(opts);
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
    since: 'since',
  };

  const getEventsWithPropertiesKPI = (filters, category = null) => {
    const filterProps = [];
    // adding fil?.extra ? fil?.extra[*] check as a hotfix for timestamp filters
    filters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : `$${_.lowerCase(fil?.props[0])}`,
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              (category == 'channels' || category == 'custom_channels')
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : 'event',
            objTy:
              (category == 'channels' || category == 'custom_channels')
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val,
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : `$${_.lowerCase(fil?.props[0])}`,
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            (category == 'channels' || category == 'custom_channels') ? '' : fil?.extra ? fil?.extra[3] : 'event',
          objTy:
            (category == 'channels' || category == 'custom_channels')
              ? fil?.extra
                ? fil?.extra[3]
                : 'event'
              : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values,
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
      } else {
        if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
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
        <div key={'init'} className={styles.composer_body__query_block}>
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
    setQueries([])
  }

  const whiteListedAccounts = [
    'junaid@factors.ai',
    'solutions@factors.ai',
  ];               

  const onFinish = (data) => {
    let payload;
    if(selKPIType === 'default') {
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
          daFie: data.kpi_dateField,
        },
      }; 
    } else {
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
        },
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
            'New KPI is created and saved successfully. You can start using it across the product shortly.',
        });
        onReset();
      })
      .catch((err) => {
        setLoading(false);
        notification.error({
          message: 'Error',
          description: err?.data?.error,
        });
        console.log('addNewCustomKPI error->', err);
      });
  };

  const deleteKPI = (item) =>{ 
      removeCustomKPI(activeProject.id,item?.id).then(()=>{
        fetchSavedCustomKPI(activeProject.id);
        notification.success({
          message: 'KPI Removed',
          description:
            'Custom KPI is removed successfully.',
        });
      }).catch((err) => {
        notification.error({
          message: 'Error',
          description: err?.data?.error,
        });
        console.log('addNewCustomKPI error->', err);
      });
}


  useEffect(() => {
    // if (!customKPIConfig) {
    fetchCustomKPIConfig(activeProject.id);
    // }
    // if (!savedCustomKPI) {
    fetchSavedCustomKPI(activeProject.id);
    // }
    fetchKPIConfigWithoutDerivedKPI(activeProject.id);
  }, [activeProject]); 

  useEffect(() => {
    let DDCategory = customKPIConfig?.result?.objTyAndProp?.find((category) => {
      if (category.objTy == selKPICategory) {
        return category;
      }
    });

    let DDvalues = DDCategory?.properties?.map((item) => {
      return [item.display_name, item.name, item.data_type, item.entity];
    });

    setFilterDDValues(DDvalues);
  }, [selKPICategory, customKPIConfig]);

  const onKPICategoryChange = (value) => {
    setKPICategory(value);
  };

  const onKPITypeChange = (value) => {
    setKPIType(value);
  };

  useEffect(() => {
    if (savedCustomKPI) {
      let savedArr = [];
      savedCustomKPI?.map((item, index) => {
        savedArr.push({
          key: index,
          name: item.name,
          desc: item.description,
          type: item.type_of_query === 1 ? 'Default': 'Derived',
          actions: item,
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

        const DNa = _.startCase(pr.prNa);

        filters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy],
        });
      } else if (pr.prDaTy === FILTER_TYPES.CATEGORICAL) {
        filters[filters.length - 1].values.push(pr.va);
      }
    });
    return filters;
  };

  return (
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className={'mb-10 pl-4'}>
            {!showForm && !viewMode && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      Custom KPIs
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className={'flex justify-end'}>
                      <Button size={'large'} onClick={() => setShowForm(true)}>
                        <SVG name={'plus'} extraClass={'mr-2'} size={16} />
                        Add New
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className={'mt-4'}>
                  <Col span={24}>
                    <div className={'mt-6'}>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0'}
                      >
                        Have a specific KPI that you measure based on the values of a CRM object's fields? Say no more — it's easy to set this up, so you can measure them over time.
                      </Text>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0 mt-2'}
                      >
                        All it takes is filtering for the CRM objects, adding your custom conditions over it, and you should be good to go!
                      </Text>

                      <Table
                        className='fa-table--basic mt-8'
                        columns={columns}
                        dataSource={tableData}
                        pagination={false}
                        loading={tableLoading}
                        tableLayout={'fixed'}
                      />
                    </div>
                  </Col>
                </Row>
              </>
            )}
            {showForm && !viewMode && (
              <>
                <Form
                  form={form}
                  onFinish={onFinish}
                  className={'w-full'}
                  onChange={onChange}
                  loading={true}
                >
                  <Row>
                    <Col span={12}>
                      <Text
                        type={'title'}
                        level={3}
                        weight={'bold'}
                        extraClass={'m-0'}
                      >
                        New Custom KPI
                      </Text>
                    </Col>
                    <Col span={12}>
                      <div className={'flex justify-end'}>
                        <Button
                          size={'large'}
                          disabled={loading}
                          onClick={() => {
                            onReset();
                          }}
                        >
                          Cancel
                        </Button>
                        <Button
                          size={'large'}
                          disabled={loading}
                          loading={loading}
                          className={'ml-2'}
                          type={'primary'}
                          htmlType='submit'
                        >
                          Save
                        </Button>
                      </div>
                    </Col>
                  </Row>
                  <Row className={'mt-8'}>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        KPI Name
                      </Text>
                      <Form.Item
                        name='name'
                        rules={[
                          { required: true, message: 'Please enter KPI name' },
                        ]}
                      >
                        <Input
                          disabled={loading}
                          size='large'
                          className={'fa-input w-full'}
                          placeholder='Display Name'
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row className={'mt-8'}>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        Description
                      </Text>
                      <Form.Item
                        name='description'
                        rules={[
                          {
                            required: true,
                            message: 'Please enter description',
                          },
                        ]}
                      >
                        <Input
                          disabled={loading}
                          size='large'
                          className={'fa-input w-full'}
                          placeholder='Description'
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  { whiteListedAccounts.includes(currentAgent?.email) &&
                  <Row className={'mt-8'}>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        KPI Type
                      </Text>
                      <Form.Item
                        name='kpi_type'
                        className={'m-0'}
                      >
                        <Select
                          className={'fa-select w-full'}
                          size={'large'}
                          onChange={(value) => onKPITypeChange(value)}
                          placeholder='KPI Type'
                          defaultValue={'default'}
                        >
                          <Option value='default'>Default</Option>
                          <Option value='derived_kpi'>Derived KPI</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                  </Row>
                  }
                  
                  {selKPIType === 'default' ?
                  <div>
                  <Row className={'mt-6'}>
                    <Col span={24}>
                      <div className={'border-top--thin-2 pt-3 mt-3'} />
                    </Col>
                  </Row>
                  <Row className={'m-0'}>
                    <Col span={18}>
                      {/* <div className={'border-top--thin-2 pt-3 mt-3'} /> */}
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        Category
                      </Text>
                      <Form.Item
                        name='kpi_category'
                        className={'m-0'}
                        rules={[
                          {
                            required: true,
                            message: 'Please select KPI Category',
                          },
                        ]}
                      >
                        <Select
                          className={'fa-select w-full'}
                          size={'large'}
                          onChange={(value) => onKPICategoryChange(value)}
                          placeholder='KPI Category'
                          showSearch
                          filterOption={(input, option) =>
                            option.children
                              .toLowerCase()
                              .indexOf(input.toLowerCase()) >= 0
                          }
                        >
                          {customKPIConfig?.result?.objTyAndProp?.map(
                            (item) => {
                              return (
                                <Option key={item.objTy} value={item.objTy}>
                                  {_.startCase(item.objTy)}
                                </Option>
                              );
                            }
                          )}
                        </Select>
                      </Form.Item>
                    </Col>
                  </Row>

                  {selKPICategory && (
                    <Row className={'mt-8'}>
                      <Col span={18}>
                        <Text type={'title'} level={7} extraClass={'m-0'}>
                          Select Function
                        </Text>
                        <Form.Item
                          name='kpi_function'
                          className={'m-0'}
                          rules={[
                            {
                              required: true,
                              message: 'Please select a Function',
                            },
                          ]}
                        >
                          <Select
                            className={'fa-select w-full'}
                            size={'large'}
                            placeholder='Function'
                            onChange={(value, details) => {
                              setKPIFn(value);
                            }}
                            showSearch
                            filterOption={(input, option) =>
                              option.children
                                .toLowerCase()
                                .indexOf(input.toLowerCase()) >= 0
                            }
                          >
                            {customKPIConfig?.result?.agFn?.map((item) => {
                              return (
                                <Option key={item} value={item}>
                                  {_.startCase(item)}
                                </Option>
                              );
                            })}
                          </Select>
                        </Form.Item>
                      </Col>
                    </Row>
                  )}

                  {KPIFn && KPIFn != 'unique' && filterDDValues && (
                    <>
                      <Row className={'mt-8'}>
                        <Col span={18}>
                          <Text type={'title'} level={7} extraClass={'m-0'}>
                            Select Property
                          </Text>
                          <Form.Item
                            name='kpi_property'
                            className={'m-0'}
                            rules={[
                              {
                                required: true,
                                message: 'Please select a property',
                              },
                            ]}
                          >
                            <Select
                              className={'fa-select w-full'}
                              size={'large'}
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
                              {customKPIConfig?.result?.objTyAndProp?.map(
                                (category) => {
                                  if (category.objTy == selKPICategory) {
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
                                }
                              )}
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    </>
                  )}

                  {filterDDValues && (
                    <>
                      <Row className={'mt-8'}>
                        <Col span={18}>
                          <div
                            className={
                              'border-top--thin-2 border-bottom--thin-2 pt-5 pb-5'
                            }
                          >
                            {/* <Collapse defaultActiveKey={['1']} ghost expandIconPosition={'right'}>
                                        <Panel header={<Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>FILTER BY</Text>} key="1">
                                         */}

                            <Text
                              type={'title'}
                              level={7}
                              weight={'bold'}
                              extraClass={'m-0'}
                            >
                              FILTER BY
                            </Text>
                            <GLobalFilter
                              filters={filterValues?.globalFilters}
                              onFiltersLoad={[
                                () => {
                                  getUserProperties(activeProject.id, null);
                                },
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

                      <Row className={'mt-8'}>
                        <Col span={18}>
                          <Text type={'title'} level={7} extraClass={'m-0'}>
                            Set time to
                          </Text>
                          <Form.Item
                            name='kpi_dateField'
                            className={'m-0'}
                            rules={[
                              {
                                required: true,
                                message: 'Please select a date field',
                              },
                            ]}
                          >
                            <Select
                              className={'fa-select w-full'}
                              size={'large'}
                              disabled={!selKPICategory}
                              placeholder='Date field'
                              showSearch
                              filterOption={(input, option) =>
                                option.children
                                  .toLowerCase()
                                  .indexOf(input.toLowerCase()) >= 0
                              }
                            >
                              {customKPIConfig?.result?.objTyAndProp?.map(
                                (category) => {
                                  if (category.objTy == selKPICategory) {
                                    return category?.properties?.map((item) => {
                                      if (item.data_type == 'datetime')
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
                                }
                              )}
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    </>
                  )}
                  </div>
                  :
                  <>
                    <Row className={'mt-6'}>
                      <Col span={24}>
                        <div className={'border-top--thin-2 pt-3 mt-3'} />
                        <Text type={'title'} level={6} extraClass={'m-0'}>
                          Select KPIs and Formula
                        </Text>
                      </Col>
                    </Row>
                    <div className={'mt-4 border rounded-lg'}>
                      <Row className={'m-0 ml-4 my-2'}>
                          <Col span={18}>
                              <Form.Item
                                  name="query_type"
                                  className={'m-0'}
                              >
                                  {queryList()}
                              </Form.Item>
                          </Col>
                      </Row>
                      <Row className={'m-0'}>
                        <Col span={24}>
                          <div className={'border-top--thin-2 pt-3 mt-3'} />
                        </Col>
                      </Row>
                      <Row className={'m-0 ml-4 my-3'}>
                          <Col>
                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 pt-2 mr-3'}>
                              Formula:
                            </Text>
                          </Col>
                          <Col span={14}>
                          <Form.Item
                            name='for'
                            rules={[
                              {
                                required: true,
                                message: 'Please enter formula',
                              },
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
                  }
                </Form>
              </>
            )}

            {viewMode && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      View Custom KPI
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className={'flex justify-end'}>
                      <Button
                        size={'large'}
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
                <Row className={'mt-8'}>
                  <Col span={18}>
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      KPI Name
                    </Text>
                    <Input
                      disabled={true}
                      size='large'
                      value={viewKPIDetails?.name}
                      className={'fa-input w-full'}
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                <Row>
                  <Col span={18}>
                    <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                      Description
                    </Text>
                    <Input
                      disabled={true}
                      size='large'
                      value={viewKPIDetails?.description}
                      className={'fa-input w-full'}
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                <Row>
                  <Col span={18}>
                    <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                      KPI Type
                    </Text>
                    <Input
                      disabled={true}
                      size='large'
                      value={viewKPIDetails?.type_of_query === 1 ? 'Default' : 'Derived'}
                      className={'fa-input w-full'}
                      placeholder='Display Name'
                    />
                  </Col>
                </Row>
                {viewKPIDetails?.type_of_query === 1 ?
                <div>
                  <Row>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                        Category
                      </Text>
                      <Input
                        disabled={true}
                        size='large'
                        value={viewKPIDetails?.obj_ty}
                        className={'fa-input w-full'}
                        placeholder='Display Name'
                      />
                    </Col>
                  </Row>
                  <Row>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                        Function
                      </Text>
                      <Input
                        disabled={true}
                        size='large'
                        value={viewKPIDetails?.transformations?.agFn}
                        className={'fa-input w-full'}
                        placeholder='Display Name'
                      />
                    </Col>
                  </Row>
                  {!_.isEmpty(viewKPIDetails?.transformations?.agPr) && (
                  <Row>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                        Property
                      </Text>
                      <Input
                        disabled={true}
                        size='large'
                        value={matchEventName(viewKPIDetails?.transformations?.agPr)}
                        className={'fa-input w-full'}
                        placeholder='Display Name'
                      />
                    </Col>
                  </Row>
                  )}
                  {!_.isEmpty(viewKPIDetails?.transformations?.fil) && (
                    <Row>
                      <Col span={18}>
                        <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
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
                            },
                          ]}
                          setGlobalFilters={setGlobalFiltersOption}
                          selKPICategory={selKPICategory}
                          DDKPIValues={filterDDValues}
                          delFilter={false}
                          viewMode={true}
                        />
                      </Col>
                    </Row>
                  )}
                  <Row>
                    <Col span={18}>
                      <Text type={'title'} level={7} extraClass={'m-0 mt-6'}>
                        Set time to
                      </Text>
                      <Input
                        disabled={true}
                        size='large'
                        value={matchEventName(viewKPIDetails?.transformations?.daFie)}
                        className={'fa-input w-full'}
                        placeholder='Display Name'
                      />
                    </Col>
                  </Row>
                </div>
                :
                <>
                    <Row className={'mt-6'}>
                      <Col span={24}>
                        <div className={'border-top--thin-2 pt-3 mt-3'} />
                        <Text type={'title'} level={6} extraClass={'m-0'}>
                          KPIs and Formula
                        </Text>
                      </Col>
                    </Row>
                    <div className={'mt-4 border rounded-lg'}>
                    {viewKPIDetails?.transformations?.qG.map((item) => (
                      <>
                      <div className={'py-2'}>
                      <Row className={'m-0 mt-1 ml-4'}>
                          <Col>
                              <Button
                              className={`mr-2`}
                              type='link'
                              disabled={true}
                              >
                                  {(item?.me[0]).replace(/_/g, ' ')}
                              </Button>
                          </Col>
                          <Col>
                              {item?.pgUrl && (
                                  <div>
                                      <span className={'mr-2'}>from</span>
                                      <Button
                                      className={`mr-2`}
                                      type='link'
                                      disabled={true}
                                      >
                                          {item?.pgUrl}
                                      </Button>
                                  </div>
                              )}
                          </Col>
                      </Row>
                      {item?.fil?.length > 0 && (
                        <Row className={'mt-2 ml-4'}>
                            <Col span={18}>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 ml-1 my-1'}>Filters</Text>
                                {getStateFromFilters(item.fil).map((filter, index) => (
                                    <div key={index} className={'mt-1'}>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.extra[0]}
                                        </Button>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.operator}
                                        </Button>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.values[0]}
                                        </Button>
                                    </div>
                                ))}
                            </Col>
                        </Row>
                      )}
                      </div>
                      </>
                    ))}
                      <Row className={'m-0'}>
                        <Col span={24}>
                          <div className={'border-top--thin-2 pt-3 mt-3'} />
                        </Col>
                      </Row>
                      <Row className={'m-0 ml-4 my-3'}>
                          <Col>
                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 pt-2 mr-3'}>
                              Formula:
                            </Text>
                          </Col>
                          <Col span={14}>
                            <Input
                              disabled={true}
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
                }
              </>
            )}
          </div>
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  customKPIConfig: state.kpi?.custom_kpi_config,
  savedCustomKPI: state.kpi?.saved_custom_kpi,
  userPropNames: state.coreQuery?.userPropNames,
  eventPropNames: state.coreQuery?.eventPropNames, 
  currentAgent: state.agent.agent_details,
});

export default connect(mapStateToProps, {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI

})(CustomKPI);
