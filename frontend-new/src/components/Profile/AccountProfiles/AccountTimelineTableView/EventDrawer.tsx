import React, { useEffect, useMemo } from 'react';
import { connect, ConnectedProps, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Drawer, Tooltip } from 'antd';
import { Text } from 'Components/factorsComponents';
import {
  eventIconsColorMap,
  getPropType,
  propValueFormat
} from 'Components/Profile/utils';
import {
  convertGroupedPropertiesToUngrouped,
  PropTextFormat
} from 'Utils/dataFormatter';
import EventIcon from './EventIcon';
const EventDrawer: React.FC<EventDrawerProps> = ({
  visible,
  onClose,
  event
}) => {
  const { eventPropertiesV2 } = useSelector((state: any) => state.coreQuery);
  const eventPropertiesModified = useMemo(() => {
    if (!event || !event.event_name || !eventPropertiesV2?.[event.event_name])
      return null;

    const eventProps: any = [];
    convertGroupedPropertiesToUngrouped(
      eventPropertiesV2[event.event_name],
      eventProps
    );
    return eventProps;
  }, [event?.event_name, eventPropertiesV2]);

  const renderEventDetails = () => {
    if (!event) return null;

    const eventIcon = eventIconsColorMap[event.icon]
      ? event.icon
      : 'calendar-star';

    const renderAliasName = () => (
      <Tooltip
        title={event.event_type === 'FE' ? event.event_name : event.alias_name}
      >
        <div className='main'>
          {event.event_type === 'FE' ? event.event_name : event.alias_name}
        </div>
      </Tooltip>
    );

    return (
      <div className='p-4'>
        <div className='top-section mb-4'>
          <div className='flex items-center w-full'>
            <EventIcon icon={eventIcon} size={28} />
            {event.alias_name ? (
              <div className='heading-with-sub ml-2'>
                <div className='sub'>{PropTextFormat(event.display_name)}</div>
                {renderAliasName()}
              </div>
            ) : (
              <Tooltip title={PropTextFormat(event.display_name)}>
                <div className='heading ml-2'>
                  {PropTextFormat(event.display_name)}
                </div>
              </Tooltip>
            )}
          </div>
        </div>
        <div>
          {Object.entries(event.properties || {}).map(([key, value]) => {
            const propType = getPropType(eventPropertiesModified, key);
            const isPageView = key === '$is_page_view' && value === true;
            const customValue = isPageView
              ? event.event_type === 'FE'
                ? event.alias_name
                : event.event_name
              : null;

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
                    {isPageView ? 'Page URL' : PropTextFormat(key)}
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    truncate
                    charLimit={44}
                    extraClass='m-0'
                    shouldTruncateURL
                  >
                    {customValue || propValueFormat(key, value, propType)}
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
      mask={true}
      maskClosable={true}
      visible={visible}
      className={'fa-drawer--right'}
      onClose={onClose}
    >
      {renderEventDetails()}
    </Drawer>
  );
};

export default EventDrawer;
