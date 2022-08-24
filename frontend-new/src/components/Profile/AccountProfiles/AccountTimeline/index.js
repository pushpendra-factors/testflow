import { Spin } from 'antd';
import React, { useMemo } from 'react';
import InfoCard from '../../../FaTimeline/InfoCard';
import { groups, hoverEvents, getLoopLength } from '../../utils';
import { CaretRightOutlined } from '@ant-design/icons';
import { SVG } from '../../../factorsComponents';

function AccountTimeline({
  timeline = [],
  granularity,
  collapse,
  setCollapse,
  loading = false,
}) {
  const compareObjTimestampsDesc = (a, b) => {
    if (a.timestamp > b.timestamp) {
      return -1;
    }
    if (a.timestamp < b.timestamp) {
      return 1;
    }
    return 0;
  };

  const formattedData = useMemo(() => {
    const groupByTimestamp = [];
    timeline.forEach((user) => {
      const newOpts = user.user_activities.map((data) => {
        return { ...data, user: user.user_name };
      });
      groupByTimestamp.push(...newOpts);
    });
    groupByTimestamp.sort(compareObjTimestampsDesc);
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
  ) : timeline.length == 0 ? (
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
            {timeline.map((data) => {
              return <th scope='col'>{data.user_name}</th>;
            })}
          </tr>
        </thead>
        <tbody>
          {Object.entries(formattedData).map(
            ([timestamp, allEvents], index) => {
              return (
                <tr>
                  <td>
                    <div className='py-4'>{timestamp}</div>
                  </td>
                  {timeline.map((data) => {
                    const loopLength = getLoopLength(allEvents);
                    const evList = [];
                    if (!allEvents[data.user_name]) {
                      for (let i = 0; i < loopLength; i++) {
                        evList.push(
                          <div className='timeline-events--event'>
                            <div
                              className='timeline-events--event--tag'
                              style={{ visibility: 'hidden' }}
                            />
                            {index ==
                              Object.entries(formattedData).length - 1 &&
                            i === loopLength - 1 ? null : (
                              <div className='timeline-events--event--tail' />
                            )}
                          </div>
                        );
                      }
                    } else {
                      allEvents[data.user_name].forEach((event, evIndex) => {
                        evList.push(
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
                            {index ==
                              Object.entries(formattedData).length - 1 &&
                            evIndex === loopLength - 1 ? null : (
                              <div className='timeline-events--event--tail' />
                            )}
                          </div>
                        );
                      });
                      while (
                        evList.length < loopLength &&
                        index !== Object.entries(formattedData).length - 1
                      ) {
                        evList.push(
                          <div className='timeline-events--event'>
                            <div
                              className='timeline-events--event--tag'
                              style={{ visibility: 'hidden' }}
                            />
                            <div className='timeline-events--event--tail' />
                          </div>
                        );
                      }
                    }
                    return (
                      <td>
                        <div className='timeline-events'>{evList}</div>
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
}
export default AccountTimeline;
