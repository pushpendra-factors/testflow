import React from 'react';
import { Popover } from 'antd';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import truncateURL from 'Utils/truncateURL';
import { propValueFormat } from '../../utils';

function InfoCard({
  title,
  eventType,
  eventSource,
  icon,
  eventName,
  properties = {},
  propertiesType,
  trigger,
  children
}) {
  const { eventPropNames } = useSelector((state) => state.coreQuery);
  const { currentProjectSettings, projectDomainsList } = useSelector(
    (state) => state.global
  );

  const renderPropRow = (key, value) => {
    if (key === '$is_page_view' && value === true) return null;

    const propType = propertiesType[key];
    const propertyValue = propValueFormat(key, value, propType) || '-';
    const urlTruncatedValue = truncateURL(propertyValue, projectDomainsList);

    return (
      <div className='event-infocard--row'>
        <TextWithOverflowTooltip
          text={eventPropNames[key] || PropTextFormat(key)}
          extraClass='prop'
        />
        <div className='value'>{urlTruncatedValue}</div>
      </div>
    );
  };

  const popoverContent = (
    <>
      <div className='event-name-section'>
        {title ? (
          <div className='heading-with-sub'>
            <div className='sub-heading truncate'>
              {PropTextFormat(eventSource)}
            </div>
            <TextWithOverflowTooltip
              text={eventType === 'FE' ? eventName : title}
              extraClass='main truncate'
            />
          </div>
        ) : (
          <TextWithOverflowTooltip
            text={PropTextFormat(eventSource)}
            extraClass='heading truncate'
          />
        )}
        <div className='source-icon'>{icon}</div>
      </div>
      <div className='properties-section'>
        {(
          currentProjectSettings?.timelines_config?.events_config?.[
            eventSource === 'Page View' ? 'PageView' : eventName
          ] || []
        ).map((key) => renderPropRow(key, properties[key]))}
      </div>
    </>
  );

  return (
    <Popover
      key={title}
      content={popoverContent}
      overlayClassName='infocard-popover'
      placement='rightBottom'
      trigger={trigger}
    >
      {children}
    </Popover>
  );
}

export default InfoCard;
