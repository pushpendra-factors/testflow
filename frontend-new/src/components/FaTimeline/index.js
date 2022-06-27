import React, { useMemo, useState } from 'react';
import { Table, Button, Timeline } from 'antd';
import styles from './index.module.scss';
import moment from 'moment';

function FaTimeline({ activities, granularity, collapse }) {

  const groups = {
    Default: (item) => moment(item.timestamp*1000).format('DD MMMM YYYY, hh:mm:ss '),
    Hourly: (item) =>
      moment(item.timestamp * 1000)
        .startOf('hour')
        .format('DD MMM YYYY, hh A') +
      ' - ' +
      moment(item.timestamp * 1000)
        .add(1, 'hour')
        .startOf('hour')
        .format('hh A'),
    Daily: (item) =>
      moment(item.timestamp * 1000)
        .startOf('day')
        .format('DD MMM YYYY'),
    Weekly: (item) =>
      moment(item.timestamp * 1000)
        .endOf('week')
        .format('DD MMM YYYY') +
      ' - ' +
      moment(item.timestamp * 1000)
        .startOf('week')
        .format('DD MMM YYYY'),
    Monthly: (item) =>
      moment(item.timestamp * 1000)
        .startOf('month')
        .format('MMM YYYY'),
  };

  const data = useMemo(() => {
    return _.groupBy(activities, groups[granularity]);
  }, [activities, granularity]);

  const renderTimeline = (data) => {
    const timeline = [];
    Object.entries(data).forEach(([key, values], index) => {
      const array = collapse ? values.slice(0, 1) : values;
      timeline.push(
        <div className={styles.timeline}>
          <div className={styles.timeline_timestamp}>
            <div className={styles.timeline_timestamp_text}>{key}</div>
          </div>
          <div className={styles.timeline_events}>
            {array.map((event, eventIndex) => {
              return (
                <div className={styles.timeline_events_event}>
                  {event ? (
                    <div className='flex'>
                      <div className={styles.tag}> {event.event_name} </div>
                      {collapse && values.length > 1 ? (
                        <div className={`${styles.num}`}>
                          {'+' + Number(values.length - 1)}
                        </div>
                      ) : null}
                    </div>
                  ) : null}
                  {index === Object.entries(data).length - 1 &&
                  eventIndex === array.length - 1 ? null : (
                    <div className={styles.timeline_events_event_tail} />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      );
    });
    return timeline;
  };

  return (
    <>
      <div className={styles.header}>Date and Time</div>
      {renderTimeline(data)}
    </>
  );
}
export default FaTimeline;
