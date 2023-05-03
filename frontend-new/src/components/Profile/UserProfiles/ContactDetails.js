import React, { useEffect, useState } from 'react';
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
import { SVG, Text } from '../../factorsComponents';
import UserTimelineBirdview from './UserTimelineBirdview';
import UserTimelineSingleview from './UserTimelineSingleview';
import {
  ALPHANUMSTR,
  DEFAULT_TIMELINE_CONFIG,
  getPropType,
  granularityOptions,
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

function ContactDetails({
  user,
  onCancel,
  userDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileUserDetails,
  userProperties,
  eventNamesMap
}) {
  const [activities, setActivities] = useState([]);
  const [granularity, setGranularity] = useState('Daily');
  const [collapse, setCollapse] = useState(true);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTLConfig] = useState({});
  const [checkListMilestones, setCheckListMilestones] = useState([]);
  const { TabPane } = Tabs;

  const { userPropNames } = useSelector((state) => state.coreQuery);

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
      setCheckListMilestones(
        checkListProps.sort((a, b) => b.enabled - a.enabled)
      );
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
        user?.identity?.id,
        user?.identity?.isAnonymous,
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
          placeholder='Search Events'
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
            onCancel();
            setCollapse(true);
            setGranularity('Daily');
            setPropSelectOpen(false);
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
          onCancel();
          setCollapse(true);
          setGranularity('Daily');
          setPropSelectOpen(false);
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
      }).then(() =>
        getProfileUserDetails(
          activeProject?.id,
          user?.identity?.id,
          user?.identity?.isAnonymous,
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
        <div key={index}>
          <LeftPanePropBlock
            property={prop}
            type={propType}
            displayName={propDisplayName}
            value={value}
            onDelete={onDelete}
          />
        </div>
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
        {user.identity.isAnonymous ? (
          <SVG
            name={`TrackedUser${user.identity.id.match(/\d/g)[0]}`}
            size={96}
          />
        ) : (
          <Avatar
            size={96}
            className='avatar'
            style={{
              backgroundColor: `${
                iconColors[
                  ALPHANUMSTR.indexOf(
                    user.identity.id.charAt(0).toUpperCase()
                  ) % 8
                ]
              }`,
              fontSize: '32px'
            }}
          >
            {user.identity.id.charAt(0).toUpperCase()}
          </Avatar>
        )}
        <div className='py-2'>
          <Text type='title' level={6} extraClass='m-0' weight='bold'>
            {userDetails.data.title}
          </Text>
          {user.identity.isAnonymous ? null : (
            <Text type='title' level={7} extraClass='m-0' color='grey'>
              {userDetails.data.subtitle}
            </Text>
          )}
        </div>
      </div>
      <div className='props'>
        {listLeftPaneProps(userDetails.data.left_pane_props)}
        <div className='px-8 pb-8 pt-2'>{renderAddNewProp()}</div>
        <div className='groups'>
          <Text type='title' level={7} extraClass='m-0 my-2' color='grey'>
            Associated Groups:
          </Text>
          {userDetails?.data?.group_infos?.map((group) => {
            return (
              <div className='flex flex-col items-start mb-2'>
                <Text type='title' level={7} extraClass='m-0'>
                  {group?.group_name}
                </Text>
                <Text type='title' level={7} extraClass='m-0' color='grey'>
                  {group?.associated_group || '-'}
                </Text>
              </div>
            );
          })}
        </div>
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
            trigger='hover'
            content={controlsPopover}
          >
            <Button
              size='large'
              className='fa-btn--custom mx-2 relative'
              type='text'
            >
              <SVG name='activity_filter' />
            </Button>
          </Popover>
          <Dropdown overlay={granularityMenu} placement='bottomRight'>
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
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileUserDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ContactDetails);
