import React from 'react';
import { Text } from '../factorsComponents';
import { Popover } from 'antd';
import styles from './index.module.scss';

function InfoCard({
  title, info, footer
}) {
  const popoverContent = () => {
    return (
      <div className={styles.infoCard}>
        <div className={styles.tabInfo}><Text extraClass="m-0" type={'title'} level={7} weight={'bold'}>{title}</Text>
          <Text extraClass={'pt-1'} mini type={'paragraph'}>{info}</Text></div>
          {footer}
      </div>
    );
  };
  return (
    <Popover
      content={popoverContent}
      overlayClassName={'fa-popupcard--wrapper--info'}
      placement='bottomLeft'
    >
      {title}
    </Popover>
  );
}
export default InfoCard;
