import { Avatar, Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import InfoCard from '../../../FaTimeline/InfoCard';
import {
  ALPHANUMSTR,
  eventIconsColorMap,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  hoverEvents,
  iconColors,
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

  const renderIcon = (event) => (
    <div
      className='icon'
      style={{
        '--border-color': `${
          eventIconsColorMap[event.icon || 'calendar_star'].borderColor
        }`,
        '--bg-color': `${
          eventIconsColorMap[event.icon || 'calendar_star'].bgColor
        }`
      }}
    >
      <SVG
        name={event.icon || 'calendar_star'}
        size={16}
        color={eventIconsColorMap[event.icon || 'calendar_star'].iconColor}
      />
    </div>
  );

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
      <div className='tag'>
        <InfoCard
          title={event?.alias_name}
          eventSource={event?.display_name}
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
          <div className='inline-flex gap--4 items-center'>
            <div className='event-name--sm'>{eventName}</div>
            {hoverConditionals ? (
              <CaretRightOutlined
                style={{ fontSize: '12px', color: '#8692A3' }}
              />
            ) : null}
          </div>
        </InfoCard>
      </div>
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
    ) : timelineEvents.length === 0 ? (
      <div className='ant-empty ant-empty-normal'>
        <div className='ant-empty-image'>
          <SVG name='nodata' />
        </div>
        <div className='ant-empty-description'>Enable Events to Show</div>
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
                  <div className='inline-flex gap--8 items-center'>
                    {user?.isAnonymous ? (
                      <SVG
                        name={`TrackedUser${user.title.match(/\d/g)[0]}`}
                        size={32}
                      />
                    ) : (
                      <Avatar
                        size={32}
                        className='userlist-avatar'
                        style={{
                          backgroundColor: `${
                            iconColors[
                              ALPHANUMSTR.indexOf(
                                user.title.charAt(0).toUpperCase()
                              ) % 8
                            ]
                          }`,
                          fontSize: '16px'
                        }}
                      >
                        {user.title.charAt(0).toUpperCase()}
                      </Avatar>
                    )}
                    <div className='flex items-start flex-col'>
                      <Text
                        type='title'
                        truncate
                        level={7}
                        weight='medium'
                        extraClass='m-0'
                      >
                        {user.title}
                      </Text>
                      <Text type='title' truncate level={8} extraClass='m-0'>
                        {user.subtitle || '-'}
                      </Text>
                    </div>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Object.entries(formattedData).map(([timestamp, allEvents]) => (
              <tr>
                <td>
                  <div className='top-64'>{timestamp}</div>
                </td>
                {timelineUsers.map((user) => {
                  if (!allEvents[user.userId])
                    return <td className='bg-gradient--44px' />;
                  const eventsList = allEvents[user.userId].collapsed
                    ? allEvents[user.userId].events.slice(0, 1)
                    : allEvents[user.userId].events;
                  return (
                    <td className='bg-gradient--44px'>
                      <div
                        className={`timeline-events multi-user--padding  ${
                          allEvents[user.userId].collapsed
                            ? 'timeline-events--collapsed'
                            : 'timeline-events--expanded'
                        }`}
                      >
                        {eventsList?.map((event) => (
                          <div className='timeline-events__event'>
                            {renderIcon(event)}
                            {renderInfoCard(event)}{' '}
                          </div>
                        ))}
                        {renderAdditionalDiv(
                          allEvents[user.userId].events.length,
                          allEvents[user.userId].collapsed,
                          () => {
                            setFormattedData(
                              toggleCellCollapse(
                                formattedData,
                                timestamp,
                                user.userId,
                                !allEvents[user.userId].collapsed
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
