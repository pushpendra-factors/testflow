import { Button } from 'antd';
import React from 'react';
import { PropTextFormat } from '../../../utils/dataFormatter';
import { Text, SVG } from '../../factorsComponents';

function LeftPanePropBlock({ property, value = 0, onDelete }) {
  return (
    <div className='leftpane-prop'>
      <div className='flex flex-col items-start truncate'>
        <Text type='title' level={8} color='grey-2'>
          {`${PropTextFormat(property)}:`}
        </Text>
        <Text type='title' level={7}>
          {value}
        </Text>
      </div>

      <Button
        type='text'
        className='del-button'
        onClick={() => onDelete(property)}
        icon={<SVG name='delete' />}
      />
    </div>
  );
}
export default LeftPanePropBlock;
