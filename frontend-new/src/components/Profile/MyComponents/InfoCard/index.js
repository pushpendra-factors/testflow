import React from 'react';
import { Popover } from 'antd';
import { Text } from 'Components/factorsComponents';
import { PropTextFormat } from 'Utils/dataFormatter';
import { propValueFormat } from '../../utils';
import { useSelector } from 'react-redux';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';

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
  const renderPropRow = (key, value) => {
    const propType = propertiesType[key];
    if (key === '$is_page_view' && value === true) return null;

    return (
      <div className='flex justify-between py-2' key={key}>
        <Text
          mini
          type='title'
          color='grey'
          extraClass='whitespace-no-wrap mr-2'
        >
          {eventPropNames[key] || PropTextFormat(key)}
        </Text>
        <Text
          mini
          type='title'
          color='grey-2'
          weight='medium'
          extraClass='break-all text-right'
          truncate
          charLimit={32}
          shouldTruncateURL
        >
          {propValueFormat(key, value, propType) || '-'}
        </Text>
      </div>
    );
  };

  const popoverContent = (
    <div className='fa-popupcard'>
      <div className='top-section mb-2'>
        {title ? (
          <div className='heading-with-sub'>
            <div className='sub'>{PropTextFormat(eventSource)}</div>
            <TextWithOverflowTooltip
              text={eventType === 'FE' ? eventName : title}
              extraClass='main'
            />
          </div>
        ) : (
          <TextWithOverflowTooltip
            text={PropTextFormat(eventSource)}
            extraClass='heading'
          />
        )}
        <div className='source-icon'>{icon}</div>
      </div>

      {Object.entries(properties).map(([key, value]) =>
        renderPropRow(key, value)
      )}
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

export default InfoCard;
