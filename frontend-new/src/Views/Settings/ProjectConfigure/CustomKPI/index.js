import React, { useState, useEffect, useMemo } from 'react';
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
  notification,
  Checkbox,
  Divider
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, PlusOutlined } from '@ant-design/icons';
import {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI
} from 'Reducers/kpi';
import {
  getUserPropertiesV2,
  deleteGroupByForEvent,
  fetchEventNames,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import _ from 'lodash';
import GLobalFilter from 'Components/KPIComposer/GlobalFilter';
import EventFilter from 'Components/GlobalFilter';
import useAutoFocus from 'hooks/useAutoFocus';
import {
  getStateFromKPIFilters,
  DefaultDateRangeFormat,
  getEventsWithPropertiesCustomKPI,
  getCustomKPIQuery,
  getStateFromCustomKPIqueryGroup
} from 'Views/CoreQuery/utils';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { convertAndAddPropertiesToGroupSelectOptions } from 'Utils/dataFormatter';
import FaSelect from 'Components/GenericComponents/FaSelect';
import EventQueryBlock from './EventQueryBlock';
import {
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE
} from '../../../../utils/constants';
import QueryBlock from './QueryBlock';
import styles from './index.module.scss';
import EmptyScreen from 'Components/EmptyScreen';

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
  getEventPropertiesV2,
  eventPropertiesV2
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
  const [pageMode, setPageMode] = useState('Initial');
  const [viewKPIDetails, setKPIDetails] = useState(false);
  const [showAsPercentage, setShowAsPercentage] = useState(false);
  const [selEventName, setEventName] = useState(false);
  const [EventPropertyDetails, setEventPropertyDetails] = useState({});
  const [EventfilterDDValues, setEventFilterDDValues] = useState();
  const [EventfilterValues, setEventFilterValues] = useState([]);
  const [EventFn, setEventFn] = useState(false);
  const inputComponentRef = useAutoFocus(
    pageMode === 'Create' || pageMode === 'Edit'
  );
  const [timePeriodRangeProperties, setTimePeriosRangeProperties] = useState(
    []
  );
  const [timePeriodRangeDDVisible, setTimePeriodRangeDDVisible] = useState([
    false,
    -1
  ]);
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

  const { groupBy, groups, groupProperties, userPropertiesV2 } = useSelector(
    (state) => state.coreQuery
  );
  const { config: kpiConfig } = useSelector((state) => state.kpi);

  const matchEventName = (item) => {
    const findItem = eventPropNames?.[item] || userPropNames?.[item];
    return findItem || item;
  };

  const handleViewCustomKPI = (item) => {
    setPageMode('View');
    setKPIDetails(item);
    if (item.metric_type === 'date_type_diff_metric') {
      setKPIType('time_period_based');
      setKPICategory(item.obj_ty);
      setKPIFn(item?.transformations?.agFn);
      setFilterValues({
        globalFilters: getStateFromKPIFilters(
          item?.transformations?.fil,
          userPropNames
        )
      });
    } else {
      setKPIType(item.type_of_query === 1 ? 'default' : 'derived_kpi');
    }
  };
  const handleCopyCustomKPI = (item) => {
    setPageMode('Edit');
    setKPIDetails(item);
    onEdit(item);
    if (item.metric_type === 'date_type_diff_metric') {
      setKPIType('time_period_based');
      setKPICategory(item.obj_ty);
      setKPIFn(item?.transformations?.agFn);

      setFilterValues({
        globalFilters: getStateFromKPIFilters(
          item?.transformations?.fil,
          userPropNames
        )
      });
    } else {
      setKPIType(item.type_of_query === 1 ? 'default' : 'derived_kpi');
    }
  };
  const menu = (item) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item key='0' onClick={() => handleViewCustomKPI(item)}>
        <SVG name='Eye' size={18} extraClass='mr-2 inline' />
        <span>View KPI</span>
      </Menu.Item>
      <Menu.Item key='1' onClick={() => handleCopyCustomKPI(item)}>
        <SVG name='Copy1' size={18} extraClass='mr-2 inline' />
        <span>Create copy</span>
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item
        key='2'
        onClick={() => {
          deleteKPI(item);
        }}
      >
        <SVG name='Delete1' size={18} color='red' extraClass='mr-2 inline' />
        <span className='text-red-600'>Remove</span>
      </Menu.Item>
    </Menu>
  );

  const alphabetIndex = 'ABCDEFGHIJK';

  const columns = [
    {
      title: 'KPI Name',
      dataIndex: 'name',
      key: 'name',
      render: (item) => (
        <Text
          type='title'
          level={7}
          truncate
          charLimit={25}
          onClick={() => handleViewCustomKPI(item)}
          extraClass='cursor-pointer'
        >
          {item?.name}
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

  const periodQueryList = () => {
    const tmpRes = customKPIConfig?.result?.filter(
      (e) => e.obj_ty === selKPICategory
    );
    let tmpOptions = (tmpRes && tmpRes[0]) || [];
    tmpOptions = (tmpOptions && tmpOptions.properties) || [];
    tmpOptions =
      tmpOptions &&
      tmpOptions
        .filter((e) => e.data_type === 'datetime')
        .map((eachOption) => ({
          label: eachOption.display_name,
          value: eachOption.name
        }));

    const tmpList = [];

    timePeriodRangeProperties.map((eachProperty, eachIndex) => {
      tmpList.push(
        <div className='flex justify-left items-center m-4'>
          <div
            key={eachIndex}
            className={`${styles.custom_kpi_order_active} flex justify-center items-center mr-2`}
          >
            <Text
              type='title'
              level={7}
              weight='bold'
              extraClass='m-0'
              color='white'
            >
              {eachIndex + 1}
            </Text>
          </div>
          <Button
            type='link'
            onClick={() => {
              setTimePeriodRangeDDVisible([true, eachIndex]);
            }}
            disabled={pageMode === 'View'}
          >
            {eachProperty.label}
          </Button>
          {pageMode !== 'View' && (
            <Button
              size='large'
              type='text'
              // onClick={deleteItem}
              className='fa-btn--custom ml-2'
              onClick={() => {
                setTimePeriosRangeProperties((prev) =>
                  prev.filter((e, i) => i !== eachIndex)
                );
              }}
            >
              <SVG name='trash' />
            </Button>
          )}
        </div>
      );
    });
    if (timePeriodRangeProperties.length < 2) {
      tmpList.push(
        <div>
          <Button
            icon={<PlusOutlined />}
            onClick={() => {
              setTimePeriodRangeDDVisible([true, -1]);
            }}
          >
            {' '}
            Add{' '}
          </Button>
        </div>
      );
    }
    if (timePeriodRangeDDVisible[0] && pageMode !== 'View')
      tmpList.push(
        <FaSelect
          key='init'
          options={tmpOptions}
          onClickOutside={(e) => {
            setTimePeriodRangeDDVisible([false, -1]);
          }}
          optionClickCallback={(value) => {
            setTimePeriosRangeProperties((p) => {
              if (p.length < 2) {
                return [...p, value];
              }
              const t = [...p];
              t[timePeriodRangeDDVisible[1]] = value;
              return t;
            });
            setTimePeriodRangeDDVisible([false, -1]);
          }}
          allowSearch
          allowSearchTextSelection
        />
      );
    return tmpList;
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

    // New Query
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

  const onEdit = (item) => {
    setPageMode('Edit');
    form.resetFields();
    setKPIType(
      item?.type_of_query === 1
        ? 'default'
        : item?.type_of_query === 2
          ? 'derived_kpi'
          : 'event_based'
    );
    if (item?.type_of_query === 1) {
      setKPICategory(item?.obj_ty);
      setKPIFn(item?.transformations?.agFn);
      setKPIPropertyDetails({
        name: item?.transformations?.agPr,
        data_type: item?.transformations?.agPrTy,
        value: matchEventName(item?.transformations?.agPr)
      });
      setGlobalFiltersOption(
        getStateFromKPIFilters(item?.transformations?.fil, userPropNames)
      );
    } else if (item?.type_of_query === 2) {
      setQueries(
        getStateFromCustomKPIqueryGroup(item?.transformations, kpiConfig)
      );
      setShowAsPercentage(item?.display_result_as !== '');
    } else {
      setEventFn(item?.transformations?.agFn);
      setEventPropertyDetails({
        name: item?.transformations?.agPr,
        data_type: item?.transformations?.agPrTy
      });
      setEventGlobalFiltersOption(
        getStateFromKPIFilters(item?.transformations?.fil, userPropNames)
      );
      setEventName(item?.transformations?.evNm);
    }
  };
  const onReset = () => {
    form.resetFields();
    setPageMode('Initial');
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
    setKPIDetails(false);
    setShowAsPercentage(false);
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
            ? getEventsWithPropertiesCustomKPI(filterValues?.globalFilters, '')
            : [],
          daFie: data.kpi_dateField
        }
      };
    } else if (selKPIType === 'derived_kpi') {
      const KPIquery = getCustomKPIQuery(
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
        display_result_as: showAsPercentage ? 'percentage_type' : '',
        transformations: {
          ...KPIquery
        }
      };
    } else if (selKPIType === 'event_based') {
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
            ? getEventsWithPropertiesCustomKPI(
                EventfilterValues?.globalFilters,
                ''
              )
            : [],
          daFie: '',
          evNm: selEventName,
          en: 'events_occurrence'
        }
      };
    } else if (selKPIType === 'time_period_based') {
      payload = {
        name: data?.name,
        description: data?.description,
        type_of_query: 1,
        obj_ty: data.kpi_category,
        metric_type: 'date_type_diff_metric',
        transformations: {
          agFn: KPIFn,
          agPr: timePeriodRangeProperties[0].value,
          agPr2: timePeriodRangeProperties[1].value,
          agPrTy: 'datetime',
          agPrTy2: 'datetime',
          fil: filterValues?.globalFilters
            ? getEventsWithPropertiesCustomKPI(filterValues?.globalFilters, '')
            : [],
          daFie: data.kpi_dateField
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
    setTableLoading(true);
    removeCustomKPI(activeProject.id, item?.id)
      .then(() => {
        setTableLoading(false);
        fetchSavedCustomKPI(activeProject.id);
        notification.success({
          message: 'KPI Removed',
          description: 'Custom KPI is removed successfully.'
        });
      })
      .catch((err) => {
        setTableLoading(false);
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
      item.entity,
      item.category
    ]);

    setFilterDDValues(DDvalues);
    setTimePeriosRangeProperties([]);
    // Below logic is added to pick the Time period based Properties
    // whenever user opens or creates a copy of it
    const pr1 = viewKPIDetails?.transformations?.agPr;
    const pr2 = viewKPIDetails?.transformations?.agPr2;
    if (pr1 && pr2) {
      const tmpRes = customKPIConfig.result?.filter(
        (e) => e.obj_ty === selKPICategory
      );
      if (tmpRes.length === 1) {
        const l = tmpRes[0]?.properties?.find((e) => e.name === pr1);
        const l2 = tmpRes[0]?.properties?.find((e) => e.name === pr2);
        if (l && l2) {
          setTimePeriosRangeProperties([
            { label: l.display_name, value: pr1 },
            { label: l2.display_name, value: pr2 }
          ]);
        }
      }
    }
  }, [selKPICategory, customKPIConfig, viewKPIDetails]);

  useEffect(() => {
    let DDCategory = {};
    for (const key of Object.keys(eventPropertiesV2)) {
      if (key === selEventName) {
        DDCategory = eventPropertiesV2[key];
      }
    }
    const DDvalues = [];
    Object.keys(DDCategory).forEach((group) => {
      DDCategory[group]?.forEach((item) => {
        DDvalues.push([item[0], item[1], item[2], 'event']);
      });
    });
    setEventFilterDDValues(DDvalues);
  }, [selEventName, eventPropertiesV2]);

  useEffect(() => {
    if (selEventName || viewKPIDetails?.transformations?.evNm) {
      getEventPropertiesV2(
        activeProject.id,
        selEventName || viewKPIDetails?.transformations?.evNm
      );
    }
  }, [selEventName, viewKPIDetails?.transformations?.evNm]);

  const onKPICategoryChange = (value) => {
    setKPICategory(value);
  };

  const onKPITypeChange = (value) => {
    setKPIType(value);
  };

  useEffect(() => {
    if (savedCustomKPI) {
      const savedArr = [];
      savedCustomKPI?.forEach((item, index) => {
        savedArr.push({
          key: index,
          name: item,
          desc: item.description,
          type:
            item.type_of_query === 1
              ? item.metric_type === 'date_type_diff_metric'
                ? 'Time Period Based'
                : 'Default'
              : item.type_of_query === 2
                ? 'Derived'
                : 'Event Based',
          actions: item
        });
      });
      setTableData(savedArr);
    }
  }, [savedCustomKPI]);

  // eslint-disable-next-line react/no-unstable-nested-components
  const TimePeriodBasedForm = (mode) => (
    <div>
      <Row className='m-0 mt-2'>
        <Col span={18}>
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
            initialValue={selKPICategory}
          >
            <Select
              className='fa-select w-full'
              size='large'
              onChange={(value) => onKPICategoryChange(value)}
              placeholder='KPI Category'
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
              defaultValue={selKPICategory}
              disabled={mode}
            >
              {customKPIConfig?.result?.map((item) => {
                if (item.type_of_query === 1) {
                  return (
                    <Option key={item.obj_ty} value={item.obj_ty}>
                      {_.startCase(item.obj_ty)}
                    </Option>
                  );
                }
              })}
            </Select>
          </Form.Item>
        </Col>
      </Row>
      {selKPICategory && (
        <Row className='mt-8'>
          <Col span={18}>
            <Text type='title' level={7} extraClass='m-0'>
              Select Functions
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
              initialValue={KPIFn}
            >
              <Select
                className='fa-select w-full'
                size='large'
                placeholder='Function'
                onChange={(value) => {
                  setKPIPropertyDetails({});
                  setKPIFn(value);
                }}
                showSearch
                filterOption={(input, option) =>
                  option.children.toLowerCase().indexOf(input.toLowerCase()) >=
                  0
                }
                defaultValue={KPIFn}
                disabled={mode}
              >
                {customKPIConfig?.result?.map((item) => {
                  if (item.obj_ty === selKPICategory) {
                    return item?.agFn.map(
                      (it) =>
                        it !== 'unique' && (
                          <Option key={it} value={it}>
                            {_.startCase(it)}
                          </Option>
                        )
                    );
                  }
                })}
              </Select>
            </Form.Item>
          </Col>
        </Row>
      )}
      {filterDDValues && (
        <Row className='my-8'>
          <Col span={18}>
            <div className='border-top--thin-2 border-bottom--thin-2 pt-5 pb-5'>
              <Text type='title' level={7} weight='bold' extraClass='m-0'>
                FILTER BY
              </Text>

              <GLobalFilter
                filters={filterValues?.globalFilters}
                setGlobalFilters={setGlobalFiltersOption}
                delFilter={false}
                viewMode={mode}
                onFiltersLoad={[
                  () => {
                    getUserPropertiesV2(activeProject.id, null);
                  }
                ]}
                selectedMainCategory={{
                  group: selKPICategory,
                  category: 'events'
                }}
                KPIConfigProps={filterDDValues}
                isSameKPIGrp // To avoid common properties in filter
              />
            </div>
          </Col>
        </Row>
      )}
      <div>
        <Text type='title' level={7} extraClass='m-0'>
          Time period between
        </Text>
        <div>
          <div className='my-2 border rounded-lg select-none'>
            <Row className='m-0 ml-4 my-2 mb-4'>
              <Col span={18}>
                <Form.Item name='query_type' className='m-0'>
                  {periodQueryList()}
                </Form.Item>
              </Col>
            </Row>
          </div>
        </div>
      </div>
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
            initialValue={
              pageMode === 'Edit'
                ? matchEventName(viewKPIDetails?.transformations?.daFie)
                : undefined
            }
          >
            <Select
              className='fa-select w-full'
              size='large'
              disabled={!selKPICategory || mode}
              placeholder='Date field'
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
              defaultValue={viewKPIDetails?.transformations?.daFie}
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
    </div>
  );
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

  function renderEventBasedKPIForm() {
    return (
      <div style={{ minHeight: '500px' }}>
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
            {pageMode === 'Edit' ? (
              <Form.Item name='event' className='m-0'>
                <EventQueryBlock
                  setEventName={setEventName}
                  selEventName={selEventName}
                />
              </Form.Item>
            ) : (
              <Form.Item name='event' className='m-0'>
                <EventQueryBlock setEventName={setEventName} />
              </Form.Item>
            )}
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
                initialValue={pageMode === 'Edit' ? EventFn : undefined}
              >
                <Select
                  className='fa-select w-full'
                  size='large'
                  placeholder='Function'
                  onChange={(value) => {
                    setEventPropertyDetails({});
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
                      return item?.agFn.map((it) => (
                        <Option key={it} value={it}>
                          {_.startCase(it)}
                        </Option>
                      ));
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
                  initialValue={
                    pageMode === 'Edit'
                      ? eventPropNames[viewKPIDetails?.transformations?.agPr]
                      : undefined
                  }
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
                    {Object.keys(eventPropertiesV2)?.map((category) => {
                      if (category === selEventName) {
                        const options = [];
                        const eventGroups = eventPropertiesV2[category];
                        Object.keys(eventGroups).forEach((group) => {
                          eventGroups[group]?.forEach((item) => {
                            if (item[2] === 'numerical') {
                              options.push(
                                <Option
                                  key={item[0]}
                                  value={item[0]}
                                  name={item[1]}
                                  data_type={item[2]}
                                  en='event'
                                >
                                  {item[0]}
                                </Option>
                              );
                            }
                          });
                        });
                        return options;
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
                  event={{ label: selEventName }}
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
              <GLobalFilter
                filters={getStateFromKPIFilters(
                  viewKPIDetails?.transformations?.fil,
                  userPropNames
                )}
                setGlobalFilters={setGlobalFiltersOption}
                delFilter={false}
                viewMode={pageMode === 'View'}
              />
            </Col>
          </Row>
        )}
      </div>
    );
  }

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          <div className='mb-10'>
            {pageMode === 'Initial' && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type='title'
                      level={3}
                      weight='bold'
                      extraClass='m-0'
                      id='fa-at-text--page-title'
                    >
                      Custom KPIs
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className='flex justify-end'>
                      <Button
                        onClick={() => {
                          form.resetFields();
                          setPageMode('Create');
                        }}
                        type='primary'
                        icon={<SVG name={'plus'} color={'white'} size={16} />}
                      >
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
                        Create personalized metrics tailored to your specific
                        objectives, whether it's conversion rates, engagement
                        metrics, or revenue targets.
                      </Text>
                      <Text
                        type='title'
                        level={7}
                        color='grey-2'
                        extraClass='m-0 mt-2'
                      >
                        Monitor progress, measure success, and gain actionable
                        insights to drive continuous improvement and achieve
                        your business milestones.{' '}
                        <a
                          href='https://help.factors.ai/en/articles/7284181-custom-kpis'
                          target='_blank'
                          rel='noreferrer'
                        >
                          Learn more
                        </a>
                      </Text>

                      {tableData.length > 0 ? (
                        <Table
                          className='fa-table--basic mt-8'
                          columns={columns}
                          dataSource={tableData}
                          pagination={false}
                          loading={tableLoading}
                          tableLayout='fixed'
                        />
                      ) : (
                        <EmptyScreen
                          loading={tableLoading}
                          title={`Define custom metrics to monitor conversion rates, track engagement metrics, and measure revenue targets tailored to your organization’s definitions.`}
                          learnMore={
                            'https://help.factors.ai/en/articles/7284181-custom-kpis'
                          }
                        />
                      )}
                    </div>
                  </Col>
                </Row>
              </>
            )}
            {(pageMode === 'Create' || pageMode === 'Edit') && (
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
                      initialValue={
                        pageMode === 'Edit'
                          ? `${viewKPIDetails?.name} - copy`
                          : ''
                      }
                    >
                      <Input
                        disabled={loading}
                        size='large'
                        className='fa-input w-full'
                        placeholder='Display Name'
                        ref={inputComponentRef}
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
                      initialValue={
                        pageMode === 'Edit' ? viewKPIDetails?.description : ''
                      }
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
                        defaultValue={selKPIType}
                      >
                        <Option value='default'>Default</Option>
                        <Option value='derived_kpi'>Derived KPI</Option>
                        <Option value='event_based'>Event Based</Option>
                        <Option value='time_period_based'>
                          Time Period Based
                        </Option>
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
                          initialValue={
                            pageMode === 'Edit' ? selKPICategory : undefined
                          }
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
                            {customKPIConfig?.result?.map((item) => {
                              if (item.type_of_query === 1) {
                                return (
                                  <Option key={item.obj_ty} value={item.obj_ty}>
                                    {_.startCase(item.obj_ty)}
                                  </Option>
                                );
                              }
                            })}
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
                            initialValue={
                              pageMode === 'Edit' ? KPIFn : undefined
                            }
                          >
                            <Select
                              className='fa-select w-full'
                              size='large'
                              placeholder='Function'
                              onChange={(value) => {
                                setKPIPropertyDetails({});
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
                                  return item?.agFn.map((it) => (
                                    <Option key={it} value={it}>
                                      {_.startCase(it)}
                                    </Option>
                                  ));
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
                            initialValue={
                              pageMode === 'Edit'
                                ? KPIPropertyDetails?.value
                                : undefined
                            }
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
                                    getUserPropertiesV2(activeProject.id, null);
                                  }
                                ]}
                                setGlobalFilters={setGlobalFiltersOption}
                                selectedMainCategory={{
                                  group: selKPICategory,
                                  category: 'events'
                                }}
                                KPIConfigProps={filterDDValues}
                                isSameKPIGrp // To avoid common properties in filter
                              />
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
                              initialValue={
                                pageMode === 'Edit'
                                  ? matchEventName(
                                      viewKPIDetails?.transformations?.daFie
                                    )
                                  : undefined
                              }
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
                    <div style={{ minHeight: '500px' }}>
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
                              color='grey-2'
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
                              initialValue={
                                pageMode === 'Edit'
                                  ? viewKPIDetails?.transformations?.for
                                  : undefined
                              }
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
                        <Row className='m-0 ml-4 -mt-2 mb-3'>
                          <Col span={14}>
                            <Form.Item name='showAspercentage'>
                              <Checkbox
                                defaultChecked={showAsPercentage}
                                onChange={(e) =>
                                  setShowAsPercentage(e.target.checked)
                                }
                              >
                                Show as percentage
                              </Checkbox>
                            </Form.Item>
                          </Col>
                        </Row>
                      </div>
                    </div>
                  </>
                ) : selKPIType === 'time_period_based' ? (
                  TimePeriodBasedForm(false)
                ) : (
                  [renderEventBasedKPIForm()]
                )}
              </Form>
            )}

            {pageMode === 'View' && (
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
                          setPageMode('Initial');
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
                        viewKPIDetails?.metric_type === 'date_type_diff_metric'
                          ? 'Time Period Based'
                          : viewKPIDetails?.type_of_query === 1
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
                {viewKPIDetails?.metric_type === 'date_type_diff_metric' ? (
                  <div>{TimePeriodBasedForm(true)}</div>
                ) : viewKPIDetails?.type_of_query === 1 ? (
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
                          <GLobalFilter
                            filters={getStateFromKPIFilters(
                              viewKPIDetails?.transformations?.fil,
                              userPropNames
                            )}
                            setGlobalFilters={setGlobalFiltersOption}
                            delFilter={false}
                            viewMode={pageMode === 'View'}
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

                                  <GLobalFilter
                                    filters={getStateFromKPIFilters(
                                      item.fil,
                                      userPropNames
                                    )}
                                    setGlobalFilters={setGlobalFiltersOption}
                                    delFilter={false}
                                    viewMode={pageMode === 'View'}
                                  />
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
                      <Row className='m-0 ml-4 -mt-2 mb-3'>
                        <Col span={14}>
                          <Checkbox
                            disabled
                            checked={viewKPIDetails?.display_result_as !== ''}
                          >
                            Show as percentage
                          </Checkbox>
                        </Col>
                      </Row>
                    </div>
                  </>
                ) : (
                  [renderEventBasedKPIView()]
                )}
                <Row className='border-top--thin-2 mt-6 pt-6'>
                  <Col span={12}>
                    {/* <a type={'link'} className={'mr-2'} onClick={() => createDuplicateAlert(viewAlertDetails)}>{'Create copy'}</a>
                <a type={'link'} color={'red'} onClick={() => confirmDeleteAlert(viewAlertDetails)}>{`Delete`}</a> */}

                    <Button
                      type='text'
                      color='blue'
                      onClick={() => {
                        setPageMode('Edit');
                        handleCopyCustomKPI(viewKPIDetails);
                      }}
                    >
                      <div className='flex items-center'>
                        <SVG
                          name='Pluscopy'
                          size={16}
                          color='grey'
                          extraClass='mr-1'
                        />
                        <Text type='title' level={7} extraClass='m-0'>
                          Create copy{' '}
                        </Text>
                      </div>
                    </Button>
                    <Button
                      type='text'
                      color='red'
                      onClick={() => {
                        setPageMode('Initial');
                        deleteKPI(viewKPIDetails);
                      }}
                    >
                      <div className='flex items-center'>
                        <SVG
                          name='Delete1'
                          size={16}
                          color='red'
                          extraClass='mr-1'
                        />
                        <Text
                          type='title'
                          level={7}
                          color='red'
                          extraClass='m-0'
                        >
                          Delete{' '}
                        </Text>
                      </div>
                    </Button>
                  </Col>
                </Row>
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
  eventPropertiesV2: state.coreQuery.eventPropertiesV2
});

export default connect(mapStateToProps, {
  fetchCustomKPIConfig,
  fetchSavedCustomKPI,
  addNewCustomKPI,
  removeCustomKPI,
  fetchKPIConfigWithoutDerivedKPI,
  fetchEventNames,
  getEventPropertiesV2
})(CustomKPI);
