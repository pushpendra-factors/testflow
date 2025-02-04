import React, {
  useState,
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef
} from 'react';
import get from 'lodash/get';
import { bindActionCreators } from 'redux';
import { connect, useSelector, useDispatch } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import { Spin } from 'antd';
import MomentTz from 'Components/MomentTz';

import { EMPTY_ARRAY, EMPTY_OBJECT, generateRandomKey } from 'Utils/global';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';
import moment from 'moment';
import {
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
} from 'Reducers/global';

import { FaErrorComp, FaErrorLog } from 'Components/factorsComponents';
import {
  deleteGroupByForEvent,
  getCampaignConfigData
} from 'Reducers/coreQuery/middleware';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_ATTRIBUTION,
  EACH_USER_TYPE,
  REPORT_SECTION,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_TYPE_PROFILE,
  apiChartAnnotations,
  presentationObj,
  DefaultChartTypes,
  CHART_TYPE_TABLE,
  QUERY_OPTIONS_DEFAULT_VALUE
} from 'Utils/constants';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import { INITIALIZE_GROUPBY } from 'Reducers/coreQuery/actions';
import CoreQueryReducer from 'Views/CoreQuery/CoreQueryReducer';
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
  DEFAULT_PIVOT_CONFIG,
  UPDATE_CORE_QUERY_REDUCER,
  SET_NAVIGATED_FROM_ANALYSE,
  DEFAULT_ATTRIBUTION_TABLE_FILTERS
} from 'Views/CoreQuery/constants';
import { getValidGranularityOptions } from 'Utils/dataFormatter';
import { getSavedPivotConfig } from 'Views/CoreQuery/coreQuery.helpers';
import { getChartChangedKey } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import _ from 'lodash';
import { fetchKPIConfig } from 'Reducers/kpi';
import { getQuickDashboardDateRange } from 'Views/Dashboard/utils';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import { getQueryData } from 'Views/PreBuildDashboard/state/services';
import {
  getPredefinedQuery,
  transformWidgetResponse
} from 'Views/PreBuildDashboard/utils';
import { selectActivePreDashboard } from 'Reducers/dashboard/selectors';
import { useLocation } from 'react-router-dom';
import ReportContent from './ReportContent';
import ReportHeader from './ReportHeader';
import { initialState, isComparisonEnabled } from '../../../CoreQuery/utils';

