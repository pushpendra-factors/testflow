import { Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import {
  convertSVGtoURL,
  eventsFormattedForGranularity,
  getEventCategory,
  getIconForCategory,
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
                  {Object.entries(allEvents).map(([user, data]) => (
                    <div>
                      <div className='timeline-user-card'>
                        <h2>{user}</h2>
                      </div>
                      <div class='user-timeline__events'>
                        {data?.events.map((event) => {
                          const category = getEventCategory(
                            event,
                            eventNamesMap
                          );
                          const icon = getIconForCategory(category);
                          const svgUrl = convertSVGtoURL(
                            singleTimelineIconSVGs[icon]
                          );
                          return (
                            <div
                              class='timeline-event__container'
                              style={{ '--svg-url': `${svgUrl}` }}
                            >
                              <div class='timestamp'>
                                {MomentTz(event?.timestamp * 1000).format(
                                  'hh:mm A'
                                )}
                              </div>
                              <div class='content'>
                                <div className='fa-popupcard'>
                                  <Text
                                    extraClass='m-0 mb-3'
                                    type='title'
                                    level={6}
                                    weight='bold'
                                    color='grey-2'
                                  >
                                    {event?.alias_name || event?.display_name}
                                  </Text>
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
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
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
