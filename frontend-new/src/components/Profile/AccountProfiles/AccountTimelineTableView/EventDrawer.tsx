import React, { useEffect, useState } from 'react';
import { Drawer, Button, message } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { EventDrawerProps } from 'Components/Profile/types';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { useSelector } from 'react-redux';
import { processProperties, PropTextFormat } from 'Utils/dataFormatter';
import getGroupIcon from 'Utils/getGroupIcon';
import { updateEventPropertiesConfig } from 'Reducers/timelines';
import styles from '../index.module.scss';
import EventDetails from './EventDetails';

function EventDrawer({
  visible,
  onClose,
  event,
  eventPropsType
}: EventDrawerProps): JSX.Element {
  const { active_project: activeProject } = useSelector(
    (state: any) => state.global
  );
  const { eventPropertiesV2 } = useSelector((state: any) => state.coreQuery);
  const { activePageView } = useSelector((state: any) => state.timelines);

  const [filterProperties, setFilterProperties] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);

  const handleUpdateEventProps = (newList: string[]) => {
    updateEventPropertiesConfig(
      activeProject?.id,
      event?.properties?.$is_page_view ? 'PageView' : event.event_name,
      newList
    )
      .then(() => {
        message.success('Updated Event Properties Configuration');
      })
      .catch((err) => {
        message.error('Error Updating Event Properties Configuration');
      });
  };

  const addNewProp = (option: any, group: any) => {
    handleUpdateEventProps([
      ...Object.keys(event.properties || {}),
      option.value
    ]);
  };

  const mapEventProperties = (properties) =>
    Object.entries(properties)
      ?.map(([group, values]) => ({
        label: PropTextFormat(group),
        iconName: group,
        values: processProperties(values)
      }))
      ?.map((opt) => ({
        iconName: getGroupIcon(opt.iconName),
        label: opt.label,
        values: opt.values
      }));

  useEffect(() => {
    let eventProps;

    if (event && !event?.properties?.$is_page_view) {
      eventProps = mapEventProperties(
        eventPropertiesV2[event.event_name] || {}
      );
    } else {
      eventProps = mapEventProperties(eventPropertiesV2[activePageView] || {});
    }

    setFilterProperties(eventProps);
  }, [event, eventPropertiesV2]);

  const selectProps = () =>
    propSelectOpen && (
      <div className={styles.account_profiles__event_selector}>
        <GroupSelect
          options={filterProperties}
          searchPlaceHolder='Select Property'
          optionClickCallback={addNewProp}
          onClickOutside={() => setPropSelectOpen(false)}
          allowSearchTextSelection={false}
          extraClass={styles.account_profiles__event_selector__select}
          allowSearch
        />
      </div>
    );

  const renderAddNewProp = () => (
    <div className='ml-2'>
      <Button
        type='link'
        icon={<SVG name='plus' color='purple' />}
        onClick={() => setPropSelectOpen(!propSelectOpen)}
      >
        Add property
      </Button>
      {selectProps()}
    </div>
  );

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
      <EventDetails
        event={event}
        eventPropsType={eventPropsType}
        onUpdate={handleUpdateEventProps}
      />
      {renderAddNewProp()}
    </Drawer>
  );
}

export default EventDrawer;
