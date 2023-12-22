import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import { eventsGroupedByGranularity } from '../../utils';
import { AccountTimelineTableViewProps } from './types';
import EventDrawer from './EventDrawer';
import TableRow from './TableRow';

const AccountTimelineTableView: React.FC<AccountTimelineTableViewProps> = ({
  timelineEvents = [],
  timelineUsers = [],
  loading
}) => {
  const [formattedData, setFormattedData] = useState<{ [key: string]: any }>(
    {}
  );
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<any | null>(null);

  useEffect(() => {
    const data = eventsGroupedByGranularity(
      timelineEvents.filter((item) => item.user !== 'milestone'),
      'Timeline'
    );
    setFormattedData(data);
  }, [timelineEvents]);

  const handleEventClick = (event: any) => {
    setSelectedEvent(event);
    setModalVisible(true);
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : (
    <>
      <div className='account-timeline-table-container'>
        <table className='account-timeline-table'>
          <tbody>
            {Object.entries(formattedData || {}).map(([timestamp, events]) => (
              <React.Fragment>
                <tr className='timestamp-row'>
                  <td>
                    <span>{timestamp}</span>
                  </td>
                </tr>
                {events.map((event: any) => {
                  const currentUser = timelineUsers.find(
                    (obj) => obj.userId === event.user
                  );
                  return (
                    currentUser && (
                      <TableRow
                        event={event}
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
      />
    </>
  );
};

export default AccountTimelineTableView;
