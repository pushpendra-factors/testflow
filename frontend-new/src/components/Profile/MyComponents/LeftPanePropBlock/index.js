import { Button } from 'antd';
import React from 'react';
import { useSelector } from 'react-redux';
import truncateURL from 'Utils/truncateURL';
import { Text, SVG } from '../../../factorsComponents';
import { propValueFormat } from '../../utils';

function LeftPanePropBlock({ property, type, displayName, value, onDelete }) {
  const { projectDomainsList } = useSelector((state) => state.global);
  const formattedValue = propValueFormat(property, value, type) || '-';
  const urlTruncatedValue = truncateURL(formattedValue, projectDomainsList);

  return (
    <div className='leftpane-prop justify-between pl-8'>
      <div className='flex flex-col items-start truncate'>
        <Text
          type='title'
          level={8}
          color='grey'
          truncate
          charLimit={30}
          extraClass='m-0'
        >
          {`${displayName}`}
        </Text>
        <Text
          type='title'
          level={7}
          truncate
          charLimit={25}
          extraClass='m-0'
          toolTipTitle={formattedValue}
        >
          {urlTruncatedValue}
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
