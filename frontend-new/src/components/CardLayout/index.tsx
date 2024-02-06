import React from 'react';
import cx from 'classnames';
import { filterCardProps } from './types';
import styles from './index.module.scss';

function CardLayout({ children, footerActions }: filterCardProps) {
  return (
    <div
      className={cx(styles['filters-box-container'], 'flex flex-col row-gap-5')}
    >
      <div className='pt-4 col-gap-5 flex flex-col row-gap-2'>{children}</div>
      <div className={cx('py-4 px-6', styles['buttons-container'])}>
        {footerActions}
      </div>
    </div>
  );
}

export default CardLayout;
