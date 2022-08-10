import React, { useEffect, useMemo } from 'react';
import MomentTz from '../../../MomentTz';

function AccountTimeline({
  timeline = [],
  granularity,
  collapse,
  setCollapse,
  loading,
}) {
  const groups = {
    Timestamp: (item) =>
      MomentTz(item.timestamp * 1000).format('DD MMMM YYYY, hh:mm:ss '),
    Hourly: (item) =>
      MomentTz(item.timestamp * 1000)
        .startOf('hour')
        .format('hh A') +
      ' - ' +
      MomentTz(item.timestamp * 1000)
        .add(1, 'hour')
        .startOf('hour')
        .format('hh A') +
      ' ' +
      MomentTz(item.timestamp * 1000)
        .startOf('hour')
        .format('DD MMM YYYY'),
    Daily: (item) =>
      MomentTz(item.timestamp * 1000)
        .startOf('day')
        .format('DD MMM YYYY'),
    Weekly: (item) =>
      MomentTz(item.timestamp * 1000)
        .endOf('week')
        .format('DD MMM YYYY') +
      ' - ' +
      MomentTz(item.timestamp * 1000)
        .startOf('week')
        .format('DD MMM YYYY'),
    Monthly: (item) =>
      MomentTz(item.timestamp * 1000)
        .startOf('month')
        .format('MMM YYYY'),
  };

  const formattedData = useMemo(() => {
    const groupByTimestamp = [];
    timeline.forEach((user) => {
      const newOpts = user.user_activities.map((data) => {
        return { ...data, user: user.user_name };
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

  return (
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
          {Object.entries(formattedData).map(([timestamp, allEvents]) => {
            return (
              <tr>
                <td>
                  <div className='py-4'>{timestamp}</div>
                </td>
                {timeline.map((data) => {
                  if (!allEvents[data.user_name]) return <td></td>;
                  return (
                    <td>
                      <div className='timeline-events'>
                        {allEvents[data.user_name].map((event) => {
                          return (
                            <div className='timeline-events--event'>
                              <div className='timeline-events--event--tag'>
                                {event.event_name}
                              </div>
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
