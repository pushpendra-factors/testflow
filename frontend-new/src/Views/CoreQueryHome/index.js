import React, { useCallback, useState, useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import {
  Row,
  Col,
  Table,
  Avatar,
  Button,
  Dropdown,
  Menu,
  Input,
  notification,
  message
} from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import { connect, useSelector, useDispatch } from 'react-redux';
import MomentTz from 'Components/MomentTz';
import _ from 'lodash';
import { fetchAgentInfo } from 'Reducers/agentActions';
import factorsai from 'factorsai';
import {
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
} from 'Reducers/global';
import useAutoFocus from 'hooks/useAutoFocus';
import { useHistory } from 'react-router-dom';
import moment from 'moment';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog
} from '../../components/factorsComponents';
// import SearchBar from '../../components/SearchBar';
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
  getProfileQueryFromRequestQuery,
  getKPIStateFromRequestQuery,
  DefaultDateRangeFormat
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
  TYPE_EVENTS_OCCURRENCE,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  REVERSE_USER_TYPES,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
  DefaultChartTypes,
  QUERY_TYPE_PROFILE
} from '../../utils/constants';
import {
  SHOW_ANALYTICS_RESULT,
  INITIALIZE_MTA_STATE,
  INITIALIZE_CAMPAIGN_STATE
} from '../../reducers/types';
import {
  SET_SHOW_CRITERIA,
  SET_PERFORMANCE_CRITERIA
} from '../../reducers/analyticsQuery';
import {
  getDashboardDateRange,
  getSavedAttributionMetrics
} from '../Dashboard/utils';
import TemplatesModal from '../CoreQuery/Templates';
import { fetchWeeklyIngishts } from '../../reducers/insights';
import { getQueryType } from '../../utils/dataFormatter';
import ShareToEmailModal from '../../components/ShareToEmailModal';
import ShareToSlackModal from '../../components/ShareToSlackModal';
import AppModal from '../../components/AppModal';
import styles from './index.module.scss';

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
    desc: 'Measure performance over time for your main objectives'
  },
  {
    title: 'Funnels',
    icon: 'funnels_cq',
    desc: 'Track how users navigate across their buying journey'
  },
  // {
  //   title: 'Attribution',
  //   icon: 'attributions_cq',
  //   desc: 'Identify the channels that contribute to conversion goals'
  // },
  {
    title: 'Profiles',
    icon: 'profiles_cq',
    desc: 'Slice and dice your visitors and users as you wish'
  },
  {
    title: 'Events',
    icon: 'events_cq',
    desc: 'Track and chart events and related properties'
  },
  // {
  //   title: 'Campaigns',
  //   icon: 'campaigns_cq',
  //   desc: 'Find the effect of your marketing campaigns',
  // },
  {
    title: 'Templates',
    icon: 'templates_cq',
    desc: 'Access pre-defined and elegant reports to quickly get started'
  }
];

const columns = [
  {
    title: 'Type',
    dataIndex: 'type',
    width: 60,
    key: 'type'
  },
  {
    title: 'Title of the Report',
    dataIndex: 'title',
    key: 'title',
    render: (text) => (
      <Text type='title' level={7} weight='bold' extraClass='m-0'>
        {text}
      </Text>
    )
  },
  {
    title: 'Created By',
    dataIndex: 'author',
    width: 240,
    key: 'author',
    render: (created_by_user) => (
      <div className='flex items-center'>
        <Avatar
          src={
            typeof created_by_user?.email === 'string' &&
            created_by_user?.email?.length !== 0 &&
            created_by_user.email.split('@')[1] === 'factors.ai'
              ? 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/factors-icon.svg'
              : created_by_user?.image
                ? created_by_user?.image
                : 'assets/avatar/avatar.png'
          }
          size={24}
          className='mr-2'
        />
        &nbsp; {created_by_user?.text}
      </div>
    )
  },
  {
    title: 'Date',
    dataIndex: 'date',
    width: 240,
    key: 'date'
  }
];

