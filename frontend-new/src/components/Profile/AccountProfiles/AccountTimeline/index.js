import { Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import InfoCard from '../../../FaTimeline/InfoCard';
import {
  eventsFormattedForGranularity,
  hoverEvents,
  toggleCellCollapse
} from '../../utils';
import { SVG, Text } from '../../../factorsComponents';

function AccountTimeline({
  timelineEvents = [],
  timelineUsers = [],
  granularity,
  collapseAll,
  setCollapseAll,
  loading = false
}) {
  const [formattedData, setFormattedData] = useState({});

  useEffect(() => {
    const data = eventsFormattedForGranularity(
      timelineEvents,
      granularity,
      collapseAll
    );
    setFormattedData(data);
  }, [timelineEvents, granularity]);

  useEffect(() => {
    const data = {};
    Object.keys(formattedData).forEach((key) => {
      data[key] = formattedData[key];
      Object.keys(formattedData[key]).forEach((username) => {
        data[key][username] = formattedData[key][username];
        data[key][username].collapsed =
          collapseAll === undefined
            ? formattedData[key][username].collapsed
            : collapseAll;
      });
    });
    setFormattedData(data);
  }, [collapseAll]);

  const renderInfoCard = (event) => {
    const eventName =
      event.display_name === 'Page View'
        ? event.event_name
        : event?.alias_name || event.display_name;
    return (
      <InfoCard
        title={event?.alias_name || event.display_name}
        eventName={event?.event_name}
        properties={event?.properties || {}}
        trigger={
          hoverEvents.includes(event.event_name) ||
          event.display_name === 'Page View' ||
          event.event_type === 'CH' ||
          event.event_type === 'CS'
            ? 'hover'
            : []
        }
      >
        <div className='flex items-center font-medium'>
          <span className='truncate mx-1'>{eventName}</span>
          {hoverEvents.includes(event.event_name) ||
          event.display_name === 'Page View' ||
          event.event_type === 'CH' ||
          event.event_type === 'CS' ? (
            <CaretRightOutlined />
          ) : null}
        </div>
      </InfoCard>
    );
  };

  const renderAdditionalDiv = (eventsCount, collapseState, onClick) =>
    eventsCount > 1 ? (
      collapseState ? (
        <div className='timeline-events__num ml-1' onClick={onClick}>
          {`+${Number(eventsCount - 1)}`}
        </div>
      ) : (
        <div className='timeline-events__num m-5' onClick={onClick}>
          <CaretUpOutlined /> Show Less
        </div>
      )
    ) : null;

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
    <div className='ant-empty ant-empty-normal'>
      <div className='ant-empty-image'>
        <SVG name='nodata' />
      </div>
      <div className='ant-empty-description'>No Associated Users</div>
    </div>
  ) : (
    <div className='table-scroll'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date and Time</th>
            {timelineUsers.map((user) => (
              <th scope='col' className='truncate'>
                <Text type='title' truncate level={7} weight='medium'>
                  {user.title}
                </Text>
                <Text type='title' truncate level={8}>
                  {user.subtitle || '-'}
                </Text>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Object.entries(formattedData).map(([timestamp, allEvents]) => (
            <tr>
              <td>
                <div className='py-4'>{timestamp}</div>
              </td>
              {timelineUsers.map((user) => {
                if (!allEvents[user.title]) return <td />;
                const eventsList = allEvents[user.title].collapsed
                  ? allEvents[user.title].events.slice(0, 1)
                  : allEvents[user.title].events;
                return (
                  <td>
                    <div
                      className={`timeline-events ${
                        allEvents[user.title].collapsed
                          ? 'flex items-center'
                          : ''
                      }`}
                    >
                      {eventsList?.map((event) => (
                        <div className='timeline-events__event'>
                          {renderInfoCard(event)}
                        </div>
                      ))}
                      {renderAdditionalDiv(
                        allEvents[user.title].events.length,
                        allEvents[user.title].collapsed,
                        () => {
                          setFormattedData(
                            toggleCellCollapse(
                              formattedData,
                              timestamp,
                              user.title,
                              !allEvents[user.title].collapsed
                            )
                          );
                          setCollapseAll(undefined);
                        }
                      )}
                    </div>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
export default AccountTimeline;
