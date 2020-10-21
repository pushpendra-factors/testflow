import React, { useCallback } from 'react';
import { Text } from '../factorsComponents';

function ChartLegends({
  events, colors, eventsMapper, focusHoveredLines, focusAllLines, setHiddenEvents, hiddenEvents
}) {
  const handleLegendClick = useCallback((event) => {
    let newHiddenEventsArr = [];
    if (hiddenEvents.indexOf(event) === -1) {
      newHiddenEventsArr = [...hiddenEvents, event];
    } else {
      newHiddenEventsArr = hiddenEvents.filter(elem => elem !== event);
    }
    if (newHiddenEventsArr.length === events.length) {
      return false;
    }
    setHiddenEvents(newHiddenEventsArr);
  }, [hiddenEvents, setHiddenEvents, events.length]);

  return (
    <div className="flex flex-wrap items-center justify-center w-full">
      {events.map((event, index) => {
        const label = event.split(',').filter(elem => elem).join(',');
        return (
          <div onClick={handleLegendClick.bind(this, event)} onMouseOver={focusHoveredLines.bind(this, eventsMapper[event])} onMouseOut={focusAllLines} key={event + index} className={`${hiddenEvents && hiddenEvents.indexOf(event) > -1 ? 'opacity-25' : 'opacity-100'} flex items-center cursor-pointer`}>
            <div style={{
              backgroundColor: colors[eventsMapper[event]], width: '16px', height: '16px', borderRadius: '8px'
            }}></div>
            <div className="px-2" key={event + index}><Text mini type="paragraph">{label}</Text></div>
          </div>
        );
      })}
    </div >
  );
}

export default ChartLegends;
