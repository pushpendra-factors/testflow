import React from 'react';
import ChartHeader from './ChartHeader';
import SparkChart from './Chart';

function SparkLineChart({
  queries, chartsData, parentClass, appliedColors, eventsMapper
}) {
  if (queries.length > 1) {
    return (
      <div className={parentClass}>
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
                <ChartHeader total={total} query={q} bgColor={appliedColors[index]} />
                <div className="mt-8">
                  <SparkChart event={eventsMapper[q]} page="totalEvents" chartData={data} chartColor={appliedColors[index]} />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  } else {
    let total = 0;
    chartsData.forEach(elem => {
      total += elem[eventsMapper[queries[0]]];
    });

    return (
      <div className={parentClass}>
        <div className="w-1/4">
          <ChartHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
        </div>
        <div className="w-3/4">
          <SparkChart event={eventsMapper[queries[0]]} page="totalEvents" chartData={chartsData} chartColor="#4D7DB4" />
        </div>
      </div>
    );
  }
}

export default SparkLineChart;
