import React from 'react';
import ChartHeader from './ChartHeader';
import SparkChart from './Chart';
import { DASHBOARD_WIDGET_SECTION } from '../../utils/constants';
import TopLegends from '../GroupedBarChart/TopLegends';
import { Text, Number as NumFormat } from '../factorsComponents';
import { values } from 'lodash';

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
  if (queries.length > 1) {
    const count = section === DASHBOARD_WIDGET_SECTION ? 3 : queries.length;
    const colors = {};
    arrayMapper.forEach((elem, index) => {
      colors[elem.mapper] = appliedColors[index];
    });
    return (
      <div
        className={`flex items-center flex-wrap justify-center w-full ${
          !cardSize ? 'flex-col' : ''
        }`}
      >
        {!cardSize ? (
          <TopLegends
            cardSize={cardSize}
            colors={values(colors)}
            legends={queries.map(
              (q) =>
                arrayMapper.find((elem) => elem.eventName === q)?.displayName
            )}
          />
        ) : null}
        {queries.slice(0, count).map((q, index) => {
          const mapper = arrayMapper.find(
            (elem) => elem.eventName === q && elem.index === index
          ).mapper;
          let total = 0;
          const data = chartsData.map((elem) => {
            return {
              date: elem.date,
              [mapper]: elem[mapper]
            };
          });
          const queryRow = resultState.data.metrics.rows.find(
            (elem) => elem[0] === index
          );
          total = queryRow ? queryRow[2] : 0;

          if (cardSize === 0) {
            return (
              <div
                key={q + index}
                className="flex items-center w-full justify-center"
              >
                <Text
                  extraClass="flex items-center w-1/4 justify-center"
                  type={'title'}
                  level={3}
                  weight={'bold'}
                >
                  <NumFormat shortHand={true} number={total} />
                </Text>
                <div className="w-2/3">
                  <SparkChart
                    frequency={frequency}
                    page={page}
                    event={mapper}
                    chartData={data}
                    chartColor={appliedColors[index]}
                    height={40}
                    title={title}
                  />
                </div>
              </div>
            );
          } else if (cardSize === 1) {
            return (
              <div
                style={{ minWidth: '300px' }}
                key={q + index}
                className="w-1/3 mt-4 px-4"
              >
                <div className="flex flex-col">
                  <ChartHeader
                    total={total}
                    query={
                      arrayMapper.find((elem) => elem.eventName === q)
                        ?.displayName
                    }
                    bgColor={appliedColors[index]}
                  />
                  <div className="mt-8">
                    <SparkChart
                      frequency={frequency}
                      page={page}
                      event={mapper}
                      chartData={data}
                      chartColor={appliedColors[index]}
                      height={height}
                      title={title}
                    />
                  </div>
                </div>
              </div>
            );
          } else {
            return (
              <div
                style={{ minWidth: '300px' }}
                key={q + index}
                className="w-1/3 mt-6 px-4"
              >
                <div className="flex flex-col">
                  <ChartHeader
                    total={total}
                    query={
                      arrayMapper.find((elem) => elem.eventName === q)
                        .displayName
                    }
                    bgColor={appliedColors[index]}
                    smallFont={true}
                  />
                </div>
              </div>
            );
          }
        })}
      </div>
    );
  } else {
    const total = resultState.data.metrics.rows.find(
      (elem) => elem[0] === 0
    )[2];

    return (
      <div
        className={`flex items-center justify-center w-full ${
          cardSize !== 1 ? 'flex-col' : ''
        }`}
      >
        <div className={cardSize === 1 ? 'w-1/4' : 'w-full'}>
          <ChartHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
        </div>
        <div className={cardSize === 1 ? 'w-3/4' : 'w-full'}>
          <SparkChart
            frequency={frequency}
            page={page}
            event={
              arrayMapper.find((elem) => elem.eventName === queries[0]).mapper
            }
            chartData={chartsData}
            chartColor="#4D7DB4"
            height={height}
            title={title}
          />
        </div>
      </div>
    );
  }
}

export default SparkLineChart;
