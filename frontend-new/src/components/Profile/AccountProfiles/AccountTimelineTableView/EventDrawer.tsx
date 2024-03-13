import React, { useEffect, useState } from 'react';
import { Drawer, Button, message, Tabs } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { EventDrawerProps } from 'Components/Profile/types';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { connect, ConnectedProps, useSelector } from 'react-redux';
import { processProperties, PropTextFormat } from 'Utils/dataFormatter';
import getGroupIcon from 'Utils/getGroupIcon';
import { updateEventPropertiesConfig } from 'Reducers/timelines';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { bindActionCreators } from 'redux';
import logger from 'Utils/logger';
import styles from '../index.module.scss';
import EventDetails from './EventDetails';
import UserDetails from './UserDetails';

function EventDrawer({
  visible,
  onClose,
  event,
  user,
  eventPropsType,
  fetchProjectSettings,
  udpateProjectSettings
}: ComponentProps): JSX.Element {
  const { active_project: activeProject } = useSelector(
    (state: any) => state.global
  );
  const { eventPropertiesV2, userPropertiesV2 } = useSelector(
    (state: any) => state.coreQuery
  );
  const { currentProjectSettings } = useSelector((state: any) => state.global);
  const { activePageView } = useSelector((state: any) => state.timelines);

  const [eventProperties, setEventProperties] = useState([]);
  const [userProperties, setUserProperties] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('user');

  useEffect(() => {
    if (!event) return;
    if (event?.is_group_event) {
      setActiveTab('event');
    } else {
      setActiveTab('user');
    }
  }, [event]);

  const handleUpdateEventProperties = (newList: string[]) => {
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

  const handleUpdateUserProperties = (newList: string[]) => {
    const timelinesConfig = { ...currentProjectSettings.timelines_config };
    timelinesConfig.user_config.table_props = newList;
    try {
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      });
      message.success('Updated User Properties Configuration');
    } catch (err) {
      logger.error(err);
      message.error('Error Updating User Properties Configuration');
    }
  };

  const addNewEventProp = (option: any, group: any) => {
    const eventPropetiesList =
      currentProjectSettings?.timelines_config?.events_config?.[
        event?.display_name === 'Page View' ? 'PageView' : event.event_name
      ] || [];

    const userPropertiesList =
      currentProjectSettings?.timelines_config?.user_config?.table_props || [];

    const currentList =
      activeTab === 'event' ? eventPropetiesList : userPropertiesList;

    if (currentList.includes(option.value)) {
      message.error('Property Already Exists');
      return;
    }
    if (activeTab === 'event')
      handleUpdateEventProperties([...currentList, option.value]);
    if (activeTab === 'user')
      handleUpdateUserProperties([...currentList, option.value]);
  };

  const mapProperties = (properties: object) =>
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
    if (!event) return;
    let eventProps;

    if (event && event?.display_name !== 'Page View') {
      eventProps = mapProperties(eventPropertiesV2[event.event_name] || {});
    } else {
      eventProps = mapProperties(eventPropertiesV2[activePageView] || {});
    }

    const userProps = mapProperties(userPropertiesV2);

    setEventProperties(eventProps);
    setUserProperties(userProps);
  }, [event, eventPropertiesV2]);

  const selectProps = (type: string) => {
    let showOptions;
    if (type === 'event') showOptions = eventProperties;
    if (type === 'user') showOptions = userProperties;
    return (
      propSelectOpen && (
        <div className={styles.account_profiles__event_selector}>
          <GroupSelect
            options={showOptions}
            searchPlaceHolder='Select Property'
            optionClickCallback={addNewEventProp}
            onClickOutside={() => setPropSelectOpen(false)}
            allowSearchTextSelection={false}
            extraClass={styles.account_profiles__event_selector__select}
            allowSearch
          />
        </div>
      )
    );
  };

  const renderAddNewProp = (type: string) => (
    <div className='ml-2'>
      <Button
        type='link'
        icon={<SVG name='plus' color='purple' />}
        onClick={() => setPropSelectOpen(!propSelectOpen)}
      >
        Add property
      </Button>
      {selectProps(type)}
    </div>
  );

  const handleTabChange = (val: string) => {
    setActiveTab(val);
  };

  return (
    <Drawer
      title={
        <div className='flex justify-between items-center'>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            More Details
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
      <Tabs
        className='timeline-view--tabs'
        defaultActiveKey={activeTab}
        size='small'
        activeKey={activeTab}
        onChange={handleTabChange}
      >
        {!event?.is_group_event && (
          <Tabs.TabPane
            tab={
              <span className='fa-activity-filter--tabname'>
                User Properties
              </span>
            }
            key='user'
          >
            <UserDetails user={user} onUpdate={handleUpdateUserProperties} />
          </Tabs.TabPane>
        )}
        {currentProjectSettings?.timelines_config?.events_config?.[
          event?.display_name === 'Page View' ? 'PageView' : event?.event_name
        ]?.length > 0 && (
          <Tabs.TabPane
            tab={
              <span className='fa-activity-filter--tabname'>
                Event Properties
              </span>
            }
            key='event'
          >
            <EventDetails
              event={event}
              eventPropsType={eventPropsType}
              onUpdate={handleUpdateEventProperties}
            />
          </Tabs.TabPane>
        )}
      </Tabs>
      {renderAddNewProp(activeTab)}
    </Drawer>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type ComponentProps = ReduxProps & EventDrawerProps;

export default connector(EventDrawer);
