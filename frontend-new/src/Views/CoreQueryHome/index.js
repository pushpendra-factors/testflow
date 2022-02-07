/* eslint-disable */
import React, { useCallback, useState, useEffect } from 'react';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog,
} from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Row, Col, Table, Avatar, Button, Dropdown, Menu, Tag } from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import { connect, useSelector, useDispatch } from 'react-redux';
import MomentTz from 'Components/MomentTz';
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
  getCampaignStateFromRequestQuery,
  getProfileQueryFromRequestQuery,
  getKPIStateFromRequestQuery,
  DefaultDateRangeFormat,
} from '../CoreQuery/utils';
import { INITIALIZE_GROUPBY } from '../../reducers/coreQuery/actions';
import ConfirmationModal from '../../components/ConfirmationModal';
import { deleteQuery } from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_TEMPLATE,
  TYPE_EVENTS_OCCURRENCE,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  reverse_user_types,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
  DefaultChartTypes,
  LOCAL_STORAGE_ITEMS,
  QUERY_TYPE_PROFILE,
  PREDEFINED_DATES,
} from '../../utils/constants';
import {
  SHOW_ANALYTICS_RESULT,
  INITIALIZE_MTA_STATE,
  INITIALIZE_CAMPAIGN_STATE,
} from '../../reducers/types';
import {
  SET_SHOW_CRITERIA,
  SET_PERFORMANCE_CRITERIA,
} from '../../reducers/analyticsQuery';
import { getDashboardDateRange } from '../Dashboard/utils';
import TemplatesModal from '../CoreQuery/Templates';
import { fetchWeeklyIngishts } from '../../reducers/insights';
import _ from 'lodash';
import { getQueryType } from '../../utils/dataFormatter';
import { fetchAgentInfo } from 'Reducers/agentActions';
import factorsai from 'factorsai';

// const whiteListedAccounts_KPI = [
//   'jitesh@factors.ai',
//   'kartheek@factors.ai',
//   'baliga@factors.ai',
//   'praveenr@factors.ai',
//   'sonali@factors.ai',
//   'solutions@factors.ai',
//   'praveen@factors.ai',
//   'ashwin@factors.ai',
// ];

const coreQueryoptions = [
  {
    title: 'KPIs',
    icon: 'KPI_cq',
    desc: 'Access your key marketing metrics',
  },
  {
    title: 'Funnels',
    icon: 'funnels_cq',
    desc: 'Find how users are navigating a defined path',
  },
  {
    title: 'Attribution',
    icon: 'attributions_cq',
    desc: 'Analyse Multi Touch Attributions',
  },
  {
    title: 'Profiles',
    icon: 'profiles_cq',
    desc: 'Explore all Profiles with filters and breakdowns',
  },
  {
    title: 'Events',
    icon: 'events_cq',
    desc: 'Create charts from events and related properties',
  },
  {
    title: 'Campaigns',
    icon: 'campaigns_cq',
    desc: 'Find the effect of your marketing campaigns',
  },
  {
    title: 'Templates',
    icon: 'templates_cq',
    desc: 'A list of advanced queries crafted by experts',
  },
];

const columns = [
  {
    title: 'Type',
    dataIndex: 'type',
    width: 60,
    key: 'type',
  },
  {
    title: 'Title of the Report',
    dataIndex: 'title',
    key: 'title',
    render: (text) => (
      <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
        {text}
      </Text>
    ),
  },
  {
    title: 'Created By',
    dataIndex: 'author',
    width: 240,
    key: 'author',
    render: (text) => (
      <div className='flex items-center'>
        <Avatar src='assets/avatar/avatar.png' size={24} className={'mr-2'} />
        &nbsp; {text}{' '}
      </div>
    ),
  },
  {
    title: 'Date',
    dataIndex: 'date',
    width: 240,
    key: 'date',
  },
];

