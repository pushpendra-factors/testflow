/* eslint-disable */
import React, { useState } from 'react';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';

function CoreQuery() {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState('event');
  const [showResult, setShowResult] = useState(false);
  // const [showResult, setShowResult] = useState(true);
  const [queries, setQueries] = useState([]);
  // const [queries, setQueries] = useState(["www.cars24.com/buy-used-cars", 'www.cars24.com/buy-used-car', 'www.cars24.com/account/appointments']);
  // const [queries, setQueries] = useState(['Paid', 'Add to Wishlist', 'Checkout'])

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

  const runQuery = () => {
    setShowResult(true);
    closeDrawer();
  };

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const title = () => {
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex'}>
          <SVG name="teamfeed"></SVG>
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>Find event funnel for</Text>
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
    eventsMapper[`${q}`] = `event${index+1}`;
    reverseEventsMapper[`event${index+1}`] = q;
  })

  let result = (
    <EventsAnalytics
      queryType={queryType}
      queries={queries.map(elem => elem.label)}
      eventsMapper={eventsMapper}
      reverseEventsMapper={reverseEventsMapper}
    />
  );

  if (queryType === 'funnel') {
    result = (
      <FunnelsResultPage
        setDrawerVisible={setDrawerVisible}
        queries={queries.map(elem => elem.label)}
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

export default CoreQuery;
