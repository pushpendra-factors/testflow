import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';

const SidebarMenuItem = ({ text, isActive, onClick }) => {
  return (
    <Tooltip placement='right' mouseEnterDelay={2} title={text}>
      <div
        role='button'
        onClick={onClick}
        className={cx(
          'cursor-pointer rounded-md p-2 flex justify-between col-gap-2 items-center',
          {
            [styles['active']]: isActive
          },
          styles['sidebar-menu-item']
        )}
      >
        <Text
          type='title'
          level={7}
          extraClass='mb-0 text-with-ellipsis'
          weight='medium'
        >
          {text}
        </Text>
        {isActive && <SVG size={16} color='#595959' name='arrowright' />}
      </div>
    </Tooltip>
  );
};

export default SidebarMenuItem;
