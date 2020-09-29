import React from 'react';
import styles from './index.module.scss';

function ChartHeader({ total, query, bgColor }) {
  return (
        <div className="flex flex-col items-center justify-center">
            <div className="flex items-center mb-4">
                <div style={{ backgroundColor: bgColor }} className={`mr-1 ${styles.eventCircle}`}></div>
                <div className={styles.eventText}>{query}</div>
            </div>
            <div className={styles.totalText}>{total}</div>
        </div>
  );
}

export default ChartHeader;
