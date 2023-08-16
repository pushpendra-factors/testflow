import React, { useEffect, useMemo } from 'react';
import { Popover } from 'antd';
import { Text } from 'Components/factorsComponents';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import {
  getPropType,
  propValueFormat,
  TimelineHoverPropDisplayNames
} from '../../utils';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { getEventPropertiesV2 } from 'Reducers/coreQuery/middleware';

function InfoCard({
  title,
  eventType,
  eventSource,
  icon,
  eventName,
  properties = {},
  trigger,
  children,
  activeProject,
  eventPropertiesV2
}) {
  useEffect(() => {
    if (!eventPropertiesV2[eventName])
      getEventPropertiesV2(activeProject?.id, eventName);
  }, [activeProject?.id, eventName]);

  const eventPropertiesModified = useMemo(() => {
    const eventProps = [];
    if (eventPropertiesV2?.[eventName]) {
      convertGroupedPropertiesToUngrouped(
        eventPropertiesV2?.[eventName],
        eventProps
      );
    }
    return eventProps;
  }, [eventName, eventPropertiesV2]);

  const popoverContent = () => (
    <div className='fa-popupcard'>
      <div className='top-section mb-2'>
        {title ? (
          <div className='heading-with-sub'>
            <div className='sub'>{PropTextFormat(eventSource)}</div>
            <div className='main'>{eventType === 'FE' ? eventName : title}</div>
          </div>
        ) : (
          <div className='heading'>{PropTextFormat(eventSource)}</div>
        )}
        <div className='source-icon'>{icon}</div>
      </div>

      {Object.entries(properties).map(([key, value]) => {
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
                {eventType === 'FE' ? title : eventName}
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
      key={title}
      content={popoverContent}
      overlayClassName='fa-infocard--wrapper'
      placement='rightBottom'
      trigger={trigger}
    >
      {children}
    </Popover>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventPropertiesV2
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(InfoCard);
