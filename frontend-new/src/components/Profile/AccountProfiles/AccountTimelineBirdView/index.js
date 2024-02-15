import { Avatar, Spin, Tooltip } from 'antd';
import React, { useState, useEffect, useMemo } from 'react';
import { CaretRightOutlined, CaretUpOutlined } from '@ant-design/icons';
import { PropTextFormat } from 'Utils/dataFormatter';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import { useSelector } from 'react-redux';
import truncateURL from 'Utils/truncateURL';
import {
  ALPHANUMSTR,
  eventIconsColorMap,
  iconColors
} from 'Components/Profile/constants';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { SVG, Text } from '../../../factorsComponents';
import {
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  toggleCellCollapse
} from '../../utils';
import InfoCard from '../../MyComponents/InfoCard';
import logger from 'Utils/logger';

function AccountTimelineBirdView({
  timelineEvents = [],
  timelineUsers = [],
  granularity,
  collapseAll,
  setCollapseAll,
  loading = false,
  propertiesType,
  eventNamesMap
}) {
  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);
  const [formattedData, setFormattedData] = useState({});

  useEffect(() => {
    if (!timelineEvents) return;

    const data = eventsFormattedForGranularity(
      timelineEvents,
      granularity,
      collapseAll
    );
    document.title = 'Accounts - FactorsAI';
    setFormattedData(data);
  }, [timelineEvents, granularity]);

  // temp
  useEffect(() => {
    logger.log(formattedData);
  }, [formattedData]);

  useEffect(() => {
    const data = Object.keys(formattedData).reduce((acc, key) => {
      acc[key] = Object.keys(formattedData[key]).reduce((userAcc, username) => {
        userAcc[username] = {
          ...formattedData[key][username],
          collapsed:
            collapseAll === undefined
              ? formattedData[key][username].collapsed
              : collapseAll
        };
        return userAcc;
      }, {});
      return acc;
    }, {});
    setFormattedData(data);
  }, [collapseAll]);

  const renderIcon = (event) => {
    const eventIcon = eventIconsColorMap[event.icon]
      ? event.icon
      : 'calendar-star';
    const { borderColor, bgColor } = eventIconsColorMap[eventIcon] || {};
    const isTrackedUser = event.user === 'new_user';

    const iconContent = isTrackedUser ? (
      <SVG name={`TrackedUser${event.id.match(/\d/g)?.[0] || 0}`} size={20} />
    ) : (
      <img
        src={`/assets/icons/${eventIcon}.svg`}
        alt=''
        height={16}
        width={16}
        loading='lazy'
      />
    );

    return (
      <div
        className='icon'
        style={{ '--border-color': borderColor, '--bg-color': bgColor }}
      >
        {iconContent}
      </div>
    );
  };

  const renderInfoCard = (event) => {
    const eventName =
      event.alias_name ||
      (event.display_name !== 'Page View' &&
        PropTextFormat(event.display_name)) ||
      event.event_name;
    const isHoverable = Object.keys(event.properties || {}).length > 0;
    const category = getEventCategory(event, eventNamesMap);
    const icon = getIconForCategory(category);

    return (
      <div className='tag'>
        <InfoCard
          eventType={event?.event_type}
          title={event?.alias_name}
          eventSource={event?.display_name}
          eventName={event?.event_name}
          properties={event?.properties || {}}
          propertiesType={propertiesType}
          trigger={isHoverable ? 'hover' : []}
          icon={
            <img
              src={`/assets/icons/${icon}.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          }
        >
          <div className='inline-flex gap--4 items-center'>
            <div className='event-name--sm'>
              <TextWithOverflowTooltip
                text={truncateURL(eventName, projectDomainsList)}
                tooltipText={eventName}
                disabled={isHoverable}
              />
            </div>
            {isHoverable && (
              <CaretRightOutlined
                style={{ fontSize: '12px', color: '#8692A3' }}
              />
            )}
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

  const renderMilestoneStrip = (milestones, showText = false) =>
    milestones.events.length ? (
      <div className='milestone-section'>
        {milestones.events.map((milestone) => (
          <div className={`green-stripe ${showText ? '' : 'opaque'}`}>
            {showText ? (
              <div className='text'>
                {groupPropNames[milestone.event_name]
                  ? groupPropNames[milestone.event_name]
                  : milestone.event_name}
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
                    <SVG name='TrackedUser1' size={32} />
                  ) : (
                    <Avatar
                      size={32}
                      className='userlist-avatar'
                      style={{
                        backgroundColor: `${
                          user.title === 'group_user'
                            ? '#BAE7FF'
                            : iconColors[
                                ALPHANUMSTR.indexOf(
                                  user.title.charAt(0).toUpperCase()
                                ) % 8
                              ]
                        }`,
                        fontSize: '16px'
                      }}
                    >
                      {user.title === 'group_user' ? (
                        <SVG name='focus' size={20} />
                      ) : (
                        user.title.charAt(0).toUpperCase()
                      )}
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
                      {user.title === 'group_user'
                        ? 'Account Activity'
                        : user.title}
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
          {Object.entries(formattedData).map(([timestamp, allEvents]) => {
            const milestones = allEvents?.milestone;
            return (
              <tr>
                <td className={`pb-${(milestones?.events?.length || 0) * 8}`}>
                  <div className='timestamp top-64'>{timestamp}</div>
                  {milestones && renderMilestoneStrip(milestones, true)}
                </td>
                {timelineUsers.map((user) => {
                  if (!allEvents[user.userId])
                    return (
                      <td className='bg-gradient--44px'>
                        {milestones && renderMilestoneStrip(milestones, false)}
                      </td>
                    );
                  const eventsList = allEvents[user.userId].collapsed
                    ? allEvents[user.userId].events.slice(0, 1)
                    : allEvents[user.userId].events;
                  return (
                    <td
                      className={`bg-gradient--44px pb-${
                        (milestones?.events?.length || 0) * 10
                      }`}
                    >
                      <div
                        className={`timeline-events account-pad ${
                          allEvents[user.userId].collapsed
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
                      {milestones && renderMilestoneStrip(milestones, false)}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
    <NoDataWithMessage message='No Associated Users' />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    renderTimeline()
  );
}

export default AccountTimelineBirdView;
