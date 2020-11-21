import React, { useState, useCallback, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';
import { runQuery as runQueryService, getFunnelData } from '../../reducers/coreQuery/services';
import {
  initialResultState, calculateFrequencyData, calculateActiveUsersData, hasApiFailed, formatApiData, getQuery, initialState, getFunnelQuery
} from './utils';

function CoreQuery({ activeProject }) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState('event');
  const [activeKey, setActiveKey] = useState('1');
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialResultState);
  const [funnelResult, updateFunnelResult] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [rowClicked, setRowClicked] = useState(false);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownTypeData, setBreakdownTypeData] = useState({
    loading: false, error: false, all: null, any: null
  });
  const [breakdownType, setBreakdownType] = useState('each');
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    groupBy: [{
      prop_category: '', // user / event
      property: '', // user/eventproperty
      prop_type: '', // categorical  /numberical
      eventValue: '', // event name (funnel only)
      eventName: '', // eventName $present for global user breakdown
      eventIndex: 0
    }],
    event_analysis_seq: '',
    session_analytics_seq: {
      start: 1,
      end: 2
    },
    date_range: {
      from: '',
      to: ''
    }
  });

  const groupBy = useSelector(state => state.coreQuery.groupBy);

  const updateResultState = useCallback((activeTab, newState) => {
    const idx = parseInt(activeTab);
    setResultState(currState => {
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

  const callRunQueryApiService = useCallback(async (activeProjectId, activeTab) => {
    try {
      const query = getQuery(activeTab, queryType, groupBy, queries);
      updateRequestQuery(query);
      const res = await runQueryService(activeProjectId, query);
      if (res.status === 200 && !hasApiFailed(res)) {
        if (activeTab !== '2') {
          updateResultState(activeTab, { loading: false, error: false, data: formatApiData(res.data.result_group[0], res.data.result_group[1]) });
        }
        return res.data;
      } else {
        updateResultState(activeTab, { loading: false, error: true, data: null });
        return null;
      }
    } catch (err) {
      console.log(err);
      updateResultState(activeTab, { loading: false, error: true, data: null });
      return null;
    }
  }, [updateResultState, queryType, groupBy, queries]);

  const runQuery = useCallback(async (activeTab, refresh = false, isQuerySaved = false) => {
    setActiveKey(activeTab);
    setBreakdownType('each');

    if (!refresh) {
      if (resultState[parseInt(activeTab)].data) {
        return false;
      }

      if (activeTab === '2') {
        updateResultState(activeTab, { loading: true, error: false, data: null });

        let activeUsersData = null; let userData = null; let sessionData = null;

        if (resultState[1].data) {
          const res = await callRunQueryApiService(activeProject.id, '2');
          userData = resultState[1].data;
          if (res) {
            sessionData = res.result_group[0];
          }
        } else {
          // combine these two and make one query group to get both session and user data
          const res1 = await callRunQueryApiService(activeProject.id, '1');
          const res2 = await callRunQueryApiService(activeProject.id, '2');
          if (res1 && res2) {
            userData = formatApiData(res1.result_group[0], res1.result_group[1]);
            sessionData = res2.result_group[0];
          }
        }

        if (userData && sessionData) {
          activeUsersData = calculateActiveUsersData(userData, sessionData, appliedBreakdown);
        }
        updateResultState(activeTab, { loading: false, error: false, data: activeUsersData });
        return false;
      }

      if (activeTab === '3') {
        let frequencyData = null; let userData = null;
        const eventData = resultState[0].data;

        if (resultState[1].data) {
          userData = resultState[1].data;
        } else {
          updateResultState(activeTab, { loading: true, error: false, data: null });
          const res = await callRunQueryApiService(activeProject.id, '1');
          if (res) {
            userData = formatApiData(res.result_group[0], res.result_group[1]);
          }
        }

        if (userData && eventData) {
          frequencyData = calculateFrequencyData(eventData, userData, appliedBreakdown);
        }

        updateResultState(activeTab, { loading: false, error: false, data: frequencyData });
        return false;
      }
    } else {
      updateResultState('1', initialState);
      updateResultState('2', initialState);
      updateResultState('3', initialState);
      setAppliedQueries(queries.map(elem => elem.label));
      setQuerySaved(isQuerySaved);
      updateAppliedBreakdown();
      setBreakdownTypeData({
        loading: false, error: false, all: null, any: null
      });
      setBreakdownType('each');
      closeDrawer();
      setShowResult(true);
    }

    updateResultState(activeTab, { loading: true, error: false, data: null });
    callRunQueryApiService(activeProject.id, activeTab);
  }, [activeProject, resultState, queries, updateResultState, callRunQueryApiService, updateAppliedBreakdown, appliedBreakdown]);

  const handleBreakdownTypeChange = useCallback(async (e) => {
    const key = e.target.value;
    setBreakdownType(key);
    if (key === 'each') {
      return false;
    }
    if (breakdownTypeData[key]) {
      return false;
    } else {
      try {
        setBreakdownTypeData(currState => {
          return { ...currState, loading: true };
        });
        const query = getQuery('1', queryType, groupBy, queries, key);
        updateRequestQuery(query);
        const res = await runQueryService(activeProject.id, query);
        if (res.status === 200 && !hasApiFailed(res)) {
          setBreakdownTypeData(currState => {
            return {
              ...currState, loading: false, error: false, [key]: res.data.result_group[0]
            };
          });
        } else {
          setBreakdownTypeData(currState => {
            return { ...currState, loading: false, error: true };
          });
        }
      } catch (err) {
        console.log(err);
        setBreakdownTypeData(currState => {
          return { ...currState, loading: false, error: true };
        });
      }
    }
  }, [activeProject.id, queries, groupBy, queryType, breakdownTypeData]);

  const runFunnelQuery = useCallback(async (isQuerySaved) => {
    try {
      closeDrawer();
      setShowResult(true);
      setQuerySaved(isQuerySaved);
      setAppliedQueries(queries.map(elem => elem.label));
      updateAppliedBreakdown();
      updateFunnelResult({ ...initialState, loading: true });
      const query = getFunnelQuery(groupBy, queries);
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
  }, [queries, updateAppliedBreakdown, activeProject.id, groupBy]);

  useEffect(() => {
    if (rowClicked) {
      if (rowClicked === 'funnel') {
        runFunnelQuery(true);
      } else {
        runQuery('0', true, true);
      }
      setRowClicked(false);
    }
  }, [rowClicked, runFunnelQuery, runQuery]);

  const queryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        queryupdated[index] = newEvent;
      } else {
        queryupdated.splice(index, 1);
      }
    }
    else if (queryType === 'event') {
      const queryExist = queryupdated.findIndex((q) => q.label === newEvent.label);
      if(queryExist < 0) {
        queryupdated.push(newEvent);
      }
    }
    else {
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

  const title = () => {
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex'}>
          <SVG name={queryType === 'funnel' ? 'funnels_cq' : 'events_cq'} size="24px"></SVG>
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{queryType === 'funnel' ? 'Find event funnel for' : 'Analyse Events'}</Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Button size={'large'} type="text"><SVG name="play"></SVG>Help</Button>
          <Button size={'large'} type="text" onClick={() => closeDrawer()}><SVG name="times"></SVG></Button>
        </div>
      </div>
    );
  };

  const eventsMapper = {};
  const reverseEventsMapper = {};

  appliedQueries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q;
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
    />
  );

  if (queryType === 'funnel') {
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
        width={'600px'}
        className={'fa-drawer'}
      >

        <QueryComposer
          queries={queries}
          runQuery={runQuery}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runFunnelQuery={runFunnelQuery}
        />
      </Drawer>

      {showResult ? (
        <>
          {result}
        </>

      ) : (
          <CoreQueryHome
            setQueryType={setQueryType}
            setDrawerVisible={setDrawerVisible}
            setQueries={setQueries}
            setQueryOptions={setExtraOptions}
            setRowClicked={setRowClicked}
          />
      )}

    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

export default connect(mapStateToProps)(CoreQuery);
