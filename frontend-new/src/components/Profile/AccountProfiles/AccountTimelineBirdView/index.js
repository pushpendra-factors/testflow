import { Avatar, Spin } from 'antd';
import React from 'react';
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
  getEventCategory,
  getIconForCategory,
  toggleCellCollapse
} from '../../utils';
import InfoCard from '../../MyComponents/InfoCard';

function AccountTimelineBirdView({
  events,
  setEvents,
  timelineUsers = [],
  setCollapseAll,
  loading = false,
  propertiesType,
  eventNamesMap
}) {
  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);

  const renderIcon = (event) => {
    const eventIcon = eventIconsColorMap[event.icon]
      ? event.icon
      : 'calendar-star';
    const { borderColor, bgColor } = eventIconsColorMap[eventIcon] || {};
    const isNewUser = event.username === 'new_user';

    const iconContent = isNewUser ? (
      <SVG
        name={`TrackedUser${event.user_id.match(/\d/g)?.[0] || 0}`}
        size={20}
      />
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
      event.name;
    const isHoverable = Object.keys(event.properties || {}).length > 0;
    const category = getEventCategory(event, eventNamesMap);
    const icon = getIconForCategory(category);

    return (
      <div className='tag'>
        <InfoCard
          eventType={event?.type}
          title={event?.alias_name}
          eventSource={event?.display_name}
          eventName={event?.name}
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
        <div className='birdview-events__num' onClick={onClick}>
          {`+${Number(eventsCount - 1)}`}
        </div>
      ) : (
        <div className='birdview-events__num' onClick={onClick}>
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
                {groupPropNames[milestone.name]
                  ? groupPropNames[milestone.name]
                  : milestone.name}
              </div>
            ) : null}
          </div>
        ))}
      </div>
    ) : null;

  const renderTimeline = () => (
    <div className='birdview-container bordered-gray--bottom'>
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
                          user.name === 'group_user'
                            ? '#BAE7FF'
                            : iconColors[
                                ALPHANUMSTR.indexOf(
                                  user.name.charAt(0).toUpperCase()
                                ) % 8
                              ]
                        }`,
                        fontSize: '16px'
                      }}
                    >
                      {user.name === 'group_user' ? (
                        <SVG name='focus' size={20} />
                      ) : (
                        user.name.charAt(0).toUpperCase()
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
                      {user.name === 'group_user'
                        ? 'Account Activity'
                        : user.name}
                    </Text>
                    <Text type='title' truncate level={8} extraClass='m-0'>
                      {user.extraProp || '-'}
                    </Text>
                  </div>
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Object.entries(events || {}).map(([timestamp, allEvents]) => {
            const milestones = allEvents?.milestone;
            return (
              <tr>
                <td
                  style={{
                    paddingBottom: `${(milestones?.events?.length || 0) * 32}px`
                  }}
                >
                  <div className='timestamp top-64'>{timestamp}</div>
                  {milestones && renderMilestoneStrip(milestones, true)}
                </td>
                {timelineUsers.map((user) => {
                  if (!allEvents[user.id])
                    return (
                      <td className='bg-gradient--44px'>
                        {milestones && renderMilestoneStrip(milestones, false)}
                      </td>
                    );
                  const eventsList = allEvents[user.id].collapsed
                    ? allEvents[user.id].events.slice(0, 1)
                    : allEvents[user.id].events;
                  return (
                    <td
                      className={`bg-gradient--44px pb-${
                        (milestones?.events?.length || 0) * 10
                      }`}
                    >
                      <div
                        className={`birdview-events account-pad ${
                          allEvents[user.id].collapsed
                            ? 'birdview-events--collapsed'
                            : 'birdview-events--expanded'
                        }`}
                      >
                        {eventsList?.map((event) => (
                          <div className='birdview-events__event'>
                            {renderIcon(event)}
                            {renderInfoCard(event)}
                          </div>
                        ))}
                        {renderAdditionalDiv(
                          allEvents[user.id].events.length,
                          allEvents[user.id].collapsed,
                          () => {
                            setEvents(
                              toggleCellCollapse(
                                events,
                                timestamp,
                                user.id,
                                !allEvents[user.id].collapsed
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
  ) : Object.keys(events).length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    renderTimeline()
  );
}

export default AccountTimelineBirdView;
