import React from 'react';
import cx from 'classnames';
import { filterCardProps } from './types';
import styles from './index.module.scss';

function CardLayout({ children, footerActions }: filterCardProps) {
  return (
    <div
      className={cx(styles['filters-box-container'], 'flex flex-col gap-y-5')}
    >
      <div className='pt-4 gap-x-5 flex flex-col gap-y-2'>{children}</div>
      <div className={cx('py-4 px-6', styles['buttons-container'])}>
        {footerActions}
      </div>
    </div>
  );
}

export default CardLayout;
