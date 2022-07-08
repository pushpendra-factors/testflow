import React, { useEffect, useMemo, useState } from 'react';
import { Spin, Button } from 'antd';
import styles from './index.module.scss';
import moment from 'moment';
import { SVG } from '../factorsComponents';
import { CaretUpOutlined } from '@ant-design/icons';

function FaTimeline({
  activities = [],
  granularity,
  collapse,
  setCollapse,
  loading,
}) {
  const [showAll, setShowAll] = useState([]);

  const groups = {
    Timestamp: (item) =>
      moment(item.timestamp * 1000).format('DD MMMM YYYY, hh:mm:ss '),
    Hourly: (item) =>
      moment(item.timestamp * 1000)
        .startOf('hour')
        .format('hh A') +
      ' - ' +
      moment(item.timestamp * 1000)
        .add(1, 'hour')
        .startOf('hour')
        .format('hh A') +
      ' ' +
      moment(item.timestamp * 1000)
        .startOf('hour')
        .format('DD MMM YYYY'),
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

  const data = _.groupBy(activities, groups[granularity]);

  useEffect(() => {
    if (collapse !== undefined) {
      const showAllState = new Array(Object.entries(data).length).fill(
        !collapse
      );
      setShowAll(showAllState);
    }
  }, [collapse]);

  const setShowAllIndex = (ind, flag) => {
    setCollapse(undefined);
    const showAllState = [...showAll];
    showAllState[ind] = flag;
    setShowAll(showAllState);
  };

  const renderTimeline = (data) => {
    if (!Object.entries(data).length)
      return (
        <div class='ant-empty ant-empty-normal'>
          <div class='ant-empty-image'>
            <SVG name='nodata' />
          </div>
          <div class='ant-empty-description'>No Activity</div>
        </div>
      );
    const timeline = [];
    Object.entries(data).forEach(([key, values], index) => {
      const arrayOpts = [];
      const groupEvents = (
        <div className={styles.timeline}>
          <div className={styles.timeline_timestamp}>
            <div className={styles.timeline_timestamp_text}>{key}</div>
          </div>
          <div className={styles.timeline_events}>
            {(() => {
              values.forEach((event, eventIndex) => {
                arrayOpts.push(
                  <>
                    <div className={styles.timeline_events_event}>
                      {event ? (
                        <div className={`flex`}>
                          <div className={styles.tag}>{event.display_name}</div>
                          {!showAll[index] && values.length > 1 ? (
                            <div
                              className={`${styles.num}`}
                              onClick={() => {
                                setShowAllIndex(index, true);
                              }}
                            >
                              {'+' + Number(values.length - 1)}
                            </div>
                          ) : null}
                        </div>
                      ) : null}
                      {index === Object.entries(data).length - 1 &&
                      !showAll[index] ? null : (
                        <div className={styles.timeline_events_event_tail} />
                      )}
                    </div>
                  </>
                );
              });
              arrayOpts.push(
                showAll[index] && arrayOpts.length > 1 ? (
                  <div className={styles.timeline_events_event}>
                    <div
                      className={`${styles.num}`}
                      onClick={() => {
                        setShowAllIndex(index, false);
                      }}
                    >
                      <CaretUpOutlined /> Show Less
                    </div>
                    {index === Object.entries(data).length - 1 ? null : (
                      <div className={styles.timeline_events_event_tail} />
                    )}
                  </div>
                ) : null
              );
              return showAll[index] ? arrayOpts : arrayOpts[0];
            })()}
          </div>
        </div>
      );
      timeline.push(groupEvents);
    });
    return timeline;
  };

  return (
    <>
      <div className={styles.header}>Date and Time</div>
      {loading ? (
        <Spin size={'large'} className={'fa-page-loader'} />
      ) : (
        renderTimeline(data)
      )}
    </>
  );
}
export default FaTimeline;
