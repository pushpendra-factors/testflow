import React from 'react';
import { Text } from '../factorsComponents';

function ChartLegends({
  events, chartData
}) {
  return (
        <div className="flex flex-wrap items-center justify-center w-full">
            {events.map((event, index) => {
              const eventObj = chartData.find(elem => elem.event === event);
              const color = eventObj ? eventObj.color : null;
              if (!color) {
                return null;
              }
              return (
                    <div key={event + index} className="opacity-100 flex items-center cursor-pointer">
                        <div style={{
                          backgroundColor: color, width: '16px', height: '16px', borderRadius: '8px'
                        }}></div>
                        <div className="px-2" key={event + index}><Text mini type="paragraph">{event}</Text></div>
                    </div>
              );
            })}
        </div >
  );
}

export default ChartLegends;
