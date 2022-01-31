import React from 'react';
import cx from 'classnames';
import { Text } from '../factorsComponents';
import { Tooltip } from 'antd';

const NonClickableTableHeader = ({ title, alignment = 'left', verticalAlignment = 'center', titleTooltip = null }) => {
  const justifyAlignment = alignment === 'left' ? 'justify-start' : 'justify-end';
  const verticalAlignmentClass = verticalAlignment === 'end' ? 'items-end' : 'items-center';
  let titleText;
  if (titleTooltip) {
    titleText = (
      <Tooltip title={titleTooltip}>
        <Text weight='bold' color='grey-2' type='title' extraClass='mb-0'>
          {title}
        </Text>
      </Tooltip>
    )
  } else {
    titleText = (
      <Text weight='bold' color='grey-2' type='title' extraClass='mb-0'>
        {title}
      </Text>
    )
  }
  return <div className={cx(`flex ${justifyAlignment} ${verticalAlignmentClass} h-full px-4`, { 'pb-1': verticalAlignment === 'end' })}>{titleText}</div>;
};

export default NonClickableTableHeader;
