import React, { useState, useCallback } from 'react';
import { connect } from 'react-redux';
import moment from 'moment';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';
import { runQuery as runQueryService } from '../../reducers/coreQuery/services';

const COND_ANY_GIVEN_EVENT = 'any_given_event';
const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
// const TYPE_UNIQUE_USERS = 'unique_users';

function CoreQuery({ activeProject }) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState('event');
  const [showResult, setShowResult] = useState(false);
  const [queries, setQueries] = useState([]);
  const [breakdown] = useState([]);
  const [resultState, setResultState] = useState({ loading: false, error: false, data: {} });

  const queryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        queryupdated[index] = newEvent;
      } else {
        queryupdated.splice(index, 1);
      }
    } else {
      queryupdated.push(newEvent);
    }
    setQueries(queryupdated);
  };

  const getEventsWithProperties = useCallback(() => {
    const ewps = [];
    queries.forEach(ev => {
      ewps.push({
        na: ev.label,
        pr: []
      });
    });
    return ewps;
  }, [queries]);

  const getQuery = useCallback(() => {
    const query = {};
    query.cl = queryType === 'event' ? 'insights' : 'funnel';
    query.ty = 'events_occurrence';
    query.ec = 'any_given_event';
    // event_occurrence supports only any_given_event.
    if (query.ty === TYPE_EVENT_OCCURRENCE) {
      query.ec = COND_ANY_GIVEN_EVENT;
    }

    // Check date range validity

    // let period = getQueryPeriod(this.state.resultDateRange[0], this.state.timeZone)

    const period = {
      from: moment().subtract(5, 'days').startOf('day').utc().unix(),
      to: moment().utc().unix()
    };

    query.fr = period.from;
    query.to = period.to;
    query.ewp = getEventsWithProperties();

    query.gbp = [];
    // for(let i=0; i < this.state.groupBys.length; i++) {
    //   let groupBy = this.state.groupBys[i];
    //   let cGroupBy = {};

    //   if (groupBy.name != '' && groupBy.type != '') {
    //     cGroupBy.pr = groupBy.name;
    //     cGroupBy.en = groupBy.type;
    //     cGroupBy.pty = groupBy.ptype

    //     // add group by event name.
    //     if (this.isEventNameRequiredForGroupBy() && groupBy.eventName != '') {
    //       let nameWithIndex = removeIndexIfExistsFromOptName(groupBy.eventName);
    //       cGroupBy.ena = nameWithIndex.name
    //       // let eni = getIndexIfExistsFromOptName(groupBy.eventName)
    //       if (!isNaN(nameWithIndex.index)) {
    //         cGroupBy.eni = nameWithIndex.index  // 1 valued index to distinguish in backend from default 0.
    //       }
    //     }
    //     query.gbp.push(cGroupBy)
    //   }
    // }

    // query.gbt = (presentation == PRESENTATION_LINE) ?
    //   getGroupByTimestampType(query.fr, query.to) : '';
    query.gbt = 'date';
    query.tz = 'Asia/Kolkata';

    // query.sse = sessionStartEvent.value
    // query.see = sessionEndEvent.value

    return query;
  }, [getEventsWithProperties, queryType]);

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const runQuery = useCallback(async () => {
    const query = getQuery();
    setResultState({ loading: true, error: false, data: {} });
    closeDrawer();
    setShowResult(true);
    try {
      const res = await runQueryService(activeProject.id, query);
      if (res.status === 200) {
        setResultState({ loading: false, error: false, data: res.data });
      } else {
        setResultState({ loading: false, error: true, data: {} });
      }
    } catch (err) {
      setResultState({ loading: false, error: true, data: {} });
    }
  }, [activeProject, getQuery]);

  const title = () => {
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex'}>
          <SVG name={queryType === 'funnel' ? 'funnels_cq' : 'events_cq'} size="24px"></SVG>
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{queryType === 'funnel' ? 'Find event funnel for' : 'Analyse Events'}</Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Button type="text"><SVG name="play"></SVG>Help</Button>
          <Button type="text" onClick={() => closeDrawer()}><SVG name="times"></SVG></Button>
        </div>
      </div>
    );
  };

  const eventsMapper = {};
  const reverseEventsMapper = {};
  queries.forEach((q, index) => {
    eventsMapper[`${q.label}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q.label;
    // eventsMapper[`${q}`] = `event${index + 1}`;
    // reverseEventsMapper[`event${index + 1}`] = q;
  });

  let result = (
    <EventsAnalytics
      queries={queries.map(elem => elem.label)}
      // queries={queries}
      eventsMapper={eventsMapper}
      reverseEventsMapper={reverseEventsMapper}
      breakdown={breakdown}
      resultState={resultState}
      setDrawerVisible={setDrawerVisible}
    />
  );

  if (queryType === 'funnel') {
    result = (
      <FunnelsResultPage
        setDrawerVisible={setDrawerVisible}
        queries={queries.map(elem => elem.label)}
        // queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
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
        />
      </Drawer>

      {showResult ? (
        <>
          {result}
        </>

      ) : (
          <CoreQueryHome setQueryType={setQueryType} setDrawerVisible={setDrawerVisible} />
      )}

    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

export default connect(mapStateToProps)(CoreQuery);