function CoreQuery({
  setDrawerVisible,
  setQueryType,
  setQueries,
  setClickedSavedReport,
  setQueryOptions,
  location,
  setNavigatedFromDashboard,
  fetchWeeklyIngishts,
  activeProject,
  updateChartTypes,
  activeAgent,
  updateSavedQuerySettings,
  setProfileQueries,
  fetchAgentInfo,
  setAttributionMetrics,
}) {
  const queriesState = useSelector((state) => state.queries);
  const [deleteModal, showDeleteModal] = useState(false);
  const [activeRow, setActiveRow] = useState(null);
  const dispatch = useDispatch();
  const { attr_dimensions } = useSelector((state) => state.coreQuery);
  const { config: kpiConfig } = useSelector((state) => state.kpi);
  const { metadata } = useSelector((state) => state.insights);
  const [templatesModalVisible, setTemplatesModalVisible] = useState(false);

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
    };
    getData();
  }, [activeProject]);

  const getFormattedRow = (q) => {
    const requestQuery = q.query;
    const queryType = getQueryType(q.query);
    const queryTypeName = {
      events: 'events_cq',
      funnel: 'funnels_cq',
      channel_v1: 'campaigns_cq',
      attribution: 'attributions_cq',
      profiles: 'profiles_cq',
      kpi: 'KPI_cq',
    };
    let svgName = '';
    Object.entries(queryTypeName).forEach(([k, v]) => {
      if (queryType === k) {
        svgName = v;
      }
    });

    return {
      key: q.id,
      type: <SVG name={svgName} size={24} />,
      title: q.title,
      author: q.created_by_name,
      settings: q.settings,
      date: (
        <div className='flex justify-between items-center'>
          <div>{MomentTz(q.created_at).format('MMM DD, YYYY')}</div>
          <div>
            <Dropdown overlay={getMenu(q)} trigger={['hover']}>
              <Button type='text' icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        </div>
      ),
      query: requestQuery,
      actions: '',
    };
  };

  useEffect(() => {
    updateChartTypes(DefaultChartTypes);
  }, [updateChartTypes]);

  const confirmDelete = useCallback(() => {
    deleteQuery(dispatch, activeRow);
    setActiveRow(null);
    showDeleteModal(false);
  }, [activeRow]);

  const handleDelete = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setActiveRow(row);
    showDeleteModal(true);
  }, []);

  const handleViewResult = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    getWeeklyIngishts(row);
    setQueryToState(getFormattedRow(row));
  }, []);

  const updateEventFunnelsState = useCallback(
    (equivalentQuery, navigatedFromDashboard) => {
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: equivalentQuery.breakdown,
      });
      setQueries(equivalentQuery.events);
      setQueryOptions((currData) => {
        let newDateRange = {};
        if (navigatedFromDashboard) {
          newDateRange = { date_range: getDashboardDateRange() };
        }

        return {
          ...currData,
          session_analytics_seq: equivalentQuery.session_analytics_seq,
          groupBy: [
            ...equivalentQuery.breakdown.global,
            ...equivalentQuery.breakdown.event,
          ],
          globalFilters: equivalentQuery.globalFilters,
          ...newDateRange,
        };
      });
    },
    [dispatch]
  );

  const updateProfileQueryState = useCallback(
    (equivalentQuery) => {
      const dateRange = { ...equivalentQuery.dateRange };
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: equivalentQuery.breakdown,
      });
      setProfileQueries(equivalentQuery.events);
      setQueryOptions((currData) => {
        let queryOpts = {};
        queryOpts = {
          ...currData,
          groupBy: [
            ...equivalentQuery.breakdown.global,
            ...equivalentQuery.breakdown.event,
          ],
          globalFilters: equivalentQuery.globalFilters,
          group_analysis: equivalentQuery.groupAnalysis,
          date_range: { ...DefaultDateRangeFormat, ...dateRange },
        };
        return queryOpts;
      });
    },
    [dispatch]
  );

  const getWeeklyIngishts = (record) => {
    if (metadata?.QueryWiseResult) {
      const insightsItem = metadata?.QueryWiseResult[record.key];
      if (insightsItem) {
        dispatch({
          type: 'SET_ACTIVE_INSIGHT',
          payload: {
            id: record?.key,
            isDashboard: false,
            ...insightsItem,
          },
        });
      } else {
        dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: false });
      }
      if (insightsItem?.Enabled) {
        if (!_.isEmpty(insightsItem?.InsightsRange)) {
          let insightsLen =
            Object.keys(insightsItem?.InsightsRange)?.length || 0;
          fetchWeeklyIngishts(
            activeProject.id,
            record.key,
            Object.keys(insightsItem.InsightsRange)[insightsLen - 1],
            insightsItem.InsightsRange[
              Object.keys(insightsItem.InsightsRange)[insightsLen - 1]
            ][0],
            false
          ).catch((e) => {
            console.log('weekly-ingishts fetch error', e);
          });
        } else {
          dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: insightsItem });
        }
      } else {
        dispatch({ type: 'RESET_WEEKLY_INSIGHTS', payload: false });
      }
    }
  };

  const setQueryToState = useCallback(
    (record, navigatedFromDashboard) => {
      try {
        let equivalentQuery;
        if (record.query.query_group) {
          if (record.query.cl && record.query.cl === QUERY_TYPE_CAMPAIGN) {
            equivalentQuery = getCampaignStateFromRequestQuery(
              record.query.query_group[0]
            );
            let newDateRange;
            if (navigatedFromDashboard) {
              newDateRange = { camp_dateRange: getDashboardDateRange() };
            }
            const usefulQuery = { ...equivalentQuery, ...newDateRange };
            delete usefulQuery.queryType;
            dispatch({ type: INITIALIZE_CAMPAIGN_STATE, payload: usefulQuery });
          }
          // else if(queryType === QUERY_TYPE_KPI) {
          //  // Vishnu convert backend formatted query to  our local state query
          // }
          else {
            equivalentQuery = getStateQueryFromRequestQuery(
              record.query.query_group[0]
            );
            updateEventFunnelsState(equivalentQuery, navigatedFromDashboard);
            if (record.query.query_group.length === 1) {
              dispatch({
                type: SET_PERFORMANCE_CRITERIA,
                payload: reverse_user_types[record.query.query_group[0].ec],
              });
              dispatch({
                type: SET_SHOW_CRITERIA,
                payload: TOTAL_USERS_CRITERIA,
              });
            } else {
              dispatch({
                type: SET_PERFORMANCE_CRITERIA,
                payload: EACH_USER_TYPE,
              });
              if (record.query.query_group.length === 2) {
                dispatch({
                  type: SET_SHOW_CRITERIA,
                  payload:
                    record.query.query_group[0].ty === TYPE_EVENTS_OCCURRENCE
                      ? TOTAL_EVENTS_CRITERIA
                      : TOTAL_USERS_CRITERIA,
                });
              } else if (record.query.query_group.length === 3) {
                dispatch({
                  type: SET_SHOW_CRITERIA,
                  payload: ACTIVE_USERS_CRITERIA,
                });
              } else {
                dispatch({
                  type: SET_SHOW_CRITERIA,
                  payload: FREQUENCY_CRITERIA,
                });
              }
            }
          }
        } else if (record.query.cl && record.query.cl === QUERY_TYPE_KPI) {
          equivalentQuery = getKPIStateFromRequestQuery(
            record.query,
            kpiConfig
          );
          updateEventFunnelsState(equivalentQuery, navigatedFromDashboard);
        } else if (
          record.query.cl &&
          record.query.cl === QUERY_TYPE_ATTRIBUTION
        ) {
          equivalentQuery = getAttributionStateFromRequestQuery(
            record.query.query,
            attr_dimensions
          );
          let newDateRange = {};
          if (navigatedFromDashboard) {
            newDateRange = { attr_dateRange: getDashboardDateRange() };
          }
          const usefulQuery = { ...equivalentQuery, ...newDateRange };
          if (record.settings && record.settings.attributionMetrics) {
            setAttributionMetrics(
              JSON.parse(record.settings.attributionMetrics)
            );
          }
          delete usefulQuery.queryType;
          dispatch({ type: INITIALIZE_MTA_STATE, payload: usefulQuery });
        } else if (record.query.cl && record.query.cl === QUERY_TYPE_PROFILE) {
          equivalentQuery = getProfileQueryFromRequestQuery(record.query);
          updateProfileQueryState(equivalentQuery);
        } else {
          equivalentQuery = getStateQueryFromRequestQuery(record.query);
          updateEventFunnelsState(equivalentQuery, navigatedFromDashboard);
        }
        updateSavedQuerySettings(record.settings || {});
        setQueryType(equivalentQuery.queryType);
        setClickedSavedReport({
          queryType: equivalentQuery.queryType,
          queryName: record.title,
          settings: record.settings,
          query_id: record.key || record.id,
        });

        //Factors VIEW_QUERY tracking
        factorsai.track('VIEW_QUERY', {
          query_type: equivalentQuery?.queryType,
          saved_query_id: record?.key || record?.id,
          query_title: record?.title,
        });
      } catch (err) {
        console.log(err);
      }
    },
    [updateEventFunnelsState, attr_dimensions, kpiConfig]
  );

  const getMenu = (row) => {
    return (
      <Menu>
        <Menu.Item key='0'>
          <a onClick={handleViewResult.bind(this, row)} href='#!'>
            View Report
          </a>
        </Menu.Item>
        <Menu.Item key='1'>
          {/* <a onClick={(e) => e.stopPropagation()} href="#!">
            Copy Link
          </a>
        </Menu.Item>
        <Menu.Item key="2"> */}
          <a onClick={handleDelete.bind(this, row)} href='#!'>
            Delete Report
          </a>
        </Menu.Item>
      </Menu>
    );
  };

  useEffect(() => {
    if (location.state && location.state.global_search) {
      setQueryToState(
        location.state.query,
        location.state.navigatedFromDashboard
      );
      setNavigatedFromDashboard(location.state.navigatedFromDashboard);
      location.state = undefined;
      window.history.replaceState(null, '');
    } else {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    }
  }, [location.state, setQueryToState]);

  const data = queriesState.data
    .filter((q) => !(q.query && q.query.cl === QUERY_TYPE_WEB))
    .map((q) => {
      return getFormattedRow(q);
    });

  const setQueryTypeTab = (item) => {
    if (item.title === 'Templates') {
      setTemplatesModalVisible(true);
      // setQueryType(QUERY_TYPE_TEMPLATE);
    } else {
      setDrawerVisible(true);
    }

    if (item.title === 'Funnels') {
      setQueryType(QUERY_TYPE_FUNNEL);
      setQueries([]);
      setQueryOptions((currData) => {
        return {
          ...currData,
          globalFilters: [],
          date_range: { ...DefaultDateRangeFormat },
        };
      });
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: [],
        },
      });
    }

    if (item.title === 'Events') {
      setQueryType(QUERY_TYPE_EVENT);
      setQueries([]);
      setQueryOptions((currData) => {
        return {
          ...currData,
          globalFilters: [],
          date_range: { ...DefaultDateRangeFormat },
        };
      });
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: [],
        },
      });
    }

    if (item.title === 'Attribution') {
      setQueryType(QUERY_TYPE_ATTRIBUTION);
    }

    if (item.title === 'KPIs') {
      setQueryType(QUERY_TYPE_KPI);
      setQueries([]);
      setQueryOptions((currData) => {
        return {
          ...currData,
          globalFilters: [],
          date_range: { ...DefaultDateRangeFormat },
        };
      });
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: [],
        },
      });
    }

    if (item.title === 'Campaigns') {
      setQueryType(QUERY_TYPE_CAMPAIGN);
    }

    if (item.title === 'Profiles') {
      setQueryType(QUERY_TYPE_PROFILE);
      setProfileQueries([]);
      setQueryOptions((currData) => {
        return {
          ...currData,
          globalFilters: [],
          group_analysis: 'users',
          date_range: { ...DefaultDateRangeFormat },
        };
      });
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: [],
        },
      });
    }
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Analyse LP Error'}
            subtitle={
              'We are facing trouble loading Analyse landing page. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        <ConfirmationModal
          visible={deleteModal}
          confirmationText='Are you sure you want to delete this report?'
          onOk={confirmDelete}
          onCancel={showDeleteModal.bind(this, false)}
          title='Delete Report'
          okText='Confirm'
          cancelText='Cancel'
        />
        <TemplatesModal
          templatesModalVisible={templatesModalVisible}
          setTemplatesModalVisible={setTemplatesModalVisible}
        />
        <Header>
          <div className='w-full h-full py-4 flex flex-col justify-center items-center'>
            <SearchBar setQueryToState={setQueryToState} />
          </div>
        </Header>
        <div className={'fa-container mt-24'}>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={20}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
                Analyse
              </Text>
              <Text
                type={'title'}
                level={6}
                weight={'regular'}
                color={'grey'}
                extraClass={'m-0'}
              >
                Use these techniques to Analyse and get a deeper understanding
                of User Behaviors and Marketing Funnels
              </Text>
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify='center' className={'mt-10'}>
            <Col span={20}>
              <div className={'flex'}>
                {coreQueryoptions.map((item, index) => {
                  // if (
                  //   item.title === 'KPIs' &&
                  //   !whiteListedAccounts_KPI.includes(activeAccount)
                  // ) {
                  //   return null;
                  // }
                  return (
                    <div
                      key={index}
                      onClick={() => setQueryTypeTab(item)}
                      className={`fai--custom-card-new flex flex-col`}
                    >
                      <div
                        className={`fai--custom-card-new--top-section flex justify-center items-center`}
                      >
                        {item.title == 'KPIs' && (
                          <Tag
                            color='orange'
                            className={'fai--custom-card--badge'}
                          >
                            BETA
                          </Tag>
                        )}
                        <SVG name={item.icon} size={40} />
                      </div>

                      <div className='fai--custom-card-new--bottom-section'>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'bold'}
                          extraClass={'m-0'}
                        >
                          {' '}
                          {item.title}{' '}
                        </Text>
                        <Text
                          type={'title'}
                          level={7}
                          color={'grey'}
                          extraClass={'m-0 mt-1 fai--custom-card-new--desc'}
                        >
                          {' '}
                          {item.desc}{' '}
                        </Text>
                      </div>
                    </div>
                  );
                })}
              </div>
            </Col>
          </Row>

          <Row justify='center' className={'mt-8'}>
            <Col span={20}>
              <Row className={'flex justify-between items-center'}>
                <Col span={10}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={'m-0 mb-2'}
                  >
                    Saved Reports
                  </Text>
                </Col>
                {/* <Col span={5}>
                <div className={"flex flex-row justify-end items-end "}>
                  <Button
                    icon={<SVG name={"help"} size={12} color={"grey"} />}
                    type="text"
                  >
                    Learn More
                  </Button>
                </div>
              </Col> */}
              </Row>
            </Col>
          </Row>
          <Row justify='center' className={'mt-2 mb-20'}>
            <Col span={20}>
              <Table
                onRow={(record) => {
                  return {
                    onClick: (e) => {
                      getWeeklyIngishts(record);
                      setQueryToState(record);
                    },
                  };
                }}
                loading={queriesState.loading}
                className='fa-table--basic'
                columns={columns}
                dataSource={data}
                pagination={true}
                rowClassName='cursor-pointer'
              />
            </Col>
          </Row>
        </div>
      </ErrorBoundary>
    </>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  activeAgent: state.agent?.agent_details?.email,
});

export default connect(mapStateToProps, {
  fetchWeeklyIngishts,
  fetchAgentInfo,
})(CoreQuery);
