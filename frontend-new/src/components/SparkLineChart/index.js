import React from 'react';
import { useSelector } from 'react-redux';
import ChartHeader from './ChartHeader';
import SparkChart from './Chart';
import { DASHBOARD_WIDGET_SECTION } from '../../utils/constants';
import { CHART_COLOR_1 } from '../../constants/color.constants';

function SparkLineChart({
  queries,
  chartsData,
  appliedColors,
  page,
  resultState,
  frequency,
  arrayMapper,
  height,
  cardSize = 1,
  section,
  title = 'chart'
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);

  if (queries.length > 1) {
    const count = section === DASHBOARD_WIDGET_SECTION ? 3 : queries.length;
    const colors = {};
    arrayMapper.forEach((elem, index) => {
      colors[elem.mapper] = appliedColors[index];
    });
    return (
      <div
        className={`flex items-center flex-wrap justify-center w-full ${
          cardSize !== 2 ? 'pt-4' : ''
        } `}
      >
        {queries.slice(0, count).map((q, index) => {
          const m = arrayMapper.find(
            (elem) => elem.eventName === q && elem.index === index
          );
          const { mapper, eventName } = m;
          let total = 0;
          const data = chartsData.map((elem) => ({
            date: elem.date,
            [mapper]: elem[mapper]
          }));
          const queryRow = resultState.data.metrics.rows.find(
            (elem) => elem[0] === index
          );
          total = queryRow ? queryRow[2] : 0;

          if (cardSize === 1 || cardSize === 0) {
            return (
              <div key={q + index} className='w-1/3 px-4 h-full'>
                <div className='flex flex-col'>
                  <ChartHeader
                    total={total}
                    query={
                      arrayMapper.find((elem) => elem.eventName === q)
                        ?.displayName
                    }
                    bgColor={appliedColors[index]}
                    eventNames={eventNames}
                    titleCharCount={cardSize === 0 ? 16 : null}
                  />
                  <div className='mt-8'>
                    <SparkChart
                      frequency={frequency}
                      page={page}
                      event={mapper}
                      chartData={data}
                      chartColor={appliedColors[index]}
                      height={height}
                      title={title}
                      eventTitle={eventName}
                    />
                  </div>
                </div>
              </div>
            );
          }
          return (
            <div
              style={{ minWidth: '300px' }}
              key={q + index}
              className='w-1/3 mt-6 px-4'
            >
              <div className='flex flex-col'>
                <ChartHeader
                  total={total}
                  query={
                    arrayMapper.find((elem) => elem.eventName === q).displayName
                  }
                  bgColor={appliedColors[index]}
                  smallFont={true}
                  eventNames={eventNames}
                />
              </div>
            </div>
          );
        })}
      </div>
    );
  }
  const total = resultState.data.metrics.rows.find((elem) => elem[0] === 0)[2];

  const m = arrayMapper.find((elem) => elem.eventName === queries[0]);
  const { mapper, eventName } = m;

  return (
    <div className='flex items-center justify-center w-full  h-full'>
      <div
        className={`flex items-center justify-center h-full ${
          cardSize == 2 ? 'flex-col w-full' : cardSize == 0 ? 'w-4/5' : 'w-3/5'
        }`}
      >
        <div className={`${cardSize === 2 ? 'w-full' : 'w-1/2'}`}>
          <ChartHeader
            bgColor={CHART_COLOR_1}
            query={queries[0]}
            total={total}
            eventNames={eventNames}
          />
        </div>
        <div
          className={`flex justify-center items-center ${
            cardSize === 2 ? 'w-full' : 'w-1/2'
          }`}
        >
          <div className={`${cardSize === 2 ? 'w-3/5' : 'w-full'}`}>
            <SparkChart
              frequency={frequency}
              page={page}
              event={mapper}
              chartData={chartsData}
              chartColor={CHART_COLOR_1}
              height={height}
              title={title}
              eventTitle={eventName}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default SparkLineChart;
