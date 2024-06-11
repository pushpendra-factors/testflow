import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import styles from './index.module.scss';

const SidebarMenuItem = ({
  text,
  isActive,
  onClick,
  icon,
  iconColor,
  iconSize = 16,
  hoverable = true
}) => (
  <Tooltip placement='right' mouseEnterDelay={2} title={text}>
    <div
      role={hoverable && 'button'}
      onClick={onClick}
      className={cx(
        'rounded-md p-2 flex justify-between gap-x-2 items-center',
        {
          [styles.active]: isActive
        },
        {
          [styles['cursor-pointer']]: hoverable
        },
        {
          [styles['sidebar-menu-item']]: hoverable
        },
        {
          [styles['font-medium']]: !hoverable
        }
      )}
    >
      <div className={cx('flex gap-x-1 items-center w-full')}>
        <ControlledComponent controller={icon != null}>
          <SVG name={icon} size={iconSize} color={iconColor} />
        </ControlledComponent>
        <Text
          type='title'
          level={7}
          extraClass='mb-0 text-with-ellipsis w-40'
          weight='medium'
        >
          {text}
        </Text>
      </div>
      {isActive && <SVG size={iconSize} color='#595959' name='arrowright' />}
    </div>
  </Tooltip>
);

export default SidebarMenuItem;
