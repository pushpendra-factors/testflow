import React from 'react';
import { useSelector } from 'react-redux';
import { Button, Drawer } from 'antd';
import { Text } from 'Components/factorsComponents';
import { eventIconsColorMap } from 'Components/Profile/constants';
import { PropTextFormat } from 'Utils/dataFormatter';
import { propValueFormat } from 'Components/Profile/utils';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { EventDrawerProps } from 'Components/Profile/types';
import EventIcon from './EventIcon';

function EventDrawer({
  visible,
  onClose,
  event,
  eventPropsType
}: EventDrawerProps): JSX.Element {
  const { eventPropNames } = useSelector((state: any) => state.coreQuery);

  const renderEventDetails = () => {
    if (!event) return null;

    const eventIcon = eventIconsColorMap[event.icon]
      ? event.icon
      : 'calendar-star';

    const renderAliasName = () => (
      <TextWithOverflowTooltip
        text={event.event_type === 'FE' ? event.event_name : event.alias_name}
        extraClass='main'
      />
    );

    return (
      <div className='py-4'>
        <div className='top-section mb-4'>
          <div className='flex items-center w-full'>
            <EventIcon icon={eventIcon} size={28} />
            {event.alias_name ? (
              <div className='heading-with-sub ml-2'>
                <div className='sub'>{PropTextFormat(event.display_name)}</div>
                {renderAliasName()}
              </div>
            ) : (
              <TextWithOverflowTooltip
                text={PropTextFormat(event.display_name)}
                extraClass='heading ml-2'
              />
            )}
          </div>
        </div>
        <div>
          {Object.entries(event.properties || {}).map(([key, value]) => {
            const propType = eventPropsType[key];
            if (key === '$is_page_view' && value === true) return null;
            return (
              <div className='leftpane-prop' key={key}>
                <div className='flex flex-col items-start truncate'>
                  <Text
                    type='title'
                    level={8}
                    color='grey'
                    truncate
                    charLimit={44}
                    extraClass='m-0'
                  >
                    {eventPropNames[key] || PropTextFormat(key)}
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    truncate
                    charLimit={44}
                    extraClass='m-0'
                  >
                    {propValueFormat(key, value, propType) || '-'}
                  </Text>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <Drawer
      title={
        <div className='flex justify-between items-center'>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Event Details
          </Text>
          <Button onClick={onClose}>Close</Button>
        </div>
      }
      placement='right'
      closable={false}
      mask
      maskClosable
      visible={visible}
      className='fa-drawer--right'
      onClose={onClose}
    >
      {renderEventDetails()}
    </Drawer>
  );
}

export default EventDrawer;
