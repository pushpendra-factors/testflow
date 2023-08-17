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
import MomentTz from 'Components/MomentTz';
import factorsai from 'factorsai';
import { updateQuery } from 'Reducers/coreQuery/services';
import {
  getDashboardDateRange,
  getSavedAttributionMetrics
} from 'Views/Dashboard/utils';
import { EMPTY_OBJECT } from 'Utils/global';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';
import moment from 'moment';
import {
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
} from 'Reducers/global';
import AttrQueryComposer from './AttrQueryComposer';
import { SVG, FaErrorComp, FaErrorLog } from 'Components/factorsComponents';
import { deleteGroupByForEvent } from 'Reducers/coreQuery/middleware';
import {
  initialState,
  getKPIQuery,
  getKPIQueryAttributionV1,
  DefaultDateRangeFormat,
  getAttributionQuery,
  isComparisonEnabled,
  getAttributionStateFromRequestQuery
} from 'Views/CoreQuery/utils';

import {
  getAttributionsData,
  getAttributionsDataV1
} from 'Reducers/coreQuery/services';
import {
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_ATTRIBUTION,
  EACH_USER_TYPE,
  INITIAL_SESSION_ANALYTICS_SEQ,
  ATTRIBUTION_METRICS,
  apiChartAnnotations,
  presentationObj,
  DefaultChartTypes,
  CHART_TYPE_TABLE,
  QUERY_OPTIONS_DEFAULT_VALUE,
  REPORT_SECTION
} from 'Utils/constants';
import { SHOW_ANALYTICS_RESULT, QUERY_UPDATED } from 'Reducers/types';
import { SET_ATTR_DATE_RANGE } from 'Reducers/coreQuery/actions';
import { CoreQueryContext } from 'Context/CoreQueryContext';
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
  SET_NAVIGATED_FROM_ANALYSE
} from 'Views/CoreQuery/constants';
import { initializeAttributionState } from 'Attribution/state/actions';
import { shouldDataFetch } from 'Utils/dataFormatter';
import { getSavedPivotConfig } from 'Views/CoreQuery/coreQuery.helpers';
import { getChartChangedKey } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import _ from 'lodash';
import AttributionHeader from './AttributionHeader';
import ReportContent from './ReportContent';
import useQuery from 'hooks/useQuery';
import { Button, notification, Spin } from 'antd';
import WeeklyInsights from 'Views/CoreQuery/WeeklyInsights';
import { useHistory } from 'react-router-dom';
import { ATTRIBUTION_ROUTES } from '../../utils/constants';

