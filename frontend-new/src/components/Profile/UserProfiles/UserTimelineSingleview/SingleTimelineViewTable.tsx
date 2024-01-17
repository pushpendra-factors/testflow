import { useSelector } from 'react-redux';
import { eventIconsColorMap } from 'Components/Profile/constants';
import EventInfoCard from 'Components/Profile/MyComponents/EventInfoCard';
import { getEventCategory, getIconForCategory } from 'Components/Profile/utils';
import React from 'react';
import { TimelineEvent } from 'Components/Profile/types';

interface SingleTimelineViewTableProps {
  data: { [key: string]: TimelineEvent[] };
  propertiesType: { [key: string]: string };
}

function SingleTimelineViewTable({
  data,
  propertiesType
}: SingleTimelineViewTableProps): JSX.Element {
  const { eventNamesMap } = useSelector((state: any) => state.coreQuery);

  return (
    <div className='table-scroll'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date</th>
            <th scope='col' />
          </tr>
        </thead>
        <tbody>
          {Object.entries(data).map(([timestamp, events]) => {
            const timelineEvents = events.filter(
              (event) => event.event_type !== 'milestone'
            );
            const milestones = events.filter(
              (event) => event.event_type === 'milestone'
            );
            if (milestones && !timelineEvents.length) return null;
            return (
              <tr key={timestamp}>
                <td>
                  <div className='timestamp top-40'>{timestamp}</div>
                </td>
                <td className={`bg-none pb-${milestones.length * 0}`}>
                  <div className='user-timeline--events'>
                    {timelineEvents.map((event, index: number) => {
                      const category = getEventCategory(event, eventNamesMap);
                      const sourceIcon = getIconForCategory(category);
                      const eventIcon = eventIconsColorMap[event.icon]
                        ? event.icon
                        : 'calendar-star';
                      return (
                        <EventInfoCard
                          key={index}
                          event={event}
                          eventIcon={eventIcon}
                          sourceIcon={sourceIcon}
                          propertiesType={propertiesType}
                        />
                      );
                    })}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default SingleTimelineViewTable;
