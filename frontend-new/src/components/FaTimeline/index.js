import React from 'react';
import { Table, Button, Timeline } from 'antd';
import styles from './index.module.scss';
import moment from 'moment';

function FaTimeline({ activities }) {
  return (
    <>
      <div className={styles.header}>Date and Time</div>
      {activities?.map((activity, index) => {
        return (
          <div className={styles.timeline}>
            <div className={styles.timeline_timestamp}>
              {moment(activity?.timestamp * 1000).format(
                'DD MMMM YYYY, hh:mm:ss'
              )}
            </div>
            <div className={styles.timeline_event}>
              {activity?.event_name ? (
                <div className={styles.tag}> {activity.event_name} </div>
              ) : null}
              {index < activities?.length - 1 ? (
                <div className={styles.timeline_event_tail} />
              ) : null}
            </div>
          </div>
        );
      })}
    </>
  );
}
export default FaTimeline;
