import React, { useState } from 'react';
import { Menu, Dropdown, Button } from 'antd';
import { getSingleEventAnalyticsData, getDataInLineChartFormat } from '../utils';
import SparkChart from '../TotalEvents/SparkChart';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import EventHeader from '../EventHeader';
import TotalEventsTable from '../TotalEvents/TotalEventsTable';
import { SVG } from '../../../components/factorsComponents';
import styles from '../index.module.scss';
import LineChart from './LineChart';

function TotalUsers({ queries }) {
  const appliedColors = generateColors(queries.length);
  const chartsData = getSingleEventAnalyticsData(queries);

  const [chartType, setChartType] = useState('sparklines');

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

  return (
    <div className="totalUsers">
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
        <div className="flex flex-wrap mt-8">
          {queries.map((q, index) => {
            let total = 0;
            const data = chartsData.map(elem => {
              return {
                date: elem.date,
                [q]: elem[q]
              };
            });
            data.forEach(elem => {
              total += elem[q];
            });

            return (
              <div key={q + index} className="w-1/3 mt-4 px-1">
                <div className="flex flex-col">
                  <EventHeader total={total} query={q} bgColor={appliedColors[index]} />
                  <div className="mt-8">
                    <SparkChart event={q} page="totalUsers" chartData={data} chartColor={appliedColors[index]} />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
          <div className="flex mt-8">
            <LineChart chartData={getDataInLineChartFormat(chartsData, queries)} appliedColors={appliedColors} queries={queries} />
          </div>
        )}

      <div className="mt-8">
        <TotalEventsTable
          data={chartsData}
          events={queries}
        />
      </div>
    </div>

  );
}

export default TotalUsers;
