import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import {
  AccountTimelineTableViewProps,
  TimelineEvent
} from 'Components/Profile/types';
import { eventsGroupedByGranularity } from '../../utils';
import EventDrawer from './EventDrawer';
import TableRow from './TableRow';

function AccountTimelineTableView({
  timelineEvents = [],
  timelineUsers = [],
  eventPropsType,
  loading
}: AccountTimelineTableViewProps) {
  const [formattedData, setFormattedData] = useState<{
    [key: string]: TimelineEvent[];
  }>({});
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<any | null>(null);

  useEffect(() => {
    const data = eventsGroupedByGranularity(
      timelineEvents.filter((item) => item.user !== 'milestone'),
      'Timeline'
    );
    setFormattedData(data);
    document.title = 'Accounts - FactorsAI';
  }, [timelineEvents]);

  const handleEventClick = (event: TimelineEvent) => {
    setSelectedEvent(event);
    setModalVisible(true);
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
    <NoDataWithMessage message='No Associated Users' />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    <>
      <div className='account-timeline-table-container'>
        <table>
          <tbody>
            {Object.entries(formattedData || {}).map(([timestamp, events]) => (
              <React.Fragment key={timestamp}>
                <tr className='timestamp-row'>
                  <td>{timestamp}</td>
                </tr>
                {events.map((event) => {
                  const currentUser = timelineUsers.find(
                    (obj) => obj.userId === event.user
                  );
                  return (
                    currentUser && (
                      <TableRow
                        event={event}
                        eventPropsType={eventPropsType}
                        user={currentUser}
                        onEventClick={() => handleEventClick(event)}
                      />
                    )
                  );
                })}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
      <EventDrawer
        visible={modalVisible}
        onClose={() => setModalVisible(false)}
        event={selectedEvent}
        eventPropsType={eventPropsType}
      />
    </>
  );
}

export default AccountTimelineTableView;
