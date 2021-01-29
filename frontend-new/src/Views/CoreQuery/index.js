import React, { useState, useCallback, useEffect } from "react";
import moment from "moment";
import { bindActionCreators } from "redux";
import { connect, useSelector, useDispatch } from "react-redux";
import FunnelsResultPage from "./FunnelsResultPage";
import QueryComposer from "../../components/QueryComposer";
import AttrQueryComposer from "../../components/AttrQueryComposer";
import CampQueryComposer from "../../components/CampQueryComposer";
import CoreQueryHome from "../CoreQueryHome";
import { Drawer, Button } from "antd";
import { SVG, Text } from "../../components/factorsComponents";
import EventsAnalytics from "./EventsAnalytics";
import {
  deleteGroupByForEvent,
  getCampaignConfigData,
} from "../../reducers/coreQuery/middleware";
import {
  calculateFrequencyData,
  calculateActiveUsersData,
  hasApiFailed,
  formatApiData,
  getQuery,
  initialState,
  getFunnelQuery,
  DefaultDateRangeFormat,
  getAttributionQuery,
  getCampaignsQuery,
} from "./utils";
import {
  runQuery as runQueryService,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
} from "../../reducers/coreQuery/services";
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_ATTRIBUTION,
} from "../../utils/constants";
import AttributionsResult from "./AttributionsResult";
import { SHOW_ANALYTICS_RESULT } from "../../reducers/types";
import CampaignAnalytics from "./CampaignAnalytics";
import AnalysisResultsPage from "./AnalysisResultsPage";
import {
  SET_CAMP_DATE_RANGE,
  SET_ATTR_DATE_RANGE,
} from "../../reducers/coreQuery/actions";

