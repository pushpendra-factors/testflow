import React, {
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  useState
} from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import get from 'lodash/get';

import cx from 'classnames';

import {
  fetchQueries,
  getEventsData,
  getFunnelData,
  getKPIData,
  updateQuery
} from 'Reducers/coreQuery/services';
import { EMPTY_ARRAY, EMPTY_OBJECT, generateRandomKey } from 'Utils/global';
import AnalysisHeader from 'Views/CoreQuery/AnalysisResultsPage/AnalysisHeader';
import {
  ACTIVE_USERS_CRITERIA,
  apiChartAnnotations,
  DefaultChartTypes,
  EACH_USER_TYPE,
  FREQUENCY_CRITERIA,
  presentationObj,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  REPORT_SECTION,
  REVERSE_USER_TYPES,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  TYPE_EVENTS_OCCURRENCE
} from 'Utils/constants';
import {
  COMPARISON_DATA_FETCHED,
  CORE_QUERY_INITIAL_STATE,
  DEFAULT_PIVOT_CONFIG,
  INITIAL_STATE as INITIAL_RESULT_STATE,
  SET_COMPARE_DURATION,
  SET_COMPARISON_SUPPORTED,
  SET_SAVED_QUERY_SETTINGS,
  UPDATE_CHART_TYPES,
  UPDATE_CORE_QUERY_REDUCER,
  UPDATE_PIVOT_CONFIG
} from 'Views/CoreQuery/constants';
import {
  formatApiData,
  getFunnelQuery,
  getKPIStateFromRequestQuery,
  getQuery,
  getStateQueryFromRequestQuery,
  isComparisonEnabled
} from 'Views/CoreQuery/utils';
import { getQueryOptionsFromEquivalentQuery } from './utils';
import { CoreQueryState, QueryParams, ResultState } from './types';
import { QUERY_UPDATED, SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import CoreQueryReducer from 'Views/CoreQuery/CoreQueryReducer';
import {
  deleteGroupByEventAction,
  INITIALIZE_GROUPBY
} from 'Reducers/coreQuery/actions';
import { ErrorBoundary } from 'react-error-boundary';
import {
  FaErrorComp,
  FaErrorLog,
  SVG,
  Text
} from 'Components/factorsComponents';
import QueryComposer from 'Components/QueryComposer';
import { Button, Spin } from 'antd';
import logger from 'Utils/logger';
import ReportContent from 'Views/CoreQuery/AnalysisResultsPage/ReportContent';
import { getDashboardDateRange } from 'Views/Dashboard/utils';
import moment from 'moment';
import {
  SET_PERFORMANCE_CRITERIA,
  SET_SHOW_CRITERIA
} from 'Reducers/analyticsQuery';
import { getValidGranularityOptions } from 'Utils/dataFormatter';
import MomentTz from 'Components/MomentTz';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';
import SaveQuery from 'Components/SaveQuery';
import { getChartChangedKey } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import _ from 'lodash';
import KPIComposer from 'Components/KPIComposer';
import { CoreQueryContext } from 'Context/CoreQueryContext';

const CoreQuery = () => {
  // Query params
  const { query_id, query_type } = useParams<QueryParams>();

  const [activeTab, setActiveTab] = useState(1);

  // Redux States
  const { active_project } = useSelector((state: any) => state.global);
  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state: any) => state.analyticsQuery);
  const { models, eventNames, groupBy } = useSelector(
    (state: any) => state.coreQuery
  );
  const savedQueries = useSelector((state: any) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );

  const [savedQueryModal, setSavedQueryModal] = useState(false);

  const [loading, setLoading] = useState(true);

  const [selectedMainCategory, setSelectedMainCategory] = useState(false);
  const [KPIConfigProps, setKPIConfigProps] = useState([]);

  // Local states
  const [coreQueryState, setCoreQueryState] = useState<CoreQueryState>(
    new CoreQueryState()
  );
  const [queryOpen, setQueryOpen] = useState(false);

  const location = useLocation();
  const history = useHistory();
  const dispatch = useDispatch();
  const [coreQueryReducerState, localDispatch] = useReducer(
    CoreQueryReducer,
    CORE_QUERY_INITIAL_STATE
  );
  const renderedCompRef = useRef<any>(null);

  const getCurrentSorter = useCallback(() => {
    if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
      return renderedCompRef.current.currentSorter;
    }
    return [];
  }, []);

  // Use Effects
  useEffect(() => {
    if (!savedQueries || !savedQueries.length) {
      fetchQueries(active_project.id);
    }
  }, [savedQueries]);

  useEffect(() => {
    if (
      query_id &&
      query_id != '' &&
      query_type &&
      savedQueries?.length &&
      !location?.state?.navigatedResultState
    ) {
      runEventsQueryFromUrl();
    }
    if (location?.state?.navigatedResultState) {
      const queryToAdd = getQueryFromHashId();
      if (query_type === QUERY_TYPE_EVENT) {
        createStateFromResult(queryToAdd);
      } else if (query_type === QUERY_TYPE_FUNNEL) {
        createFunnelStateFromResult(queryToAdd);
      } else if (query_type === QUERY_TYPE_KPI) {
        createKPIStateFromResult(queryToAdd);
      }
    }
  }, [query_id, query_type, savedQueries, location]);

  useEffect(() => {
    if (coreQueryState.resultState?.data) {
      const qState = _.cloneDeep(coreQueryState);
      qState.queryOptions = {
        ...qState.queryOptions,
        groupBy
      };
      // setAppliedBreakdowns(groupBy, qState);
      setCoreQueryState(qState);
    }
  }, [groupBy]);

  useEffect(() => {
    setKPIConfigProps(findKPIitem(selectedMainCategory?.group));
  }, [selectedMainCategory]);

  const KPI_config = useSelector((state) => state.kpi.config);

  const findKPIitem = useCallback(
    (groupName) => {
      const KPIlist = KPI_config || [];
      const selGroup = KPIlist.find(
        (item) => item.display_category === groupName
      );

      const DDvalues = selGroup?.properties?.map((item) => {
        if (item == null) return null;
        const ddName = item.display_name ? item.display_name : item.name;
        const ddtype =
          selGroup?.category === 'channels' ||
          selGroup?.category === 'custom_channels'
            ? item.object_type
            : item.entity
              ? item.entity
              : item.object_type;
        return [ddName, item.name, item.data_type, ddtype, item.category];
      });
      return DDvalues;
    },
    [KPI_config]
  );

  const getQueryFromHashId = () =>
    savedQueries?.find((quer: any) => quer.id_text === query_id);

  const updateCoreQueryReducer = useCallback((payload) => {
    localDispatch({
      type: UPDATE_CORE_QUERY_REDUCER,
      payload
    });
  }, []);

  const updateEventFunnelsState = useCallback(
    (
      equivalentQuery,
      navigatedFromDashboard = false,
      qState: CoreQueryState
    ) => {
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
      let queryDateRange;
      if (navigatedFromDashboard && location?.state?.navigatedResultState) {
        queryDateRange = { date_range: dashboardDateRange };
      } else queryDateRange = { date_range: savedDateRange };

      const queryOpts = {
        ...coreQueryState.queryOptions,
        session_analytics_seq: equivalentQuery.session_analytics_seq,
        groupBy: {
          global: [...equivalentQuery.breakdown.global],
          event: [...equivalentQuery.breakdown.event]
        },
        globalFilters: equivalentQuery.globalFilters,
        group_analysis: equivalentQuery.groupAnalysis,
        ...queryDateRange,
        events_condition: equivalentQuery.eventsCondition
      };

      qState.setItem('queryOptions', queryOpts);

      // setCoreQueryState(coreQueryState);
      //   setQueryOptions((currData) => {
      //     let queryDateRange = {};

      //     let queryOpts = {};
      //     queryOpts = {
      //       ...currData,
      //       session_analytics_seq: equivalentQuery.session_analytics_seq,
      //       groupBy: [
      //         ...equivalentQuery.breakdown.global,
      //         ...equivalentQuery.breakdown.event
      //       ],
      //       globalFilters: equivalentQuery.globalFilters,
      //       group_analysis: equivalentQuery.groupAnalysis,
      //       ...queryDateRange,
      //       events_condition: equivalentQuery.eventsCondition
      //     };
      //     return queryOpts;
      //   });
    },
    [dispatch]
  );

  const updateResultFromSavedQuery = (res: any, qState: CoreQueryState) => {
    const data = res.data.result || res.data;
    let resultSt;
    if (result_criteria === TOTAL_EVENTS_CRITERIA) {
      resultSt = {
        ...INITIAL_RESULT_STATE,
        data: formatApiData(data.result_group[0], data.result_group[1]),
        apiCallStatus: res.status
      };
    } else if (result_criteria === TOTAL_USERS_CRITERIA) {
      if (user_type === EACH_USER_TYPE) {
        resultSt = {
          ...INITIAL_RESULT_STATE,
          data: formatApiData(data.result_group[0], data.result_group[1]),
          apiCallStatus: res.status
        };
      } else {
        resultSt = {
          ...INITIAL_RESULT_STATE,
          data: data.result_group[0],
          apiCallStatus: res.status
        };
      }
    }
    qState.setItem('resultState', resultSt);
  };

  const setAppliedBreakdowns = (breakdown: any, qState: CoreQueryState) => {
    const newAppliedBreakdown = [...breakdown.event, ...breakdown.global];
    qState.appliedBreakdown = newAppliedBreakdown;
  };

  const createKPIStateFromResult = (
    queryToAdd: (typeof savedQueries)[0],
    resultState?: ResultState | null,
    dateRange?: any | null,
    isSavedQuery = true
  ) => {
    const equivalentQuery = getKPIStateFromRequestQuery(queryToAdd?.query);
    const queryState = new CoreQueryState();
    queryState.queryType = QUERY_TYPE_KPI;
    if (isSavedQuery) {
      queryState.querySaved = { name: queryToAdd.title, id: queryToAdd.id };
    } else {
      queryState.querySaved = {};
    }

    queryState.requestQuery = queryToAdd?.query;
    queryState.showResult = true;
    queryState.loading = false;
    setLoading(false);
    queryState.queries = equivalentQuery.events;
    if (queryState.queryType === QUERY_TYPE_KPI) {
      queryState.appliedQueries = queryState.queries.map((q) => {
        const category = KPI_config.find(
          (elem) =>
            elem.category === q.category && q.group === elem.display_category
        );
        const metric = category?.metrics.find((m) => m.name === q.metric);
        return {
          ...q,
          metricType: metric?.type != null ? metric.type : q.metricType
        };
      });
      // updateAppliedBreakdown();
    }
    queryState.queryOptions = getQueryOptionsFromEquivalentQuery(
      queryState.queryOptions,
      equivalentQuery
    );

    if (dateRange) {
      queryState.queryOptions.date_range = dateRange;
    }

    queryState.breakdownType = REVERSE_USER_TYPES[queryState.requestQuery.ec];

    // if (queryState.requestQuery) {
    //   updateEventFunnelsState(
    //     equivalentQuery,
    //     location?.state?.navigatedFromDashboard,
    //     queryState
    //   );
    //   if (queryState.requestQuery.length === 1) {
    //     dispatch({
    //       type: SET_PERFORMANCE_CRITERIA,
    //       payload: REVERSE_USER_TYPES[queryState.requestQuery.ec]
    //     });
    //     dispatch({
    //       type: SET_SHOW_CRITERIA,
    //       payload: TOTAL_USERS_CRITERIA
    //     });
    //   } else {
    //     dispatch({
    //       type: SET_PERFORMANCE_CRITERIA,
    //       payload: EACH_USER_TYPE
    //     });
    //     if (queryState.requestQuery.length === 2) {
    //       dispatch({
    //         type: SET_SHOW_CRITERIA,
    //         payload:
    //           queryState.requestQuery.ty === TYPE_EVENTS_OCCURRENCE
    //             ? TOTAL_EVENTS_CRITERIA
    //             : TOTAL_USERS_CRITERIA
    //       });
    //     }
    //     // else if (queryState.requestQuery.query.length === 3) {
    //     //   dispatch({
    //     //     type: SET_SHOW_CRITERIA,
    //     //     payload: ACTIVE_USERS_CRITERIA
    //     //   });
    //     // }
    //     else {
    //       dispatch({
    //         type: SET_SHOW_CRITERIA,
    //         payload: FREQUENCY_CRITERIA
    //       });
    //     }
    //   }
    // }

    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    dispatch({
      type: SET_COMPARISON_SUPPORTED,
      payload: isComparisonEnabled(
        queryState.queryType,
        equivalentQuery.events,
        equivalentQuery.breakdown,
        models
      )
    });

    setAppliedBreakdowns(equivalentQuery.breakdown, queryState);

    // updateAppliedBreakdown();
    dispatch({
      type: UPDATE_PIVOT_CONFIG,
      payload: { ...DEFAULT_PIVOT_CONFIG }
    });
    dispatch({ type: SET_SAVED_QUERY_SETTINGS, payload: EMPTY_OBJECT });

    if (resultState) {
      queryState.resultState = {
        ...INITIAL_RESULT_STATE,
        data: resultState.data.result || resultState.data,
        status: resultState.status
      };
      // updateResultState();
      // updateResultFromSavedQuery(resultState, queryState);
    }
    if (location?.state?.navigatedResultState && !resultState) {
      queryState.resultState = location.state.navigatedResultState;
    }

    setCoreQueryState(queryState);
  };

  const createFunnelStateFromResult = (
    queryToAdd: (typeof savedQueries)[0],
    resultState?: ResultState | null,
    dateRange?: any | null,
    isSavedQuery = true
  ) => {
    const equivalentQuery = getStateQueryFromRequestQuery(queryToAdd?.query);
    const queryState = new CoreQueryState();
    queryState.queryType = QUERY_TYPE_FUNNEL;
    if (isSavedQuery) {
      queryState.querySaved = { name: queryToAdd.title, id: queryToAdd.id };
    } else {
      queryState.querySaved = {};
    }

    queryState.requestQuery = queryToAdd?.query;
    queryState.showResult = true;
    queryState.loading = false;
    setLoading(false);
    queryState.queries = equivalentQuery.events;
    queryState.appliedQueries = equivalentQuery.events.map((elem: any) =>
      elem.alias ? elem.alias : elem.label
    );
    queryState.queryOptions = getQueryOptionsFromEquivalentQuery(
      queryState.queryOptions,
      equivalentQuery
    );

    if (dateRange) {
      queryState.queryOptions.date_range = dateRange;
    }

    queryState.breakdownType = REVERSE_USER_TYPES[queryState.requestQuery.ec];

    if (queryState.requestQuery) {
      updateEventFunnelsState(
        equivalentQuery,
        location?.state?.navigatedFromDashboard,
        queryState
      );
      if (queryState.requestQuery.length === 1) {
        dispatch({
          type: SET_PERFORMANCE_CRITERIA,
          payload: REVERSE_USER_TYPES[queryState.requestQuery.ec]
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
        if (queryState.requestQuery.length === 2) {
          dispatch({
            type: SET_SHOW_CRITERIA,
            payload:
              queryState.requestQuery.ty === TYPE_EVENTS_OCCURRENCE
                ? TOTAL_EVENTS_CRITERIA
                : TOTAL_USERS_CRITERIA
          });
        }
        // else if (queryState.requestQuery.query.length === 3) {
        //   dispatch({
        //     type: SET_SHOW_CRITERIA,
        //     payload: ACTIVE_USERS_CRITERIA
        //   });
        // }
        else {
          dispatch({
            type: SET_SHOW_CRITERIA,
            payload: FREQUENCY_CRITERIA
          });
        }
      }
    }

    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    dispatch({
      type: SET_COMPARISON_SUPPORTED,
      payload: isComparisonEnabled(
        queryState.queryType,
        equivalentQuery.events,
        equivalentQuery.breakdown,
        models
      )
    });

    setAppliedBreakdowns(equivalentQuery.breakdown, queryState);

    // updateAppliedBreakdown();
    dispatch({
      type: UPDATE_PIVOT_CONFIG,
      payload: { ...DEFAULT_PIVOT_CONFIG }
    });

    dispatch({ type: SET_SAVED_QUERY_SETTINGS, payload: EMPTY_OBJECT });

    if (resultState) {
      queryState.resultState = {
        ...INITIAL_RESULT_STATE,
        data: resultState.data.result || resultState.data,
        status: resultState.status
      };
      // updateResultState();
      // updateResultFromSavedQuery(resultState, queryState);
    }
    if (location?.state?.navigatedResultState && !resultState) {
      queryState.resultState = location.state.navigatedResultState;
    }

    setCoreQueryState(queryState);
  };

  const createStateFromResult = (
    queryToAdd: (typeof savedQueries)[0],
    resultState?: ResultState | null
  ) => {
    const equivalentQuery = getStateQueryFromRequestQuery(
      queryToAdd?.query?.query_group[0]
    );
    const queryState = new CoreQueryState();
    queryState.queryType = QUERY_TYPE_EVENT;
    queryState.querySaved = { name: queryToAdd.title, id: queryToAdd.id };
    queryState.requestQuery = queryToAdd?.query?.query_group;
    queryState.showResult = true;
    queryState.loading = false;
    setLoading(false);
    queryState.queries = equivalentQuery.events;
    queryState.appliedQueries = equivalentQuery.events.map((elem: any) =>
      elem.alias ? elem.alias : elem.label
    );
    queryState.queryOptions = getQueryOptionsFromEquivalentQuery(
      queryState.queryOptions,
      equivalentQuery
    );
    queryState.breakdownType =
      REVERSE_USER_TYPES[queryState.requestQuery[0].ec];

    if (queryState.requestQuery) {
      updateEventFunnelsState(
        equivalentQuery,
        location?.state?.navigatedFromDashboard,
        queryState
      );
      if (queryState.requestQuery.length === 1) {
        dispatch({
          type: SET_PERFORMANCE_CRITERIA,
          payload: REVERSE_USER_TYPES[queryState.requestQuery[0].ec]
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
        if (queryState.requestQuery.length === 2) {
          dispatch({
            type: SET_SHOW_CRITERIA,
            payload:
              queryState.requestQuery[0].ty === TYPE_EVENTS_OCCURRENCE
                ? TOTAL_EVENTS_CRITERIA
                : TOTAL_USERS_CRITERIA
          });
        } else if (queryState.requestQuery.query.length === 3) {
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
    }

    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    dispatch({
      type: SET_COMPARISON_SUPPORTED,
      payload: isComparisonEnabled(
        queryState.queryType,
        equivalentQuery.events,
        equivalentQuery.breakdown,
        models
      )
    });

    setAppliedBreakdowns(equivalentQuery.breakdown, queryState);

    // updateAppliedBreakdown();
    dispatch({
      type: UPDATE_PIVOT_CONFIG,
      payload: { ...DEFAULT_PIVOT_CONFIG }
    });
    dispatch({ type: SET_SAVED_QUERY_SETTINGS, payload: EMPTY_OBJECT });

    if (resultState) {
      updateResultFromSavedQuery(resultState, queryState);
    }
    if (location?.state?.navigatedResultState && !resultState) {
      queryState.resultState = location.state.navigatedResultState;
    }

    setCoreQueryState(queryState);
  };

  const runEventsQueryFromUrl = () => {
    const queryToAdd = getQueryFromHashId();
    if (queryToAdd) {
      // updateResultState({ ...initialState, loading: true });
      // dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
      if (query_type === QUERY_TYPE_FUNNEL) {
        getFunnelData(active_project.id, null, null, false, query_id).then(
          (res) => {
            createFunnelStateFromResult(queryToAdd, res);
          },
          (err) => {
            logger.error(err);
          }
        );
      } else if (query_type === QUERY_TYPE_EVENT) {
        getEventsData(active_project.id, null, null, false, query_id).then(
          (res) => {
            createStateFromResult(queryToAdd, res);
          },
          (err) => {
            logger.error(err);
          }
        );
      } else if (query_type === QUERY_TYPE_KPI) {
        getKPIData(active_project.id, null, null, false, query_id).then(
          (res) => {
            createKPIStateFromResult(queryToAdd, res);
          },
          (err) => {
            logger.error(err);
          }
        );
      }
    }
  };

  const arrayMapper = coreQueryState.appliedQueries.map((q, index) => ({
    eventName: q,
    index,
    mapper: `event${index + 1}`,
    displayName: eventNames[q] ? eventNames[q] : q
  }));

  const renderQueryComposerNew = () => (
    <CoreQueryContext.Provider
      value={{
        coreQueryState: coreQueryReducerState,
        updateCoreQueryReducer: updateCoreQueryReducer
      }}
    >
      <div
        className={`query_card_cont ${
          queryOpen ? `query_card_open` : `query_card_close`
        }`}
        onClick={() => !queryOpen && setQueryOpen(true)}
      >
        <div className='query_composer'>{renderComposer()}</div>
        <Button size='large' className='query_card_expand'>
          <SVG name='expand' size={20} />
          Expand
        </Button>
      </div>
    </CoreQueryContext.Provider>
  );

  const updateLocalReducer = useCallback((type, payload) => {
    localDispatch({ type, payload });
  }, []);

  const configActionsOnRunningQuery = (
    isQuerySaved = false,
    qState: CoreQueryState
  ) => {
    setQueryOpen(false);
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    // setQuerySaved(isQuerySaved);
    if (!isQuerySaved) {
      // reset pivot config
      updateLocalReducer(UPDATE_PIVOT_CONFIG, { ...DEFAULT_PIVOT_CONFIG });
      updateLocalReducer(SET_SAVED_QUERY_SETTINGS, EMPTY_OBJECT);
    }

    updateLocalReducer(
      SET_COMPARISON_SUPPORTED,
      isComparisonEnabled(
        coreQueryState.queryType,
        coreQueryState.queries,
        coreQueryState.queryOptions.groupBy,
        models
      )
    );

    if (
      coreQueryState.queryType === QUERY_TYPE_FUNNEL ||
      coreQueryState.queryType === QUERY_TYPE_EVENT
    ) {
      qState.appliedQueries = coreQueryState.queries.map((elem) =>
        elem.alias ? elem.alias : elem.label
      );
    }
  };

  const runQuery = async (dateRange: any = undefined) => {
    try {
      let durationObj;
      const qState = _.cloneDeep(coreQueryState);

      if (!dateRange) {
        durationObj = coreQueryState.queryOptions.date_range;
      } else {
        durationObj = dateRange;
      }
      const query = getQuery(
        coreQueryState.queryOptions.groupBy,
        coreQueryState.queries,
        result_criteria,
        user_type,
        durationObj,
        coreQueryState.queryOptions.globalFilters,
        coreQueryState.queryOptions.group_analysis
      );

      // if (!isQuerySaved) {
      //   // Factors RUN_QUERY tracking
      //   factorsai.track('RUN-QUERY', {
      //     email_id: currentAgent?.email,
      //     query_type: QUERY_TYPE_EVENT,
      //     project_id: activeProject?.id,
      //     project_name: activeProject?.name
      //   });
      // }

      //if (!isCompareQuery) {

      setAppliedBreakdowns(coreQueryState.queryOptions?.groupBy, qState);
      qState.showResult = true;
      qState.loading = true;
      setLoading(true);
      configActionsOnRunningQuery(false, qState);
      qState.requestQuery = query;
      qState.resultState = { ...qState.resultState, loading: true };
      qState.querySaved = {};
      setCoreQueryState(qState);
      // resetComparisonData();
      //}

      const res: any = await getEventsData(
        active_project.id,
        query,
        null, // we need to call fresh query when granularity is changed
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
      }

      if (dateRange) {
        qState.queryOptions.date_range = dateRange;
      }

      qState.loading = false;
      setLoading(false);
      qState.resultState = {
        ...qState.resultState,
        data: resultantData,
        loading: false,
        apiCallStatus: res.status
      };
      setCoreQueryState(qState);
      // if (isCompareQuery) {
      //   updateLocalReducer(COMPARISON_DATA_FETCHED, resultantData);
      // } else {
      //   setLoading(false);
      //   updateResultState({
      //     ...initialState,
      //     data: resultantData,
      //     status: res.status
      //   });
      // }
    } catch (err) {
      logger.error(err);
      const qState = { ...coreQueryState };
      qState.loading = false;
      setLoading(false);
      qState.resultState = {
        ...qState.resultState,
        loading: false,
        error: true
      };
      setCoreQueryState(qState);
    }
  };

  const handleDurationChange = (dates: any, isCompareDate: boolean = false) => {
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

    frequency = getValidGranularityOptions({ from, to })[0];

    const startDate = moment(from).startOf('day').utc().unix() * 1000;
    const endDate = moment(to).endOf('day').utc().unix() * 1000 + 1000;
    const daysDiff = moment(endDate).diff(startDate, 'days');
    if (daysDiff > 1) {
      frequency =
        coreQueryState.queryOptions.date_range.frequency === 'hour' ||
        frequency === 'hour'
          ? 'date'
          : coreQueryState.queryOptions.date_range.frequency;
    } else frequency = 'hour';

    const payload = {
      from: MomentTz(from).startOf('day'),
      to: MomentTz(to).endOf('day'),
      frequency,
      dateType
    };

    const qState = _.cloneDeep(coreQueryState);

    if (!isCompareDate) {
      qState.queryOptions = {
        ...qState.queryOptions,
        date_range: {
          ...qState.queryOptions.date_range,
          ...payload
        }
      };
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
      ...qState.queryOptions.date_range,
      ...payload
    };

    setCoreQueryState(qState);

    if (qState.queryType === QUERY_TYPE_EVENT) {
      runQuery(qState.queryOptions.date_range);
    } else if (qState.queryType === QUERY_TYPE_FUNNEL) {
      runFunnelQuery(false, qState.queryOptions.date_range, isCompareDate);
    }
  };

  const handleRunQuery = () => {
    switch (coreQueryState.queryType) {
      case QUERY_TYPE_EVENT: {
        runQuery();
        break;
      }
      default: {
        return false;
      }
    }
  };

  const queryChange = (
    newEvent: any,
    index: number,
    changeType: string = 'add',
    flag = null
  ) => {
    const queryupdated = [...coreQueryState.queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
          deleteGroupByEventAction(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else if (changeType === 'filters_updated') {
        // dont remove group by if filter is changed
        queryupdated[index] = newEvent;
      } else {
        deleteGroupByEventAction(newEvent, index);
        queryupdated.splice(index, 1);
      }
    } else {
      if (flag) {
        Object.assign(newEvent, { pageViewVal: flag });
      }
      queryupdated.push(newEvent);
    }
    setQueries(
      queryupdated.map((q) => {
        return {
          ...q,
          key: q.key || generateRandomKey()
        };
      })
    );
  };

  const setQueries = useCallback(
    (q: any[]) => {
      const qState = _.cloneDeep(coreQueryState);
      qState.queries = q;
      setCoreQueryState(qState);
    },
    [setCoreQueryState, coreQueryState]
  );

  const setQueryOptions = (opts: {} | any) => {
    const qState = _.cloneDeep(coreQueryState);

    qState.queryOptions = opts;
    if (opts?.globalFilters) {
      qState.queryOptions.globalFilters = opts.globalFilters;
    }
    setCoreQueryState(qState);
  };

  const handleGranularityChange = useCallback(
    ({ key: frequency }) => {
      // resetComparisonData();
      if (
        coreQueryState.queryType === QUERY_TYPE_EVENT ||
        coreQueryState.queryType === QUERY_TYPE_KPI
      ) {
        const qState = _.cloneDeep(coreQueryState);
        const appliedDateRange = {
          ...coreQueryState.queryOptions.date_range,
          frequency
        };
        qState.queryOptions = {
          ...qState.queryOptions,
          date_range: appliedDateRange
        };
        setCoreQueryState(qState);
        // setQueryOptions((currState) => ({
        //   ...currState,
        //   date_range: appliedDateRange
        // }));
        if (coreQueryState.queryType === QUERY_TYPE_EVENT) {
          runQuery(appliedDateRange);
        }
        // if (queryType === QUERY_TYPE_KPI) {
        //   runKPIQuery(querySaved, appliedDateRange, true);
        // }
      }
    },
    [coreQueryState.queryOptions, runQuery]
  );

  const updateChartTypes = useCallback(
    (payload) => {
      updateLocalReducer(UPDATE_CHART_TYPES, payload);
    },
    [updateLocalReducer]
  );

  const handleChartTypeChange = useCallback(
    ({ key, callUpdateService = true }) => {
      //#TODO fix

      // console.log(coreQueryReducerState);
      const qType = coreQueryState.queryType;
      const changedKey = getChartChangedKey({
        queryType: coreQueryState.queryType,
        breakdown: coreQueryState.appliedBreakdown,
        campaignGroupBy: {},
        attributionModels: []
      });

      updateChartTypes({
        ...DefaultChartTypes,
        [qType]: {
          ...DefaultChartTypes[qType as keyof typeof DefaultChartTypes],
          [changedKey]: key
        }
      });

      if (coreQueryState.querySaved.id && callUpdateService) {
        const queryGettingUpdated = savedQueries.find(
          (elem: any) => elem.id === coreQueryState.querySaved.id
        );

        const settings = {
          ...queryGettingUpdated.settings,
          chart: apiChartAnnotations[key as keyof typeof apiChartAnnotations]
        };

        const reqBody = {
          title: queryGettingUpdated.title,
          settings
        };

        updateQuery(active_project.id, coreQueryState.querySaved.id, reqBody);

        // #Todo Disabled for now. The query is getting rerun again. Have to figure out a way around it.
        if (!qType) {
          dispatch({
            type: QUERY_UPDATED,
            queryId: coreQueryState.querySaved.id,
            payload: reqBody
          });
        }
      }
    },
    [
      coreQueryState.queryType,
      coreQueryState.appliedBreakdown,
      coreQueryReducerState
    ]
  );

  const runFunnelQuery = useCallback(
    async (isQuerySaved, dateRange, isCompareQuery) => {
      try {
        let durationObj;
        const qState = _.cloneDeep(coreQueryState);

        if (!dateRange) {
          durationObj = coreQueryState.queryOptions.date_range;
        } else {
          durationObj = dateRange;
        }

        const query = getFunnelQuery(
          coreQueryState.queryOptions.groupBy,
          coreQueryState.queries,
          coreQueryState.queryOptions.session_analytics_seq,
          durationObj,
          coreQueryState.queryOptions.globalFilters,
          coreQueryState.queryOptions.events_condition,
          coreQueryState.queryOptions.group_analysis,
          coreQueryReducerState.funnelConversionDurationNumber,
          coreQueryReducerState.funnelConversionDurationUnit
        );

        if (!isQuerySaved) {
          // Factors RUN_QUERY tracking
          // factorsai.track('RUN-QUERY', {
          //   email_id: currentAgent?.email,
          //   query_type: QUERY_TYPE_FUNNEL,
          //   project_id: activeProject?.id,
          //   project_name: activeProject?.name
          // });
        }

        if (!isCompareQuery) {
          setLoading(true);
          configActionsOnRunningQuery(isQuerySaved, qState);
          setAppliedBreakdowns(coreQueryState.queryOptions?.groupBy, qState);
          qState.showResult = true;
          qState.loading = true;
          setLoading(true);
          // configActionsOnRunningQuery(false, qState);
          qState.requestQuery = query;
          qState.resultState = { ...qState.resultState, loading: true };
          qState.querySaved = {};
          setCoreQueryState(qState);
        }

        const res = await getFunnelData(active_project.id, query, null, true);

        let queryToAdd = getQueryFromHashId();
        if (!isQuerySaved) {
          queryToAdd.query = query;
        }

        createFunnelStateFromResult(queryToAdd, res, dateRange, false);
      } catch (err) {
        logger.error(err);
        setLoading(false);
        // updateResultState({ ...initialState, error: true, status: err.status });
      }
    },
    [
      coreQueryState,
      coreQueryReducerState.funnelConversionDurationNumber,
      coreQueryReducerState.funnelConversionDurationUnit
    ]
  );

  const renderComposer = () => {
    if (
      coreQueryState.queryType === QUERY_TYPE_FUNNEL ||
      coreQueryState.queryType === QUERY_TYPE_EVENT
    ) {
      return (
        <QueryComposer
          queries={coreQueryState.queries}
          setQueries={setQueries}
          runQuery={handleRunQuery}
          eventChange={queryChange}
          queryType={coreQueryState.queryType}
          queryOptions={coreQueryState.queryOptions}
          setQueryOptions={setQueryOptions}
          runFunnelQuery={runFunnelQuery}
          collapse={coreQueryState.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }
    if (coreQueryState.queryType === QUERY_TYPE_KPI) {
      return (
        <KPIComposer
          queries={coreQueryState.queries}
          setQueries={setQueries}
          eventChange={queryChange}
          queryType={coreQueryState.queryType}
          queryOptions={coreQueryState.queryOptions}
          setQueryOptions={setQueryOptions}
          handleRunQuery={handleRunQuery}
          selectedMainCategory={selectedMainCategory}
          setSelectedMainCategory={setSelectedMainCategory}
          KPIConfigProps={KPIConfigProps}
          setKPIConfigProps={setKPIConfigProps}
        />
      );
    }
  };

  const renderSpinner = () => {
    return (
      <div className='w-full h-64 flex items-center justify-center'>
        <Spin size='large' />
      </div>
    );
  };

  const renderSaveQueryComp = () => (
    <SaveQuery
      queryType={coreQueryState.queryType}
      requestQuery={coreQueryState.requestQuery}
      queryTitle={
        coreQueryState.querySaved ? coreQueryState.querySaved.name : null
      }
      setQuerySaved={(v: any) => coreQueryState.setItem('querySaved', v)}
      getCurrentSorter={getCurrentSorter}
      savedQueryId={
        coreQueryState.querySaved ? coreQueryState.querySaved.id : null
      }
      breakdown={coreQueryState.appliedBreakdown}
      dateFromTo={{
        from:
          coreQueryState.queryOptions.date_range.from ||
          coreQueryState.requestQuery?.fr,
        to:
          coreQueryState.queryOptions.date_range.to ||
          coreQueryState.requestQuery?.to
      }}
      attributionsState={undefined}
      campaignState={undefined}
      showSaveQueryModal={savedQueryModal}
      setShowSaveQueryModal={setSavedQueryModal}
      showUpdateQuery={undefined}
    />
  );

  const handleCloseDashboardQuery = () => {
    history.push({
      pathname: '/',
      state: {
        dashboardWidgetId: coreQueryReducerState.navigatedFromDashboard.id
      }
    });
    // handleBreadCrumbClick();
  };

  const renderEmptyHeader = () => {
    return (
      <div
        id='app-header'
        className={cx('bg-white z-50 flex-col  px-8 w-full', {
          fixed: true
        })}
        style={{
          borderBottom: true ? '1px solid lightgray' : 'none'
        }}
      >
        <div className='items-center flex justify-between w-full pt-3 pb-3'>
          <div
            role='button'
            tabIndex={0}
            className='flex items-center cursor-pointer'
          >
            <Button
              size='large'
              type='text'
              onClick={() => {
                history.push('/');
              }}
              icon={<SVG size={32} name='Brand' />}
            />
            <Text
              type='title'
              level={5}
              weight='bold'
              extraClass='m-0 mt-1'
              lineHeight='small'
            >
              {coreQueryState.querySaved
                ? `Reports / ${coreQueryState.queryType} / ${coreQueryState.querySaved.name}`
                : `Reports / ${
                    coreQueryState.queryType
                  } / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
            </Text>
          </div>

          <div className='flex items-center gap-x-2'>
            <div className='pr-2 border-r'>{renderSaveQueryComp()}</div>
            <Button
              size='large'
              type='default'
              onClick={handleCloseDashboardQuery}
            >
              Close
            </Button>
          </div>
        </div>
      </div>
    );
  };

  const renderAnalysisHeader = () => (
    <AnalysisHeader
      isFromAnalysisPage={false}
      requestQuery={coreQueryState.requestQuery}
      onBreadCrumbClick={() => {
        console.log('breadcrumb click');
      }}
      queryType={coreQueryState.queryType}
      queryTitle={
        coreQueryState.querySaved ? coreQueryState.querySaved?.name : null
      }
      setQuerySaved={(v: any) => coreQueryState.setItem('querySaved', v)}
      breakdownType={coreQueryState.breakdownType}
      changeTab={(v: any) => coreQueryState.setItem('activeTab', v)}
      activeTab={coreQueryState.activeTab}
      getCurrentSorter={getCurrentSorter}
      savedQueryId={
        coreQueryState.querySaved ? coreQueryState.querySaved.id : null
      }
      breakdown={coreQueryState.appliedBreakdown}
      dateFromTo={{
        from:
          coreQueryState.queryOptions.date_range.from ||
          coreQueryState.requestQuery?.fr,
        to:
          coreQueryState.queryOptions.date_range.to ||
          coreQueryState.requestQuery?.to
      }}
    />
  );

  const renderReportContent = () => {
    // if (coreQueryState.queryType === QUERY_TYPE_KPI) return null;
    return (
      <ReportContent
        coreQueryReducerState={coreQueryReducerState}
        breakdownType={coreQueryState.breakdownType}
        queryType={coreQueryState.queryType}
        renderedCompRef={renderedCompRef}
        breakdown={coreQueryState.appliedBreakdown}
        handleChartTypeChange={handleChartTypeChange}
        queryOptions={coreQueryState.queryOptions}
        arrayMapper={arrayMapper}
        resultState={coreQueryState.resultState}
        queryTitle={coreQueryState.querySaved.name}
        section={REPORT_SECTION}
        eventPage={result_criteria}
        handleDurationChange={handleDurationChange}
        onReportClose={() => {
          console.log('Close report');
        }}
        handleGranularityChange={handleGranularityChange}
        setDrawerVisible={() => {
          console.log('Drawer visible');
        }}
        queries={coreQueryState.appliedQueries}
      />
    );
  };

  const renderMain = () => {
    if (coreQueryState.loading) {
      return (
        <>
          {renderEmptyHeader()}
          {renderQueryComposerNew()}
          <div className='mt-24 px-8'>{renderSpinner()}</div>
        </>
      );
    }
    if (coreQueryState.showResult && !coreQueryState.loading) {
      return (
        <>
          {renderAnalysisHeader()}
          <div className='mt-24 px-8'>
            <ErrorBoundary
              fallback={
                <FaErrorComp
                  size='medium'
                  title='Analyse Results Error'
                  subtitle='We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
                  className={undefined}
                  type={undefined}
                />
              }
              onError={FaErrorLog}
            >
              {Number(coreQueryState.activeTab) === 1 && (
                <>
                  {renderQueryComposerNew()}
                  {coreQueryState.requestQuery && renderReportContent()}
                </>
              )}
            </ErrorBoundary>
          </div>
        </>
      );
    } else if (
      !coreQueryState.showResult &&
      coreQueryState.resultState.loading
    ) {
      return <PageSuspenseLoader />;
    } else {
      return null;
    }
  };

  return renderMain();
};

export default CoreQuery;
