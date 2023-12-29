import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import { eventsGroupedByGranularity } from '../../utils';
import { AccountTimelineTableViewProps } from './types';
import EventDrawer from './EventDrawer';
import TableRow from './TableRow';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

const AccountTimelineTableView: React.FC<AccountTimelineTableViewProps> = ({
  timelineEvents = [],
  timelineUsers = [],
  eventPropsType,
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
    document.title = 'Accounts - FactorsAI';
  }, [timelineEvents]);

  const handleEventClick = (event: any) => {
    setSelectedEvent(event);
    setModalVisible(true);
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
    <NoDataWithMessage message={'No Associated Users'} />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message={'No Events Enabled to Show'} />
  ) : (
    <>
      <div className='account-timeline-table-container'>
        <table>
          <tbody>
            {Object.entries(formattedData || {}).map(([timestamp, events]) => (
              <React.Fragment>
                <tr className='timestamp-row'>
                  <td>{timestamp}</td>
                </tr>
                {events.map((event: any) => {
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
};

export default AccountTimelineTableView;
