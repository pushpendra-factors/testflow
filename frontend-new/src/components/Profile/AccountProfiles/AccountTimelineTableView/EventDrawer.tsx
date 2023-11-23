import { Button, Drawer } from 'antd';
import { Text } from 'Components/factorsComponents';
import {
  eventIconsColorMap,
  TimelineHoverPropDisplayNames
} from 'Components/Profile/utils';
import React, { useEffect, useState } from 'react';
import { PropTextFormat } from 'Utils/dataFormatter';
import EventIcon from './EventIcon';
import { EventDrawerProps } from './types';

const EventDrawer: React.FC<EventDrawerProps> = ({
  visible,
  onClose,
  selectedEvent
}) => {

  const renderEventDetails = () => {
    if (!selectedEvent) return null;

    const eventIcon = eventIconsColorMap[selectedEvent.icon]
      ? selectedEvent.icon
      : 'calendar-star';

    return (
      <div className='p-4'>
        <div className='top-section mb-4'>
          <div className='flex items-center'>
            <EventIcon icon={eventIcon} size={20} />
            {selectedEvent.alias_name ? (
              <div className='heading-with-sub ml-2'>
                <div className='sub'>
                  {PropTextFormat(selectedEvent.display_name)}
                </div>
                <div className='main'>
                  {selectedEvent.event_type === 'FE'
                    ? selectedEvent.event_name
                    : selectedEvent.alias_name}
                </div>
              </div>
            ) : (
              <div className='heading ml-2'>
                {PropTextFormat(selectedEvent.display_name)}
              </div>
            )}
          </div>
        </div>
        <div>
          {Object.entries(selectedEvent.properties || {}).map(
            ([key, value]) => (
              <div className='leftpane-prop' key={key}>
                <div className='flex flex-col items-start truncate'>
                  <Text
                    type='title'
                    level={8}
                    color='grey'
                    truncate
                    charLimit={40}
                    extraClass='m-0'
                  >
                    {key === '$is_page_view' && value === true
                      ? 'Page URL'
                      : TimelineHoverPropDisplayNames[key]
                      ? TimelineHoverPropDisplayNames[key]
                      : PropTextFormat(key)}
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    truncate
                    charLimit={36}
                    extraClass='m-0'
                    shouldTruncateURL
                  >
                    {key === '$is_page_view' && value === true
                      ? selectedEvent.event_type === 'FE'
                        ? selectedEvent.alias_name
                        : selectedEvent.event_name
                      : value
                      ? value
                      : '-'}
                  </Text>
                </div>
              </div>
            )
          )}
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
      mask={false}
      maskClosable={true}
      visible={visible}
      className={'fa-drawer--right'}
    >
      {renderEventDetails()}
    </Drawer>
  );
};

export default EventDrawer;
