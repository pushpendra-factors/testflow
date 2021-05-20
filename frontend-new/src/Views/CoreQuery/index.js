import React, { useState, useCallback, useEffect, useMemo } from 'react';
import moment from 'moment';
import { bindActionCreators } from 'redux';
import { connect, useSelector, useDispatch } from 'react-redux';
import QueryComposer from '../../components/QueryComposer';
import AttrQueryComposer from '../../components/AttrQueryComposer';
import CampQueryComposer from '../../components/CampQueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog,
} from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import {
  deleteGroupByForEvent,
  getCampaignConfigData,
} from '../../reducers/coreQuery/middleware';
import {
  calculateFrequencyData,
  calculateActiveUsersData,
  formatApiData,
  getQuery,
  initialState,
  getFunnelQuery,
  DefaultDateRangeFormat,
  getAttributionQuery,
  getCampaignsQuery,
} from './utils';
import {
  getEventsData,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
} from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_ATTRIBUTION,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  EACH_USER_TYPE,
  REPORT_SECTION,
  INITIAL_SESSION_ANALYTICS_SEQ,
  ATTRIBUTION_METRICS,
} from '../../utils/constants';
import { SHOW_ANALYTICS_RESULT } from '../../reducers/types';
import AnalysisResultsPage from './AnalysisResultsPage';
import {
  SET_CAMP_DATE_RANGE,
  SET_ATTR_DATE_RANGE,
} from '../../reducers/coreQuery/actions';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';

