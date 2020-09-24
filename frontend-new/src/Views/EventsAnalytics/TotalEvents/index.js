import React, { useState } from 'react';
import { Menu, Dropdown, Button } from 'antd';
import SparkChart from './SparkChart';
import TotalEventsTable from './TotalEventsTable';
import { getSingleEventAnalyticsData, getDataInLineChartFormat, getMultiEventsAnalyticsData } from '../utils';
import EventHeader from '../EventHeader';
import styles from '../index.module.scss';
import { SVG } from '../../../components/factorsComponents';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import LineChart from '../TotalUsers/LineChart';


function TotalEvents({ queries }) {
  const appliedColors = generateColors(queries.length);
  const [chartType, setChartType] = useState('sparklines');

  const eventsMapper = {};
  const reverseEventsMapper = {};
  queries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index}`;
    reverseEventsMapper[`event${index}`] = q;
  })

  let chartsData;
  if (queries.length === 1) {
    chartsData = getSingleEventAnalyticsData(queries[0], eventsMapper);
  } else {
    chartsData = getMultiEventsAnalyticsData(queries, eventsMapper);
  }

  if (!chartsData.length) {
    return null;
  }

  const menuItems = [
    {
      key: 'sparklines',
      onClick: setChartType,
      name: 'Sparkline',
    },
    {
      key: 'linechart',
      onClick: setChartType,
      name: 'Line Chart',
    }
  ]

  const menu = (
    <Menu className={styles.dropdownMenu}>
      {menuItems.map(item => {
        return (
          <Menu.Item key={item.key} onClick={setChartType.bind(this, item.key)} className={`${styles.dropdownMenuItem} ${chartType === item.key ? styles.active : ''}`}>
            <div className={`flex items-center`}>
              <SVG extraClass="mr-1" name={item.key} size={25} color={chartType === item.key ? '#8692A3' : '#3E516C'} />
              <span className="mr-3">{item.name}</span>
              {chartType === item.key ? (
                <SVG name="checkmark" size={17} color="#8692A3" />
              ) : null}
            </div>
          </Menu.Item>
        )
      })}
    </Menu>
  );

  let sparkLinesJsx;

  if (queries.length === 1) {

    let total = 0;

    chartsData.forEach(elem => {
      total += elem[eventsMapper[queries[0]]];
    });

    sparkLinesJsx = (
      <div className="flex justify-center items-center mt-8">
        <div className="w-1/4">
          <EventHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
        </div>
        <div className="w-3/4">
          <SparkChart event={eventsMapper[queries[0]]} page="totalEvents" chartData={chartsData} chartColor="#4D7DB4" />
        </div>
      </div>
    )
  } else {
    sparkLinesJsx = (
      <div className="flex flex-wrap mt-8">
        {queries.map((q, index) => {
          let total = 0;
          const data = chartsData.map(elem => {
            return {
              date: elem.date,
              [eventsMapper[q]]: elem[eventsMapper[q]]
            };
          });
          data.forEach(elem => {
            total += elem[eventsMapper[q]];
          });

          return (
            <div key={q + index} className="w-1/3 mt-4 px-1">
              <div className="flex flex-col">
                <EventHeader total={total} query={q} bgColor={appliedColors[index]} />
                <div className="mt-8">
                  <SparkChart event={eventsMapper[q]} page="totalEvents" chartData={data} chartColor={appliedColors[index]} />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    )
  }

  return (
    <div className="total-events">
      <div className="flex items-center justify-between">
        <div className="filters-info">

        </div>
        <div className="user-actions">
          <Dropdown overlay={menu}>
            <Button className={`ant-dropdown-link flex items-center ${styles.dropdownBtn}`}>
              <SVG name={chartType} size={25} color="#0E2647" />
              <SVG name={'dropdown'} size={25} color="#3E516C" />
            </Button>
          </Dropdown>
        </div>
      </div>
      {chartType === 'sparklines' ? (
        <div>{sparkLinesJsx}</div>

      ) : (
          <div className="flex mt-8">
            <LineChart chartData={getDataInLineChartFormat(chartsData, queries, eventsMapper)} appliedColors={appliedColors} queries={queries} />
          </div>
        )}
      <div className="mt-8">
        <TotalEventsTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
        />
      </div>
    </div>
  );
}

export default TotalEvents;
