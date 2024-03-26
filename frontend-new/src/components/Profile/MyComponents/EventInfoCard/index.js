import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import { eventIconsColorMap } from 'Components/Profile/constants';
import { propValueFormat } from 'Components/Profile/utils';
import React from 'react';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';
import truncateURL from 'Utils/truncateURL';

function EventInfoCard({ event, eventIcon, sourceIcon, propertiesType }) {
  const { eventPropNames } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);
  return (
    <div className='timeline-event__container'>
      <div className='timestamp'>
        {MomentTz((event?.timestamp || 0) * 1000).format('hh:mm A')}
      </div>
      <div
        className='event-icon'
        style={{
          '--border-color': `${eventIconsColorMap[eventIcon]?.borderColor}`,
          '--bg-color': `${eventIconsColorMap[eventIcon]?.bgColor}`,
          '--icon-size': '32px'
        }}
      >
        <img
          src={`/assets/icons/${eventIcon}.svg`}
          alt=''
          height={20}
          width={20}
          loading='lazy'
        />
      </div>
      <div className='card'>
        <div className='top-section mb-2'>
          {event.alias_name ? (
            <div className='heading-with-sub'>
              <div className='sub'>{PropTextFormat(event.display_name)}</div>
              <div className='main'>
                {event.event_type === 'FE'
                  ? event.event_name
                  : event.alias_name}
              </div>
            </div>
          ) : (
            <div className='heading'>{PropTextFormat(event.display_name)}</div>
          )}
          <div className='source-icon'>
            <img
              src={`/assets/icons/${sourceIcon}.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          </div>
        </div>

        {Object.entries(event?.properties || {}).map(([key, value]) => {
          const propType = propertiesType[key];
          if (key === '$is_page_view' && value === true) return null;
          const formattedValue = propValueFormat(key, value, propType) || '-';
          const urlTruncatedValue = truncateURL(
            formattedValue,
            projectDomainsList
          );
          return (
            <div className='flex justify-between py-2'>
              <Text
                mini
                type='title'
                color='grey'
                extraClass={`${
                  key.length > 20 ? 'break-words' : 'whitespace-nowrap'
                } max-w-xs mr-2`}
              >
                {eventPropNames[key] || PropTextFormat(key)}
              </Text>
              <Text
                mini
                type='title'
                color='grey-2'
                weight='medium'
                extraClass={`${
                  value?.length > 30 ? 'break-words' : 'whitespace-nowrap'
                }  text-right`}
                truncate
                charLimit={40}
                toolTipTitle={formattedValue}
              >
                {urlTruncatedValue}
              </Text>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default EventInfoCard;
