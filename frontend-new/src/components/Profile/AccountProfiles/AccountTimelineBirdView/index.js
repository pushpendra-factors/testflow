import { Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import InfoCard from '../../../FaTimeline/InfoCard';
import {
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  getIconForEvent,
  hoverEvents,
  toggleCellCollapse
} from '../../utils';
import { SVG, Text } from '../../../factorsComponents';
import { PropTextFormat } from 'Utils/dataFormatter';

function AccountTimelineBirdView({
  timelineEvents = [],
  timelineUsers = [],
  granularity,
  collapseAll,
  setCollapseAll,
  loading = false,
  eventNamesMap
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
        : event?.alias_name || PropTextFormat(event.display_name);
    const hoverConditionals =
      hoverEvents.includes(event.event_name) ||
      event.display_name === 'Page View' ||
      event.event_type === 'CH' ||
      event.event_type === 'CS';
    const category = getEventCategory(event, eventNamesMap);
    const icon = getIconForCategory(category);
    return (
      <InfoCard
        title={event?.alias_name || event.display_name}
        eventName={event?.event_name}
        properties={event?.properties || {}}
        trigger={hoverConditionals ? 'hover' : []}
        icon={
          <SVG
            name={icon}
            size={24}
            color={icon === 'events_cq' ? 'blue' : null}
          />
        }
      >
        <div className='inline-flex-gap--6 items-center'>
          <div>
            <SVG
              name={icon}
              size={16}
              color={icon === 'events_cq' ? 'blue' : null}
            />
          </div>
          <div className='event-name--sm'>{eventName}</div>
          {hoverConditionals ? <CaretRightOutlined /> : null}
        </div>
      </InfoCard>
    );
  };

  const renderAdditionalDiv = (eventsCount, collapseState, onClick) =>
    eventsCount > 1 ? (
      collapseState ? (
        <div className='timeline-events__num' onClick={onClick}>
          {`+${Number(eventsCount - 1)}`}
        </div>
      ) : (
        <div className='timeline-events__num' onClick={onClick}>
          <CaretUpOutlined /> Show Less
        </div>
      )
    ) : null;

  const renderTimeline = () =>
    timelineUsers.length === 0 ? (
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
              <th
                scope='col'
                className={`${timelineUsers.length > 1 ? '' : 'single-user'}`}
              >
                Date and Time
              </th>
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
                  <div className='top-75'>{timestamp}</div>
                </td>
                {timelineUsers.map((user) => {
                  if (!allEvents[user.title])
                    return <td className='bg-gradient--44px' />;
                  const eventsList = allEvents[user.title].collapsed
                    ? allEvents[user.title].events.slice(0, 1)
                    : allEvents[user.title].events;
                  return (
                    <td className='bg-gradient--44px'>
                      <div
                        className={`timeline-events multi-user--padding  ${
                          allEvents[user.title].collapsed
                            ? 'timeline-events--collapsed'
                            : 'timeline-events--expanded'
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

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : (
    renderTimeline()
  );
}
export default AccountTimelineBirdView;
