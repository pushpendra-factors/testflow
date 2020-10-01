/* eslint-disable */
import React, { useState } from 'react';
import { connect } from 'react-redux';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';

import { runQuery as runQueryService } from '../../reducers/coreQuery/services';


const COND_ANY_GIVEN_EVENT = 'any_given_event';
const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
const TYPE_UNIQUE_USERS = 'unique_users';

function CoreQuery({ activeProject }) {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState('event');
  const [showResult, setShowResult] = useState(false);
  const [queries, setQueries] = useState([]);
  const [breakdown, setBreakdown] = useState([]);
  const [resultState, setResultState] = useState('none');
  const [queryResult, setQueryResult] = useState({})


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

  const getEventsWithProperties = (events) => {

    let ewps = [];
    events.forEach(ev => {
      ewps.push({
        na: ev.label,
        pr: []
      })
    })
    return ewps;
  }


  const getQuery = (presentation, toSave = false) => {
    let query = {};
    query.cl = queryType === 'event' ? 'insights' : 'funnel';
    query.ty = 'events_occurrence';
    query.ec = 'any_given_event';
    // event_occurrence supports only any_given_event.
    if (query.ty == TYPE_EVENT_OCCURRENCE) {
      query.ec = COND_ANY_GIVEN_EVENT;
    }

    // Check date range validity

    // let period = getQueryPeriod(this.state.resultDateRange[0], this.state.timeZone)

    const period = {
      from: 1601145000,
      to: 1601404199
    }

    query.fr = period.from
    query.to = period.to
    query.ewp = getEventsWithProperties(queries);

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

    query.tz = "Asia/Kolkata";

    // query.sse = sessionStartEvent.value
    // query.see = sessionEndEvent.value

    return query
  }

  const runQuery = () => {
    const query = getQuery();
    setResultState('loading');
    setShowResult(true);
    closeDrawer();
    runQueryService(activeProject.id, query).then(res => {
      if (res.status === 200) {
        setQueryResult(res.data);
        setResultState('success');
        setShowResult(true);
        closeDrawer();
      } else {
        setResultError()
      }
    }, err => {
      setResultError();
    })

  };

  const setResultError = () => {
    console.log(err);
    setResultState('error');
  }

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const title = () => {
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex'}>
          <SVG name={queryType === 'funnel' ? "funnels_cq" : "events_cq"} size="24px"></SVG>
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{queryType === 'funnel' ? "Find event funnel for" : "Analyse Events"}</Text>
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
    // eventsMapper[`${q.label}`] = `event${index+1}`;
    // reverseEventsMapper[`event${index+1}`] = q.label;
    eventsMapper[`${q}`] = `event${index + 1}`;
    reverseEventsMapper[`event${index + 1}`] = q;
  })

  let result = (
    <EventsAnalytics
      //queries={queries.map(elem => elem.label)}
      queries={queries}
      eventsMapper={eventsMapper}
      reverseEventsMapper={reverseEventsMapper}
      breakdown={breakdown}
    />
  );

  if (queryType === 'funnel') {
    result = (
      <FunnelsResultPage
        setDrawerVisible={setDrawerVisible}
        // queries={queries.map(elem => elem.label)}
        queries={queries}
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
