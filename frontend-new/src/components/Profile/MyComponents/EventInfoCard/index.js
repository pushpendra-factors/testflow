import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import {
  eventIconsColorMap,
  getPropType,
  propValueFormat,
  TimelineHoverPropDisplayNames
} from 'Components/Profile/utils';
import React, { useMemo } from 'react';
import { connect } from 'react-redux';
import {
  convertGroupedPropertiesToUngrouped,
  PropTextFormat
} from 'Utils/dataFormatter';

const EventInfoCard = ({ event, eventIcon, sourceIcon, eventPropertiesV2 }) => {
  const eventPropertiesModified = useMemo(() => {
    const eventProps = [];
    if (eventPropertiesV2?.[event?.event_name]) {
      convertGroupedPropertiesToUngrouped(
        eventPropertiesV2?.[event?.event_name],
        eventProps
      );
    }
    return eventProps;
  }, [event, eventPropertiesV2]);
  return (
    <div className='timeline-event__container'>
      <div className='timestamp'>
        {MomentTz(event?.timestamp * 1000).format('hh:mm A')}
      </div>
      <div
        className='event-icon'
        style={{
          '--border-color': `${eventIconsColorMap[eventIcon]?.borderColor}`,
          '--bg-color': `${eventIconsColorMap[eventIcon]?.bgColor}`
        }}
      >
        <img
          src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${eventIcon}.svg`}
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
              src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${sourceIcon}.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          </div>
        </div>

        {Object.entries(event?.properties || {}).map(([key, value]) => {
          const propType = getPropType(eventPropertiesModified, key);
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
                  {event.event_type === 'FE'
                    ? event.alias_name
                    : event.event_name}
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
};

const mapStateToProps = (state) => ({
  eventPropertiesV2: state.coreQuery.eventPropertiesV2
});

export default connect(mapStateToProps, null)(EventInfoCard);
