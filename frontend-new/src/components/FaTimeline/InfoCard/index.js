import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Popover } from 'antd';
import { PropTextFormat } from '../../../utils/dataFormatter';
import humanizeDuration from 'humanize-duration';
import MomentTz from '../../MomentTz';

function InfoCard({ title, properties = {}, trigger, children }) {
  const popoverContent = () => {
    return (
      <div className='fa-popupcard'>
        <Text
          extraClass='m-0 mb-3'
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
        >
          {title}
        </Text>
        {Object.entries(properties).map(([key, value]) => {
          return (
            <div className='flex justify-between py-2'>
              <Text mini type={'paragraph'} color={'grey'}>
                {key === '$timestamp' ? 'Date and Time' : PropTextFormat(key)}
              </Text>
              <Text mini type={'paragraph'} color={'grey-2'} weight={'medium'}>
                {key === '$timestamp'
                  ? MomentTz(value * 1000).format('DD MMMM YYYY, hh:mm A')
                  : key.includes('_time')
                  ? humanizeDuration(value * 1000, { largest: 2 })
                  : value}
              </Text>
            </div>
          );
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
