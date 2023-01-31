import { Avatar, Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import {
  ALPHANUMSTR,
  convertSVGtoURL,
  eventIconsColorMap,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
  iconColors,
  propValueFormat,
  singleTimelineIconSVGs,
  TimelineHoverPropDisplayNames
} from '../../utils';
import { SVG, Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import { PropTextFormat } from 'Utils/dataFormatter';

function AccountTimelineSingleView({
  timelineEvents = [],
  timelineUsers = [],
  loading = false,
  eventNamesMap
}) {
  const [formattedData, setFormattedData] = useState({});
  useEffect(() => {
    const data = eventsFormattedForGranularity(timelineEvents, 'Daily', true);
    setFormattedData(data);
  }, [timelineEvents]);

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
              <th scope='col'>Date and Time</th>
              <th scope='col'></th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(formattedData).map(([timestamp, allEvents]) => (
              <tr>
                <td>
                  <div className='top-40'>{timestamp}</div>
                </td>
                <td className='bg-none pt-6'>
                  {Object.entries(allEvents).map(([user, data]) => {
                    const currentUser = timelineUsers.find(
                      (obj) => obj.title === user
                    );
                    const isAnonymous = currentUser.isAnonymous;
                    return (
                      <div>
                        <div className='timeline-user-card'>
                          <div className='inline-flex gap--8 items-center'>
                            {isAnonymous ? (
                              <SVG
                                name={`TrackedUser${user.match(/\d/g)[0]}`}
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
                                        user.charAt(0).toUpperCase()
                                      ) % 8
                                    ]
                                  }`,
                                  fontSize: '16px'
                                }}
                              >
                                {user.charAt(0).toUpperCase()}
                              </Avatar>
                            )}
                            <h2 className='m-0'>{user}</h2>
                          </div>
                        </div>
                        <div class='user-timeline__events'>
                          {data?.events.map((event) => {
                            const category = getEventCategory(
                              event,
                              eventNamesMap
                            );
                            const sourceIcon = getIconForCategory(category);
                            const eventIcon = singleTimelineIconSVGs[event.icon]
                              ? event.icon
                              : 'calendar_star';
                            const svgUrl = convertSVGtoURL(
                              singleTimelineIconSVGs[eventIcon]
                            );
                            return (
                              <div
                                class='timeline-event__container'
                                style={{
                                  '--svg-url': `${svgUrl}`,
                                  '--svg-bg': `${eventIconsColorMap[eventIcon]?.bgColor}`,
                                  '--svg-border-color': `${eventIconsColorMap[eventIcon]?.borderColor}`
                                }}
                              >
                                <div class='timestamp'>
                                  {MomentTz(event?.timestamp * 1000).format(
                                    'hh:mm A'
                                  )}
                                </div>
                                <div class='card'>
                                  <div className='top-section'>
                                    {event.alias_name ? (
                                      <div className='heading-with-sub'>
                                        <div className='sub'>
                                          {PropTextFormat(event.display_name)}
                                        </div>
                                        <div className='main'>
                                          {event.alias_name}
                                        </div>
                                      </div>
                                    ) : (
                                      <div className='heading'>
                                        {PropTextFormat(event.display_name)}
                                      </div>
                                    )}

                                    <div className='icon'>
                                      <SVG
                                        name={sourceIcon}
                                        size={24}
                                        color={
                                          sourceIcon === 'events_cq'
                                            ? 'blue'
                                            : null
                                        }
                                      />
                                    </div>
                                  </div>

                                  {Object.entries(event?.properties || {}).map(
                                    ([key, value]) => {
                                      if (
                                        key === '$is_page_view' &&
                                        value === true
                                      )
                                        return (
                                          <div className='flex justify-between py-2'>
                                            <Text
                                              mini
                                              type='title'
                                              color='grey'
                                              extraClass='whitespace-no-wrap mr-2'
                                            >
                                              Page URL
                                            </Text>
                                            <Text
                                              mini
                                              type='title'
                                              color='grey-2'
                                              weight='medium'
                                              extraClass='break-all text-right'
                                              truncate
                                              charLimit={40}
                                            >
                                              {event.event_name}
                                            </Text>
                                          </div>
                                        );
                                      return (
                                        <div className='flex justify-between py-2'>
                                          <Text
                                            mini
                                            type='title'
                                            color='grey'
                                            extraClass={`${
                                              key.length > 20
                                                ? 'break-words'
                                                : 'whitespace-no-wrap'
                                            } max-w-xs mr-2`}
                                          >
                                            {TimelineHoverPropDisplayNames[
                                              key
                                            ] || PropTextFormat(key)}
                                          </Text>
                                          <Text
                                            mini
                                            type='title'
                                            color='grey-2'
                                            weight='medium'
                                            extraClass={`${
                                              value?.length > 30
                                                ? 'break-words'
                                                : 'whitespace-no-wrap'
                                            }  text-right`}
                                            truncate
                                            charLimit={40}
                                          >
                                            {propValueFormat(key, value) || '-'}
                                          </Text>
                                        </div>
                                      );
                                    }
                                  )}
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    );
                  })}
                </td>
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

export default AccountTimelineSingleView;
