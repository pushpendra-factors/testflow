import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Avatar,
  Menu,
  Dropdown,
  Popover,
  Tabs,
  notification
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from '../../factorsComponents';
import UserTimelineBirdview from './UserTimelineBirdview';
import UserTimelineSingleview from './UserTimelineSingleview';
import {
  ALPHANUMSTR,
  DEFAULT_TIMELINE_CONFIG,
  getPropType,
  granularityOptions,
  hoverEvents,
  iconColors
} from '../utils';
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
import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import { PropTextFormat } from 'Utils/dataFormatter';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import { useHistory, useLocation } from 'react-router-dom';
import { getEventProperties } from 'Reducers/coreQuery/middleware';

function ContactDetails({
  userDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileUserDetails,
  userProperties,
  eventNamesMap,
  eventProperties,
  getEventProperties
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const [activities, setActivities] = useState([]);
  const [granularity, setGranularity] = useState('Daily');
  const [collapse, setCollapse] = useState(true);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTLConfig] = useState({});
  const [checkListMilestones, setCheckListMilestones] = useState([]);
  const { TabPane } = Tabs;

  const [openPopover, setOpenPopover] = useState(false);

  const handleOpenPopoverChange = (value) => {
    setOpenPopover(value);
  };

  const { userPropNames } = useSelector((state) => state.coreQuery);

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  useEffect(() => {
    return () => {
      setGranularity('Daily');
      setCollapse(true);
      setPropSelectOpen(false);
    };
  }, []);

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
    const lsitDatetimeProperties = userProperties.filter(
      (item) => item[2] === 'datetime'
    );
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      lsitDatetimeProperties,
      currentProjectSettings.timelines_config?.user_config?.milestones
    );
    setCheckListMilestones(userPropsWithEnableKey);
  }, [currentProjectSettings, userProperties]);

  useEffect(() => {
    if (currentProjectSettings?.timelines_config) {
      const timelinesConfig = {};
      timelinesConfig.disabled_events = [
        ...currentProjectSettings?.timelines_config?.disabled_events
      ];
      timelinesConfig.user_config = {
        ...DEFAULT_TIMELINE_CONFIG.user_config,
        ...currentProjectSettings?.timelines_config?.user_config
      };
      timelinesConfig.account_config = {
        ...DEFAULT_TIMELINE_CONFIG.account_config,
        ...currentProjectSettings?.timelines_config?.account_config
      };
      setTLConfig(timelinesConfig);
    }
  }, [currentProjectSettings]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  const generateUserProps = () => {
    const groupProps = [{ label: 'User Properties', icon: 'user', values: [] }];
    groupProps[0].values = userProperties;
    return groupProps;
  };

  useEffect(() => {
    hoverEvents.forEach((event) => {
      if (!eventProperties[event] &&
        userDetails?.data?.user_activities?.some(
            (activity) => activity?.event_name === event
          )
      ) {
        getEventProperties(activeProject?.id, event);
      }
    });
  }, [activeProject?.id, eventProperties, userDetails?.data?.user_activities]);

  useEffect(() => {
    const listActivities = addEnabledFlagToActivities(
      userDetails?.data?.user_activities,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, userDetails]);

  const granularityMenu = (
    <Menu>
      {granularityOptions.map((option) => (
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
    setOpenPopover(false)
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
      setOpenPopover(false)
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
    })
      .then(() => fetchProjectSettings(activeProject.id))
      .then(() =>
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

  const renderModalHeader = () => (
    <div className='fa-timeline-modal--header'>
      <div className='flex items-center'>
        <Button
          style={{ padding: 0 }}
          type='text'
          icon={<SVG name='brand' size={36} />}
          size='large'
          onClick={() => {
            history.goBack();
          }}
        />
        <Text type='title' level={4} weight='bold' extraClass='m-0'>
          Contact Details
        </Text>
      </div>
      <Button
        size='large'
        type='text'
        onClick={() => {
          history.goBack();
        }}
        icon={<SVG name='times' />}
      />
    </div>
  );

  const handleOptionClick = (group, value) => {
    const timelinesConfig = { ...tlConfig };
    if (!timelinesConfig.user_config.leftpane_props.includes(value[1])) {
      timelinesConfig.user_config.leftpane_props.push(value[1]);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      })
        .then(() => fetchProjectSettings(activeProject.id))
        .then(() =>
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
    timelinesConfig.user_config.leftpane_props.splice(
      timelinesConfig.user_config.leftpane_props.indexOf(option),
      1
    );
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const listLeftPaneProps = (props = []) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.user_config?.leftpane_props ||
      [];

    showProps.forEach((prop, index) => {
      const propType = getPropType(userProperties, prop);
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
      <div className='relative'>
        <GroupSelect2
          groupedProperties={generateUserProps()}
          placeholder='Select Property'
          optionClick={handleOptionClick}
          onClickOutside={() => setPropSelectOpen(false)}
        />
      </div>
    );

  const renderAddNewProp = () =>
    !currentProjectSettings?.timelines_config?.user_config?.leftpane_props ||
    currentProjectSettings?.timelines_config?.user_config?.leftpane_props
      ?.length < 8 ? (
      <div>
        <Button
          type='link'
          icon={<SVG name='plus' color='purple' />}
          onClick={() => setPropSelectOpen(!propSelectOpen)}
        >
          Add property
        </Button>
        {selectProps()}
      </div>
    ) : null;

  const renderLeftPane = () => (
    <div className='leftpane'>
      <div className='user'>
        {isAnonymous ? (
          <SVG name={`TrackedUser${userID.match(/\d/g)?.[0] || 0}`} size={96} />
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
      <div className='props'>
        {listLeftPaneProps(userDetails.data.left_pane_props)}
        <div className='px-8 pb-8 pt-2'>{renderAddNewProp()}</div>
      </div>
    </div>
  );

  const renderSingleTimelineView = () => (
    <UserTimelineSingleview
      activities={activities?.filter((activity) => activity.enabled === true)}
      milestones={userDetails.data?.milestones || {}}
      loading={userDetails.isLoading}
      eventNamesMap={eventNamesMap}
      listProperties={userProperties}
    />
  );

  const renderBirdviewWithActions = () => (
    <div className='flex flex-col'>
      <div className='timeline-actions flex-row-reverse'>
        <div className='timeline-actions__group'>
          <div className='timeline-actions__group__collapse'>
            <Button
              className='collapse-btn collapse-btn--left'
              type='text'
              onClick={() => setCollapse(false)}
            >
              <SVG name='line_height' size={22} />
            </Button>
            <Button
              className='collapse-btn collapse-btn--right'
              type='text'
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
            <Button
              size='large'
              className='fa-btn--custom mx-2 relative'
              type='text'
            >
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
      </div>
      <UserTimelineBirdview
        activities={activities?.filter((activity) => activity.enabled === true)}
        milestones={userDetails.data?.milestones || {}}
        loading={userDetails.isLoading}
        granularity={granularity}
        collapse={collapse}
        setCollapse={setCollapse}
        eventNamesMap={eventNamesMap}
        listProperties={userProperties}
      />
    </div>
  );

  const renderTimelineView = () => {
    return (
      <div className='timeline-view'>
        <Tabs
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
  };

  return (
    <div>
      {renderModalHeader()}
      <div className='fa-timeline'>
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
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileUserDetails,
      fetchProjectSettings,
      udpateProjectSettings,
      getEventProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ContactDetails);
