import React, {
  useState,
  useCallback,
  useEffect,
  useMemo,
  useReducer
} from 'react';
import get from 'lodash/get';
import { bindActionCreators } from 'redux';
import { useParams } from 'react-router';
import { connect, useSelector, useDispatch } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import { Drawer, Button, Modal, Row, Col, Spin } from 'antd';
import MomentTz from 'Components/MomentTz';
import factorsai from 'factorsai';

import { EMPTY_ARRAY } from 'Utils/global';
import KPIComposer from 'Components/KPIComposer';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';
import QueryComposer from '../../components/QueryComposer';
import AttrQueryComposer from '../../components/AttrQueryComposer';
import CampQueryComposer from '../../components/CampQueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog
} from '../../components/factorsComponents';
import {
  deleteGroupByForEvent,
  getCampaignConfigData
} from '../../reducers/coreQuery/middleware';
import {
  calculateFrequencyData,
  calculateActiveUsersData,
  formatApiData,
  getQuery,
  initialState,
  getFunnelQuery,
  getKPIQuery,
  DefaultDateRangeFormat,
  getAttributionQuery,
  getCampaignsQuery,
  isComparisonEnabled,
  getProfileQuery,
  getStateQueryFromRequestQuery
} from './utils';
import {
  getEventsData,
  getEventsDataFromId,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
  getProfileData,
  getKPIData
} from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_ATTRIBUTION,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  EACH_USER_TYPE,
  REPORT_SECTION,
  INITIAL_SESSION_ANALYTICS_SEQ,
  ATTRIBUTION_METRICS,
  QUERY_TYPE_PROFILE,
  apiChartAnnotations,
  presentationObj,
  DefaultChartTypes,
  CHART_TYPE_TABLE,
  QUERY_OPTIONS_DEFAULT_VALUE
} from '../../utils/constants';
import { SHOW_ANALYTICS_RESULT } from '../../reducers/types';
import AnalysisResultsPage from './AnalysisResultsPage';
import AnalysisHeader from './AnalysisResultsPage/AnalysisHeader';
import {
  SET_CAMP_DATE_RANGE,
  SET_ATTR_DATE_RANGE,
  INITIALIZE_GROUPBY
} from '../../reducers/coreQuery/actions';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import CoreQueryReducer from './CoreQueryReducer';
import {
  CORE_QUERY_INITIAL_STATE,
  SET_COMPARISON_ENABLED,
  COMPARISON_DATA_LOADING,
  COMPARISON_DATA_FETCHED,
  RESET_COMPARISON_DATA,
  SET_COMPARISON_SUPPORTED,
  SET_COMPARE_DURATION,
  SET_NAVIGATED_FROM_DASHBOARD,
  UPDATE_CHART_TYPES,
  SET_SAVED_QUERY_SETTINGS,
  UPDATE_PIVOT_CONFIG,
  DEFAULT_PIVOT_CONFIG
} from './constants';
import {
  getValidGranularityOptions,
  shouldDataFetch
} from '../../utils/dataFormatter';
import ProfileComposer from '../../components/ProfileComposer';
import {
  IconAndTextSwitchQueryType,
  getSavedPivotConfig
} from './coreQuery.helpers';
import { getChartChangedKey } from './AnalysisResultsPage/analysisResultsPage.helpers';
import { EMPTY_OBJECT } from '../../utils/global';
import moment from 'moment';
import {
  fetchDemoProject,
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
} from 'Reducers/global';
import { meetLink } from '../../utils/hubspot';
import NewProject from '../Settings/SetupAssist/Modals/NewProject';
import userflow from 'userflow.js';
import { useHistory } from 'react-router-dom';
import AnalyseBeforeIntegration from './AnalyseBeforeIntegration';

