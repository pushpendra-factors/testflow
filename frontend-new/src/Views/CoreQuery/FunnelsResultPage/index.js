import React, { useState, useEffect } from 'react';
import GroupedChart from './GroupedChart';
import {
  generateGroupedChartsData, generateDummyData, generateGroups, generateUngroupedChartsData
} from './utils';
import FiltersInfo from './FiltersInfo';
import UngroupedChart from './UngroupedChart';
import Header from '../../AppLayout/Header';
import EventsInfo from './EventsInfo';
import { Button } from 'antd';
import { PoweroffOutlined } from '@ant-design/icons';
import FunnelsResultTable from './FunnelsResultTable';

function FunnelsResultPage({
  queries, setDrawerVisible, setShowFunnels, showFunnels
}) {
  const [eventsData, setEventsData] = useState([]);
  const [groups, setGroups] = useState([]);
  const [grouping, setGrouping] = useState(true);

  useEffect(() => {
    const dummyData = generateDummyData(queries);
    setEventsData(dummyData);
    setGroups(generateGroups(dummyData));
  }, [queries]);

  const groupedChartData = generateGroupedChartsData(eventsData, groups);
  const ungroupedChartsData = generateUngroupedChartsData(eventsData);

  if (!eventsData.length) {
    return null;
  }

  return (
    <>
            <Header>
                <div className="flex py-4 justify-end">
                    <Button type="primary" icon={<PoweroffOutlined />} >Save query as</Button>
                </div>
                <div className="py-4">
                    <EventsInfo queries={queries} />
                </div>
                <div className="pb-2 flex justify-end">
                    <FiltersInfo grouping={grouping} setGrouping={setGrouping} setDrawerVisible={setDrawerVisible} />
                </div>
            </Header>
            <div className="mt-40 mb-8 fa-container">
                {grouping ? (
                    <GroupedChart
                        chartData={groupedChartData}
                        groups={groups.filter(elem => elem.is_visible)}
                        eventsData={eventsData}
                    />
                ) : (
                        <UngroupedChart
                            chartData={ungroupedChartsData}
                        />
                )}

                <div className="mt-8">
                    <FunnelsResultTable
                        eventsData={eventsData}
                        groups={groups}
                        setGroups={setGroups}
                    />
                </div>
            </div>
    </>
  );
}

export default FunnelsResultPage;
