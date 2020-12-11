import React, { useState, useCallback, useEffect } from "react";
import moment from "moment";
import { bindActionCreators } from "redux";
import { connect, useSelector } from "react-redux";
import FunnelsResultPage from "./FunnelsResultPage";
import QueryComposer from "../../components/QueryComposer";
import CoreQueryHome from "../CoreQueryHome";
import { Drawer, Button } from "antd";
import { SVG, Text } from "../../components/factorsComponents";
import EventsAnalytics from "./EventsAnalytics";
import { deleteGroupByForEvent } from "../../reducers/coreQuery/middleware";
import {
  initialResultState, calculateFrequencyData, calculateActiveUsersData, hasApiFailed, formatApiData, getQuery, initialState, getFunnelQuery, DefaultDateRangeFormat
} from './utils';
import {   runQuery as runQueryService,
getFunnelData,
} from "../../reducers/coreQuery/services";
import {QUERY_TYPE_FUNNEL, QUERY_TYPE_EVENT, QUERY_TYPE_CAMPAIGN, QUERY_TYPE_ATTRIBUTION} from 'Utils/constants';

function CoreQuery({ activeProject, deleteGroupByForEvent, location }) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeKey, setActiveKey] = useState("1");
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialResultState);
  const [funnelResult, updateFunnelResult] = useState(initialState);
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

  const groupBy = useSelector((state) => state.coreQuery.groupBy);
  const dateRange = queryOptions.date_range;

  const updateResultState = useCallback((activeTab, newState) => {
    const idx = parseInt(activeTab);
    setResultState((currState) => {
      return currState.map((elem, index) => {
        if (index === idx) {
          return newState;
        }
        return elem;
      });
    });
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
          breakdownType,
          appliedDateRange
        );
        updateRequestQuery(query);
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
    [updateResultState, groupBy, queries, breakdownType]
  );

  const runQuery = useCallback(
    async (
      activeTab,
      refresh = false,
      isQuerySaved = false,
      appliedDateRange
    ) => {
      if (!appliedDateRange) {
        appliedDateRange = dateRange;
      }
      setActiveKey(activeTab);
      setBreakdownType("each");

      if (!refresh) {
        if (resultState[parseInt(activeTab)].data) {
          return false;
        }

        if (activeTab === "2") {
          updateResultState(activeTab, {
            loading: true,
            error: false,
            data: null,
          });

          let activeUsersData = null;
          let userData = null;
          let sessionData = null;

          if (resultState[1].data) {
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
          return false;
        }

        if (activeTab === "3") {
          let frequencyData = null;
          let userData = null;
          const eventData = resultState[0].data;

          if (resultState[1].data) {
            userData = resultState[1].data;
          } else {
            updateResultState(activeTab, {
              loading: true,
              error: false,
              data: null,
            });
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
          return false;
        }
      } else {
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
        setShowResult(true);
      }

      updateResultState(activeTab, { loading: true, error: false, data: null });
      callRunQueryApiService(activeProject.id, activeTab, appliedDateRange);
    },
    [
      activeProject,
      dateRange,
      resultState,
      queries,
      updateResultState,
      callRunQueryApiService,
      updateAppliedBreakdown,
      appliedBreakdown,
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
          updateRequestQuery(query);
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
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        setAppliedQueries(queries.map((elem) => elem.label));
        updateAppliedBreakdown();
        updateFunnelResult({ ...initialState, loading: true });
        const query = getFunnelQuery(groupBy, queries, appliedDateRange);
        updateRequestQuery(query);
        const res = await getFunnelData(activeProject.id, query);
        if (res.status === 200) {
          updateFunnelResult({ ...initialState, data: res.data });
        } else {
          updateFunnelResult({ ...initialState, error: true });
        }
      } catch (err) {
        console.log(err);
        updateFunnelResult({ ...initialState, error: true });
      }
    },
    [queries, updateAppliedBreakdown, activeProject.id, groupBy, dateRange]
  );

  const handleDurationChange = useCallback(
    (dates) => {
      if (dates && dates.selected) {
        let frequency = "date";
        if (
          moment(dates.selected.endDate).diff(
            dates.selected.startDate,
            "hours"
          ) <= 24
        ) {
          frequency = "hour";
        }
        setQueryOptions((currState) => {
          return {
            ...currState,
            date_range: {
              ...currState.date_range,
              from: dates.selected.startDate,
              to: dates.selected.endDate,
              frequency,
            },
          };
        });
        const appliedDateRange = {
          ...queryOptions.date_range,
          from: dates.selected.startDate,
          to: dates.selected.endDate,
          frequency,
        };

        if (queryType === QUERY_TYPE_FUNNEL) {
          runFunnelQuery(querySaved, appliedDateRange);
        } else {
          runQuery("0", true, querySaved, appliedDateRange);
        }
      }
    },
    [queryType, runFunnelQuery, runQuery, querySaved, queryOptions.date_range]
  );

  useEffect(() => {
    if (rowClicked) {
      if (rowClicked === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(true);
      } else {
        runQuery("0", true, true);
      }
      setRowClicked(false);
    }
  }, [rowClicked, runFunnelQuery, runQuery]);

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

const IconAndTextSwitchQueryType = (queryType) =>{
    switch (queryType){
      case QUERY_TYPE_EVENT: return{
        text:'Analyse Events',
        icon: 'funnels_cq'
      };
      case QUERY_TYPE_FUNNEL: return {
        text:'Find event funnel for',
        icon: 'events_dashboard_cq'
      };
      case QUERY_TYPE_CAMPAIGN: return {
        text:'Campaign Analytics',
        icon: 'funnels_cq'
      };
      case QUERY_TYPE_ATTRIBUTION: return {
        text:'Attributions',
        icon: 'funnels_cq'
      };
      default: return {
        text:'Templates',
        icon: 'funnels_cq'
      };
    }
  }

  const title = () => {
    const IconAndText = IconAndTextSwitchQueryType(queryType);
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex items-center'}>
          <SVG name={IconAndText.icon} size="24px"></SVG>
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{IconAndText.text}</Text>
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

  const eventsMapper = {};
  const reverseEventsMapper = {};
  const arrayMapper = [];

  appliedQueries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q;
    arrayMapper.push({
      eventName: q,
      index,
      mapper: `event${index + 1}`,
    });
  });

  let result = (
    <EventsAnalytics
      queries={appliedQueries}
      eventsMapper={eventsMapper}
      reverseEventsMapper={reverseEventsMapper}
      breakdown={appliedBreakdown}
      resultState={resultState}
      setDrawerVisible={setDrawerVisible}
      runQuery={runQuery}
      activeKey={activeKey}
      breakdownType={breakdownType}
      handleBreakdownTypeChange={handleBreakdownTypeChange}
      breakdownTypeData={breakdownTypeData}
      queryType={queryType}
      requestQuery={requestQuery}
      setShowResult={setShowResult}
      querySaved={querySaved}
      setQuerySaved={setQuerySaved}
      durationObj={queryOptions.date_range}
      handleDurationChange={handleDurationChange}
      arrayMapper={arrayMapper}
    />
  );

  if (queryType === QUERY_TYPE_FUNNEL) {
    result = (
      <FunnelsResultPage
        setDrawerVisible={setDrawerVisible}
        queries={appliedQueries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        resultState={funnelResult}
        breakdown={appliedBreakdown}
        requestQuery={requestQuery}
        setShowResult={setShowResult}
        querySaved={querySaved}
        setQuerySaved={setQuerySaved}
        durationObj={queryOptions.date_range}
        handleDurationChange={handleDurationChange}
      />
    );
  }

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
        {(queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) && 
        <QueryComposer
          queries={queries}
          runQuery={runQuery}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runFunnelQuery={runFunnelQuery}
        />
        }
      </Drawer>

      {showResult ? (
        <>{result}</>
      ) : (
        <CoreQueryHome
          setQueryType={setQueryType}
          setDrawerVisible={setDrawerVisible}
          setQueries={setQueries}
          setQueryOptions={setExtraOptions}
          setRowClicked={setRowClicked}
          location={location}
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
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