function CoreQuery({
  activeProject,
  deleteGroupByForEvent,
  location,
  getCampaignConfigData,
}) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeKey, setActiveKey] = useState('0');
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialState);
  const [cmprResultState, setCmprResultState] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [rowClicked, setRowClicked] = useState(false);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownType, setBreakdownType] = useState(EACH_USER_TYPE);
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    groupBy: [
      {
        prop_category: '', // user / event
        property: '', // user/eventproperty
        prop_type: '', // categorical  /numberical
        eventValue: '', // event name (funnel only)
        eventName: '', // eventName $present for global user breakdown
        eventIndex: 0,
      },
    ],
    event_analysis_seq: '',
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat },
  });
  const [attributionsState, setAttributionsState] = useState({
    eventGoal: {},
    touchpoint: '',
    models: [],
    linkedEvents: [],
    date_range: {},
  });

  const [cmprAttrDurationObj, setcmprAttrDurationObj] = useState({});

  const [campaignState, setCampaignState] = useState({
    channel: '',
    select_metrics: [],
    filters: [],
    group_by: [],
    date_range: {},
  });

  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS,
  ]);

  const dispatch = useDispatch();
  const {
    groupBy,
    eventGoal,
    touchpoint,
    touchpoint_filters,
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
  } = useSelector((state) => state.coreQuery);

  const {
    show_criteria: result_criteria,
    performance_criteria: user_type,
  } = useSelector((state) => state.analyticsQuery);

  const dateRange = queryOptions.date_range;
  const { session_analytics_seq } = queryOptions;

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

  const runQuery = useCallback(
    async (isQuerySaved, durationObj) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        setAppliedQueries(queries.map((elem) => elem.label));
        updateAppliedBreakdown();
        setBreakdownType(user_type);
        updateResultState({ ...initialState, loading: true });
        const query = getQuery(
          groupBy,
          queries,
          result_criteria,
          user_type,
          durationObj
        );
        updateRequestQuery(query);
        const res = await getEventsData(activeProject.id, query);
        if (result_criteria === TOTAL_EVENTS_CRITERIA) {
          updateResultState({
            ...initialState,
            data: formatApiData(
              res.data.result_group[0],
              res.data.result_group[1]
            ),
          });
        } else if (result_criteria === TOTAL_USERS_CRITERIA) {
          if (user_type === EACH_USER_TYPE) {
            updateResultState({
              ...initialState,
              data: formatApiData(
                res.data.result_group[0],
                res.data.result_group[1]
              ),
            });
          } else {
            updateResultState({
              ...initialState,
              data: res.data.result_group[0],
            });
          }
        } else if (result_criteria === ACTIVE_USERS_CRITERIA) {
          const userData = formatApiData(
            res.data.result_group[0],
            res.data.result_group[1]
          );
          const sessionsData = res.data.result_group[2];
          const activeUsersData = calculateActiveUsersData(
            userData,
            sessionsData,
            [...groupBy.global, ...groupBy.event]
          );
          updateResultState({ ...initialState, data: activeUsersData });
        } else if (result_criteria === FREQUENCY_CRITERIA) {
          const eventData = formatApiData(
            res.data.result_group[0],
            res.data.result_group[1]
          );
          const userData = formatApiData(
            res.data.result_group[2],
            res.data.result_group[3]
          );
          const frequencyData = calculateFrequencyData(eventData, userData, [
            ...groupBy.global,
            ...groupBy.event,
          ]);
          updateResultState({ ...initialState, data: frequencyData });
        }
      } catch (err) {
        console.log(err);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      queries,
      dateRange,
      result_criteria,
      user_type,
      activeProject.id,
      dispatch,
      groupBy,
      updateAppliedBreakdown,
      updateResultState,
    ]
  );

  const runFunnelQuery = useCallback(
    async (isQuerySaved, durationObj) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        setAppliedQueries(queries.map((elem) => elem.label));
        updateAppliedBreakdown();
        updateResultState({ ...initialState, loading: true });
        const query = getFunnelQuery(
          groupBy,
          queries,
          session_analytics_seq,
          durationObj
        );
        updateRequestQuery(query);
        const res = await getFunnelData(activeProject.id, query);
        if (res.status === 200) {
          updateResultState({ ...initialState, data: res.data });
        } else {
          updateResultState({ ...initialState, error: true });
        }
      } catch (err) {
        console.log(err);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      queries,
      session_analytics_seq,
      updateAppliedBreakdown,
      activeProject.id,
      groupBy,
      dateRange,
      dispatch,
      updateResultState,
    ]
  );

  const runAttrCmprQuery = (cmprDuration) => {
    if (!cmprDuration) {
      setCmprResultState({
        ...initialState,
      });
    }

    setcmprAttrDurationObj(cmprDuration);
    const query = getAttributionQuery(
      eventGoal,
      touchpoint,
      touchpoint_filters,
      attr_query_type,
      models,
      window,
      linkedEvents,
      cmprDuration
    );

    setCmprResultState({
      ...initialState,
      loading: true,
      data: null,
    });

    getAttributionsData(activeProject.id, query).then(
      (res) => {
        setCmprResultState({
          ...initialState,
          data: res.data,
        });
      },
      (err) => {
        setCmprResultState({
          ...initialState,
          loading: false,
          error: true,
          data: null,
        });
      }
    );
  };

  const runAttributionQuery = useCallback(
    async (isQuerySaved, durationObj) => {
      try {
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        updateResultState({
          ...initialState,
          loading: true,
        });
        if (!durationObj) {
          durationObj = attr_dateRange;
        }
        const query = getAttributionQuery(
          eventGoal,
          touchpoint,
          touchpoint_filters,
          attr_query_type,
          models,
          window,
          linkedEvents,
          durationObj
        );
        updateRequestQuery(query);
        setAttributionsState({
          eventGoal,
          touchpoint,
          models,
          linkedEvents,
          date_range: { ...durationObj },
        });
        const res = await getAttributionsData(activeProject.id, query);
        updateResultState({
          ...initialState,
          data: res.data,
        });
      } catch (err) {
        console.log(err);
        updateResultState({
          ...initialState,
          error: true,
        });
      }
    },
    [
      dispatch,
      activeProject.id,
      eventGoal,
      linkedEvents,
      models,
      touchpoint,
      touchpoint_filters,
      attr_query_type,
      window,
      attr_dateRange,
      updateResultState,
    ]
  );

  const runCampaignsQuery = useCallback(
    async (isQuerySaved, durationObj = null) => {
      try {
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        updateResultState({
          ...initialState,
          loading: true,
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
        setCampaignState({
          channel: query.query_group[0].channel,
          filters: query.query_group[0].filters,
          select_metrics: query.query_group[0].select_metrics,
          group_by: query.query_group[0].group_by,
          date_range: { ...durationObj },
        });
        updateRequestQuery(query);
        const res = await getCampaignsData(activeProject.id, query);
        updateResultState({
          ...initialState,
          data: res.data.result ? res.data.result : res.data,
        });
      } catch (err) {
        console.log(err);
        updateResultState({
          ...initialState,
          error: true,
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
    ]
  );

  const handleDurationChange = useCallback(
    (dates) => {
      let from,
        to,
        frequency = 'date';
      if (Array.isArray(dates.startDate)) {
        from = dates.startDate[0];
        to = dates.startDate[1];
      } else {
        from = dates.startDate;
        to = dates.endDate;
      }
      if (moment(to).diff(from, 'hours') < 24) {
        frequency = 'hour';
      }
      setQueryOptions((currState) => {
        return {
          ...currState,
          date_range: {
            ...currState.date_range,
            from,
            to,
            frequency,
          },
        };
      });
      const appliedDateRange = {
        ...queryOptions.date_range,
        from,
        to,
        frequency,
      };

      if (queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(querySaved, appliedDateRange);
      }
      if (queryType === QUERY_TYPE_EVENT) {
        runQuery(querySaved, appliedDateRange);
      }

      if (queryType === QUERY_TYPE_CAMPAIGN) {
        const payload = {
          from: moment(from).startOf('day'),
          to: moment(to).endOf('day'),
          frequency: 'date',
        };
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload);
      }

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        const payload = {
          from: moment(from).startOf('day'),
          to: moment(to).endOf('day'),
          frequency: 'date',
        };
        dispatch({ type: SET_ATTR_DATE_RANGE, payload });
        runAttributionQuery(querySaved, payload);
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
    ]
  );

  useEffect(() => {
    if (rowClicked) {
      if (rowClicked.queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(rowClicked.queryName);
      } else if (rowClicked.queryType === QUERY_TYPE_ATTRIBUTION) {
        runAttributionQuery(rowClicked.queryName);
      } else if (rowClicked.queryType === QUERY_TYPE_CAMPAIGN) {
        runCampaignsQuery(rowClicked.queryName);
      } else {
        runQuery(rowClicked.queryName);
      }
      setRowClicked(false);
    }
  }, [
    rowClicked,
    runFunnelQuery,
    runQuery,
    runAttributionQuery,
    runCampaignsQuery,
  ]);

  useEffect(() => {
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  const queryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
          deleteGroupByForEvent(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else {
        deleteGroupByForEvent(newEvent, index);
        queryupdated.splice(index, 1);
      }
    } else {
      queryupdated.push(newEvent);
    }
    setQueries(queryupdated);
  };

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const setExtraOptions = (options) => {
    setQueryOptions(options);
  };

  const IconAndTextSwitchQueryType = (queryType) => {
    switch (queryType) {
      case QUERY_TYPE_EVENT:
        return {
          text: 'Analyse Events',
          icon: 'events_cq',
        };
      case QUERY_TYPE_FUNNEL:
        return {
          text: 'Find event funnel for',
          icon: 'funnels_cq',
        };
      case QUERY_TYPE_CAMPAIGN:
        return {
          text: 'Campaign Analytics',
          icon: 'campaigns_cq',
        };
      case QUERY_TYPE_ATTRIBUTION:
        return {
          text: 'Attributions',
          icon: 'attributions_cq',
        };
      default:
        return {
          text: 'Templates',
          icon: 'templates_cq',
        };
    }
  };

  const title = () => {
    const IconAndText = IconAndTextSwitchQueryType(queryType);
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex items-center'}>
          <SVG name={IconAndText.icon} size='24px'></SVG>
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
          {/* <Button size={"large"} type="text">
            <SVG name="play"></SVG>Help
          </Button> */}
          <Button size={'large'} type='text' onClick={() => closeDrawer()}>
            <SVG name='times'></SVG>
          </Button>
        </div>
      </div>
    );
  };

  let eventsMapper = {};
  let reverseEventsMapper = {};
  let arrayMapper = [];

  const campaignsArrayMapper = useMemo(() => {
    return campaignState.select_metrics.map((metric, index) => {
      return {
        eventName: metric,
        index,
        mapper: `event${index + 1}`,
      };
    });
  }, [campaignState.select_metrics]);

  appliedQueries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q;
    arrayMapper.push({
      eventName: q,
      index,
      mapper: `event${index + 1}`,
      displayName: eventNames[q] || q,
    });
  });

  const renderQueryComposer = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
      return (
        <QueryComposer
          queries={queries}
          runQuery={runQuery}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runFunnelQuery={runFunnelQuery}
          activeKey={activeKey}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return <AttrQueryComposer runAttributionQuery={runAttributionQuery} />;
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return (
        <CampQueryComposer
          handleRunQuery={runCampaignsQuery}
        ></CampQueryComposer>
      );
    }
  };

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
        <Drawer
          title={title()}
          placement='left'
          closable={false}
          visible={drawerVisible}
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

        {showResult ? (
          <CoreQueryContext.Provider
            value={{ attributionMetrics, setAttributionMetrics }}
          >
            <AnalysisResultsPage
              queryType={queryType}
              resultState={resultState}
              setDrawerVisible={setDrawerVisible}
              requestQuery={requestQuery}
              queries={appliedQueries}
              breakdown={appliedBreakdown}
              setShowResult={setShowResult}
              querySaved={querySaved}
              setQuerySaved={setQuerySaved}
              durationObj={queryOptions.date_range}
              cmprDuration={cmprAttrDurationObj}
              handleDurationChange={handleDurationChange}
              arrayMapper={arrayMapper}
              queryOptions={queryOptions}
              attributionsState={attributionsState}
              breakdownType={breakdownType}
              campaignState={campaignState}
              eventPage={result_criteria}
              section={REPORT_SECTION}
              runAttrCmprQuery={runAttrCmprQuery}
              cmprResultState={cmprResultState}
              campaignsArrayMapper={campaignsArrayMapper}
            />
          </CoreQueryContext.Provider>
        ) : (
          <CoreQueryHome
            setQueryType={setQueryType}
            setDrawerVisible={setDrawerVisible}
            setQueries={setQueries}
            setQueryOptions={setExtraOptions}
            setRowClicked={setRowClicked}
            location={location}
            setActiveKey={setActiveKey}
            setBreakdownType={setBreakdownType}
          />
        )}
      </ErrorBoundary>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      deleteGroupByForEvent,
      getCampaignConfigData,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
