import { Avatar, Spin } from 'antd';
import React, { useState, useEffect, useMemo } from 'react';
import {
  ALPHANUMSTR,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  iconColors,
  timestampToString
} from '../../utils';
import { SVG } from 'Components/factorsComponents';
import EventInfoCard from 'Components/Profile/MyComponents/EventInfoCard';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

function AccountTimelineSingleView({
  timelineEvents = [],
  timelineUsers = [],
  milestones,
  loading = false,
  eventNamesMap,
  listProperties
}) {
  const [formattedData, setFormattedData] = useState({});
  useEffect(() => {
    const data = eventsFormattedForGranularity(timelineEvents, 'Daily', true);
    setFormattedData(data);
  }, [timelineEvents]);

  const formattedMilestones = useMemo(() => {
    return Object.entries(milestones || {}).map(([key, value]) => [
      key,
      timestampToString['Daily'](value)
    ]);
  }, [milestones]);

  const UsernameWithIcon = ({ username, isAnonymous }) => (
    <div className='user-card'>
      <div className='inline-flex gap--8 items-center'>
        {isAnonymous ? (
          <SVG name={`TrackedUser${username.match(/\d/g)[0]}`} size={32} />
        ) : (
          <Avatar
            size={32}
            className='userlist-avatar'
            style={{
              backgroundColor: `${
                iconColors[
                  ALPHANUMSTR.indexOf(username.charAt(0).toUpperCase()) % 8
                ]
              }`,
              fontSize: '16px'
            }}
          >
            {username.charAt(0).toUpperCase()}
          </Avatar>
        )}
        <h2 className='m-0'>{username}</h2>
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
            const milestones = formattedMilestones.filter(
              (milestone) => milestone[1] === timestamp
            );
            return (
              <tr>
                <td>
                  <div className='timestamp top-40'>{timestamp}</div>
                  {/* {milestones.length ? (
                    <div className='milestone-section'>
                      {milestones.map((milestone) => (
                        <div className='green-stripe'>
                          <div className='text'>{milestone[0]}</div>
                        </div>
                      ))}
                    </div>
                  ) : null} */}
                </td>
                <td className={`bg-none pt-6 pb-${milestones.length * 10}`}>
                  {Object.entries(allEvents).map(([user, data]) => {
                    const currentUser = timelineUsers.find(
                      (obj) => obj.title === user
                    );
                    return (
                      <div className='relative'>
                        <div className='user-card--wrapper'>
                          <UsernameWithIcon
                            username={user}
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
                            const eventIcon = event.icon
                              ? event.icon
                              : 'calendar_star';
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
                  {/* {milestones.length ? (
                    <div className='milestone-section'>
                      {milestones.map((milestone) => (
                        <div className={`green-stripe opaque`} />
                      ))}
                    </div>
                  ) : null} */}
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
