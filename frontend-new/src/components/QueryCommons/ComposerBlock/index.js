import React from 'react';
import { Collapse, Tooltip } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import styles from './index.module.scss';

import { SVG, Text } from '../../factorsComponents';

import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';

const { Panel } = Collapse;

function ComposerBlock({
  blockTitle,
  disabled = false,
  isOpen,
  showIcon = true,
  onClick,
  children,
  extraClass
}) {
  let tooltipContent = '';
  if (blockTitle === 'FILTER BY' || blockTitle === 'BREAKDOWN') {
    tooltipContent = `This global ${
      blockTitle === 'FILTER BY' ? 'filter' : 'breakdown'
    } impacts all added events above`;
  } else if (blockTitle === 'CONVERSION GOAL') {
    tooltipContent =
      'The primary user action which a campaign is expected to drive';
  } else if (blockTitle === 'CRITERIA') {
    tooltipContent = 'Pick an attribution model for your analysis';
  } else if (blockTitle === 'FUNNEL CRITERIA') {
    tooltipContent = 'Specify an advanced criteria such as conversion window';
  } else if (blockTitle === 'LINKED EVENTS') {
    tooltipContent =
      'Select events you expect to occur after the conversion goal you defined above.';
  }
  const renderHeader = () =>
    blockTitle && (
      <div className={`${styles.cmpBlock__title}`}>
        <div className='mb-0'>
          <Text
            type='title'
            level={7}
            weight='bold'
            disabled={disabled}
            extraClass='m-0 mb-2 inline'
          >
            {blockTitle}
            {tooltipContent.length > 0 ? (
              <Tooltip
                className='p-1'
                title={tooltipContent}
                placement='right'
                color={TOOLTIP_CONSTANTS.DARK}
              >
                <InfoCircleOutlined />
              </Tooltip>
            ) : (
              ''
            )}
          </Text>
        </div>
        {showIcon && (
          <div className={`${styles.cmpBlock__title__icon}`}>
            <SVG
              name={isOpen ? 'minus' : 'plus'}
              color={disabled ? 'black' : 'gray'}
              onClick={() => onClick()}
            />
          </div>
        )}
      </div>
    );

  return (
    <div
      className={`${styles.cmpBlock} fa--query_block bordered ${extraClass}`}
    >
      <Collapse
        bordered={false}
        activeKey={isOpen ? [1] : [0]}
        expandIcon={() => {}}
        onChange={() => !disabled && onClick()}
      >
        <Panel header={renderHeader()} key={1}>
          {children}
        </Panel>
      </Collapse>
    </div>
  );
}

export default ComposerBlock;
