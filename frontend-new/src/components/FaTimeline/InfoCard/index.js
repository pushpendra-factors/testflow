import React from 'react';
import { Popover } from 'antd';
import { Text } from 'Components/factorsComponents';
import { PropTextFormat } from 'Utils/dataFormatter';
import {
  propValueFormat,
  TimelineHoverPropDisplayNames
} from '../../Profile/utils';

function InfoCard({
  title,
  icon,
  eventName,
  properties = {},
  trigger,
  children
}) {
  const popoverContent = () => (
    <div className='fa-popupcard'>
      <div className='inline-flex-gap--6 items-start mb-3'>
        {icon}
        <Text
          extraClass='m-0'
          type='title'
          level={6}
          weight='bold'
          color='grey-2'
        >
          {PropTextFormat(title)}
        </Text>
      </div>

      {Object.entries(properties).map(([key, value]) => {
        if (key === '$is_page_view' && value === true)
          return (
            <div className='flex justify-between py-2'>
              <Text
                mini
                type='title'
                color='grey'
                extraClass='whitespace-no-wrap mr-2'
              >
                Page URL
              </Text>

              <Text
                mini
                type='title'
                color='grey-2'
                weight='medium'
                extraClass='break-all text-right'
                truncate
                charLimit={40}
              >
                {eventName}
              </Text>
            </div>
          );
        return (
          <div className='flex justify-between py-2'>
            <Text
              mini
              type='title'
              color='grey'
              extraClass={`${
                key.length > 20 ? 'break-words' : 'whitespace-no-wrap'
              } max-w-xs mr-2`}
            >
              {TimelineHoverPropDisplayNames[key] || PropTextFormat(key)}
            </Text>
            <Text
              mini
              type='title'
              color='grey-2'
              weight='medium'
              extraClass={`${
                value?.length > 30 ? 'break-words' : 'whitespace-no-wrap'
              }  text-right`}
              truncate
              charLimit={40}
            >
              {propValueFormat(key, value) || '-'}
            </Text>
          </div>
        );
      })}
    </div>
  );
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
