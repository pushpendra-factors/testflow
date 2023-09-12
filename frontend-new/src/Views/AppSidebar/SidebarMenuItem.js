import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';

const SidebarMenuItem = ({
  text,
  isActive,
  onClick,
  icon,
  iconColor,
  iconSize = 16
}) => {
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
        <div className={cx('flex col-gap-1 items-center w-full')}>
          <ControlledComponent controller={icon != null}>
            <SVG name={icon} size={20} color={iconColor} />
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
};

export default SidebarMenuItem;
