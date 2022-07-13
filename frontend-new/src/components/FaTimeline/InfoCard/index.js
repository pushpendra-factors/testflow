import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Popover } from 'antd';
import { property } from 'lodash';

function InfoCard({ title, properties = {}, trigger, children }) {
  const popoverContent = () => {
    return (
      <div className='fa-popupcard'>
        <Text
          extraClass='m-0'
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
        >
          {title}
        </Text>
        {Object.entries(properties).forEach(([key, value]) => {
          <div className='flex justify-between py-2'>
            <Text mini type={'paragraph'} color={'grey'}>
              {key}
            </Text>
            <Text mini type={'paragraph'} color={'grey-2'}>
              {value}
            </Text>
          </div>;
        })}
      </div>
    );
  };
  return (
    <Popover
      content={popoverContent}
      overlayClassName='fa-infocard--wrapper'
      placement='rightBottom'
      trigger={trigger}
    >
      {children}
    </Popover>
  );
}
export default InfoCard;
