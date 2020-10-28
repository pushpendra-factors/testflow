import React, { useEffect, useState } from 'react';
import { PoweroffOutlined } from '@ant-design/icons';
import { generateUngroupedChartsData } from '../utils';
import Header from '../../../AppLayout/Header';
import { Button } from 'antd';
import EventsInfo from '../EventsInfo';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';

function UngroupedChart({ resultState, queries, setDrawerVisible, eventsMapper }) {

  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(resultState.data, queries)
    setChartData(formattedData);
  }, [queries, resultState.data])

  if (!chartData.length) {
    return null;
  }

  return (
    <>
      <Header>
        <div className="flex py-4 justify-end">
          <Button type="primary" icon={<PoweroffOutlined />} >Save query as</Button>
        </div>
        <div className="py-4">
          <EventsInfo setDrawerVisible={setDrawerVisible} queries={queries} />
        </div>
      </Header>

      <div className="mt-40 mb-8 fa-container">

        <Chart
          chartData={chartData}
        />

        <div className="mt-8">
          <FunnelsResultTable
            chartData={chartData}
            breakdown={[]}
            queries={queries}
            groups={[]}
            eventsMapper={eventsMapper}
          />
        </div>
      </div>
    </>
  )
}

export default UngroupedChart;