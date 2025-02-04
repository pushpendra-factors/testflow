import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Avatar,
  Menu,
  Dropdown,
  Popover,
  Tabs,
  notification
} from 'antd';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  PropTextFormat,
  convertAndAddPropertiesToGroupSelectOptions,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import { useHistory, useLocation } from 'react-router-dom';
import { getEventPropertiesV2 } from 'Reducers/coreQuery/middleware';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import useKey from 'hooks/useKey';
import { PathUrls } from 'Routes/pathUrls';
import styles from './index.module.scss';
import { SVG, Text } from '../../factorsComponents';
import UserTimelineBirdview from './UserTimelineBirdview';
import UserTimelineSingleview from './UserTimelineSingleview';
import { getPropType } from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileUserDetails } from '../../../reducers/timelines/middleware';
import {
  addEnabledFlagToActivities,
  formatUserPropertiesToCheckList
} from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../MyComponents/LeftPanePropBlock';
import { ALPHANUMSTR, GranularityOptions, iconColors } from '../constants';

function ContactDetails({
  userDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileUserDetails,
  userPropertiesV2,
  eventPropertiesV2,
  getEventPropertiesV2,
  eventNamesMap
}) {
  const history = useHistory();
  const location = useLocation();
  const [activities, setActivities] = useState([]);
  const [granularity, setGranularity] = useState('Daily');
  const [collapse, setCollapse] = useState(true);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTLConfig] = useState({});
  const [checkListMilestones, setCheckListMilestones] = useState([]);
  const [requestedEvents, setRequestedEvents] = useState({});
  const [eventPropertiesType, setEventPropertiesType] = useState({});
  const { TabPane } = Tabs;

  const [openPopover, setOpenPopover] = useState(false);

  const handleOpenPopoverChange = (value) => {
    setOpenPopover(value);
  };

  const { userPropNames } = useSelector((state) => state.coreQuery);

  useEffect(
    () => () => {
      setGranularity('Daily');
      setCollapse(true);
      setPropSelectOpen(false);
    },
    []
  );

  const uniqueEventNames = useMemo(() => {
    const userEvents = userDetails.data?.user_activities || [];

    const eventsArray = userEvents
      .filter((event) => event?.display_name === 'Page View')
      .map((event) => event.event_name);

    const pageViewEvent = userEvents.find(
      (event) => event?.display_name === 'Page View'
    );

    if (pageViewEvent) {
      eventsArray.push(pageViewEvent.event_name);
    }

    return Array.from(new Set(eventsArray));
  }, [userDetails.data?.user_activities]);

  const fetchEventPropertiesType = async () => {
    const promises = uniqueEventNames.map(async (eventName) => {
      if (!requestedEvents[eventName]) {
        setRequestedEvents((prevRequestedEvents) => ({
          ...prevRequestedEvents,
          [eventName]: true
        }));
        if (!eventPropertiesV2[eventName])
          getEventPropertiesV2(activeProject?.id, eventName);
      }
    });

    await Promise.allSettled(promises);

    const typeMap = {};
    Object.values(eventPropertiesV2).forEach((propertyGroup) => {
      Object.values(propertyGroup || {}).forEach((arr) => {
        arr.forEach((property) => {
          typeMap[property[1]] = property[2];
        });
      });
    });
    setEventPropertiesType(typeMap);
  };

  useEffect(() => {
    fetchEventPropertiesType();
  }, [uniqueEventNames, requestedEvents, activeProject?.id, eventPropertiesV2]);

  const [userID, isAnonymous] = useMemo(() => {
    const decodedUserID = atob(location.pathname.split('/').pop());
    const isUserAnonymous = location.search.split('=').pop() === 'true';
    document.title = 'People - FactorsAI';
    return [decodedUserID, isUserAnonymous];
  }, [location]);

  useEffect(() => {
    if (userID && userID !== '')
      getProfileUserDetails(
        activeProject.id,
        userID,
        isAnonymous,
        currentProjectSettings?.timelines_config
      );
  }, [activeProject.id, userID, isAnonymous]);

  useEffect(() => {
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    const lsitDatetimeProperties = userPropertiesModified.filter(
      (item) => item[2] === 'datetime'
    );
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      lsitDatetimeProperties,
      currentProjectSettings.timelines_config?.user_config?.milestones
    );
    setCheckListMilestones(userPropsWithEnableKey);
  }, [currentProjectSettings, userPropertiesV2]);

  useEffect(() => {
    if (!currentProjectSettings?.timelines_config) return;
    setTLConfig(currentProjectSettings.timelines_config);
  }, [currentProjectSettings?.timelines_config]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  const generateUserProps = () => {
    const filterOptsObj = {};
    if (userPropertiesV2) {
      convertAndAddPropertiesToGroupSelectOptions(
        userPropertiesV2,
        filterOptsObj,
        'user'
      );
    }
    return Object.values(filterOptsObj);
  };

  useEffect(() => {
    const listActivities = addEnabledFlagToActivities(
      userDetails?.data?.user_activities,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, userDetails]);

  const granularityMenu = (
    <Menu>
      {GranularityOptions.map((option) => (
        <Menu.Item key={option} onClick={(key) => setGranularity(key.key)}>
          <div className='flex items-center'>
            <span className='mr-3'>{option}</span>
          </div>
        </Menu.Item>
      ))}
    </Menu>
  );

  const handleEventsChange = (option) => {
    const timelinesConfig = { ...tlConfig };
    if (option.enabled) {
      timelinesConfig.disabled_events.push(option.display_name);
    } else if (!option.enabled) {
      timelinesConfig.disabled_events.splice(
        timelinesConfig.disabled_events.indexOf(option.display_name),
        1
      );
    }
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
    setOpenPopover(false);
  };

  const handleMilestonesChange = (option) => {
    if (
      option.enabled ||
      checkListMilestones.filter((item) => item.enabled === true).length < 5
    ) {
      const checkListProps = [...checkListMilestones];
      const optIndex = checkListProps.findIndex(
        (obj) => obj.prop_name === option.prop_name
      );
      checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
      setCheckListMilestones(checkListProps);
      setOpenPopover(false);
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum of 5 Milestones Selection Reached.',
        duration: 2
      });
    }
  };

  const applyMilestones = () => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.user_config.milestones = checkListMilestones
      .filter((item) => item.enabled === true)
      .map((item) => item?.prop_name);
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    }).then(() =>
      getProfileUserDetails(
        activeProject?.id,
        userID,
        isAnonymous,
        currentProjectSettings?.timelines_config
      )
    );
  };

  const controlsPopover = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <Tabs.TabPane
        tab={<span className='fa-activity-filter--tabname'>Events</span>}
        key='events'
      >
        <SearchCheckList
          placeholder='Select Events to Show'
          mapArray={activities}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handleEventsChange}
        />
      </Tabs.TabPane>
      <Tabs.TabPane
        tab={<span className='fa-activity-filter--tabname'>Milestones</span>}
        key='milestones'
      >
        <SearchCheckList
          placeholder='Select Upto 5 Milestones'
          mapArray={checkListMilestones}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handleMilestonesChange}
          showApply
          onApply={applyMilestones}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const handleOptionBackClick = useCallback(() => {
    history.replace(PathUrls.ProfilePeople, {
      activeSegment: location.state?.activeSegment,
      fromDetails: true,
      timelinePayload: location.state?.timelinePayload,
      currentPage: location.state?.currentPage,
      currentPageSize: location.state?.currentPageSize,
      activeSorter: location.state?.activeSorter,
      appliedFilters: location.state?.appliedFilters,
      peoplesTableRow: location.state?.peoplesTableRow
    });
  }, []);

  const renderHeader = () => (
    <div className='fa-timeline--header'>
      <div className='flex items-center'>
        <div
          className='flex items-center cursor-pointer'
          onClick={handleOptionBackClick}
        >
          <Text type='title' level={6} weight='bold' extraClass='m-0 underline'>
            User Profiles
          </Text>
        </div>
        {userDetails.data?.title && (
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            {`\u00A0/ ${userDetails.data.title}`}
          </Text>
        )}
      </div>
      <Button size='large' onClick={handleOptionBackClick}>
        Close
      </Button>
    </div>
  );

  const handleOptionClick = (option, group) => {
    const timelinesConfig = { ...tlConfig };
    if (!timelinesConfig.account_config.table_props) {
      timelinesConfig.account_config.table_props = [];
    }

    if (!timelinesConfig.user_config.table_props.includes(option?.value)) {
      timelinesConfig.user_config.table_props.push(option?.value);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      }).then(() =>
        getProfileUserDetails(
          activeProject?.id,
          userID,
          isAnonymous,
          currentProjectSettings?.timelines_config
        )
      );
    }
    setPropSelectOpen(false);
  };

  const onDelete = (option) => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.user_config.table_props.splice(
      timelinesConfig.user_config.table_props.indexOf(option),
      1
    );
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const listLeftPaneProps = (props = []) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.user_config?.table_props?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      ) || [];
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    showProps.forEach((prop, index) => {
      const propType = getPropType(userPropertiesModified, prop);
      const propDisplayName = userPropNames[prop]
        ? userPropNames[prop]
        : PropTextFormat(prop);
      const value = props[prop] || '-';
      propsList.push(
        <LeftPanePropBlock
          property={prop}
          type={propType}
          displayName={propDisplayName}
          value={value}
          onDelete={onDelete}
        />
      );
    });
    return propsList;
  };

  const selectProps = () =>
    propSelectOpen && (
      <div className={styles.user_profiles__event_selector}>
        <GroupSelect
          options={generateUserProps()}
          searchPlaceHolder='Select Property'
          optionClickCallback={handleOptionClick}
          onClickOutside={() => setPropSelectOpen(false)}
          allowSearchTextSelection={false}
          extraClass={styles.user_profiles__event_selector__select}
          allowSearch
        />
      </div>
    );

  const renderAddNewProp = () => (
    <>
      <Button
        type='link'
        icon={<SVG name='plus' color='purple' />}
        onClick={() => setPropSelectOpen(!propSelectOpen)}
      >
        Add property
      </Button>
      {selectProps()}
    </>
  );

  const renderLeftPane = () => (
    <div className='leftpane'>
      <div className='header'>
        <div className='user'>
          {isAnonymous ? (
            <SVG
              name={`TrackedUser${userID.match(/\d/g)?.[0] || 0}`}
              size={96}
            />
          ) : (
            <Avatar
              size={96}
              className='avatar'
              style={{
                backgroundColor: `${
                  iconColors[
                    ALPHANUMSTR.indexOf(userID.charAt(0).toUpperCase()) % 8
                  ]
                }`,
                fontSize: '32px'
              }}
            >
              {userID.charAt(0).toUpperCase()}
            </Avatar>
          )}
          <div className='py-2'>
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {userDetails.data.title}
            </Text>
            {isAnonymous ? null : (
              <Text type='title' level={7} extraClass='m-0' color='grey'>
                {userDetails.data.subtitle}
              </Text>
            )}
          </div>
        </div>
        <div className='account inline-flex gap--8'>
          <div className='icon'>
            <SVG name='globe' size={20} />
          </div>
          <div className='flex flex-col items-start'>
            <Text type='title' level={8} color='grey' extraClass='m-0'>
              Account:
            </Text>
            <Text type='title' level={7} extraClass='m-0'>
              {userDetails.data.account || '-'}
            </Text>
          </div>
        </div>
      </div>
      <div className='props'>
        {listLeftPaneProps(userDetails.data.leftpane_props)}
      </div>
      {!currentProjectSettings?.timelines_config?.user_config?.table_props ||
      currentProjectSettings?.timelines_config?.user_config?.table_props
        ?.length < 8 ? (
        <div className='add-prop-btn'>{renderAddNewProp()}</div>
      ) : null}
    </div>
  );

  const renderSingleTimelineView = () => (
    <>
      <div className='h-6' />
      <UserTimelineSingleview
        activities={activities?.filter((activity) => activity.enabled === true)}
        milestones={userDetails.data?.milestones || {}}
        loading={userDetails.isLoading}
        propertiesType={eventPropertiesType}
        eventNamesMap={eventNamesMap}
      />
    </>
  );

  const renderBirdviewWithActions = () => (
    <div className='flex flex-col'>
      <div className='tl-actions-row'>
        <div className='collapse-btns'>
          <Button
            className='collapse-btns--btn'
            onClick={() => setCollapse(false)}
          >
            <SVG name='line_height' size={22} />
          </Button>
          <Button
            className='collapse-btns--btn'
            onClick={() => setCollapse(true)}
          >
            <SVG name='grip_lines' size={22} />
          </Button>
        </div>
        <Popover
          overlayClassName='fa-activity--filter'
          placement='bottomLeft'
          trigger='click'
          content={controlsPopover}
          open={openPopover}
          onOpenChange={handleOpenPopoverChange}
        >
          <Button type='text'>
            <SVG name='activity_filter' />
          </Button>
        </Popover>
        <Dropdown
          overlay={granularityMenu}
          placement='bottomRight'
          trigger={['click']}
        >
          <Button type='text' className='flex items-center'>
            {granularity}
            <SVG name='caretDown' size={16} extraClass='ml-1' />
          </Button>
        </Dropdown>
      </div>
      <UserTimelineBirdview
        activities={activities?.filter((activity) => activity.enabled === true)}
        milestones={userDetails.data?.milestones || {}}
        loading={userDetails.isLoading}
        granularity={granularity}
        collapse={collapse}
        setCollapse={setCollapse}
        eventNamesMap={eventNamesMap}
        propertiesType={eventPropertiesType}
      />
    </div>
  );

  const renderTimelineView = () => (
    <div className='timeline-view'>
      <Tabs
        className='timeline-view--tabs'
        defaultActiveKey='birdview'
        size='small'
        onChange={() => setGranularity(granularity)}
      >
        <TabPane
          tab={<span className='fa-activity-filter--tabname'>Timeline</span>}
          key='timeline'
        >
          {renderSingleTimelineView()}
        </TabPane>
        <TabPane
          tab={<span className='fa-activity-filter--tabname'>Birdview</span>}
          key='birdview'
        >
          {renderBirdviewWithActions()}
        </TabPane>
      </Tabs>
    </div>
  );
  useKey(['Escape'], handleOptionBackClick);

  return (
    <div className='fa-timeline'>
      {renderHeader()}
      <div className='fa-timeline--content'>
        {renderLeftPane()}
        {renderTimelineView()}
      </div>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  userDetails: state.timelines.contactDetails,
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileUserDetails,
      fetchProjectSettings,
      udpateProjectSettings,
      getEventPropertiesV2
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ContactDetails);
