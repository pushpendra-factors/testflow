import { Spin } from 'antd';
import React, { useMemo, useState, useEffect } from 'react';
import InfoCard from '../../../FaTimeline/InfoCard';
import {
  eventsFormattedForGranularity,
  hoverEvents,
  toggleCellCollapse,
} from '../../utils';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import { SVG } from '../../../factorsComponents';
import { PropTextFormat } from '../../../../utils/dataFormatter';

function AccountTimeline({
  timelineEvents = [],
  timelineUsers = [],
  granularity,
  collapseAll,
  setCollapseAll,
  loading = false,
}) {
  const [formattedData, setFormattedData] = useState({});

  useEffect(() => {
    if (collapseAll !== undefined) {
      const data = eventsFormattedForGranularity(
        timelineEvents,
        granularity,
        collapseAll
      );
      setFormattedData(data);
    }
  }, [timelineEvents, granularity, collapseAll]);

  const renderInfoCard = (event) => {
    const eventName =
      event.display_name === 'Page View'
        ? event.event_name
        : event?.alias_name || event.display_name;
    return (
      <InfoCard
        title={event?.alias_name || event.display_name}
        event_name={event?.event_name}
        properties={event?.properties || {}}
        trigger={
          hoverEvents.includes(event.event_name) ||
          event.display_name === 'Page View'
            ? 'hover'
            : []
        }
      >
        <div className={`flex items-center font-medium`}>
          <span className='truncate mx-1'>{eventName}</span>
          {hoverEvents.includes(event.event_name) ||
          event.display_name === 'Page View' ? (
            <CaretRightOutlined />
          ) : null}
        </div>
      </InfoCard>
    );
  };

  const renderAdditionalDiv = (events_count, collapseState, onClick) => {
    return events_count > 1 ? (
      collapseState ? (
        <div className='timeline-events--num ml-1' onClick={onClick}>
          {'+' + Number(events_count - 1)}
        </div>
      ) : (
        <div className='timeline-events--num m-5' onClick={onClick}>
          <CaretUpOutlined /> Show Less
        </div>
      )
    ) : null;
  };

  return loading ? (
    <Spin size={'large'} className={'fa-page-loader'} />
  ) : timelineUsers.length == 0 ? (
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
            {timelineUsers.map((name) => {
              return (
                <th scope='col' className='truncate'>
                  {name}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {Object.entries(formattedData).map(
            ([timestamp, allEvents], rowIndex) => {
              return (
                <tr>
                  <td>
                    <div className='py-4'>{timestamp}</div>
                  </td>
                  {timelineUsers.map((username, columnIndex) => {
                    if (!allEvents[username]) return <td></td>;
                    let eventsList = allEvents[username].collapsed
                      ? allEvents[username].events.slice(0, 1)
                      : allEvents[username].events;
                    return (
                      <td>
                        <div
                          className={`timeline-events ${
                            allEvents[username].collapsed
                              ? 'flex items-center'
                              : ''
                          }`}
                        >
                          {eventsList?.map((event) => {
                            return (
                              <div className='timeline-events--event'>
                                {renderInfoCard(event)}
                              </div>
                            );
                          })}
                          {renderAdditionalDiv(
                            allEvents[username].events.length,
                            allEvents[username].collapsed,
                            () => {
                              setFormattedData(
                                toggleCellCollapse(
                                  formattedData,
                                  timestamp,
                                  username,
                                  !allEvents[username].collapsed
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
              );
            }
          )}
        </tbody>
      </table>
    </div>
  );
}
export default AccountTimeline;
