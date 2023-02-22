import { SVG, Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import {
  eventIconsColorMap,
  getPropType,
  iconMap,
  propValueFormat,
  TimelineHoverPropDisplayNames
} from 'Components/Profile/utils';
import React from 'react';
import { PropTextFormat } from 'Utils/dataFormatter';

const EventInfoCard = ({
  event,
  eventIcon,
  sourceIcon,
  listProperties
}) => (
  <div className='timeline-event__container'>
    <div className='timestamp'>
      {MomentTz(event?.timestamp * 1000).format('hh:mm A')}
    </div>
    <div
      className='event-icon'
      style={{
        '--border-color': `${eventIconsColorMap[eventIcon].borderColor}`,
        '--bg-color': `${eventIconsColorMap[eventIcon].bgColor}`
      }}
    >
      <img
        src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
          iconMap[eventIcon] ? iconMap[eventIcon] : eventIcon
        }.svg`}
        alt=''
        height={20}
        width={20}
        loading='lazy'
      />
    </div>
    <div className='card'>
      <div className='top-section'>
        {event.alias_name ? (
          <div className='heading-with-sub'>
            <div className='sub'>{PropTextFormat(event.display_name)}</div>
            <div className='main'>{event.alias_name}</div>
          </div>
        ) : (
          <div className='heading'>{PropTextFormat(event.display_name)}</div>
        )}
        <div className='source-icon'>
          <img
            src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
              iconMap[sourceIcon] ? iconMap[sourceIcon] : sourceIcon
            }.svg`}
            alt=''
            height={24}
            width={24}
            loading='lazy'
          />
        </div>
      </div>

      {Object.entries(event?.properties || {}).map(([key, value]) => {
        const propType = getPropType(listProperties, key);
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
                {event.event_name}
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
  </div>
);

export default EventInfoCard;
