import React, { useState, useEffect, useRef } from 'react';
import { Spin } from 'antd';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import {
  AccountTimelineTableViewProps,
  NewEvent
} from 'Components/Profile/types';
import { SVG } from 'Components/factorsComponents';
import { eventsGroupedByGranularity } from '../../utils';
import EventDrawer from './EventDrawer';
import TableRow from './TableRow';

function AccountTimelineTableView({
  timelineEvents = [],
  eventPropsType,
  userPropsType,
  loading,
  extraClass,
  eventDrawerVisible,
  setEventDrawerVisible,
  hasScrollAction,
  setScrollPercent
}: AccountTimelineTableViewProps) {
  const [formattedData, setFormattedData] = useState<{
    [key: string]: NewEvent[];
  }>({});
  const [selectedEvent, setSelectedEvent] = useState<NewEvent>();
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const handleScroll = () => {
    const container = scrollContainerRef.current;
    if (!container) {
      setScrollPercent(0);
      return;
    }

    const { scrollTop, scrollHeight, clientHeight } = container;
    const percentage = (scrollTop / (scrollHeight - clientHeight)) * 100;
    setScrollPercent(percentage);
  };

  useEffect(() => {
    let flag = false;
    const scrollContainer = scrollContainerRef.current;

    if (hasScrollAction && scrollContainer) {
      flag = true;
      scrollContainer.addEventListener('scroll', handleScroll);
    }

    return () => {
      if (flag && scrollContainer)
        scrollContainer.removeEventListener('scroll', handleScroll);
    };
  }, [scrollContainerRef, timelineEvents]);

  useEffect(() => {
    const data = eventsGroupedByGranularity(
      timelineEvents.filter((item) => item.username !== 'milestone'),
      'Timeline'
    );
    setFormattedData(data);
    document.title = 'Accounts - FactorsAI';
  }, [timelineEvents]);

  const handleEventClick = (event: NewEvent) => {
    setSelectedEvent(event);
    setEventDrawerVisible(true);
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    <>
      <div
        ref={scrollContainerRef}
        className={`timeline-table-container bordered-gray--bottom ${
          extraClass || ''
        }`}
      >
        <table>
          <tbody>
            {Object.entries(formattedData || {}).map(([date, events]) => (
              <>
                <tr className='timestamp-row'>
                  <td className='inline-flex gap--4'>
                    <SVG name='calendar' />
                    {date}
                  </td>
                </tr>
                {events.map((event) => (
                  <TableRow
                    event={event}
                    eventPropsType={eventPropsType}
                    onEventClick={() => handleEventClick(event)}
                  />
                ))}
              </>
            ))}
          </tbody>
        </table>
      </div>
      <EventDrawer
        visible={eventDrawerVisible}
        onClose={() => setEventDrawerVisible(false)}
        event={selectedEvent}
        eventPropsType={eventPropsType}
        userPropsType={userPropsType}
      />
    </>
  );
}

export default AccountTimelineTableView;
