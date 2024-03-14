import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import {
  AccountTimelineTableViewProps,
  TimelineEvent,
  TimelineUser
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
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<any | null>(null);
  const [selectedUser, setSelectedUser] = useState<any | null>(null);

  useEffect(() => {
    const data = eventsGroupedByGranularity(
      timelineEvents.filter((item) => item.username !== 'milestone'),
      'Timeline'
    );
    setFormattedData(data);
    document.title = 'Accounts - FactorsAI';
  }, [timelineEvents]);

  const handleEventClick = (event: TimelineEvent, user: TimelineUser) => {
    setSelectedEvent(event);
    setSelectedUser(user);
    setDrawerVisible(true);
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
            {Object.entries(formattedData || {}).map(([date, events]) => (
              <>
                <tr className='timestamp-row'>
                  <td>{date}</td>
                </tr>
                {events.map((event) => {
                  const currentUser = timelineUsers.find(
                    (obj) => obj.id === event.user_id
                  );
                  return (
                    currentUser && (
                      <TableRow
                        event={event}
                        eventPropsType={eventPropsType}
                        user={currentUser}
                        onEventClick={() =>
                          handleEventClick(event, currentUser)
                        }
                      />
                    )
                  );
                })}
              </>
            ))}
          </tbody>
        </table>
      </div>
      <EventDrawer
        visible={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        event={selectedEvent}
        user={selectedUser}
        eventPropsType={eventPropsType}
      />
    </>
  );
}

export default AccountTimelineTableView;