function CoreQuery({
  activeProject,
  deleteGroupByForEvent,
  location,
  getCampaignConfigData,
  KPI_config,
  fetchDemoProject,
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  existingQueries
}) {
  const { query_id, query_type } = useParams();

  const queriesState = useSelector((state) => state.queries);
  const savedQueries = useSelector((state) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );

  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);

  const [coreQueryState, localDispatch] = useReducer(
    CoreQueryReducer,
    CORE_QUERY_INITIAL_STATE
  );
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeKey, setActiveKey] = useState('0');
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [clickedSavedReport, setClickedSavedReport] = useState(false);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownType, setBreakdownType] = useState(EACH_USER_TYPE);
  const [queriesA, setQueries] = useState([]);
  const [selectedMainCategory, setSelectedMainCategory] = useState(false);
  const [KPIConfigProps, setKPIConfigProps] = useState([]);
  const [loading, setLoading] = useState(true);

  const [demoProjectId, setDemoProjectId] = useState(null);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const { projects } = useSelector((state) => state.global);
  const currentAgent = useSelector((state) => state.agent.agent_details);

  const history = useHistory();

  const [profileQueries, setProfileQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });
  const [attributionsState, setAttributionsState] = useState({
    eventGoal: {},
    touchpoint: '',
    models: [],
    tacticOfferType: '',
    linkedEvents: [],
    date_range: {},
    attr_dimensions: [],
    content_groups: []
  });

  const [campaignState, setCampaignState] = useState({
    channel: '',
    select_metrics: [],
    filters: [],
    group_by: [],
    date_range: {}
  });

  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS
  ]);

  const dispatch = useDispatch();
  const {
    groupBy,
    eventGoal,
    touchpoint,
    touchpoint_filters,
    tacticOfferType,
    attr_query_type,
    models,
    window,
    linkedEvents,
    camp_channels,
    camp_measures,
    camp_filters,
    camp_groupBy,
    camp_dateRange,
    attr_dateRange,
    eventNames,
    attr_dimensions,
    attrQueries,
    content_groups
  } = useSelector((state) => state.coreQuery);

  const [activeTab, setActiveTab] = useState(1);

  const [queryOpen, setQueryOpen] = useState(false);

  const [dataResp, setDataResp] = useState(null);

  const [dateFromTo, setDateFromTo] = useState({from: '', to: ''});

  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state) => state.analyticsQuery);

  const dateRange = queryOptions.date_range;
  const { session_analytics_seq } = queryOptions;
  const { globalFilters } = queryOptions;
  const groupAnalysis = queryOptions.group_analysis;

  useEffect(() => {
    fetchDemoProject()
      .then((res) => {
        setDemoProjectId(res.data[0]);
      })
      .catch((err) => {
        console.log(err.data.error);
      });
  }, [activeProject]);

  const handleTour = () => {
    history.push('/');
    userflow.start('c162ed75-0983-41f3-ae56-8aedd7dbbfbd');
  };

  useEffect(() => {
    closeDrawer();
    if (query_type === 'event') {
      updateResultState({ ...initialState, loading: true });
      setLoading(true);
      runEventsQueryFromUrl();
    } else if (query_type === 'funnel') {
      updateResultState({ ...initialState, loading: true });
      setLoading(true);
      runFunnelsQueryFromUrl();
    }
  }, [query_id, query_type, queriesState]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    fetchBingAdsIntegration(activeProject.id);
    fetchMarketoIntegration(activeProject.id);
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  const isIntegrationEnabled =
    integration?.int_segment ||
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit ||
    integrationV1?.int_completed ||
    bingAds?.accounts ||
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null;

  const getQueryFromHashId = () => {
    return queriesState.data.find(function (quer) {
      return quer.id_text === query_id;
    });
  };

  const getQueryOptionsFromEquivalentQuery = (currOpts, equivalentQuery) => {
    return {
      ...currOpts,
      date_range: {
        from: MomentTz(equivalentQuery.dateRange.from),
        to: MomentTz(equivalentQuery.dateRange.to),
        frequency: equivalentQuery.dateRange.frequency
      },
      session_analytics_seq: equivalentQuery.session_analytics_seq,
      groupBy: {
        global: [...equivalentQuery.breakdown.global],
        event: [...equivalentQuery.breakdown.event]
      },
      globalFilters: equivalentQuery.globalFilters
    };
  };

  const runEventsQueryFromUrl = () => {
    const queryToAdd = getQueryFromHashId();
    if (queryToAdd) {
      // updateResultState({ ...initialState, loading: true });
      getEventsData(activeProject.id, null, null, false, query_id).then(
        (res) => {
          const queryLabels = queryToAdd?.query?.query_group[0].ewp.map((ev) =>
            ev.an ? ev.an : ev.na
          );
          const equivalentQuery = getStateQueryFromRequestQuery(
            queryToAdd?.query?.query_group[0]
          );
          setQueryType(QUERY_TYPE_EVENT);
          setQuerySaved({ name: queryToAdd.title, id: queryToAdd.id });
          updateRequestQuery(queryToAdd?.query?.query_group);
          dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
          localDispatch({
            type: SET_COMPARISON_SUPPORTED,
            payload: isComparisonEnabled(
              queryType,
              queriesA,
              equivalentQuery.breakdown,
              models
            )
          });
          setShowResult(true);
          setLoading(false);
          dispatch({
            type: INITIALIZE_GROUPBY,
            payload: equivalentQuery.breakdown
          });
          setQueries(equivalentQuery.events);
          setAppliedQueries(
            equivalentQuery.events.map((elem) =>
              elem.alias ? elem.alias : elem.label
            )
          );
          setQueryOptions((currOpts) =>
            getQueryOptionsFromEquivalentQuery(currOpts, equivalentQuery)
          );
          const newAppliedBreakdown = [
            ...equivalentQuery.breakdown.event,
            ...equivalentQuery.breakdown.global
          ];
          setAppliedBreakdown(newAppliedBreakdown);
          // updateAppliedBreakdown();
          updatePivotConfig({ ...DEFAULT_PIVOT_CONFIG });
          updateSavedQuerySettings(EMPTY_OBJECT);
          updateResultFromSavedQuery(res);
        },
        (err) => {
          console.log(err);
        }
      );
    }
  };

  const runFunnelsQueryFromUrl = () => {
    const queryToAdd = getQueryFromHashId();
    updateResultState({ ...initialState, loading: true });
    getFunnelData(activeProject.id, null, null, false, query_id).then(
      (res) => {
        const queryLabels = queryToAdd?.query?.ewp.map((ev) =>
          ev.an ? ev.an : ev.na
        );
        const equivalentQuery = getStateQueryFromRequestQuery(
          queryToAdd?.query
        );
        setQueryType(QUERY_TYPE_FUNNEL);
        closeDrawer();
        setLoading(false);
        updateRequestQuery(queryToAdd?.query);
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved({ name: queryToAdd.title, id: queryToAdd.id });
        dispatch({
          type: INITIALIZE_GROUPBY,
          payload: equivalentQuery.breakdown
        });
        localDispatch({
          type: SET_COMPARISON_SUPPORTED,
          payload: isComparisonEnabled(
            queryType,
            queriesA,
            equivalentQuery.breakdown,
            models
          )
        });
        setQueries(equivalentQuery.events);
        setAppliedQueries(
          equivalentQuery.events.map((elem) =>
            elem.alias ? elem.alias : elem.label
          )
        );
        setQueryOptions((currOpts) =>
          getQueryOptionsFromEquivalentQuery(currOpts, equivalentQuery)
        );
        const newAppliedBreakdown = [
          ...equivalentQuery.breakdown.event,
          ...equivalentQuery.breakdown.global
        ];
        setAppliedBreakdown(newAppliedBreakdown);
        // updateAppliedBreakdown();
        updateResultState({
          ...initialState,
          data: res.data.result || res.data
        });
      },
      (err) => {
        console.log(err);
      }
    );
  };

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getCampaignConfigData(activeProject.id, 'all_ads');
    }
  }, [activeProject, getCampaignConfigData]);

  const updateResultState = useCallback((newState) => {
    setResultState(newState);
  }, []);

  const updateAppliedBreakdown = useCallback(() => {
    const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
    setAppliedBreakdown(newAppliedBreakdown);
  }, [groupBy]);

  const updateLocalReducer = useCallback((type, payload) => {
    localDispatch({ type, payload });
  }, []);

  const updateChartTypes = useCallback(
    (payload) => {
      updateLocalReducer(UPDATE_CHART_TYPES, payload);
    },
    [updateLocalReducer]
  );

  const updatePivotConfig = useCallback(
    (payload) => {
      updateLocalReducer(UPDATE_PIVOT_CONFIG, payload);
    },
    [updateLocalReducer]
  );

  const resetComparisonData = useCallback(() => {
    updateLocalReducer(RESET_COMPARISON_DATA);
  }, [updateLocalReducer]);

  const handleCompareWithClick = useCallback(() => {
    updateLocalReducer(SET_COMPARISON_ENABLED, true);
  }, [updateLocalReducer]);

  const setNavigatedFromDashboard = useCallback(
    (payload) => {
      updateLocalReducer(SET_NAVIGATED_FROM_DASHBOARD, payload);
    },
    [updateLocalReducer]
  );

  const updateSavedQuerySettings = useCallback(
    (payload) => {
      updateLocalReducer(SET_SAVED_QUERY_SETTINGS, payload);
    },
    [updateLocalReducer]
  );

  const configActionsOnRunningQuery = useCallback(
    (isQuerySaved) => {
      closeDrawer();
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
      setShowResult(true);
      setQuerySaved(isQuerySaved);
      if (!isQuerySaved) {
        // reset pivot config
        updatePivotConfig({ ...DEFAULT_PIVOT_CONFIG });
        setNavigatedFromDashboard(false);
        updateSavedQuerySettings(EMPTY_OBJECT);
        setAttributionMetrics([...ATTRIBUTION_METRICS]);
      } else {
        if (queryType !== QUERY_TYPE_CAMPAIGN) {
          const selectedReport = savedQueries.find(
            (elem) => elem.id === isQuerySaved.id
          );

          // update pivot config
          const pivotConfig = getSavedPivotConfig({
            queryType,
            selectedReport
          });

          updatePivotConfig(pivotConfig);

          // update the chart type to the saved chart type
          const savedChartType = get(
            selectedReport,
            'settings.chart',
            apiChartAnnotations[CHART_TYPE_TABLE]
          );

          // even though new queries wont have saved chart type as table but old queries can have saved chart type as table!
          if (savedChartType !== apiChartAnnotations[CHART_TYPE_TABLE]) {
            const changedKey = getChartChangedKey({
              queryType,
              breakdown: [...groupBy.event, ...groupBy.global],
              attributionModels: models
            });
            updateChartTypes({
              ...DefaultChartTypes,
              [queryType]: {
                ...DefaultChartTypes[queryType],
                [changedKey]: presentationObj[savedChartType]
              }
            });
          }
        }
      }
      localDispatch({
        type: SET_COMPARISON_SUPPORTED,
        payload: isComparisonEnabled(queryType, queriesA, groupBy, models)
      });
      if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
        setAppliedQueries(
          queriesA.map((elem) => (elem.alias ? elem.alias : elem.label))
        );
        updateAppliedBreakdown();
      }
      if (queryType === QUERY_TYPE_KPI) {
        setAppliedQueries(queriesA);
        updateAppliedBreakdown();
      }
      if (queryType === QUERY_TYPE_PROFILE) {
        setAppliedQueries(
          profileQueries.map((elem) => (elem.alias ? elem.alias : elem.label))
        );
        updateAppliedBreakdown();
      }
    },
    [
      dispatch,
      groupBy,
      queriesA,
      profileQueries,
      queryType,
      models,
      savedQueries,
      updateAppliedBreakdown,
      setNavigatedFromDashboard,
      updateSavedQuerySettings,
      updateChartTypes
    ]
  );

  const getDashboardConfigs = useCallback(
    (isQuerySaved) => {
      // use cache urls when user expands the dashboard widget
      if (isQuerySaved && coreQueryState.navigatedFromDashboard) {
        return {
          id: coreQueryState.navigatedFromDashboard.dashboard_id,
          unit_id: coreQueryState.navigatedFromDashboard.id,
          refresh: false
        };
      }
      return null;
    },
    [coreQueryState.navigatedFromDashboard]
  );

  const updateResultFromSavedQuery = (res) => {
    const data = res.data.result || res.data;
    if (result_criteria === TOTAL_EVENTS_CRITERIA) {
      updateResultState({
        ...initialState,
        data: formatApiData(data.result_group[0], data.result_group[1])
      });
    } else if (result_criteria === TOTAL_USERS_CRITERIA) {
      if (user_type === EACH_USER_TYPE) {
        updateResultState({
          ...initialState,
          data: formatApiData(data.result_group[0], data.result_group[1])
        });
      } else {
        updateResultState({
          ...initialState,
          data: data.result_group[0]
        });
      }
    } else if (result_criteria === ACTIVE_USERS_CRITERIA) {
      const userData = formatApiData(
        data.result_group[0],
        data.result_group[1]
      );
      const sessionsData = data.result_group[2];
      const activeUsersData = calculateActiveUsersData(userData, sessionsData, [
        ...groupBy.global,
        ...groupBy.event
      ]);
      updateResultState({ ...initialState, data: activeUsersData });
    } else if (result_criteria === FREQUENCY_CRITERIA) {
      const eventData = formatApiData(
        data.result_group[0],
        data.result_group[1]
      );
      const userData = formatApiData(
        data.result_group[2],
        data.result_group[3]
      );
      const frequencyData = calculateFrequencyData(eventData, userData, [
        ...groupBy.global,
        ...groupBy.event
      ]);
      updateResultState({ ...initialState, data: frequencyData });
    }
  };

  const runQuery = useCallback(
    async (
      isQuerySaved,
      durationObj,
      isGranularityChange = false,
      isCompareQuery = false
    ) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        const query = getQuery(
          groupBy,
          queriesA,
          result_criteria,
          user_type,
          durationObj,
          globalFilters
        );

        setDateFromTo({ from: query[0]?.fr, to: query[0]?.to });

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_EVENT,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }
        if (!isCompareQuery) {
          setShowResult(true);
          setLoading(true);
          configActionsOnRunningQuery(isQuerySaved);
          setBreakdownType(user_type);
          updateRequestQuery(query);
          updateResultState({ ...initialState, loading: true });
          resetComparisonData();
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        const res = await getEventsData(
          activeProject.id,
          query,
          getDashboardConfigs(isGranularityChange ? false : isQuerySaved), // we need to call fresh query when granularity is changed
          true
        );
        const data = res.data.result || res.data;
        let resultantData = null;
        if (result_criteria === TOTAL_EVENTS_CRITERIA) {
          resultantData = formatApiData(
            data.result_group[0],
            data.result_group[1]
          );
        } else if (result_criteria === TOTAL_USERS_CRITERIA) {
          if (user_type === EACH_USER_TYPE) {
            resultantData = formatApiData(
              data.result_group[0],
              data.result_group[1]
            );
          } else {
            resultantData = data.result_group[0];
          }
        } else if (result_criteria === ACTIVE_USERS_CRITERIA) {
          const userData = formatApiData(
            data.result_group[0],
            data.result_group[1]
          );
          const sessionsData = data.result_group[2];
          resultantData = calculateActiveUsersData(userData, sessionsData, [
            ...groupBy.global,
            ...groupBy.event
          ]);
        } else if (result_criteria === FREQUENCY_CRITERIA) {
          const eventData = formatApiData(
            data.result_group[0],
            data.result_group[1]
          );
          const userData = formatApiData(
            data.result_group[2],
            data.result_group[3]
          );
          resultantData = calculateFrequencyData(eventData, userData, [
            ...groupBy.global,
            ...groupBy.event
          ]);
        }
        if (isCompareQuery) {
          updateLocalReducer(COMPARISON_DATA_FETCHED, resultantData);
        } else {
          setLoading(false);
          updateResultState({
            ...initialState,
            data: resultantData
          });
        }
      } catch (err) {
        console.log(err);
        setLoading(false);
        updateResultState({ ...initialState, loading: false, error: true });
      }
    },
    [
      queriesA,
      dateRange,
      result_criteria,
      user_type,
      activeProject.id,
      groupBy,
      globalFilters,
      updateResultState,
      configActionsOnRunningQuery,
      updateLocalReducer,
      getDashboardConfigs
    ]
  );

  const runFunnelQuery = useCallback(
    async (isQuerySaved, durationObj, isCompareQuery) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
          resetComparisonData();
        }
        const query = getFunnelQuery(
          groupBy,
          queriesA,
          session_analytics_seq,
          durationObj,
          globalFilters
        );

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_FUNNEL,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }

        if (!isCompareQuery) {
          setLoading(true);
          configActionsOnRunningQuery(isQuerySaved);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(query);
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        const res = await getFunnelData(
          activeProject.id,
          query,
          getDashboardConfigs(isQuerySaved),
          true
        );
        if (isCompareQuery) {
          updateLocalReducer(
            COMPARISON_DATA_FETCHED,
            res.data.result || res.data
          );
        } else {
          setLoading(false);
          updateResultState({
            ...initialState,
            data: res.data.result || res.data
          });
        }
      } catch (err) {
        console.log(err);
        setLoading(false);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      queriesA,
      session_analytics_seq,
      activeProject.id,
      groupBy,
      globalFilters,
      dateRange,
      updateResultState,
      configActionsOnRunningQuery,
      updateLocalReducer,
      resetComparisonData,
      getDashboardConfigs
    ]
  );

  const runAttributionQuery = useCallback(
    async (isQuerySaved, durationObj, isCompareQuery) => {
      try {
        if (!durationObj) {
          durationObj = attr_dateRange;
          resetComparisonData();
        }
        const query = getAttributionQuery(
          eventGoal,
          touchpoint,
          attr_dimensions,
          content_groups,
          touchpoint_filters,
          attr_query_type,
          models,
          window,
          linkedEvents,
          durationObj,
          tacticOfferType
        );

        if (
          queryOptions.group_analysis &&
          queryOptions.group_analysis !== 'users'
        ) {
          const dtRange = { ...durationObj, frequency: 'hour' };
          const kpiQuery = getKPIQuery(
            attrQueries,
            dtRange,
            { event: [], global: [] },
            queryOptions,
            []
          );
          if (queryOptions.group_analysis === 'hubspot_deals') {
            kpiQuery.gGBy = [
              {
                gr: '',
                prNa: '$hubspot_deal_hs_object_id',
                prDaTy: 'numerical',
                en: 'user',
                objTy: '',
                gbty: 'raw_values'
              }
            ];
          } else if (
            queryOptions.group_analysis === 'salesforce_opportunities'
          ) {
            kpiQuery.gGBy = [
              {
                gr: '',
                prNa: '$salesforce_opportunity_id',
                prDaTy: 'numerical',
                en: 'user',
                objTy: '',
                gbty: 'raw_values'
              }
            ];
          }
          query.query.analyze_type = queryOptions.group_analysis;
          query.query.kpi_query_group = kpiQuery;
        }

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_ATTRIBUTION,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }

        if (!isCompareQuery) {
          configActionsOnRunningQuery(isQuerySaved);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(query);
          setAttributionsState({
            eventGoal,
            touchpoint,
            models,
            linkedEvents,
            attr_dimensions,
            content_groups,
            tacticOfferType,
            date_range: { ...durationObj }
          });
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        let apiCallStatus = { required: true, message: null };
        if (
          isQuerySaved &&
          coreQueryState.navigatedFromDashboard &&
          !isCompareQuery
        ) {
          apiCallStatus = shouldDataFetch(durationObj);
        }
        if (apiCallStatus.required) {
          if (!isCompareQuery) {
            setLoading(true);
          }
          const res = await getAttributionsData(
            activeProject.id,
            query,
            getDashboardConfigs(isQuerySaved),
            true
          );
          if (isCompareQuery) {
            updateLocalReducer(
              COMPARISON_DATA_FETCHED,
              res.data.result || res.data
            );
          } else {
            setLoading(false);
            updateResultState({
              ...initialState,
              data: res.data.result || res.data,
              apiCallStatus
            });
          }
        } else {
          updateResultState({
            ...initialState,
            apiCallStatus
          });
        }
      } catch (err) {
        console.log(err);
        setLoading(false);
        updateResultState({
          ...initialState,
          error: true
        });
      }
    },
    [
      activeProject.id,
      eventGoal,
      linkedEvents,
      models,
      touchpoint,
      touchpoint_filters,
      attr_query_type,
      tacticOfferType,
      window,
      attr_dateRange,
      updateResultState,
      attr_dimensions,
      content_groups,
      getDashboardConfigs,
      configActionsOnRunningQuery,
      resetComparisonData,
      updateLocalReducer,
      coreQueryState.navigatedFromDashboard
    ]
  );

  const runKPIQuery = useCallback(
    async (
      isQuerySaved,
      durationObj = null,
      isGranularityChange = false,
      isCompareQuery = false
    ) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        const KPIquery = getKPIQuery(
          queriesA,
          durationObj,
          groupBy,
          queryOptions
        );

        setDateFromTo({ from: KPIquery?.qG[0]?.fr, to: KPIquery?.qG[0]?.to });

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_KPI,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }

        if (!isCompareQuery) {
          setLoading(true);
          configActionsOnRunningQuery(isQuerySaved);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(KPIquery);
          resetComparisonData();
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }

        const res = await getKPIData(
          activeProject.id,
          KPIquery,
          getDashboardConfigs(isGranularityChange ? false : isQuerySaved),
          true
        );

        if (isCompareQuery) {
          updateLocalReducer(
            COMPARISON_DATA_FETCHED,
            res.data.result || res.data
          );
        } else {
          setLoading(false);
          updateResultState({
            ...initialState,
            data: res.data.result || res.data
          });
        }
      } catch (err) {
        console.log(err);
        setLoading(false);
        updateResultState({
          ...initialState,
          error: true
        });
      }
    },
    [
      queriesA,
      activeProject.id,
      groupBy,
      globalFilters,
      dateRange,
      updateResultState,
      configActionsOnRunningQuery,
      updateLocalReducer,
      resetComparisonData,
      getDashboardConfigs
    ]
  );

  const runCampaignsQuery = useCallback(
    async (isQuerySaved, durationObj = null, isGranularityChange = false) => {
      try {
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        if (!isQuerySaved) {
          setNavigatedFromDashboard(false);
        }
        updateResultState({
          ...initialState,
          loading: true
        });
        if (!durationObj) {
          durationObj = camp_dateRange;
        }
        const query = getCampaignsQuery(
          camp_channels,
          camp_measures,
          camp_filters,
          camp_groupBy,
          durationObj
        );

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_CAMPAIGN,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }

        setCampaignState({
          channel: query.query_group[0].channel,
          filters: query.query_group[0].filters,
          select_metrics: query.query_group[0].select_metrics,
          group_by: query.query_group[0].group_by,
          date_range: { ...durationObj }
        });
        updateRequestQuery(query);
        setLoading(true);
        const res = await getCampaignsData(
          activeProject.id,
          query,
          getDashboardConfigs(isGranularityChange ? false : isQuerySaved), // we need to call fresh query when granularity is changed
          true
        );
        setLoading(false);
        updateResultState({
          ...initialState,
          data: res.data.result || res.data
        });
      } catch (err) {
        console.log(err);
        setLoading(false);
        updateResultState({
          ...initialState,
          error: true
        });
      }
    },
    [
      dispatch,
      activeProject.id,
      camp_measures,
      camp_filters,
      camp_groupBy,
      camp_channels,
      camp_dateRange,
      updateResultState,
      setNavigatedFromDashboard,
      getDashboardConfigs
    ]
  );

  const runProfileQuery = useCallback(
    async (isQuerySaved, durationObj) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        const query = getProfileQuery(
          profileQueries,
          groupBy,
          globalFilters,
          durationObj,
          groupAnalysis
        );

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            query_type: QUERY_TYPE_PROFILE,
            project_id: activeProject?.id,
            project_name: activeProject?.name
          });
        }

        configActionsOnRunningQuery(isQuerySaved);
        updateRequestQuery(query);
        updateResultState({ ...initialState, loading: true });
        setLoading(true);
        const res = await getProfileData(
          activeProject.id,
          query,
          getDashboardConfigs(isQuerySaved),
          true
        );
        setLoading(false);
        updateResultState({
          ...initialState,
          data: res.data.result || res.data
        });
      } catch (err) {
        setLoading(false);
        console.log(err);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      profileQueries,
      activeProject.id,
      groupBy,
      globalFilters,
      dateRange,
      groupAnalysis,
      updateResultState,
      getDashboardConfigs
    ]
  );

  const handleGranularityChange = useCallback(
    ({ key: frequency }) => {
      resetComparisonData();
      if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_KPI) {
        const appliedDateRange = {
          ...queryOptions.date_range,
          frequency
        };
        setQueryOptions((currState) => {
          return {
            ...currState,
            date_range: appliedDateRange
          };
        });
        if (queryType === QUERY_TYPE_EVENT) {
          runQuery(querySaved, appliedDateRange, true);
        }
        if (queryType === QUERY_TYPE_KPI) {
          runKPIQuery(querySaved, appliedDateRange, true);
        }
      }
      if (queryType === QUERY_TYPE_CAMPAIGN) {
        const payload = {
          ...camp_dateRange,
          frequency
        };
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload, true);
      }
    },
    [
      queryOptions.date_range,
      querySaved,
      runQuery,
      camp_dateRange,
      dispatch,
      queryType,
      runCampaignsQuery,
      resetComparisonData
    ]
  );

  const handleDurationChange = useCallback(
    (dates, isCompareDate) => {
      let from;
      let to;
      let frequency;
      const { dateType, selectedOption } = dates;

      if (Array.isArray(dates.startDate)) {
        from = dates.startDate[0];
        to = dates.startDate[1];
      } else {
        from = dates.startDate;
        to = dates.endDate;
      }

      if (
        queryType === QUERY_TYPE_EVENT ||
        queryType === QUERY_TYPE_CAMPAIGN ||
        queryType === QUERY_TYPE_KPI
      ) {
        frequency = getValidGranularityOptions({ from, to })[0];
      }

      const startDate = moment(from).startOf('day').utc().unix() * 1000;
      const endDate = moment(to).endOf('day').utc().unix() * 1000 + 1000;
      const daysDiff = moment(endDate).diff(startDate, 'days');
      if (daysDiff > 1) {
        frequency =
          queryOptions.date_range.frequency === 'hour' || frequency === 'hour'
            ? 'date'
            : queryOptions.date_range.frequency;
      } else frequency = 'hour';

      const payload = {
        from: MomentTz(from).startOf('day'),
        to: MomentTz(to).endOf('day'),
        frequency,
        dateType
      };

      if (!isCompareDate) {
        setQueryOptions((currState) => {
          return {
            ...currState,
            date_range: {
              ...currState.date_range,
              ...payload
            }
          };
        });
      }

      if (isCompareDate) {
        localDispatch({
          type: SET_COMPARE_DURATION,
          payload: {
            from,
            to,
            frequency,
            dateType,
            selectedOption
          }
        });
      }

      const appliedDateRange = {
        ...queryOptions.date_range,
        ...payload
      };

      if (queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(querySaved, appliedDateRange, isCompareDate);
      }

      if (queryType === QUERY_TYPE_EVENT) {
        runQuery(querySaved, appliedDateRange, false, isCompareDate);
      }
      if (queryType === QUERY_TYPE_KPI) {
        runKPIQuery(querySaved, appliedDateRange, false, isCompareDate);
      }

      if (queryType === QUERY_TYPE_CAMPAIGN) {
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload);
      }

      if (queryType === QUERY_TYPE_PROFILE) {
        runProfileQuery(querySaved, payload);
      }

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        if (!isCompareDate) {
          // set range in reducer only when original date is changed and not the comparisom date
          dispatch({ type: SET_ATTR_DATE_RANGE, payload });
        }
        runAttributionQuery(querySaved, payload, isCompareDate);
      }
    },
    [
      queryType,
      runFunnelQuery,
      runQuery,
      querySaved,
      queryOptions.date_range,
      dispatch,
      runCampaignsQuery,
      runAttributionQuery,
      runProfileQuery
    ]
  );

  useEffect(() => {
    if (clickedSavedReport) {
      if (clickedSavedReport.queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      } else if (clickedSavedReport.queryType === QUERY_TYPE_ATTRIBUTION) {
        runAttributionQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      } else if (clickedSavedReport.queryType === QUERY_TYPE_CAMPAIGN) {
        runCampaignsQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      } else if (clickedSavedReport.queryType === QUERY_TYPE_KPI) {
        runKPIQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      } else if (clickedSavedReport.queryType === QUERY_TYPE_PROFILE) {
        runProfileQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      } else {
        runQuery({
          id: clickedSavedReport.query_id,
          name: clickedSavedReport.queryName
        });
      }
      setClickedSavedReport(false);
    }
  }, [
    clickedSavedReport,
    runFunnelQuery,
    runQuery,
    runAttributionQuery,
    runCampaignsQuery,
    runKPIQuery,
    runProfileQuery
  ]);

  useEffect(() => {
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  const queryChange = (newEvent, index, changeType = 'add', flag = null) => {
    const queryupdated = [...queriesA];
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

  const profileQueryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...profileQueries];
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
      queryupdated.push(newEvent);
    }
    setProfileQueries(queryupdated);
  };

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const setExtraOptions = (options) => {
    setQueryOptions(options);
  };

  const handleRunQuery = useCallback(() => {
    switch (queryType) {
      case QUERY_TYPE_EVENT: {
        runQuery(false);
        break;
      }
      case QUERY_TYPE_FUNNEL: {
        runFunnelQuery(false);
        break;
      }
      case QUERY_TYPE_KPI: {
        runKPIQuery(false);
        break;
      }
      case QUERY_TYPE_ATTRIBUTION: {
        runAttributionQuery(false);
        break;
      }
      case QUERY_TYPE_PROFILE: {
        runProfileQuery(false);
        break;
      }
    }
  }, [
    queryType,
    runQuery,
    runFunnelQuery,
    runKPIQuery,
    runAttributionQuery,
    runProfileQuery
  ]);

  const title = () => {
    const IconAndText = IconAndTextSwitchQueryType(queryType);
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex items-center'}>
          <SVG name={IconAndText.icon} size="24px"></SVG>
          <Text
            type={'title'}
            level={4}
            weight={'bold'}
            extraClass={'ml-2 m-0'}
          >
            {IconAndText.text}
          </Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Button size={'large'} type="text" onClick={() => closeDrawer()}>
            <SVG name="times"></SVG>
          </Button>
        </div>
      </div>
    );
  };

  const campaignsArrayMapper = useMemo(() => {
    return campaignState.select_metrics.map((metric, index) => {
      return {
        eventName: metric,
        index,
        mapper: `event${index + 1}`
      };
    });
  }, [campaignState.select_metrics]);

  const arrayMapper = useMemo(() => {
    return appliedQueries.map((q, index) => {
      return {
        eventName: q,
        index,
        mapper: `event${index + 1}`,
        displayName: eventNames[q] ? eventNames[q] : q
      };
    });
  }, [appliedQueries, eventNames]);

  const checkIfnewComposer = () => {
    return (
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_ATTRIBUTION ||
      queryType === QUERY_TYPE_KPI ||
      queryType === QUERY_TYPE_PROFILE
    );
  };

  const renderQueryComposer = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
      return (
        <QueryComposer
          queries={queriesA}
          runQuery={handleRunQuery}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runFunnelQuery={handleRunQuery}
          activeKey={activeKey}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return (
        <AttrQueryComposer
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runAttributionQuery={handleRunQuery}
        />
      );
    }

    if (queryType === QUERY_TYPE_KPI) {
      return (
        <KPIComposer
          queries={queriesA}
          setQueries={setQueries}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          activeKey={activeKey}
          handleRunQuery={handleRunQuery}
          selectedMainCategory={selectedMainCategory}
          setSelectedMainCategory={setSelectedMainCategory}
          KPIConfigProps={KPIConfigProps}
          setKPIConfigProps={setKPIConfigProps}
        />
      );
    }
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return (
        <CampQueryComposer
          handleRunQuery={runCampaignsQuery}
        ></CampQueryComposer>
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      return (
        <ProfileComposer
          queries={profileQueries}
          setQueries={setProfileQueries}
          runProfileQuery={handleRunQuery}
          eventChange={profileQueryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
        ></ProfileComposer>
      );
    }
  };

  const renderQueryComposerNew = () => {
    if (
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_ATTRIBUTION ||
      queryType === QUERY_TYPE_KPI ||
      queryType === QUERY_TYPE_PROFILE
    ) {
      return (
        <div
          className={queryOpen ? 'query_card_open-add' : 'query_card_close'}
          onClick={() => !queryOpen && setQueryOpen(true)}
        >
          {renderQueryComposer()}
        </div>
      );
    }
    return null;
  };

  const handleBreadCrumbClick = () => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
    setQuerySaved(false);
    updateRequestQuery(null);
    closeDrawer();

    if (queryType === QUERY_TYPE_KPI) {
      setQueries([]);
    }
  };

  function changeTab(key) {
    setActiveTab(key);
  }

  const renderCreateQFlow = () => {
    return (
      <CoreQueryContext.Provider
        value={{
          coreQueryState,
          attributionMetrics,
          setAttributionMetrics,
          setNavigatedFromDashboard,
          resetComparisonData,
          handleCompareWithClick
        }}
      >
        <Modal
          title={
            <AnalysisHeader
              requestQuery={requestQuery}
              onBreadCrumbClick={handleBreadCrumbClick}
              queryType={queryType}
              queryTitle={querySaved ? querySaved.name : null}
              setQuerySaved={setQuerySaved}
              breakdownType={breakdownType}
              changeTab={changeTab}
              activeTab={activeTab}
              savedQueryId={querySaved ? querySaved.id : null}
            />
          }
          visible={drawerVisible}
          footer={null}
          centered={false}
          mask={false}
          closable={false}
          className={'fa-modal--full-width'}
        >
          <div className="px-20">
            <ErrorBoundary
              fallback={
                <FaErrorComp
                  size={'medium'}
                  title={'Analyse Results Error'}
                  subtitle={
                    'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
                  }
                />
              }
              onError={FaErrorLog}
            >
              {Number(activeTab) === 1 && <>{renderQueryComposerNew()}</>}
            </ErrorBoundary>
          </div>
        </Modal>
      </CoreQueryContext.Provider>
    );
  };

  const closeResultPage = (flag = false) => {
    setQuerySaved(false);
    setDrawerVisible(flag);
  };

  useEffect(() => {
    setKPIConfigProps(findKPIitem(selectedMainCategory?.group));
  }, [selectedMainCategory]);

  const findKPIitem = (groupName) => {
    const KPIlist = KPI_config || [];
    const selGroup = KPIlist.find((item) => {
      return item.display_category === groupName;
    });

    const DDvalues = selGroup?.properties?.map((item) => {
      if (item == null) return;
      const ddName = item.display_name ? item.display_name : item.name;
      const ddtype =
        selGroup?.category === 'channels' ||
        selGroup?.category === 'custom_channels'
          ? item.object_type
          : item.entity
          ? item.entity
          : item.object_type;
      return [ddName, item.name, item.data_type, ddtype];
    });
    return DDvalues;
  };

  const closePage = () => {
    history.goBack();
  };

  const handleCloseDashboardQuery = () => {
    history.push({
      pathname: '/',
      state: { dashboardWidgetId: coreQueryState.navigatedFromDashboard.id }
    });
    handleBreadCrumbClick();
  };

  const handleCloseToAnalyse = () => {
    history.push({
      pathname: '/analyse'
    });
    handleBreadCrumbClick();
  };

  if (loading) {
    return (
      <div className="flex justify-center flex-col items-center w-full">
        <div className="w-full flex center">
          <div
            id="app-header"
            className="bg-white z-50 flex-col  px-8 w-full"
          >
            <div
              className={'items-center flex justify-between w-full pt-3 pb-3'}
            >
              <div
                role="button"
                tabIndex={0}
                className="flex items-center cursor-pointer"
              >
                <Button
                  size={'large'}
                  type="text"
                  onClick={() => {
                    history.push('/');
                  }}
                  icon={<SVG size={32} name="Brand" />}
                />
                <Text
                  type={'title'}
                  level={5}
                  weight={`bold`}
                  extraClass={'m-0 mt-1'}
                  lineHeight={'small'}
                >
                  {querySaved
                    ? `Reports / ${queryType} / ${querySaved.name}`
                    : `Reports / ${queryType} / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
                </Text>
              </div>

              <div className="flex items-center gap-x-2">
                <Button
                  size={'large'}
                  type="text"
                  icon={<SVG size={20} name={'close'} />}
                  onClick={
                    coreQueryState.navigatedFromDashboard
                      ? handleCloseDashboardQuery
                      : handleCloseToAnalyse
                  }
                />
              </div>
            </div>

            {/* {renderReportTabs()} */}
            {showResult ? (
              <div
                className={`query_card_cont ${
                  queryOpen ? `query_card_open` : `query_card_close`
                }`}
                onClick={(e) => !queryOpen && setQueryOpen(true)}
              >
                {renderQueryComposer()}
                <Button size={'large'} className={`query_card_expand`}>
                  <SVG name={'expand'} size={20}></SVG>Expand
                </Button>
              </div>
            ) : null}
          </div>
        </div>
        <div className="w-full h-64 flex items-center justify-center">
          <Spin size="large" />
        </div>
      </div>
    );
  }

  if (isIntegrationEnabled || activeProject.id === demoProjectId) {
    return (
      <>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Analyse Error'}
              subtitle={
                'We are facing trouble loading Analyse. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
          {
            <Drawer
              title={title()}
              placement="left"
              closable={false}
              visible={drawerVisible && !checkIfnewComposer()}
              onClose={closeDrawer}
              getContainer={false}
              width={'650px'}
              className={'fa-drawer'}
            >
              <ErrorBoundary
                fallback={
                  <FaErrorComp subtitle={'Facing issues with Query Builder'} />
                }
                onError={FaErrorLog}
              >
                {renderQueryComposer()}
              </ErrorBoundary>
            </Drawer>
          }

          {!showResult &&
          !resultState.data &&
          !resultState.loading &&
          activeProject.id === demoProjectId ? (
            <div className={'rounded-lg border-2 h-20 mt-20 -mb-20 mx-20'}>
              <Row justify={'space-between'} className={'m-0 p-3'}>
                <Col span={projects.length === 1 ? 12 : 18}>
                  <img
                    src="assets/icons/welcome.svg"
                    style={{ float: 'left', marginRight: '20px' }}
                  />
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={'m-0'}
                  >
                    Welcome! You just entered a Factors demo project
                  </Text>
                  {projects.length === 1 ? (
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      These reports have been built with a sample dataset. Use
                      this to start exploring!
                    </Text>
                  ) : (
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      To jump back into your Factors project, click on your
                      account card on the{' '}
                      <span className={'font-bold'}>top right</span> of the
                      screen.
                    </Text>
                  )}
                </Col>
                <Col className={'mr-2 mt-2'}>
                  {projects.length === 1 ? (
                    <Button
                      type={'default'}
                      style={{
                        background: 'white',
                        border: '1px solid #E7E9ED',
                        height: '40px'
                      }}
                      className={'m-0 mr-2'}
                      onClick={() => setShowProjectModal(true)}
                    >
                      Set up my own Factors project
                    </Button>
                  ) : null}

                  <Button
                    type={'link'}
                    style={{
                      background: 'white',
                      // border: '1px solid #E7E9ED',
                      height: '40px'
                    }}
                    className={'m-0 mr-2'}
                    onClick={() => handleTour()}
                  >
                    Take the tour{' '}
                    <SVG
                      name={'Arrowright'}
                      size={16}
                      extraClass={'ml-1'}
                      color={'blue'}
                    />
                  </Button>
                </Col>
              </Row>
            </div>
          ) : null}

          {!showResult && resultState.loading ? <PageSuspenseLoader /> : null}

          {!showResult && drawerVisible && checkIfnewComposer()
            ? renderCreateQFlow()
            : !showResult &&
              !resultState.loading && (
                <CoreQueryHome
                  setQueryType={setQueryType}
                  setDrawerVisible={closeResultPage}
                  setQueries={setQueries}
                  setProfileQueries={setProfileQueries}
                  setQueryOptions={setExtraOptions}
                  setClickedSavedReport={setClickedSavedReport}
                  location={location}
                  setActiveKey={setActiveKey}
                  setBreakdownType={setBreakdownType}
                  setNavigatedFromDashboard={setNavigatedFromDashboard}
                  updateChartTypes={updateChartTypes}
                  updateSavedQuerySettings={updateSavedQuerySettings}
                  setAttributionMetrics={setAttributionMetrics}
                  dateFromTo={dateFromTo}
                />
              )}

          {showResult && !resultState.loading ? (
            <CoreQueryContext.Provider
              value={{
                coreQueryState,
                attributionMetrics,
                setAttributionMetrics,
                setNavigatedFromDashboard,
                resetComparisonData,
                handleCompareWithClick,
                updatePivotConfig,
                queriesA,
                profileQueries,
                queryOptions,
                selectedMainCategory,
                setSelectedMainCategory,
                runQuery,
                queryChange,
                profileQueryChange,
                setExtraOptions,
                runFunnelQuery,
                runKPIQuery,
                runProfileQuery,
                activeKey,
                showResult,
                KPIConfigProps,
                setQueries,
                setProfileQueries,
                runAttributionQuery
              }}
            >
              <AnalysisResultsPage
                queryType={queryType}
                resultState={resultState}
                setDrawerVisible={closeResultPage}
                requestQuery={requestQuery}
                queries={appliedQueries}
                breakdown={appliedBreakdown}
                setShowResult={() => {
                  setShowResult(false);
                  updateRequestQuery(false);
                }}
                queryTitle={querySaved ? querySaved.name : null}
                savedQueryId={querySaved ? querySaved.id : null}
                setQuerySaved={setQuerySaved}
                durationObj={queryOptions.date_range}
                handleDurationChange={handleDurationChange}
                arrayMapper={arrayMapper}
                queryOptions={queryOptions}
                attributionsState={attributionsState}
                breakdownType={breakdownType}
                campaignState={campaignState}
                eventPage={result_criteria}
                section={REPORT_SECTION}
                runAttrCmprQuery={null}
                cmprResultState={null}
                campaignsArrayMapper={campaignsArrayMapper}
                handleGranularityChange={handleGranularityChange}
                updateChartTypes={updateChartTypes}
                dateFromTo={dateFromTo}
              />
            </CoreQueryContext.Provider>
          ) : null}
          {/* create project modal */}
          <NewProject
            visible={showProjectModal}
            handleCancel={() => setShowProjectModal(false)}
          />
        </ErrorBoundary>
      </>
    );
  } else {
    return (
      <>
        <AnalyseBeforeIntegration />
      </>
    );
  }
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  KPI_config: state.kpi?.config,
  existingQueries: state.queries
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      deleteGroupByForEvent,
      getCampaignConfigData,
      fetchDemoProject,
      getHubspotContact,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