function CoreQuery({
  activeProject,
  deleteGroupByForEvent,
  location,
  getCampaignConfigData,
}) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeKey, setActiveKey] = useState("0");
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialState);
  // const [funnelResult, updateFunnelResult] = useState(initialState);
  // const [attributionResult, updateAttributionResult] = useState(initialState);
  // const [campaignsResult, updateCampaignsResult] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [rowClicked, setRowClicked] = useState(false);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownTypeData, setBreakdownTypeData] = useState({
    loading: false,
    error: false,
    all: null,
    any: null,
  });
  const [breakdownType, setBreakdownType] = useState("each");
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    groupBy: [
      {
        prop_category: "", // user / event
        property: "", // user/eventproperty
        prop_type: "", // categorical  /numberical
        eventValue: "", // event name (funnel only)
        eventName: "", // eventName $present for global user breakdown
        eventIndex: 0,
      },
    ],
    event_analysis_seq: "",
    session_analytics_seq: {
      start: 1,
      end: 2,
    },
    date_range: { ...DefaultDateRangeFormat },
  });
  const [attributionsState, setAttributionsState] = useState({
    eventGoal: {},
    touchpoint: "",
    models: [],
    linkedEvents: [],
    date_range: {},
  });

  const [campaignState, setCampaignState] = useState({
    channel: "",
    select_metrics: [],
    filters: [],
    group_by: [],
    date_range: {},
  });

  const dispatch = useDispatch();
  const {
    groupBy,
    eventGoal,
    touchpoint,
    models,
    window,
    linkedEvents,
    camp_channels,
    camp_measures,
    camp_filters,
    camp_groupBy,
    camp_dateRange,
    attr_dateRange,
  } = useSelector((state) => state.coreQuery);

  const dateRange = queryOptions.date_range;

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getCampaignConfigData(activeProject.id, "all_ads");
    }
  }, [activeProject, getCampaignConfigData]);

  const updateResultState = useCallback((newState) => {
    setResultState(newState);
  }, []);

  const updateAppliedBreakdown = useCallback(() => {
    const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
    setAppliedBreakdown(newAppliedBreakdown);
  }, [groupBy]);

  const callRunQueryApiService = useCallback(
    async (activeProjectId, activeTab, appliedDateRange) => {
      try {
        const query = getQuery(
          activeTab,
          groupBy,
          queries,
          "each",
          appliedDateRange
        );
        if (activeTab !== "2") {
          updateRequestQuery(query);
        }

        const res = await runQueryService(activeProjectId, query);
        if (res.status === 200 && !hasApiFailed(res)) {
          if (activeTab !== "2") {
            updateResultState(activeTab, {
              loading: false,
              error: false,
              data: formatApiData(
                res.data.result_group[0],
                res.data.result_group[1]
              ),
            });
          }
          return res.data;
        } else {
          updateResultState(activeTab, {
            loading: false,
            error: true,
            data: null,
          });
          return null;
        }
      } catch (err) {
        console.log(err);
        updateResultState(activeTab, {
          loading: false,
          error: true,
          data: null,
        });
        return null;
      }
    },
    [updateResultState, groupBy, queries]
  );

  const runQuery = useCallback(
    async (refresh, isQuerySaved, appliedDateRange, activeTab) => {
      if (!appliedDateRange) {
        appliedDateRange = dateRange;
      }
      if (!activeTab) {
        activeTab = activeKey;
      }
      if (refresh) {
        updateResultState("0", initialState);
        updateResultState("1", initialState);
        updateResultState("2", initialState);
        updateResultState("3", initialState);
        setAppliedQueries(queries.map((elem) => elem.label));
        setQuerySaved(isQuerySaved);
        updateAppliedBreakdown();
        setBreakdownTypeData({
          loading: false,
          error: false,
          all: null,
          any: null,
        });
        setBreakdownType("each");
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
      } else {
        if (resultState[parseInt(activeTab)].data) {
          return false;
        }
      }

      updateResultState(activeTab, {
        loading: true,
        error: false,
        data: null,
      });

      if (parseInt(activeTab) === 2) {
        try {
          let activeUsersData = null;
          let userData = null;
          let sessionData = null;

          if (!refresh && resultState[1].data) {
            const res = await callRunQueryApiService(
              activeProject.id,
              "2",
              appliedDateRange
            );
            userData = resultState[1].data;
            if (res) {
              sessionData = res.result_group[0];
            }
          } else {
            // combine these two and make one query group to get both session and user data
            const res1 = await callRunQueryApiService(
              activeProject.id,
              "1",
              appliedDateRange
            );
            const res2 = await callRunQueryApiService(
              activeProject.id,
              "2",
              appliedDateRange
            );
            if (res1 && res2) {
              userData = formatApiData(
                res1.result_group[0],
                res1.result_group[1]
              );
              sessionData = res2.result_group[0];
            }
          }

          if (userData && sessionData) {
            activeUsersData = calculateActiveUsersData(
              userData,
              sessionData,
              appliedBreakdown
            );
          }

          updateResultState(activeTab, {
            loading: false,
            error: false,
            data: activeUsersData,
          });
        } catch (err) {
          console.log(err);
          console.log(err.response);
          updateResultState(activeKey, {
            loading: false,
            error: true,
            data: null,
          });
        }
      } else if (parseInt(activeTab) === 3) {
        try {
          let frequencyData = null;
          let userData = null;
          let eventData = null;
          if (!refresh && resultState[0].data) {
            eventData = resultState[0].data;
          } else {
            const res = await callRunQueryApiService(
              activeProject.id,
              "0",
              appliedDateRange
            );
            if (res) {
              eventData = formatApiData(
                res.result_group[0],
                res.result_group[1]
              );
            }
          }

          if (!refresh && resultState[1].data) {
            userData = resultState[1].data;
          } else {
            const res = await callRunQueryApiService(
              activeProject.id,
              "1",
              appliedDateRange
            );
            if (res) {
              userData = formatApiData(
                res.result_group[0],
                res.result_group[1]
              );
            }
          }

          if (userData && eventData) {
            frequencyData = calculateFrequencyData(
              eventData,
              userData,
              appliedBreakdown
            );
          }

          updateResultState(activeTab, {
            loading: false,
            error: false,
            data: frequencyData,
          });
        } catch (err) {
          console.log(err);
          console.log(err.response);
          updateResultState(activeKey, {
            loading: false,
            error: true,
            data: null,
          });
        }
      } else {
        callRunQueryApiService(activeProject.id, activeTab, appliedDateRange);
      }
    },
    [
      activeProject,
      activeKey,
      dateRange,
      resultState,
      queries,
      updateResultState,
      callRunQueryApiService,
      updateAppliedBreakdown,
      appliedBreakdown,
      dispatch,
    ]
  );

  const handleBreakdownTypeChange = useCallback(
    async (e) => {
      const key = e.target.value;
      setBreakdownType(key);
      if (key === "each") {
        return false;
      }
      if (breakdownTypeData[key]) {
        return false;
      } else {
        try {
          setBreakdownTypeData((currState) => {
            return { ...currState, loading: true };
          });
          const query = getQuery("1", groupBy, queries, key, dateRange);
          // updateRequestQuery(query);
          const res = await runQueryService(activeProject.id, query);
          if (res.status === 200 && !hasApiFailed(res)) {
            setBreakdownTypeData((currState) => {
              return {
                ...currState,
                loading: false,
                error: false,
                [key]: res.data.result_group[0],
              };
            });
          } else {
            setBreakdownTypeData((currState) => {
              return { ...currState, loading: false, error: true };
            });
          }
        } catch (err) {
          console.log(err);
          setBreakdownTypeData((currState) => {
            return { ...currState, loading: false, error: true };
          });
        }
      }
    },
    [activeProject.id, queries, groupBy, breakdownTypeData, dateRange]
  );

  const runFunnelQuery = useCallback(
    async (isQuerySaved, appliedDateRange) => {
      try {
        if (!appliedDateRange) {
          appliedDateRange = dateRange;
        }
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        setAppliedQueries(queries.map((elem) => elem.label));
        updateAppliedBreakdown();
        updateResultState({ ...initialState, loading: true });
        const query = getFunnelQuery(groupBy, queries, appliedDateRange);
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
      updateAppliedBreakdown,
      activeProject.id,
      groupBy,
      dateRange,
      dispatch,
      updateResultState,
    ]
  );

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
      window,
      attr_dateRange,
      updateResultState,
    ]
  );

  const runCampaignsQuery = useCallback(
    async (isQuerySaved, durationObj = null) => {
      console.log(durationObj)
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
      let frequency = "date";
      if (moment(dates.endDate).diff(dates.startDate, "hours") <= 24) {
        frequency = "hour";
      }
      setQueryOptions((currState) => {
        return {
          ...currState,
          date_range: {
            ...currState.date_range,
            from: dates.startDate,
            to: dates.endDate,
            frequency,
          },
        };
      });
      const appliedDateRange = {
        ...queryOptions.date_range,
        from: dates.startDate,
        to: dates.endDate,
        frequency,
      };

      if (queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(querySaved, appliedDateRange);
      }
      if (queryType === QUERY_TYPE_EVENT) {
        runQuery(true, querySaved, appliedDateRange);
      }

      if (queryType === QUERY_TYPE_CAMPAIGN) {
        const payload = {
          from: moment(dates.startDate).startOf("day"),
          to: moment(dates.endDate).endOf("day"),
          frequency: "date",
        };
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload);
      }

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        const payload = {
          from: moment(dates.startDate).startOf("day"),
          to: moment(dates.endDate).endOf("day"),
          frequency: "date",
        };
        dispatch({ type: SET_ATTR_DATE_RANGE, payload });
        runAttributionQuery(querySaved, payload);
      }
    },
    [queryType, runFunnelQuery, runQuery, querySaved, queryOptions.date_range, dispatch, runCampaignsQuery, runAttributionQuery]
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
        if (rowClicked.settings) {
          runQuery(
            true,
            rowClicked.queryName,
            null,
            rowClicked.settings.activeKey
          );
        } else {
          runQuery(true, rowClicked.queryName);
        }
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

  const queryChange = (newEvent, index, changeType = "add") => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === "add") {
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
          text: "Analyse Events",
          icon: "funnels_cq",
        };
      case QUERY_TYPE_FUNNEL:
        return {
          text: "Find event funnel for",
          icon: "events_dashboard_cq",
        };
      case QUERY_TYPE_CAMPAIGN:
        return {
          text: "Campaign Analytics",
          icon: "funnels_cq",
        };
      case QUERY_TYPE_ATTRIBUTION:
        return {
          text: "Attributions",
          icon: "funnels_cq",
        };
      default:
        return {
          text: "Templates",
          icon: "funnels_cq",
        };
    }
  };

  const title = () => {
    const IconAndText = IconAndTextSwitchQueryType(queryType);
    return (
      <div className={"flex justify-between items-center"}>
        <div className={"flex items-center"}>
          <SVG name={IconAndText.icon} size="24px"></SVG>
          <Text
            type={"title"}
            level={4}
            weight={"bold"}
            extraClass={"ml-2 m-0"}
          >
            {IconAndText.text}
          </Text>
        </div>
        <div className={"flex justify-end items-center"}>
          <Button size={"large"} type="text">
            <SVG name="play"></SVG>Help
          </Button>
          <Button size={"large"} type="text" onClick={() => closeDrawer()}>
            <SVG name="times"></SVG>
          </Button>
        </div>
      </div>
    );
  };

  let eventsMapper = {};
  let reverseEventsMapper = {};
  let arrayMapper = [];

  appliedQueries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q;
    arrayMapper.push({
      eventName: q,
      index,
      mapper: `event${index + 1}`,
    });
  });

  // let result = (
  //   <EventsAnalytics
  //     queries={appliedQueries}
  //     eventsMapper={eventsMapper}
  //     reverseEventsMapper={reverseEventsMapper}
  //     breakdown={appliedBreakdown}
  //     resultState={resultState}
  //     setDrawerVisible={setDrawerVisible}
  //     runQuery={runQuery}
  //     activeKey={activeKey}
  //     breakdownType={breakdownType}
  //     handleBreakdownTypeChange={handleBreakdownTypeChange}
  //     breakdownTypeData={breakdownTypeData}
  //     queryType={queryType}
  //     requestQuery={requestQuery}
  //     setShowResult={setShowResult}
  //     querySaved={querySaved}
  //     setQuerySaved={setQuerySaved}
  //     durationObj={queryOptions.date_range}
  //     handleDurationChange={handleDurationChange}
  //     arrayMapper={arrayMapper}
  //     setActiveKey={setActiveKey}
  //   />
  // );

  // if (queryType === QUERY_TYPE_FUNNEL) {
  //   result = (
  //     <FunnelsResultPage
  //       setDrawerVisible={setDrawerVisible}
  //       queries={appliedQueries}
  //       resultState={funnelResult}
  //       breakdown={appliedBreakdown}
  //       requestQuery={requestQuery}
  //       setShowResult={setShowResult}
  //       querySaved={querySaved}
  //       setQuerySaved={setQuerySaved}
  //       durationObj={queryOptions.date_range}
  //       handleDurationChange={handleDurationChange}
  //       arrayMapper={arrayMapper}
  //     />
  //   );
  // }

  // if (queryType === QUERY_TYPE_ATTRIBUTION) {
  //   result = (
  //     <AttributionsResult
  //       setShowResult={setShowResult}
  //       requestQuery={requestQuery}
  //       querySaved={querySaved}
  //       setQuerySaved={setQuerySaved}
  //       resultState={attributionResult}
  //       setDrawerVisible={setDrawerVisible}
  //       attributionsState={attributionsState}
  //     />
  //   );
  // }

  // if (queryType === QUERY_TYPE_CAMPAIGN) {
  //   arrayMapper = campaignState.select_metrics.map((metric, index) => {
  //     return {
  //       eventName: metric,
  //       index,
  //       mapper: `event${index + 1}`,
  //     };
  //   });
  //   result = (
  //     <CampaignAnalytics
  //       setShowResult={setShowResult}
  //       requestQuery={requestQuery}
  //       querySaved={querySaved}
  //       setQuerySaved={setQuerySaved}
  //       resultState={campaignsResult}
  //       setDrawerVisible={setDrawerVisible}
  //       arrayMapper={arrayMapper}
  //       campaignState={campaignState}
  //       // attributionsState={attributionsState}
  //     />
  //   );
  // }

  // result = <AnalysisResultsPage queryType={queryType} />;

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
      <Drawer
        title={title()}
        placement="left"
        closable={false}
        visible={drawerVisible}
        onClose={closeDrawer}
        getContainer={false}
        width={"600px"}
        className={"fa-drawer"}
      >
        {renderQueryComposer()}
      </Drawer>

      {showResult ? (
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
          handleDurationChange={handleDurationChange}
          arrayMapper={arrayMapper}
          queryOptions={queryOptions}
          attributionsState={attributionsState}
          breakdownType={breakdownType}
          campaignState={campaignState}
        />
      ) : (
        <CoreQueryHome
          setQueryType={setQueryType}
          setDrawerVisible={setDrawerVisible}
          setQueries={setQueries}
          setQueryOptions={setExtraOptions}
          setRowClicked={setRowClicked}
          location={location}
          setActiveKey={setActiveKey}
        />
      )}
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
