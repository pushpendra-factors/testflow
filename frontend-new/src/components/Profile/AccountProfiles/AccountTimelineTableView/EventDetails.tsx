import React, { useEffect, useState } from 'react';
import { ReactSortable } from 'react-sortablejs';
import { useSelector } from 'react-redux';
import { HolderOutlined } from '@ant-design/icons';
import { SVG, Text } from 'Components/factorsComponents';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { eventIconsColorMap } from 'Components/Profile/constants';
import { propValueFormat } from 'Components/Profile/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import { EventDetailsProps } from 'Components/Profile/types';
import { Button } from 'antd';
import EventIcon from './EventIcon';

function EventDetails({ event, eventPropsType, onUpdate }: EventDetailsProps) {
  const [sortableItems, setSortableItems] = useState<[string, unknown][]>([]);
  const { eventPropNames } = useSelector((state: any) => state.coreQuery);

  const eventIcon = eventIconsColorMap[event.icon]
    ? event.icon
    : 'calendar-star';

  const renderAliasName = () => (
    <TextWithOverflowTooltip
      text={event.event_type === 'FE' ? event.event_name : event.alias_name}
      extraClass='main'
    />
  );

  useEffect(() => {
    if (event) {
      setSortableItems(
        Object.entries(event.properties || {}).filter(
          (item) => item[0] !== '$is_page_view'
        )
      );
    }
  }, [event]);

  const handleSortableItems = (newOrder: [string, unknown][]) => {
    setSortableItems(newOrder);
    onUpdate(newOrder.map((item) => item[0]));
  };

  const handleDelete = (index: number) => {
    const updatedItems = [...sortableItems];
    updatedItems.splice(index, 1);
    setSortableItems(updatedItems);
    onUpdate(updatedItems.map((item) => item[0]));
  };

  if (!event) return null;

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
        <ReactSortable list={sortableItems} setList={handleSortableItems}>
          {sortableItems.map(([key, value], index) => {
            const propType = eventPropsType[key];
            return (
              <div className='leftpane-prop justify-between' key={key}>
                <div className='flex items-center justify-start'>
                  <div className='del-button mr-4' style={{ cursor: 'grab' }}>
                    <HolderOutlined />
                  </div>
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

                <Button
                  type='text'
                  className='del-button'
                  onClick={() => handleDelete(index)}
                  icon={<SVG name='delete' />}
                />
              </div>
            );
          })}
        </ReactSortable>
      </div>
    </div>
  );
}

export default EventDetails;
