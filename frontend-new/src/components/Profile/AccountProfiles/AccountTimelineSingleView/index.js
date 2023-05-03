import { Avatar, Spin } from 'antd';
import React, { useState, useEffect, useMemo } from 'react';
import {
  ALPHANUMSTR,
  eventIconsColorMap,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  iconColors
} from '../../utils';
import { SVG } from 'Components/factorsComponents';
import EventInfoCard from 'Components/Profile/MyComponents/EventInfoCard';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

function AccountTimelineSingleView({
  timelineEvents = [],
  timelineUsers = [],
  loading = false,
  eventNamesMap,
  listProperties
}) {
  const [formattedData, setFormattedData] = useState({});
  useEffect(() => {
    const data = eventsFormattedForGranularity(timelineEvents, 'Daily', true);
    setFormattedData(data);
  }, [timelineEvents]);

  const UsernameWithIcon = ({ title, subtitle, isAnonymous }) => (
    <div className='user-card'>
      <div className='inline-flex gap--8 items-center w-full'>
        {isAnonymous ? (
          <SVG name={`TrackedUser1`} size={32} />
        ) : (
          <Avatar
            size={32}
            className='userlist-avatar'
            style={{
              backgroundColor: `${
                iconColors[
                  ALPHANUMSTR.indexOf(title.charAt(0).toUpperCase()) % 8
                ]
              }`,
              fontSize: '16px'
            }}
          >
            {title.charAt(0).toUpperCase()}
          </Avatar>
        )}
        <div className='top-section'>
          {subtitle ? (
            <div className='heading-with-sub'>
              <div className='main'>{title}</div>
              <div className='sub'>{subtitle}</div>
            </div>
          ) : (
            <div className='heading'>{title}</div>
          )}
        </div>
      </div>
    </div>
  );

  const SingleTimelineViewTable = ({ data = {} }) => (
    <div className='table-scroll'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date</th>
            <th scope='col' />
          </tr>
        </thead>
        <tbody>
          {Object.entries(data).map(([timestamp, allEvents]) => {
            return (
              <tr>
                <td>
                  <div className='timestamp top-40'>{timestamp}</div>
                </td>
                <td className={`bg-none pt-6`}>
                  {Object.entries(allEvents).map(([user, data]) => {
                    const currentUser = timelineUsers.find(
                      (obj) => obj.userId === user
                    );
                    if (!currentUser) return null;
                    return (
                      <div className='relative'>
                        <div className='user-card--wrapper'>
                          <UsernameWithIcon
                            title={currentUser.title}
                            subtitle={currentUser.subtitle}
                            isAnonymous={currentUser.isAnonymous}
                          />
                        </div>
                        <div class='user-timeline--events'>
                          {data?.events.map((event, index) => {
                            const category = getEventCategory(
                              event,
                              eventNamesMap
                            );
                            const sourceIcon = getIconForCategory(category);
                            const eventIcon = eventIconsColorMap[event.icon]
                              ? event.icon
                              : 'calendar-star';
                            return (
                              <EventInfoCard
                                event={event}
                                sourceIcon={sourceIcon}
                                eventIcon={eventIcon}
                                listProperties={listProperties}
                              />
                            );
                          })}
                        </div>
                      </div>
                    );
                  })}
                </td>
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
    <NoDataWithMessage message={'No Associated Users'} />
  ) : timelineEvents.length === 0 ? (
    <NoDataWithMessage message={'No Events Enabled to Show'} />
  ) : (
    <SingleTimelineViewTable data={formattedData} />
  );
}

export default AccountTimelineSingleView;
