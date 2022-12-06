import { Button } from 'antd';
import React from 'react';
import { PropTextFormat } from '../../../utils/dataFormatter';
import { Text, SVG } from '../../factorsComponents';
import { propValueFormat } from '../utils';

function LeftPanePropBlock({ property, value = 0, onDelete }) {
  return (
    <div className='leftpane-prop'>
      <div className='flex flex-col items-start truncate'>
        <Text
          type='title'
          level={8}
          color='grey-2'
          truncate
          charLimit={30}
          extraClass='m-0'
        >
          {`${PropTextFormat(property)}:`}
        </Text>
        <Text type='title' level={7} truncate charLimit={25} extraClass='m-0'>
          {propValueFormat(property, value)}
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
