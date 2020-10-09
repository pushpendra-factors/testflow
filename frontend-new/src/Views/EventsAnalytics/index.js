import React from 'react';
import Header from '../AppLayout/Header';
import { Button } from 'antd';
import { PoweroffOutlined } from '@ant-design/icons';
import EventsInfo from '../CoreQuery/FunnelsResultPage/EventsInfo';
import ContentTabs from '../../components/ContentTabs';
import TotalEvents from './TotalEvents';
import { SVG } from '../../components/factorsComponents';
import TotalUsers from './TotalUsers';
import Frequency from './Frequency';

function EventsAnalytics({
  queries, eventsMapper, reverseEventsMapper, breakdown, resultState, setDrawerVisible, runQuery, activeKey
}) {
  const handleTabChange = (tabKey) => {
    runQuery(tabKey);
  };

  const tabItems = [
    {
      key: '0',
      title: 'Total Events',
      titleIcon: <SVG name={'totalevents'} size={24} color={activeKey === '1' ? '#3E516C' : '#8692A3'} />,
      content: <TotalEvents index="0" page="totalEvents" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} />
    },
    {
      key: '1',
      title: 'Total Users',
      titleIcon: <SVG name={'totalusers'} size={24} color={activeKey === '2' ? '#3E516C' : '#8692A3'} />,
      content: <TotalUsers index="1" page="totalUsers" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} />
    },
    {
      key: '2',
      title: 'Active Users',
      titleIcon: <SVG name={'activeusers'} size={24} color={activeKey === '3' ? '#3E516C' : '#8692A3'} />,
      content: <TotalUsers index="2" page="activeUsers" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} />
    },
    {
      key: '3',
      title: 'Frequency',
      titleIcon: <SVG name={'frequency'} size={24} color={activeKey === '4' ? '#3E516C' : '#8692A3'} />,
      content: <Frequency index="3" page="frequency" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} />
    }
  ];

  return (
    <>
      <Header>
        <div className="flex py-4 justify-end">
          <Button type="primary" icon={<PoweroffOutlined />} >Save query as</Button>
        </div>
        <div className="py-4">
          <EventsInfo
            setDrawerVisible={setDrawerVisible}
            queries={queries}
          />
        </div>
      </Header>
      <div className="mt-40 mb-8 fa-container">
        <ContentTabs resultState={resultState} onChange={handleTabChange} activeKey={activeKey} tabItems={tabItems} />
      </div>
    </>
  );
}

export default EventsAnalytics;
