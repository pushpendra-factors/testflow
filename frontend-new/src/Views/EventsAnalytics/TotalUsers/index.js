import React from 'react';
import { getSpikeChartData } from '../utils';
import SpikeChart from '../TotalEvents/SpikeChart';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import EventHeader from '../EventHeader';
import TotalEventsTable from '../TotalEvents/TotalEventsTable';

function TotalUsers({ queries }) {
  const appliedColors = generateColors(queries.length);
  const chartsData = getSpikeChartData(queries);

  return (
        <div className="totalUsers">
            <div className="flex flex-wrap">
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
                                <SpikeChart event={q} page="totalUsers" chartData={data} chartColor={appliedColors[index]} />
                            </div>
                        </div>
                  );
                })}
            </div>
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
