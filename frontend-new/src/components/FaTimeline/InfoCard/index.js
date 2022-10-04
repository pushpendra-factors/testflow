import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Popover } from 'antd';
import {
  formatDurationIntoString,
  PropTextFormat,
} from '../../../utils/dataFormatter';
import MomentTz from '../../MomentTz';
import { TimelineHoverPropDisplayNames } from '../../Profile/utils';

function InfoCard({ title, event_name, properties = {}, trigger, children }) {
  const popoverPropValueFormat = (key, value) => {
    if (
      key.includes('timestamp') ||
      key.includes('starttime') ||
      key.includes('endtime')
    ) {
      return MomentTz(value * 1000).format('DD MMMM YYYY, hh:mm A');
    } else if (key.includes('_time')) {
      return formatDurationIntoString(value);
    } else if (key.includes('durationmilliseconds')) {
      return formatDurationIntoString(parseInt(value / 1000));
    } else return value;
  };
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
                  type={'title'}
                  color={'grey'}
                  extraClass={'whitespace-no-wrap mr-2'}
                >
                  Page URL
                </Text>

                <Text
                  mini
                  type={'title'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={`break-all text-right`}
                  truncate={true}
                  charLimit={40}
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
                  type={'title'}
                  color={'grey'}
                  extraClass={`${
                    key.length > 20 ? 'break-words' : 'whitespace-no-wrap'
                  } max-w-xs mr-2`}
                >
                  {TimelineHoverPropDisplayNames[key] || PropTextFormat(key)}
                </Text>
                <Text
                  mini
                  type={'title'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={`${
                    value?.length > 30 ? 'break-words' : 'whitespace-no-wrap'
                  }  text-right`}
                  truncate={true}
                  charLimit={40}
                >
                  {popoverPropValueFormat(key, value) || '-'}
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