function CoreQuery({
  setDrawerVisible,
  setQueryType,
  setQueries,
  setClickedSavedReport,
  setQueryOptions,
  location,
  setNavigatedFromDashboard,
  setNavigatedFromAnalyse,
  fetchWeeklyIngishts,
  activeProject,
  updateChartTypes,
  updateSavedQuerySettings,
  setProfileQueries,
  fetchAgentInfo,
  setAttributionMetrics,
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  dateFromTo,
  updateCoreQueryReducer
}) {
  const activeProjectProfilePicture = useSelector(
    (state) => state.global.active_project.profile_picture
  );

  const queriesState = useSelector((state) => state.queries);
  const [deleteModal, showDeleteModal] = useState(false);
  const [activeRow, setActiveRow] = useState(null);
  const dispatch = useDispatch();
  const { attr_dimensions, content_groups } = useSelector(
    (state) => state.coreQuery
  );
  const { config: kpiConfig } = useSelector((state) => state.kpi);
  const { metadata } = useSelector((state) => state.insights);
  const [templatesModalVisible, setTemplatesModalVisible] = useState(false);
  const [showShareToEmailModal, setShowShareToEmailModal] = useState(false);
  const [showShareToSlackModal, setShowShareToSlackModal] = useState(false);
  const [tableData, setTableData] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [channelOpts, setChannelOpts] = useState([]);
  const [allChannels, setAllChannels] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedRow, setSelectedRow] = useState(null);
  const [overrideDate, setOverrideDate] = useState(false);

  const { slack } = useSelector((state) => state.global);
  const { projectSettingsV1 } = useSelector((state) => state.global);
  const { agent_details } = useSelector((state) => state.agent);
  const inputComponentRef = useAutoFocus(showSearch);
  const history = useHistory();

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
    };
    getData();
  }, [activeProject, fetchAgentInfo]);

  useEffect(() => {
    if (dateFromTo?.to === undefined || dateFromTo?.to === '') {
      setOverrideDate(false);
    } else {
      setOverrideDate(true);
    }
  }, [dateFromTo]);

  const pushDataToLocation = (data) => {
    const stateWithNoFunctions = {};
    for (const key of Object.keys(data)) {
      if (key === 'type' || key === 'date') {
        // empty
      } else {
        stateWithNoFunctions[key] = data[key];
      }
    }

    history.push({
      pathname: '/analyse',
      state: {
        coreQuery: true,
        navigatedFromAnalyse: stateWithNoFunctions
      }
    });
  };

  const getFormattedRow = (q) => {
    const requestQuery = q.query;
    const queryType = getQueryType(q.query);
    const queryTypeName = {
      events: 'events_cq',
      funnel: 'funnels_cq',
      channel_v1: 'campaigns_cq',
      attribution: 'attributions_cq',
      profiles: 'profiles_cq',
      kpi: 'KPI_cq'
    };
    let svgName = '';
    Object.entries(queryTypeName).forEach(([k, v]) => {
      if (queryType === k) {
        svgName = v;
      }
    });

    return {
      key: q.id,
      id_text: q.id_text,
      type: <SVG name={svgName} size={24} color='blue' />,
      title: q.title,
      author: {
        image: activeProjectProfilePicture,
        text: q.created_by_name,
        email: q.created_by_email
      },
      settings: q.settings,
      date: (
        <div className='flex justify-between items-center'>
          <div>{MomentTz(q.created_at).format('MMM DD, YYYY')}</div>
          <div>
            <Dropdown overlay={getMenu(q)} placement='bottomRight'>
              <Button type='text' icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        </div>
      ),
      query: requestQuery,
      actions: ''
    };
  };

  const confirmDelete = useCallback(() => {
    const queryDetails = {
      ...activeRow,
      project_id: activeProject?.id
    };
    dispatch(deleteQuery(queryDetails));
    setActiveRow(null);
    showDeleteModal(false);
  }, [activeProject?.id, activeRow, dispatch]);

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
    pushDataToLocation(getFormattedRow(row));
  }, []);

  const showEmailModal = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setShowShareToEmailModal(true);
    setSelectedRow(row);
  }, []);

  const showSlackModal = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setShowShareToSlackModal(true);
    setSelectedRow(row);
  }, []);

  const updateEventFunnelsState = useCallback(
    (equivalentQuery, navigatedFromDashboard) => {
      const savedDateRange = { ...equivalentQuery.dateRange };
      const newDateRange = getDashboardDateRange();
      const dashboardDateRange = {
        ...newDateRange,
        frequency:
          moment(newDateRange.to).diff(newDateRange.from, 'days') <= 1
            ? 'hour'
            : equivalentQuery.dateRange.frequency
      };
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: equivalentQuery.breakdown
      });
      setQueries(equivalentQuery.events);
      setQueryOptions((currData) => {
        let queryDateRange = {};
        if (navigatedFromDashboard) {
          queryDateRange = { date_range: dashboardDateRange };
        } else queryDateRange = { date_range: savedDateRange };

        let queryOpts = {};
        queryOpts = {
          ...currData,
          session_analytics_seq: equivalentQuery.session_analytics_seq,
          groupBy: [
            ...equivalentQuery.breakdown.global,
            ...equivalentQuery.breakdown.event
          ],
          globalFilters: equivalentQuery.globalFilters,
          group_analysis: equivalentQuery.groupAnalysis,
          ...queryDateRange,
          events_condition: equivalentQuery.eventsCondition
        };
        return queryOpts;
      });
    },
    [dispatch, setQueries, setQueryOptions]
  );

  const updateProfileQueryState = useCallback(
    (equivalentQuery) => {
      const dateRange = { ...equivalentQuery.dateRange };
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: equivalentQuery.breakdown
      });
      setProfileQueries(equivalentQuery.events);
      setQueryOptions((currData) => {
        let queryOpts = {};
        queryOpts = {
          ...currData,
          groupBy: [
            ...equivalentQuery.breakdown.global,
            ...equivalentQuery.breakdown.event
          ],
          globalFilters: equivalentQuery.globalFilters,
          group_analysis: equivalentQuery.groupAnalysis,
          date_range: { ...DefaultDateRangeFormat, ...dateRange }
        };
        return queryOpts;
      });
    },
    [dispatch, setProfileQueries, setQueryOptions]
  );

  const updateKPIQueryState = useCallback(
    (equivalentQuery, navigatedFromDashboard) => {
      const savedDateRange = { ...equivalentQuery.dateRange };
      const newDateRange = getDashboardDateRange();
      const dashboardDateRange = {
        ...newDateRange,
        frequency: equivalentQuery.dateRange.frequency
      };
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: equivalentQuery.breakdown
      });
      setQueries(equivalentQuery.events);
      setQueryOptions((currData) => {
        let queryDateRange = {};
        if (navigatedFromDashboard) {
          queryDateRange = { date_range: dashboardDateRange };
        } else queryDateRange = { date_range: savedDateRange };

        let queryOpts = {};
        queryOpts = {
          ...currData,
          session_analytics_seq: equivalentQuery.session_analytics_seq,
          groupBy: [
            ...equivalentQuery.breakdown.global,
            ...equivalentQuery.breakdown.event
          ],
          globalFilters: equivalentQuery.globalFilters,
          ...queryDateRange
        };
        return queryOpts;
      });
    },
    [dispatch, setQueries, setQueryOptions]
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
            ...insightsItem
          }
        });
      } else {
        dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: false });
      }
      if (insightsItem?.Enabled) {
        if (!_.isEmpty(insightsItem?.InsightsRange)) {
          const insightsLen =
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
        // if(record?.type?.props?.name === 'events_cq' ) {
        //   history.push('/analyse/event/' + record.id_text);
        //   return null;
        // }
        // else if (record?.type?.props?.name === 'funnels_cq') {
        //   history.push('/analyse/funnel/' + record.id_text);
        //   return null;

        // }
        // else if (record?.type?.props?.name === 'attributions_cq') {
        //   window.location.replace("/analyse/attributions/" + record.id_text);
        //   return null;
        // }
        let equivalentQuery;
        if (record.query.query_group) {
          equivalentQuery = getStateQueryFromRequestQuery(
            record.query.query_group[0]
          );
          updateEventFunnelsState(equivalentQuery, navigatedFromDashboard);
          if (record.query.query_group.length === 1) {
            dispatch({
              type: SET_PERFORMANCE_CRITERIA,
              payload: REVERSE_USER_TYPES[record.query.query_group[0].ec]
            });
            dispatch({
              type: SET_SHOW_CRITERIA,
              payload: TOTAL_USERS_CRITERIA
            });
          } else {
            dispatch({
              type: SET_PERFORMANCE_CRITERIA,
              payload: EACH_USER_TYPE
            });
            if (record.query.query_group.length === 2) {
              dispatch({
                type: SET_SHOW_CRITERIA,
                payload:
                  record.query.query_group[0].ty === TYPE_EVENTS_OCCURRENCE
                    ? TOTAL_EVENTS_CRITERIA
                    : TOTAL_USERS_CRITERIA
              });
            } else if (record.query.query_group.length === 3) {
              dispatch({
                type: SET_SHOW_CRITERIA,
                payload: ACTIVE_USERS_CRITERIA
              });
            } else {
              dispatch({
                type: SET_SHOW_CRITERIA,
                payload: FREQUENCY_CRITERIA
              });
            }
          }
        } else if (record.query.cl && record.query.cl === QUERY_TYPE_KPI) {
          equivalentQuery = getKPIStateFromRequestQuery(
            record.query,
            kpiConfig
          );
          updateKPIQueryState(equivalentQuery, navigatedFromDashboard);
        } else if (
          record.query.cl &&
          record.query.cl === QUERY_TYPE_ATTRIBUTION
        ) {
          equivalentQuery = getAttributionStateFromRequestQuery(
            record.query.query,
            attr_dimensions,
            content_groups,
            kpiConfig
          );
          let newDateRange = {};
          if (navigatedFromDashboard) {
            newDateRange = { attr_dateRange: getDashboardDateRange() };
          }
          const usefulQuery = { ...equivalentQuery, ...newDateRange };
          if (record.settings && record.settings.attributionMetrics) {
            setAttributionMetrics(
              getSavedAttributionMetrics(
                JSON.parse(record.settings.attributionMetrics)
              )
            );
          }
          if (record.settings && record.settings.tableFilters) {
            updateCoreQueryReducer({
              attributionTableFilters: JSON.parse(record.settings.tableFilters)
            });
          }
          delete usefulQuery.queryType;
          dispatch({ type: INITIALIZE_MTA_STATE, payload: usefulQuery });
          setQueryOptions((currData) => ({
            ...currData,
            group_analysis: record.query.query.analyze_type
          }));
        } else if (record.query.cl && record.query.cl === QUERY_TYPE_PROFILE) {
          equivalentQuery = getProfileQueryFromRequestQuery(record.query);
          updateProfileQueryState(equivalentQuery);
        } else {
          equivalentQuery = getStateQueryFromRequestQuery(record.query);
          updateEventFunnelsState(equivalentQuery, navigatedFromDashboard);
          updateCoreQueryReducer({
            funnelConversionDurationNumber:
              equivalentQuery.funnelConversionDurationNumber,
            funnelConversionDurationUnit:
              equivalentQuery.funnelConversionDurationUnit
          });
        }
        updateSavedQuerySettings(record.settings || {});
        setQueryType(equivalentQuery.queryType);
        setClickedSavedReport({
          queryType: equivalentQuery.queryType,
          queryName: record.title,
          settings: record.settings,
          query_id: record.key || record.id
        });

        // Factors VIEW_QUERY tracking
        factorsai.track('VIEW_QUERY', {
          email_id: agent_details?.email,
          query_type: equivalentQuery?.queryType,
          saved_query_id: record?.key || record?.id,
          query_title: record?.title,
          project_id: activeProject?.id,
          project_name: activeProject?.name
        });
      } catch (err) {
        console.log(err);
      }
    },
    [
      updateSavedQuerySettings,
      setQueryType,
      setClickedSavedReport,
      agent_details?.email,
      activeProject?.id,
      activeProject?.name,
      dispatch,
      updateEventFunnelsState,
      kpiConfig,
      updateKPIQueryState,
      attr_dimensions,
      content_groups,
      setQueryOptions,
      setAttributionMetrics,
      updateCoreQueryReducer,
      updateProfileQueryState
    ]
  );

  const getMenu = (row) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item key='0'>
        <a onClick={handleViewResult.bind(this, row)} href='#!'>
          <SVG name='eye' size={18} color='grey' extraClass='inline mr-2' />
          View Report
        </a>
      </Menu.Item>
      {getQueryType(row.query) === QUERY_TYPE_KPI ||
      getQueryType(row.query) === QUERY_TYPE_EVENT ? (
        <Menu.Item key='1'>
          <a onClick={showEmailModal.bind(this, row)} href='#!'>
            <SVG
              name='envelope'
              size={18}
              color='grey'
              extraClass='inline mr-2'
            />
            Email this report
          </a>
        </Menu.Item>
      ) : null}
      {getQueryType(row.query) === QUERY_TYPE_KPI ||
      getQueryType(row.query) === QUERY_TYPE_EVENT ? (
        <Menu.Item key='2'>
          <a onClick={showSlackModal.bind(this, row)} href='#!'>
            <SVG
              name='SlackStroke'
              size={18}
              color='grey'
              extraClass='inline mr-2'
            />
            Share to Slack
          </a>
        </Menu.Item>
      ) : null}
      <Menu.Item key='3'>
        {/* <a onClick={(e) => e.stopPropagation()} href="#!">
            Copy Link
          </a>
        </Menu.Item>
        <Menu.Item key="2"> */}
        <a onClick={handleDelete.bind(this, row)} href='#!'>
          <SVG name='trash' size={18} color='grey' extraClass='inline mr-2' />
          Delete Report
        </a>
      </Menu.Item>
    </Menu>
  );

  useEffect(() => {
    if (location.state && location.state.global_search) {
      setQueryToState(
        location.state.query,
        location.state.navigatedFromDashboard
      );
      setNavigatedFromDashboard(location.state.navigatedFromDashboard);
      location.state = undefined;
      window.history.replaceState(null, '');
    } else if (location.state && location.state.coreQuery) {
      setNavigatedFromAnalyse(location.state.navigatedFromAnalyse);
      location.state = undefined;
      window.history.replaceState(null, '');
    } else if (location.state && location.state.navigatedFromAIChartPrompt) {
      setQueryToState(location.state.query, false);
    } else {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
      history.push('/reports');
    }
  }, [
    dispatch,
    location,
    setNavigatedFromDashboard,
    setNavigatedFromAnalyse,
    setQueryToState
  ]);

  const data = queriesState.data
    .filter((q) => !(q.query && q.query.cl === QUERY_TYPE_WEB))
    .map((q) => getFormattedRow(q));

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
      setQueryOptions((currData) => ({
        ...currData,
        globalFilters: [],
        group_analysis: 'users',
        date_range: { ...DefaultDateRangeFormat }
      }));
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: []
        }
      });
    }

    if (item.title === 'Events') {
      setQueryType(QUERY_TYPE_EVENT);
      setQueries([]);
      setQueryOptions((currData) => ({
        ...currData,
        globalFilters: [],
        group_analysis: 'users',
        date_range: { ...DefaultDateRangeFormat }
      }));
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: []
        }
      });
    }

    if (item.title === 'Attribution') {
      setQueryType(QUERY_TYPE_ATTRIBUTION);
    }

    if (item.title === 'KPIs') {
      setQueryType(QUERY_TYPE_KPI);
      setQueries([]);
      setQueryOptions((currData) => ({
        ...currData,
        globalFilters: [],
        date_range: { ...DefaultDateRangeFormat }
      }));
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: []
        }
      });
    }

    if (item.title === 'Campaigns') {
      setQueryType(QUERY_TYPE_CAMPAIGN);
    }

    if (item.title === 'Profiles') {
      setQueryType(QUERY_TYPE_PROFILE);
      setProfileQueries([]);
      setQueryOptions((currData) => ({
        ...currData,
        globalFilters: [],
        group_analysis: 'users',
        date_range: { ...DefaultDateRangeFormat }
      }));
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: []
        }
      });
    }
  };

  const searchReport = (e) => {
    const term = e.target.value;
    const searchResults = data.filter((item) =>
      item?.title?.toLowerCase().includes(searchTerm.toLowerCase())
    );
    setSearchTerm(term);
    setTableData(searchResults);
  };

  const onConnectSlack = () => {
    enableSlackIntegration(activeProject.id)
      .then((r) => {
        if (r.status === 200) {
          window.open(r.data.redirectURL, '_blank');
          setShowShareToSlackModal(false);
        }
        if (r.status >= 400) {
          message.error('Error fetching Slack redirect url');
        }
      })
      .catch((err) => {
        console.log('Slack error-->', err);
      });
  };

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettingsV1?.int_slack) {
      fetchSlackChannels(activeProject.id);
    }
  }, [
    activeProject,
    fetchProjectSettingsV1,
    fetchSlackChannels,
    projectSettingsV1?.int_slack,
    showShareToSlackModal
  ]);

  useEffect(() => {
    if (slack?.length > 0) {
      const tempArr = [];
      const allArr = [];
      for (let i = 0; i < slack.length; i++) {
        tempArr.push({ label: `#${slack[i].name}`, value: slack[i].id });
        allArr.push({
          name: slack[i].name,
          id: slack[i].id,
          is_private: slack[i].is_private
        });
      }
      setChannelOpts(tempArr);
      setAllChannels(allArr);
    }
  }, [activeProject, agent_details, slack]);

  const handleEmailClick = ({ data, frequency, onSuccess }) => {
    setLoading(true);

    let emails = [];
    if (data?.emails) {
      emails = data.emails.map((item) => item.email);
    }
    if (data.email) {
      emails.push(data.email);
    }

    const payload = {
      alert_name: selectedRow?.title || data?.subject,
      alert_type: 3,
      // "query_id": selectedRow?.key || selectedRow?.id,
      alert_description: {
        message: data?.message,
        date_range: frequency === 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: true,
        slack_enabled: false,
        emails,
        slack_channels_and_user_groups: {}
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id,
        dateFromTo,
        overrideDate
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected emails',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id
      )
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    setLoading(false);
    onSuccess();
  };

  const handleSlackClick = ({ data, frequency, onSuccess }) => {
    setLoading(true);

    let slackChannels = {};
    const selected = allChannels.filter((c) => c.id === data.channel);
    const map = new Map();
    map.set(agent_details.uuid, selected);
    for (const [key, value] of map) {
      slackChannels = { ...slackChannels, [key]: value };
    }

    const payload = {
      alert_name: selectedRow?.title || data?.subject,
      alert_type: 3,
      // "query_id": selectedRow?.key || selectedRow?.id,
      alert_description: {
        message: data?.message,
        date_range: frequency === 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: false,
        slack_enabled: true,
        emails: [],
        slack_channels_and_user_groups: slackChannels
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id,
        dateFromTo,
        overrideDate
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected Slack channel',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id
      )
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    setLoading(false);
    onSuccess();
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Analyse LP Error'
          subtitle='We are facing trouble loading Analyse landing page. Drop us a message on the in-app chat.'
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
      {/* <FaHeader>
          <SearchBar setQueryToState={setQueryToState} />
        </FaHeader> */}
      <div>
        <div className='fa-container'>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={20}>
              <Row gutter={[24, 24]}>
                <Col span={24}>
                  <div className='flex space-between w-full items-center'>
                    <div className='flex flex-col w-full'>
                      <Text
                        type='title'
                        level={3}
                        weight='bold'
                        extraClass='m-0'
                      >
                        Analyse
                      </Text>
                      <Text
                        type='title'
                        level={6}
                        weight='regular'
                        color='grey'
                        extraClass='m-0'
                      >
                        Here's where all the action happens. Use these modules
                        to get a deeper understanding of your marketing and
                        revenue activities. <a href='#!'>Learn more</a>
                      </Text>
                    </div>
                    {/* <div className='flex justify-end'>
                        <Button
                          type='link'
                          icon={
                            <SVG name={`Handshake`} size={24} color={'blue'} />
                          }
                          onClick={() => {
                            userflow.start(USERFLOW_CONFIG_ID?.AnalysePage);
                          }}
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center'
                          }}
                        >
                          Walk me through
                        </Button>
                      </div> */}
                  </div>
                </Col>
                <Col span={24}>
                  <div className='flex justify-between mt-4'>
                    {coreQueryoptions.map((item, index) => (
                      // if (
                      //   item.title === 'KPIs' &&
                      //   !whiteListedAccounts_KPI.includes(activeAccount)
                      // ) {
                      //   return null;
                      // }
                      <div
                        key={index}
                        onClick={() => setQueryTypeTab(item)}
                        className='fai--custom-card-new flex flex-col'
                      >
                        <div className='fai--custom-card-new--top-section flex justify-center items-center'>
                          {/* {item.title == 'KPIs' && (
                          <Tag
                            color='orange'
                            className={'fai--custom-card--badge'}
                          >
                            BETA
                          </Tag>
                        )} */}
                          <SVG name={item.icon} size={40} color='blue' />
                        </div>

                        <div className='fai--custom-card-new--bottom-section'>
                          <Text
                            type='title'
                            level={7}
                            weight='bold'
                            extraClass='m-0'
                          >
                            {item.title}
                          </Text>
                          <Text
                            type='title'
                            level={7}
                            color='grey'
                            extraClass='m-0 mt-1 fai--custom-card-new--desc'
                          >
                            {item.desc}
                          </Text>
                        </div>
                      </div>
                    ))}
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={24}>
                  <div className='flex items-center space-between w-full  mt-8 mb-2'>
                    <div className='flex items-center w-full'>
                      <Text
                        type='title'
                        level={6}
                        weight='bold'
                        extraClass='m-0'
                      >
                        Saved Reports
                      </Text>
                    </div>

                    <div className='flex items-center justify-between'>
                      {showSearch ? (
                        <Input
                          onChange={searchReport}
                          className=''
                          placeholder='Search reports'
                          style={{ width: '220px', 'border-radius': '5px' }}
                          prefix={<SVG name='search' size={16} color='grey' />}
                          ref={inputComponentRef}
                        />
                      ) : null}
                      <Button
                        type='text'
                        ghost
                        className='p-2 bg-white'
                        onClick={() => {
                          setShowSearch(!showSearch);
                          if (showSearch) {
                            setSearchTerm('');
                          }
                        }}
                      >
                        <SVG
                          name={!showSearch ? 'search' : 'close'}
                          size={20}
                          color='grey'
                        />
                      </Button>
                    </div>
                  </div>
                </Col>
              </Row>
              <Row className='mt-2 mb-20'>
                <Col span={24}>
                  <Table
                    onRow={(record) => ({
                      onClick: () => {
                        getWeeklyIngishts(record);
                        setQueryToState(record);
                        pushDataToLocation(record);
                      }
                    })}
                    loading={queriesState.loading}
                    className='fa-table--basic'
                    columns={columns}
                    dataSource={searchTerm ? tableData : data}
                    pagination
                    rowClassName='cursor-pointer'
                  />
                </Col>
              </Row>
            </Col>
          </Row>
        </div>
      </div>

      <ShareToEmailModal
        visible={showShareToEmailModal}
        onSubmit={handleEmailClick}
        isLoading={loading}
        setShowShareToEmailModal={setShowShareToEmailModal}
        queryTitle={selectedRow?.title}
      />

      {projectSettingsV1?.int_slack ? (
        <ShareToSlackModal
          visible={showShareToSlackModal}
          onSubmit={handleSlackClick}
          channelOpts={channelOpts}
          isLoading={loading}
          setShowShareToSlackModal={setShowShareToSlackModal}
          queryTitle={selectedRow?.title}
        />
      ) : (
        <AppModal
          title={null}
          visible={showShareToSlackModal}
          footer={null}
          centered
          mask
          maskClosable={false}
          maskStyle={{ backgroundColor: 'rgb(0 0 0 / 70%)' }}
          closable
          isLoading={loading}
          onCancel={() => setShowShareToSlackModal(false)}
          className='fa-modal--regular'
          width='470px'
        >
          <div className='m-0 mb-2'>
            <Row className='m-0'>
              <Col>
                <SVG name='Slack' size={25} extraClass='inline mr-2 -mt-2' />
                <Text
                  type='title'
                  level={5}
                  weight='bold'
                  extraClass='inline m-0'
                >
                  Slack Integration
                </Text>
              </Col>
            </Row>
            <Row className='m-0 mt-4'>
              <Col>
                <Text
                  type='title'
                  level={6}
                  color='grey-2'
                  weight='regular'
                  extraClass='m-0'
                >
                  Slack is not integrated, Do you want to integrate with your
                  Slack account now?
                </Text>
              </Col>
            </Row>
            <Col>
              <Row justify='end' className='w-full mb-1 mt-4'>
                <Col className='mr-2'>
                  <Button
                    type='default'
                    onClick={() => setShowShareToSlackModal(false)}
                  >
                    Cancel
                  </Button>
                </Col>
                <Col className='mr-2'>
                  <Button type='primary' onClick={onConnectSlack}>
                    Connect to Slack
                  </Button>
                </Col>
              </Row>
            </Col>
          </div>
        </AppModal>
      )}
    </ErrorBoundary>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  activeAgent: state.agent?.agent_details?.email
});

export default connect(mapStateToProps, {
  fetchWeeklyIngishts,
  fetchAgentInfo,
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
})(CoreQuery);
