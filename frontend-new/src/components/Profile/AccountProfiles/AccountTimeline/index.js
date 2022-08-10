import { Spin } from 'antd';
import React, { useEffect, useMemo } from 'react';
import InfoCard from '../../../FaTimeline/InfoCard';
import { groups, hoverEvents } from '../../utils';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';

function AccountTimeline({
  timeline = [],
  granularity,
  collapse,
  setCollapse,
  loading = false,
}) {
  const formattedData = useMemo(() => {
    const groupByTimestamp = [];
    timeline.forEach((user) => {
      const newOpts = user.user_activities.map((data) => {
        return { ...data, user: user.user_id };
      });
      groupByTimestamp.push(...newOpts);
    });
    const data = _.groupBy(groupByTimestamp, groups[granularity]);
    let retData = {};
    Object.entries(data).forEach(([key, value]) => {
      const ret = _.groupBy(value, (item) => item.user);
      const obj = new Object();
      obj[key] = ret;
      retData = { ...retData, ...obj };
    });
    return retData;
  }, [timeline, granularity]);

  return loading ? (
    <Spin size={'large'} className={'fa-page-loader'} />
  ) : (
    <div className='table-scroll'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date and Time</th>
            {timeline.map((data) => {
              return <th scope='col'>{data.user_name || data.user_id}</th>;
            })}
          </tr>
        </thead>
        <tbody>
          {Object.entries(formattedData).map(([timestamp, allEvents]) => {
            return (
              <tr>
                <td>
                  <div className='py-4'>{timestamp}</div>
                </td>
                {timeline.map((data) => {
                  if (!allEvents[data.user_id]) return <td></td>;
                  return (
                    <td>
                      <div className='timeline-events'>
                        {allEvents[data.user_id].map((event) => {
                          return (
                            <div className='timeline-events--event'>
                              <InfoCard
                                title={event.display_name}
                                event_name={event.event_name}
                                properties={event?.properties || {}}
                                trigger={
                                  hoverEvents.includes(event.display_name)
                                    ? 'hover'
                                    : []
                                }
                              >
                                <div className='timeline-events--event--tag truncate'>
                                  {event.display_name === 'Page View'
                                    ? event.event_name
                                    : event.display_name}
                                  {hoverEvents.includes(event.display_name) ? (
                                    <CaretRightOutlined />
                                  ) : null}
                                </div>
                              </InfoCard>

                              <div className='timeline-events--event--tail' />
                            </div>
                          );
                        })}
                      </div>
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
}
export default AccountTimeline;
