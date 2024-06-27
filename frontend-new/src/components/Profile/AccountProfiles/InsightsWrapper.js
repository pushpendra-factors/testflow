import React from 'react';
import cx from 'classnames';
import styles from './index.module.scss';

const InsightsWrapper = ({ children }) => (
  <div
    className={cx(
      'w-full h-full overflow-scroll flex flex-col',
      styles['insights-wrapper']
    )}
  >
    {children}
  </div>
);

export default InsightsWrapper;
