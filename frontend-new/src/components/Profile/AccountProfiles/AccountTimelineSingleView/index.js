import { Spin } from 'antd';
import React, { useState, useEffect } from 'react';
import {
  eventsFormattedForGranularity,
  propValueFormat,
  TimelineHoverPropDisplayNames
} from '../../utils';
import { SVG, Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import { PropTextFormat } from 'Utils/dataFormatter';

function AccountTimelineSingleView({
  timelineEvents = [],
  timelineUsers = [],
  loading = false
}) {
  const [formattedData, setFormattedData] = useState({});

  useEffect(() => {
    const data = eventsFormattedForGranularity(timelineEvents, 'Daily', true);
    setFormattedData(data);
  }, [timelineEvents]);

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : timelineUsers.length === 0 ? (
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
                <div className='py-4'>{timestamp}</div>
              </td>
              <td style={{ paddingTop: '24px', background: 'none' }}>
                {Object.entries(allEvents).map(([user, data]) => (
                  <div>
                    <div className='timeline-user-card'>
                      <h2>{user}</h2>
                    </div>
                    <div class='user-timeline__events'>
                      {data?.events.map((event) => (
                        <div class='timeline-event__container'>
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
                                  if (key === '$is_page_view' && value === true)
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
                                        {TimelineHoverPropDisplayNames[key] ||
                                          PropTextFormat(key)}
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
                      ))}
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
}

export default AccountTimelineSingleView;
