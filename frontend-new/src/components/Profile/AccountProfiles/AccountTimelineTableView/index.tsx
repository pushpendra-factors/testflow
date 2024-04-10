import React, { useState, useEffect } from 'react';
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
  loading,
  extraClass
}: AccountTimelineTableViewProps) {
  const [formattedData, setFormattedData] = useState<{
    [key: string]: NewEvent[];
  }>({});
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<NewEvent>();

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
    setDrawerVisible(true);
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    <>
      <div className={`account-timeline-table-container ${extraClass}`}>
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
        visible={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        event={selectedEvent}
        eventPropsType={eventPropsType}
      />
    </>
  );
}

export default AccountTimelineTableView;
