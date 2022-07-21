import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Popover } from 'antd';
import {
  formatDurationIntoString,
  PropTextFormat,
} from '../../../utils/dataFormatter';
import MomentTz from '../../MomentTz';

function InfoCard({ title, event_name, properties = {}, trigger, children }) {
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
          if (key === '$is_page_view' && value === true)
            return (
              <div className='flex justify-between py-2'>
                <Text
                  mini
                  type={'paragraph'}
                  color={'grey'}
                  extraClass={'max-w-2/3 break-words mr-2'}
                >
                  Page URL
                </Text>

                <Text
                  mini
                  type={'paragraph'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={'break-words text-right'}
                >
                  {event_name}
                </Text>
              </div>
            );
          else
            return (
              <div className='flex justify-between py-2'>
                <Text
                  mini
                  type={'paragraph'}
                  color={'grey'}
                  extraClass={'max-w-2/3 truncate mr-2'}
                >
                  {key === '$timestamp' ? 'Date and Time' : PropTextFormat(key)}
                </Text>
                <Text
                  mini
                  type={'paragraph'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={'break-words text-right'}
                >
                  {key === '$timestamp'
                    ? MomentTz(value * 1000).format('DD MMMM YYYY, hh:mm A')
                    : key.includes('_time')
                    ? formatDurationIntoString(value)
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
