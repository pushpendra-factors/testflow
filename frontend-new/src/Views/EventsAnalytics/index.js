import React from 'react';
import Header from '../AppLayout/Header';
import { Button } from 'antd';
import { PoweroffOutlined } from '@ant-design/icons';
import EventsInfo from '../CoreQuery/FunnelsResultPage/EventsInfo';
import ContentTabs from '../../components/ContentTabs';
import ResultTab from './ResultTab.js';
import { SVG } from '../../components/factorsComponents';
import EventBreakdown from './EventBreakdown';

function EventsAnalytics({
  queries, eventsMapper, reverseEventsMapper, breakdown, resultState, setDrawerVisible, runQuery, activeKey, breakdownType, handleBreakdownTypeChange, breakdownTypeData, queryType
}) {
  const handleTabChange = (tabKey) => {
    runQuery(tabKey);
  };

  let totalUsersTabContent = null;

  if (activeKey === '1' && breakdownType === 'each') {
    totalUsersTabContent = (
      <ResultTab handleBreakdownTypeChange={handleBreakdownTypeChange} breakdownType={breakdownType} activeKey={activeKey} index={1} page="totalUsers" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} />
    );
  } else if (activeKey === '1' && breakdownType !== 'each') {
    totalUsersTabContent = (
      <EventBreakdown data={breakdownTypeData} queries={queries} breakdown={breakdown} breakdownType={breakdownType} handleBreakdownTypeChange={handleBreakdownTypeChange} />
    );
  }

  const tabItems = [
    {
      key: '0',
      title: 'Total Events',
      titleIcon: <SVG name={'totalevents'} size={24} color={activeKey === '1' ? '#3E516C' : '#8692A3'} />,
      content: activeKey === '0' ? <ResultTab handleBreakdownTypeChange={handleBreakdownTypeChange} breakdownType={breakdownType} index={0} page="totalEvents" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} /> : null
    },
    {
      key: '1',
      title: 'Total Users',
      titleIcon: <SVG name={'totalusers'} size={24} color={activeKey === '2' ? '#3E516C' : '#8692A3'} />,
      content: activeKey === '1' ? totalUsersTabContent : null
    },
    {
      key: '2',
      title: 'Active Users',
      titleIcon: <SVG name={'activeusers'} size={24} color={activeKey === '3' ? '#3E516C' : '#8692A3'} />,
      content: activeKey === '2' ? <ResultTab handleBreakdownTypeChange={handleBreakdownTypeChange} breakdownType={breakdownType} activeKey={activeKey} index={2} page="activeUsers" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} /> : null
    },
    {
      key: '3',
      title: 'Frequency',
      titleIcon: <SVG name={'frequency'} size={24} color={activeKey === '4' ? '#3E516C' : '#8692A3'} />,
      content: activeKey === '3' ? <ResultTab handleBreakdownTypeChange={handleBreakdownTypeChange} breakdownType={breakdownType} activeKey={activeKey} index={3} page="frequency" resultState={resultState} breakdown={breakdown} eventsMapper={eventsMapper} reverseEventsMapper={reverseEventsMapper} queries={queries} /> : null
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
            queryType={queryType}
          />
        </div>
      </Header>
      <div className="mt-32 mb-8 fa-container">
        <ContentTabs breakdownTypeData={breakdownTypeData} resultState={resultState} onChange={handleTabChange} activeKey={activeKey} tabItems={tabItems} />
      </div>
    </>
  );
}

export default EventsAnalytics;