function CoreQuery({
  activeProject,
  deleteGroupByForEvent,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  initializeAttributionState,
  location,
  currentAgent
}) {
  const { data: savedQueries, loading: QueriesLoading } = useSelector(
    (state) => state.attributionDashboard.attributionQueries
  );
  const { config: kpiConfig } = useSelector((state) => state.kpi);

  const [coreQueryState, localDispatch] = useReducer(
    CoreQueryReducer,
    CORE_QUERY_INITIAL_STATE
  );
  //   const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const queryType = QUERY_TYPE_ATTRIBUTION;
  const [showResult, setShowResult] = useState(false);
  //   const [appliedQueries, setAppliedQueries] = useState([]);
  const appliedQueries = [];
  //   const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const appliedBreakdown = [];
  const [resultState, setResultState] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [querySaved, setQuerySaved] = useState(false);
  //   const [breakdownType, setBreakdownType] = useState(EACH_USER_TYPE);
  const breakdownType = EACH_USER_TYPE;
  const [queriesA, setQueries] = useState([]);
  const [loading, setLoading] = useState(false);
  const renderedCompRef = useRef(null);
  const [queryOpen, setQueryOpen] = useState(true);
  //for tracking if for cetain queryId data is loaded or not
  const [queryDataLoaded, setQueryDataLoaded] = useState('');

  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat },
    group_analysis: 'all'
  });
  const [attributionsState, setAttributionsState] = useState({
    eventGoal: {},
    touchpoint: '',
    models: [],
    tacticOfferType: '',
    linkedEvents: [],
    date_range: {},
    attr_dimensions: [],
    content_groups: [],
    attrQueries: []
  });

  const campaignState = {
    channel: '',
    select_metrics: [],
    filters: [],
    group_by: [],
    date_range: {}
  };

  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS
  ]);

  const [savedReportLoaded, setSavedReportLoaded] = useState(false);

  const dispatch = useDispatch();
  const history = useHistory();
  const {
    eventGoal,
    touchpoint,
    touchpoint_filters,
    tacticOfferType,
    attr_query_type,
    models,
    window,
    linkedEvents,
    attr_dateRange,
    attr_dimensions,
    attrQueries,
    content_groups
  } = useSelector((state) => state.attributionDashboard);

  const { groupBy } = useSelector((state) => state.coreQuery);

  const [activeTab, setActiveTab] = useState(1);

  const routerQuery = useQuery();
  const queryId = routerQuery.get('queryId');

  //   const [dateFromTo, setDateFromTo] = useState({ from: '', to: '' });
  const dateFromTo = { from: '', to: '' };

  const { dashboards } = useSelector((state) => state.dashboard);

  const updateResultState = useCallback((newState) => {
    setResultState(newState);
  }, []);

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
      setQuerySaved(isQuerySaved);
      if (!isQuerySaved) {
        // reset pivot config
        updatePivotConfig({ ...DEFAULT_PIVOT_CONFIG });
        // setNavigatedFromDashboard(false);
        updateSavedQuerySettings(EMPTY_OBJECT);
        setAttributionMetrics([...ATTRIBUTION_METRICS]);
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
      localDispatch({
        type: SET_COMPARISON_SUPPORTED,
        payload: isComparisonEnabled(queryType, queriesA, groupBy, models)
      });
    },
    [
      dispatch,
      groupBy,
      queriesA,
      queryType,
      models,
      savedQueries,
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

  const runAttributionQuery = useCallback(
    async (isQuerySaved, durationObj, isCompareQuery, v1 = true) => {
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
          null,
          models,
          window,
          linkedEvents,
          durationObj,
          tacticOfferType,
          true
        );

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
        } else if (queryOptions.group_analysis === 'salesforce_opportunities') {
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
        // query.query.analyze_type = queryOptions.group_analysis;
        if (v1) {
          query.query.kpi_queries = getKPIQueryAttributionV1(
            attrQueries,
            dtRange,
            { event: [], global: [] },
            queryOptions,
            []
          );
        } else {
          query.query.kpi_query_group = kpiQuery;
        }

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          factorsai.track('RUN-QUERY', {
            email_id: currentAgent?.email,
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
            date_range: { ...durationObj },
            attrQueries
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

          let res;

          if (v1) {
            res = await getAttributionsDataV1(
              activeProject.id,
              query,
              getDashboardConfigs(isQuerySaved),
              true
            );
          } else {
            res = await getAttributionsData(
              activeProject.id,
              query,
              getDashboardConfigs(isQuerySaved),
              true
            );
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
      attrQueries,
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

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        if (!isCompareDate) {
          // set range in reducer only when original date is changed and not the comparisom date
          dispatch({ type: SET_ATTR_DATE_RANGE, payload });
        }
        runAttributionQuery(querySaved, payload, isCompareDate, true);
      }
    },
    [
      queryType,
      querySaved,
      queryOptions.date_range,
      dispatch,
      runAttributionQuery
    ]
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
      setQueries(queryupdated);
    },
    [queriesA]
  );

  const setExtraOptions = useCallback((options) => {
    setQueryOptions(options);
  }, []);

  const handleRunQuery = useCallback(() => {
    runAttributionQuery(false, null, null, true);
    setQueryOpen(false);
  }, [queryType, runAttributionQuery]);

  const getCurrentSorter = useCallback(() => {
    if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
      return renderedCompRef.current.currentSorter;
    }
    return [];
  }, []);

  const { chartTypes } = coreQueryState;

  const savedQueryId = querySaved ? querySaved.id : null;

  const handleChartTypeChange = useCallback(
    ({ key, callUpdateService = true }) => {
      const changedKey = getChartChangedKey({
        queryType,
        appliedBreakdown,
        campaignGroupBy: campaignState.group_by,
        attributionModels: attributionsState.models
      });

      updateChartTypes({
        ...chartTypes,
        [queryType]: {
          ...chartTypes[queryType],
          [changedKey]: key
        }
      });

      if (savedQueryId && callUpdateService) {
        const queryGettingUpdated = savedQueries.find(
          (elem) => elem.id === savedQueryId
        );

        const settings = {
          ...queryGettingUpdated.settings,
          chart: apiChartAnnotations[key]
        };

        const reqBody = {
          title: queryGettingUpdated.title,
          settings
        };

        updateQuery(activeProject.id, savedQueryId, reqBody);

        // #Todo Disabled for now. The query is getting rerun again. Have to figure out a way around it.
        if (!queryType) {
          dispatch({
            type: QUERY_UPDATED,
            queryId: savedQueryId,
            payload: reqBody
          });
        }
      }
    },
    [
      queryType,
      updateChartTypes,
      appliedBreakdown,
      chartTypes,
      campaignState.group_by,
      attributionsState.models,
      savedQueryId,
      savedQueries
    ]
  );

  const contextValue = useMemo(
    () => ({
      coreQueryState,
      attributionMetrics,
      queriesA,
      queryOptions,
      showResult,
      setAttributionMetrics,
      setNavigatedFromDashboard,
      setNavigatedFromAnalyse,
      resetComparisonData,
      handleCompareWithClick,
      updatePivotConfig,
      queryChange,
      setExtraOptions,
      setQueries,
      runAttributionQuery
    }),
    [
      coreQueryState,
      attributionMetrics,
      queriesA,
      queryOptions,
      showResult,
      setAttributionMetrics,
      resetComparisonData,
      handleCompareWithClick,
      updatePivotConfig,
      queryChange,
      setExtraOptions,
      runAttributionQuery
    ]
  );

  useEffect(() => {
    if (querySaved && querySaved?.id && !queryId) {
      history.replace({
        pathname: ATTRIBUTION_ROUTES.report,
        search: `?${new URLSearchParams({
          queryId: querySaved.id
        }).toString()}`
      });
    }
    if (queryId) {
      if (querySaved && querySaved.id !== queryId) {
        history.push({
          pathname: ATTRIBUTION_ROUTES.report,
          search: `?${new URLSearchParams({
            queryId: querySaved.id
          }).toString()}`
        });
      }
    }
  }, [querySaved, queryId]);

  useEffect(() => {
    const handleQueryIdChange = () => {
      if (queryDataLoaded === queryId || querySaved.id == queryId) return;
      const record = savedQueries.find((sq) => sq.id == queryId);
      if (
        !record ||
        !record?.query?.cl ||
        record.query.cl !== QUERY_TYPE_ATTRIBUTION
      ) {
        notification.error({
          message: `Attribution Report with id=${queryId} Not Found`,
          duration: 5
        });
        history.replace({
          pathname: ATTRIBUTION_ROUTES.reports
        });
        return;
      }

      setQueryOpen(false);
      const equivalentQuery = getAttributionStateFromRequestQuery(
        record.query.query,
        attr_dimensions,
        content_groups,
        kpiConfig
      );
      const newDateRange = { attr_dateRange: getDashboardDateRange() };
      const usefulQuery = { ...equivalentQuery, ...newDateRange };
      if (record.settings && record.settings.attributionMetrics) {
        setAttributionMetrics(
          getSavedAttributionMetrics(
            JSON.parse(record.settings.attributionMetrics)
          )
        );
      }
      delete usefulQuery.queryType;
      initializeAttributionState(usefulQuery);

      setQueryOptions((currData) => ({
        ...currData,
        group_analysis: 'all'
      }));
      updateSavedQuerySettings(record.settings || {});
      setSavedReportLoaded({
        queryType: equivalentQuery.queryType,
        queryName: record.title,
        settings: record.settings,
        query_id: record.key || record.id
      });
      setQueryDataLoaded(queryId);
    };
    if (queryId && !QueriesLoading) handleQueryIdChange();
  }, [queryId, savedQueries, QueriesLoading, querySaved]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    if (_.isEmpty(dashboards?.data)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
  }, [activeProject]);

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, []);

  useEffect(() => {
    if (savedReportLoaded) {
      runAttributionQuery({
        id: savedReportLoaded.query_id,
        name: savedReportLoaded.queryName
      });
      setSavedReportLoaded(false);
    }
  }, [savedReportLoaded, runAttributionQuery]);

  useEffect(() => {
    if (location?.state?.navigatedFromDashboard)
      setNavigatedFromDashboard(location.state.navigatedFromDashboard);
    if (location?.state?.navigatedFromAnalyse)
      setNavigatedFromAnalyse(location.state.navigatedFromAnalyse);
  }, [location, setNavigatedFromDashboard, setNavigatedFromAnalyse]);

  if (queryId && QueriesLoading)
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
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
        <AttributionHeader
          isFromAnalysisPage={false}
          requestQuery={requestQuery}
          queryType={QUERY_TYPE_ATTRIBUTION}
          queryTitle={querySaved ? querySaved.name : null}
          setQuerySaved={setQuerySaved}
          breakdownType={breakdownType}
          changeTab={setActiveTab}
          activeTab={activeTab}
          getCurrentSorter={getCurrentSorter}
          savedQueryId={savedQueryId}
          breakdown={appliedBreakdown}
          attributionsState={attributionsState}
          campaignState={campaignState}
          dateFromTo={dateFromTo}
        />

        <div className='mt-24 px-8'>
          <ErrorBoundary
            fallback={
              <FaErrorComp
                size='medium'
                title='Attribution Results Error'
                subtitle='We are facing trouble loading Attribution results. Drop us a message on the in-app chat.'
              />
            }
            onError={FaErrorLog}
          >
            {Number(activeTab) === 1 && (
              <>
                <div
                  className={`query_card_cont ${
                    queryOpen ? `query_card_open` : `query_card_close`
                  }`}
                  onClick={() => !queryOpen && setQueryOpen(true)}
                >
                  <div className='query_composer'>
                    <AttrQueryComposer
                      queryOptions={queryOptions}
                      setQueryOptions={setQueryOptions}
                      runAttributionQuery={handleRunQuery}
                      collapse={showResult}
                      setCollapse={() => setQueryOpen(false)}
                    />
                  </div>
                  <Button size='large' className='query_card_expand'>
                    <SVG name='expand' size={20} />
                    Expand
                  </Button>
                </div>
                {loading ? (
                  <div className='w-full h-full flex items-center justify-center'>
                    <div className='w-full h-64 flex items-center justify-center'>
                      <Spin size='large' />
                    </div>
                  </div>
                ) : requestQuery ? (
                  <ReportContent
                    breakdownType={breakdownType}
                    queryType={QUERY_TYPE_ATTRIBUTION}
                    renderedCompRef={renderedCompRef}
                    breakdown={appliedBreakdown}
                    attributionsState={attributionsState}
                    campaignState={campaignState}
                    savedQueryId={savedQueryId}
                    handleChartTypeChange={handleChartTypeChange}
                    queryOptions={queryOptions}
                    resultState={resultState}
                    queries={appliedQueries}
                    handleDurationChange={handleDurationChange}
                    queryTitle={querySaved ? querySaved.name : null}
                    section={REPORT_SECTION}
                    runAttrCmprQuery={null}
                  />
                ) : null}
              </>
            )}

            {Number(activeTab) === 2 && (
              <WeeklyInsights
                requestQuery={requestQuery}
                queryType={queryType}
                savedQueryId={savedQueryId}
              />
            )}
          </ErrorBoundary>
        </div>
      </CoreQueryContext.Provider>
      {/* create project modal */}
      {/* <NewProject
          visible={showProjectModal}
          handleCancel={() => setShowProjectModal(false)}
        /> */}
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  KPI_config: state.kpi?.config,
  currentAgent: state.agent.agent_details,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      deleteGroupByForEvent,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      initializeAttributionState
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