function CoreQuery({ activeProject }) {
  const query_type = 'kpi';

  const savedQueries = useSelector((state) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );
  const activeDashboard = useSelector((state) =>
    selectActivePreDashboard(state)
  );
  const filtersData = useSelector(
    (state) => state.preBuildDashboardConfig.reportFilters
  );
  const [coreQueryState, localDispatch] = useReducer(
    CoreQueryReducer,
    CORE_QUERY_INITIAL_STATE
  );
  const [queryType, setQueryType] = useState(QUERY_TYPE_KPI);
  const [activeKey, setActiveKey] = useState('0');
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownType, setBreakdownType] = useState(EACH_USER_TYPE);
  const [queriesA, setQueries] = useState([]);
  const [loading, setLoading] = useState(true);
  const renderedCompRef = useRef(null);

  const location = useLocation();

  const [profileQueries, setProfileQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...getQuickDashboardDateRange() }
  });

  const dispatch = useDispatch();
  const { groupBy, models, window, camp_dateRange } = useSelector(
    (state) => state.coreQuery
  );

  const [activeTab, setActiveTab] = useState(1);

  const [dateFromTo, setDateFromTo] = useState({ from: '', to: '' });

  const dateRange = queryOptions.date_range;
  /*
    This use Effect checks which route drawer we need to open
    when we goto route /analyse/:query_type

  */
  useEffect(() => {
    switch (query_type) {
      case QUERY_TYPE_KPI:
        setQueryType(QUERY_TYPE_KPI);
        break;
      case QUERY_TYPE_FUNNEL:
        setQueryType(QUERY_TYPE_FUNNEL);
        break;
      case QUERY_TYPE_ATTRIBUTION:
        setQueryType(QUERY_TYPE_ATTRIBUTION);
        break;
      case QUERY_TYPE_PROFILE:
        setQueryType(QUERY_TYPE_PROFILE);
        break;
      case QUERY_TYPE_EVENT:
        setQueryType(QUERY_TYPE_EVENT);
        break;
      default:
        break;
    }
    if (query_type && query_type.length > 0) {
      setQueries([]);
      dispatch({
        type: INITIALIZE_GROUPBY,
        payload: {
          global: [],
          event: []
        }
      });
    }
  }, [dispatch, query_type]);

  const updateResultState = useCallback((newState) => {
    setResultState(newState);
  }, []);

  const updateAppliedBreakdown = useCallback((data) => {
    setAppliedBreakdown(data);
  }, []);

  const updateLocalReducer = useCallback((type, payload) => {
    localDispatch({ type, payload });
  }, []);

  const updateCoreQueryReducer = useCallback((payload) => {
    localDispatch({
      type: UPDATE_CORE_QUERY_REDUCER,
      payload
    });
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

  const setNavigatedFromAnalyse = useCallback(
    (payload) => {
      updateLocalReducer(SET_NAVIGATED_FROM_ANALYSE, payload);
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
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
      setShowResult(true);
      if (!isQuerySaved) {
        // reset pivot config
        updatePivotConfig({ ...DEFAULT_PIVOT_CONFIG });
        // setNavigatedFromDashboard(false);
        updateSavedQuerySettings(EMPTY_OBJECT);
        // reset attribution table filters
        updateCoreQueryReducer({
          attributionTableFilters: DEFAULT_ATTRIBUTION_TABLE_FILTERS
        });
      } else if (queryType !== QUERY_TYPE_CAMPAIGN) {
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
          isQuerySaved,
          'chart_setting.ty',
          apiChartAnnotations[CHART_TYPE_TABLE]
        );

        // even though new queries wont have saved chart type as table but old queries can have saved chart type as table!
        if (savedChartType !== apiChartAnnotations[CHART_TYPE_TABLE]) {
          const changedKey = getChartChangedKey({
            queryType,
            breakdown: isQuerySaved?.g_by
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
      localDispatch({
        type: SET_COMPARISON_SUPPORTED,
        payload: isComparisonEnabled(queryType, queriesA, groupBy, models)
      });
      if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
        setAppliedQueries(
          queriesA.map((elem) => (elem.alias ? elem.alias : elem.label))
        );
      }
      if (queryType === QUERY_TYPE_PROFILE) {
        setAppliedQueries(
          profileQueries.map((elem) => (elem.alias ? elem.alias : elem.label))
        );
      }
    },
    [
      dispatch,
      queryType,
      queriesA,
      groupBy,
      models,
      updatePivotConfig,
      updateSavedQuerySettings,
      updateCoreQueryReducer,
      savedQueries,
      updateChartTypes,
      profileQueries
    ]
  );

  const runKPIQuery = useCallback(
    async (
      query,
      breakdown = {},
      filter = [],
      durationObj = null,
      isGranularityChange = false,
      isCompareQuery = false
    ) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }

        setQuerySaved(query);
        const kpiData = query.me.map((obj) => {
          const { inter_e_type, ty, na, d_na, ...rest } = obj;
          return { ...rest, metric: na, label: d_na, metricType: ty };
        });
        setAppliedQueries(kpiData);

        const payload = getPredefinedQuery(
          query,
          durationObj,
          filter,
          breakdown
        );

        setDateFromTo({ from: payload?.q_g[0]?.fr, to: payload?.q_g[0]?.to });

        if (!isCompareQuery) {
          setLoading(true);
          configActionsOnRunningQuery(query);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(payload);
          resetComparisonData();
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }

        const res = await getQueryData(
          activeProject.id,
          payload,
          activeDashboard?.inter_id
        );

        if (query?.inter_id === 1) {
          res.data = transformWidgetResponse(res.data.result || res.data);
        }

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
      activeDashboard?.inter_id,
      activeProject.id,
      configActionsOnRunningQuery,
      dateRange,
      resetComparisonData,
      updateLocalReducer,
      updateResultState
    ]
  );

  useEffect(() => {
    if (location.state && location.state.web_analytics) {
      runKPIQuery(
        location.state.query,
        location.state.query.g_by?.[0],
        filtersData
      );
      setAppliedBreakdown([location.state.query.g_by?.[0]]);

      setNavigatedFromDashboard(location.state.navigatedFromDashboard);
      location.state = undefined;
      // window.history.replaceState(null, '');
    } else if (location.state && location.state.web_analytics) {
      setNavigatedFromAnalyse(location.state.navigatedFromDashboard);
      location.state = undefined;
      window.history.replaceState(null, '');
    } else {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    }
  }, []);

  const handleGranularityChange = useCallback(
    ({ key: frequency }) => {
      resetComparisonData();
      if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_KPI) {
        const appliedDateRange = {
          ...queryOptions.date_range,
          frequency
        };
        setQueryOptions((currState) => ({
          ...currState,
          date_range: appliedDateRange
        }));
        if (queryType === QUERY_TYPE_KPI) {
          runKPIQuery(
            querySaved,
            appliedBreakdown?.[0],
            filtersData,
            appliedDateRange,
            true
          );
        }
      }
    },
    [
      queryOptions.date_range,
      querySaved,
      camp_dateRange,
      dispatch,
      queryType,
      resetComparisonData,
      runKPIQuery
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
        setQueryOptions((currState) => ({
          ...currState,
          date_range: {
            ...currState.date_range,
            ...payload
          }
        }));
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

      runKPIQuery(
        querySaved,
        appliedBreakdown?.[0],
        filtersData,
        appliedDateRange,
        false,
        isCompareDate
      );
    },
    [
      queryType,
      queryOptions.date_range,
      runKPIQuery,
      querySaved,
      appliedBreakdown,
      filtersData
    ]
  );

  useEffect(
    () => () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    },
    [dispatch]
  );

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queriesA];
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
      setQueries(
        queryupdated.map((q) => ({
          ...q,
          key: q.key || generateRandomKey()
        }))
      );
    },
    [queriesA, deleteGroupByForEvent]
  );

  const setExtraOptions = useCallback((options) => {
    setQueryOptions(options);
  }, []);

  const handleBreadCrumbClick = () => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
    setNavigatedFromAnalyse(false);
    setQuerySaved(false);
    updateRequestQuery(null);

    if (queryType === QUERY_TYPE_KPI) {
      setQueries([]);
    }
  };

  const getCurrentSorter = useCallback(() => {
    if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
      return renderedCompRef.current.currentSorter;
    }
    return [];
  }, []);

  const { chartTypes } = coreQueryState;

  const handleChartTypeChange = useCallback(
    ({ key, callUpdateService = true }) => {
      const breakdown = appliedBreakdown;
      const changedKey = getChartChangedKey({
        queryType,
        breakdown
      });

      updateChartTypes({
        ...chartTypes,
        [queryType]: {
          ...chartTypes[queryType],
          [changedKey]: key
        }
      });
    },
    [queryType, updateChartTypes, appliedBreakdown, chartTypes]
  );

  const contextValue = useMemo(
    () => ({
      coreQueryState,
      queryOptions,
      activeKey,
      showResult,
      setNavigatedFromDashboard,
      setNavigatedFromAnalyse,
      resetComparisonData,
      handleCompareWithClick,
      updatePivotConfig,
      queryChange,
      setExtraOptions,
      updateCoreQueryReducer
    }),
    [
      coreQueryState,
      queryOptions,
      activeKey,
      showResult,
      resetComparisonData,
      handleCompareWithClick,
      updatePivotConfig,
      queryChange,
      setExtraOptions,
      setNavigatedFromDashboard,
      setNavigatedFromAnalyse,
      updateCoreQueryReducer
    ]
  );

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Analyse Error'
          subtitle='We are facing trouble loading Analyse. Drop us a message on the in-app chat.'
        />
      }
      onError={FaErrorLog}
    >
      {!showResult && resultState.loading ? <PageSuspenseLoader /> : null}
      <CoreQueryContext.Provider value={contextValue}>
        <ReportHeader
          isFromAnalysisPage={false}
          requestQuery={requestQuery}
          onBreadCrumbClick={handleBreadCrumbClick}
          queryType={queryType}
          queryTitle={querySaved ? querySaved?.d_na : null}
          setQuerySaved={setQuerySaved}
          breakdownType={breakdownType}
          changeTab={setActiveTab}
          activeTab={activeTab}
          getCurrentSorter={getCurrentSorter}
          savedQueryId={querySaved ? querySaved?.inter_id : null}
          querySaved={querySaved}
          breakdown={appliedBreakdown}
          dateFromTo={dateFromTo}
        />

        <div className='mt-24 px-8'>
          <ErrorBoundary
            fallback={
              <FaErrorComp
                size='medium'
                title='Quick board Results Error'
                subtitle='We are facing trouble loading Quick board results. Drop us a message on the in-app chat.'
              />
            }
            onError={FaErrorLog}
          >
            {Number(activeTab) === 1 && (
              <>
                {loading ? (
                  <div className='w-full h-full flex items-center justify-center'>
                    <div className='w-full h-64 flex items-center justify-center'>
                      <Spin size='large' />
                    </div>
                  </div>
                ) : requestQuery ? (
                  <ReportContent
                    breakdownType={breakdownType}
                    runKPIQuery={runKPIQuery}
                    queryType={QUERY_TYPE_KPI}
                    renderedCompRef={renderedCompRef}
                    breakdown={appliedBreakdown}
                    updateAppliedBreakdown={updateAppliedBreakdown}
                    savedQueryId={querySaved ? querySaved?.inter_id : null}
                    handleChartTypeChange={handleChartTypeChange}
                    queryOptions={queryOptions}
                    resultState={resultState}
                    queries={appliedQueries}
                    querySaved={querySaved}
                    handleDurationChange={handleDurationChange}
                    handleGranularityChange={handleGranularityChange}
                    queryTitle={querySaved ? querySaved?.d_na : null}
                    section={REPORT_SECTION}
                    runAttrCmprQuery={null}
                  />
                ) : null}
              </>
            )}
          </ErrorBoundary>
        </div>
      </CoreQueryContext.Provider>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  KPI_config: state.kpi?.config,
  existingQueries: state.queries,
  currentAgent: state.agent.agent_details
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      deleteGroupByForEvent,
      getCampaignConfigData,
      getHubspotContact,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      fetchKPIConfig
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
