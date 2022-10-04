import React, { useEffect, useState } from 'react';
import { Spin } from 'antd';
import styles from './index.module.scss';
import { SVG } from '../factorsComponents';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';
import InfoCard from './InfoCard';
import { groups, hoverEvents } from '../Profile/utils';

function FaTimeline({
  activities = [],
  granularity,
  collapse,
  setCollapse,
  loading,
}) {
  const [showAll, setShowAll] = useState([]);

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
        <div className='ant-empty ant-empty-normal'>
          <div className='ant-empty-image'>
            <SVG name='nodata' />
          </div>
          <div className='ant-empty-description'>No Activity</div>
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
                          <InfoCard
                            title={event?.alias_name || event.display_name}
                            event_name={event.event_name}
                            properties={event?.properties || {}}
                            trigger={
                              hoverEvents.includes(event.event_name) ||
                              event.display_name === 'Page View'
                                ? 'hover'
                                : []
                            }
                          >
                            <div className={`${styles.tag}`}>
                              <span className='truncate'>
                                {event.display_name === 'Page View'
                                  ? event.event_name
                                  : event.display_name}
                              </span>
                              {hoverEvents.includes(event.event_name) ||
                              event.display_name === 'Page View' ? (
                                <CaretRightOutlined />
                              ) : null}
                            </div>
                          </InfoCard>
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
