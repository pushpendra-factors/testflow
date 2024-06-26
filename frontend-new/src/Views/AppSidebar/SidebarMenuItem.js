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
  hoverable = true,
  extraClass = ''
}) => {
  const containerClasses = cx(
    'rounded-md p-2 flex justify-between gap-x-2 items-center',
    extraClass,
    {
      [styles.active]: isActive,
      [styles['cursor-pointer']]: hoverable,
      [styles['sidebar-menu-item']]: hoverable,
      [styles['font-medium']]: !hoverable
    }
  );

  const contentClasses = 'flex gap-x-1 items-center w-full';

  return (
    <Tooltip placement='right' mouseEnterDelay={2} title={text}>
      <div
        role={hoverable ? 'button' : undefined}
        onClick={onClick}
        className={containerClasses}
      >
        <div className={contentClasses}>
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
      </div>
    </Tooltip>
  );
};

export default SidebarMenuItem;
