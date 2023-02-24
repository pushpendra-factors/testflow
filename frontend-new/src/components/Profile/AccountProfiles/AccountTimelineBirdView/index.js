import { Avatar, Spin } from 'antd';
import React, { useState, useEffect, useMemo } from 'react';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import InfoCard from '../../MyComponents/InfoCard';
import {
  ALPHANUMSTR,
  eventIconsColorMap,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  hoverEvents,
  iconColors,
  iconMap,
  timestampToString,
  toggleCellCollapse
} from '../../utils';
import { SVG, Text } from '../../../factorsComponents';
import { PropTextFormat } from 'Utils/dataFormatter';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import { useSelector } from 'react-redux';

function AccountTimelineBirdView({
  timelineEvents = [],
  timelineUsers = [],
  milestones,
  granularity,
  collapseAll,
  setCollapseAll,
  loading = false,
  eventNamesMap,
  listProperties
}) {
  const [formattedData, setFormattedData] = useState({});
  const { groupPropNames } = useSelector((state) => state.coreQuery);

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

  const formattedMilestones = useMemo(() => {
    return Object.entries(milestones || {}).map(([key, value]) => [
      key,
      timestampToString[granularity](value)
    ]);
  }, [milestones, granularity]);

  console.log(formattedMilestones)

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
      <img
        src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
          iconMap[event.icon] ? iconMap[event.icon] : event.icon
        }.svg`}
        alt=''
        height={16}
        width={16}
        loading='lazy'
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
            <img
              src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
                iconMap[icon] ? iconMap[icon] : icon
              }.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          }
          listProperties={listProperties}
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

  const renderMilestoneStrip = (milestonesList, showText = false) =>
    milestonesList.length ? (
      <div className='milestone-section'>
        {milestonesList.map((milestone) => (
          <div className={`green-stripe ${showText ? '' : 'opaque'}`}>
            {showText ? (
              <div className='text'>
                {groupPropNames[milestone[0]]
                  ? groupPropNames[milestone[0]]
                  : milestone[0]}
              </div>
            ) : null}
          </div>
        ))}
      </div>
    ) : null;

  const renderTimeline = () => (
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
          {Object.entries(formattedData).map(
            ([timestamp, allEvents], index) => {
              const milestones = formattedMilestones.filter(
                (milestone) => milestone[1] === timestamp
              );
              return (
                <tr>
                  <td>
                    <div className='timestamp top-64'>{timestamp}</div>
                    {renderMilestoneStrip(milestones, true)}
                  </td>
                  {timelineUsers.map((user) => {
                    if (!allEvents[user.title])
                      return (
                        <td className='bg-gradient--44px'>
                          {renderMilestoneStrip(milestones, false)}
                        </td>
                      );
                    const eventsList = allEvents[user.title].collapsed
                      ? allEvents[user.title].events.slice(0, 1)
                      : allEvents[user.title].events;
                    return (
                      <td
                        className={`bg-gradient--44px pb-${
                          milestones.length * 10
                        }`}
                      >
                        <div
                          className={`timeline-events account-pad ${
                            allEvents[user.title].collapsed
                              ? 'timeline-events--collapsed'
                              : 'timeline-events--expanded'
                          }`}
                        >
                          {eventsList?.map((event) => (
                            <div className='timeline-events__event'>
                              {renderIcon(event)}
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
                        {renderMilestoneStrip(milestones, false)}
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

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
    <NoDataWithMessage message={'No Associated Users'} />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message={'No Events Enabled to Show'} />
  ) : (
    renderTimeline()
  );
}
export default AccountTimelineBirdView;
