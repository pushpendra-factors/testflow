import React, { useEffect, useState } from 'react';
import { ReactSortable } from 'react-sortablejs';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { HolderOutlined } from '@ant-design/icons';
import { SVG, Text } from 'Components/factorsComponents';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { eventIconsColorMap } from 'Components/Profile/constants';
import { propValueFormat } from 'Components/Profile/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import { EventDetailsProps } from 'Components/Profile/types';
import { Button } from 'antd';
import _ from 'lodash';
import { getConfiguredEventProperties } from 'Reducers/timelines/middleware';
import { bindActionCreators } from 'redux';
import EventIcon from './EventIcon';

function EventDetails({
  event,
  eventPropsType,
  onUpdate,
  getConfiguredEventProperties
}: ComponentProps) {
  const [sortableItems, setSortableItems] = useState<string[]>([]);
  const { eventPropNames } = useSelector((state: any) => state.coreQuery);
  const { active_project: activeProject, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );
  const { eventConfigProperties } = useSelector(
    (state: any) => state.timelines
  );

  useEffect(() => {
    if (!event) return;
    if (!eventConfigProperties[event?.id]) {
      const eventName =
        event.display_name === 'Page View' ? 'PageView' : event.name;
      getConfiguredEventProperties(activeProject.id, event.id, eventName);
    }
  }, [activeProject, event, eventConfigProperties]);

  const eventIcon = eventIconsColorMap[event.icon]
    ? event.icon
    : 'calendar-star';

  const renderAliasName = () => (
    <TextWithOverflowTooltip
      text={event.type === 'FE' ? event.name : event.alias_name}
      extraClass='main'
    />
  );

  useEffect(() => {
    if (event && currentProjectSettings?.timelines_config?.events_config) {
      const eventName =
        event.display_name === 'Page View' ? 'PageView' : event.name;
      setSortableItems(
        currentProjectSettings.timelines_config.events_config[eventName]
      );
    }
  }, [event, currentProjectSettings]);

  const compareOrder = (newOrder: string[]) => {
    const existingOrder =
      currentProjectSettings?.timelines_config?.events_config?.[
        event.display_name === 'Page View' ? 'PageView' : event.name
      ];
    if (_.isEqual(existingOrder, newOrder)) return;
    onUpdate(newOrder);
  };

  const handleSortableItems = (newOrder: string[]) => {
    if (!newOrder || !newOrder.length) return;
    setSortableItems(newOrder);
    compareOrder(newOrder);
  };

  const handleDelete = (index: number) => {
    const updatedItems = [...sortableItems];
    updatedItems.splice(index, 1);
    setSortableItems(updatedItems);
    onUpdate(updatedItems);
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
      <div className='event-drawer-items'>
        {sortableItems && (
          <ReactSortable list={sortableItems} setList={handleSortableItems}>
            {sortableItems.map((property, index) => {
              const propType = eventPropsType[property];
              return (
                <div className='leftpane-prop justify-between'>
                  <div className='flex items-center justify-start'>
                    <div className='del-button mr-4' style={{ cursor: 'grab' }}>
                      <HolderOutlined />
                    </div>
                    <div className='flex flex-col items-start'>
                      <Text
                        type='title'
                        level={8}
                        color='grey'
                        truncate
                        charLimit={40}
                        extraClass='m-0'
                      >
                        {eventPropNames[property] || PropTextFormat(property)}
                      </Text>
                      <Text
                        type='title'
                        level={7}
                        truncate
                        charLimit={36}
                        extraClass='m-0'
                      >
                        {propValueFormat(
                          property,
                          eventConfigProperties?.[event.id]?.[property],
                          propType
                        ) || '-'}
                      </Text>
                    </div>
                  </div>

                  {sortableItems.length > 1 && (
                    <Button
                      type='text'
                      className='del-button'
                      onClick={() => handleDelete(index)}
                      icon={<SVG name='delete' />}
                    />
                  )}
                </div>
              );
            })}
          </ReactSortable>
        )}
      </div>
    </div>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      getConfiguredEventProperties
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type ComponentProps = ReduxProps & EventDetailsProps;

export default connector(EventDetails);
