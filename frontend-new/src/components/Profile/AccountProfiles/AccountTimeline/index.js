import { Spin } from 'antd';
import React, { useMemo, useState, useEffect } from 'react';
import InfoCard from '../../../FaTimeline/InfoCard';
import {
  eventsFormattedForGranularity,
  groups,
  hoverEvents,
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
  const formattedData = useMemo(() => {
    return eventsFormattedForGranularity(
      timelineEvents,
      granularity,
      collapseAll
    );
  }, [timelineEvents, granularity, collapseAll]);

  const renderInfoCard = (event) => {
    return (
      <InfoCard
        title={event?.display_name}
        event_name={event?.event_name}
        properties={event?.properties || {}}
        trigger={hoverEvents.includes(event.display_name) ? 'hover' : []}
      >
        <div className={`flex items-center font-medium`}>
          <span className='truncate mx-1'>
            {event?.display_name === 'Page View'
              ? event?.event_name
              : PropTextFormat(event?.display_name)}
          </span>
          {hoverEvents.includes(event?.display_name) ? (
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
        <div className='timeline-events--num m-5'>
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
                    let eventsList = collapseAll
                      ? allEvents[username].slice(0, 1)
                      : allEvents[username];
                    return (
                      <td>
                        <div
                          className={`timeline-events ${
                            collapseAll ? 'flex items-center' : ''
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
                            allEvents[username].length,
                            collapseAll,
                            () => console.log('Clicked')
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
