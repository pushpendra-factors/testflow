import React, { useState, useEffect } from 'react';
import { useDispatch } from 'react-redux';
import GroupedChart from './GroupedChart';
import DataTable from './DataTable';
import {
  generateGroupedChartsData, generateDummyData, generateGroups, generateUngroupedChartsData
} from './utils';
import FiltersInfo from './FiltersInfo';
import UngroupedChart from './UngroupedChart';
import { FUNNEL_RESULTS_AVAILABLE, FUNNEL_RESULTS_UNAVAILABLE } from '../../../reducers/types';
import Header from '../../AppLayout/Header';
import EventsInfo from './EventsInfo';
import { Button } from 'antd';
import { PoweroffOutlined } from '@ant-design/icons';

function PageContent({ queries, setDrawerVisible }) {
  const [eventsData, setEventsData] = useState([]);
  const [groups, setGroups] = useState([]);
  const [grouping, setGrouping] = useState(true);
  const dispatch = useDispatch();

  useEffect(() => {
    dispatch({ type: FUNNEL_RESULTS_AVAILABLE, payload: queries });
    return () => {
      dispatch({ type: FUNNEL_RESULTS_UNAVAILABLE });
    };
  }, [queries, dispatch]);

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
          <EventsInfo />
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
          <DataTable
            eventsData={eventsData}
            groups={groups}
            setGroups={setGroups}
          />
        </div>
      </div>
    </>
  );
}

export default PageContent;
