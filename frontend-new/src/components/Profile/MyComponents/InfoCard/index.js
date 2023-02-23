import React from 'react';
import { Popover } from 'antd';
import { Text } from 'Components/factorsComponents';
import { PropTextFormat } from 'Utils/dataFormatter';
import {
  getPropType,
  propValueFormat,
  TimelineHoverPropDisplayNames
} from '../../utils';

function InfoCard({
  title,
  eventSource,
  icon,
  eventName,
  properties = {},
  trigger,
  children,
  listProperties
}) {
  const popoverContent = () => (
    <div className='fa-popupcard'>
      <div className='top-section'>
        {title ? (
          <div className='heading-with-sub'>
            <div className='sub'>{PropTextFormat(eventSource)}</div>
            <div className='main'>{title}</div>
          </div>
        ) : (
          <div className='heading'>{PropTextFormat(eventSource)}</div>
        )}
        <div className='source-icon'>{icon}</div>
      </div>

      {Object.entries(properties).map(([key, value]) => {
        const propType = getPropType(listProperties, key)
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
              {propValueFormat(key, value, propType) || '-'}
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
