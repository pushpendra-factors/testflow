import React, { useEffect, useState } from 'react';
import { Drawer, Button, message } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { EventDrawerProps } from 'Components/Profile/types';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { connect, ConnectedProps, useSelector } from 'react-redux';
import { processProperties, PropTextFormat } from 'Utils/dataFormatter';
import getGroupIcon from 'Utils/getGroupIcon';
import { updateEventPropertiesConfig } from 'Reducers/timelines';
import { fetchProjectSettings } from 'Reducers/global';
import { bindActionCreators } from 'redux';
import logger from 'Utils/logger';
import styles from '../index.module.scss';
import EventDetails from './EventDetails';

function EventDrawer({
  visible,
  onClose,
  event,
  eventPropsType,
  fetchProjectSettings
}: ComponentProps): JSX.Element {
  const { active_project: activeProject } = useSelector(
    (state: any) => state.global
  );
  const { eventPropertiesV2 } = useSelector((state: any) => state.coreQuery);
  const { currentProjectSettings } = useSelector((state: any) => state.global);
  const { activePageView } = useSelector((state: any) => state.timelines);

  const [filterProperties, setFilterProperties] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);

  const handleUpdateEventProps = (newList: string[]) => {
    updateEventPropertiesConfig(
      activeProject?.id,
      event?.display_name === 'Page View' ? 'PageView' : event.event_name,
      newList
    )
      .then(() => {
        fetchProjectSettings(activeProject?.id);
        message.success('Updated Event Properties Configuration');
      })
      .catch((err) => {
        logger.error(err);
        message.error('Error Updating Event Properties Configuration');
      });
  };

  const addNewProp = (option: any, group: any) => {
    const currentList =
      currentProjectSettings?.timelines_config?.events_config?.[
        event?.display_name === 'Page View' ? 'PageView' : event.event_name
      ] || [];

    if (currentList.includes(option.value)) {
      message.error('Property Already Exists');
      return;
    }
    handleUpdateEventProps([...currentList, option.value]);
  };

  const mapEventProperties = (properties: object) =>
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

    if (event && event?.display_name !== 'Page View') {
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

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      fetchProjectSettings
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type ComponentProps = ReduxProps & EventDrawerProps;

export default connector(EventDrawer);
