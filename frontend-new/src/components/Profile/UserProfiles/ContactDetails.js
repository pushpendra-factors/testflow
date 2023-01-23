import React, { useEffect, useState } from 'react';
import { Button, Avatar, Menu, Dropdown, Popover, Tabs } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from '../../factorsComponents';
import FaTimeline from '../../FaTimeline';
import {
  DEFAULT_TIMELINE_CONFIG,
  granularityOptions,
  iconColors
} from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileUserDetails } from '../../../reducers/timelines/middleware';
import { getActivitiesWithEnableKeyConfig } from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../LeftPanePropBlock';
import GroupSelect2 from '../../QueryComposer/GroupSelect2';

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
    const listActivities = getActivitiesWithEnableKeyConfig(
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

  const handleChange = (option) => {
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
          onChange={handleChange}
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
      const value = props[prop] || '-';
      propsList.push(
        <div key={index}>
          <LeftPanePropBlock
            property={prop}
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
          placeholder='Select Event'
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
      <div className='leftpane__user'>
        {user.identity.isAnonymous ? (
          <SVG
            name={`TrackedUser${user.identity.id.match(/\d/g)[0]}`}
            size={80}
          />
        ) : (
          <Avatar
            size={72}
            className='leftpane__user__avatar'
            style={{
              '--avatar-bg': `${iconColors[Math.floor(Math.random() * 8)]}`,
            }}
          >
            {user.identity.id.charAt(0).toUpperCase()}
          </Avatar>
        )}
        <div className='py-2'>
          <Text type='title' level={6} extraClass='m-0' weight='bold'>
            {userDetails.data.title}
          </Text>
          <Text type='title' level={7} extraClass='m-0' color='grey'>
            {userDetails.data.subtitle}
          </Text>
        </div>
      </div>
      <div className='leftpane__props'>
        {listLeftPaneProps(userDetails.data.left_pane_props)}
      </div>
      <div className='px-8 pb-8'>{renderAddNewProp()}</div>
      <div className='leftpane__groups'>
        <Text type='title' level={7} extraClass='m-0 my-2' color='grey'>
          Associated Groups:
        </Text>
        {userDetails?.data?.group_infos?.map((group) => (
          <Text type='title' level={7} extraClass='m-0 mb-2'>
            {group.group_name}
          </Text>
        )) || '-'}
      </div>
    </div>
  );

  const renderTimelineWithActions = () => (
    <div className='timeline-view'>
      <div className='timeline-actions'>
        <Text type='title' level={3} weight='bold'>
          Timeline
        </Text>
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
      <FaTimeline
        activities={activities?.filter((activity) => activity.enabled === true)}
        loading={userDetails.isLoading}
        granularity={granularity}
        collapse={collapse}
        setCollapse={setCollapse}
        eventNamesMap={eventNamesMap}
      />
    </div>
  );

  return (
    <div>
      {renderModalHeader()}
      <div className='fa-timeline'>
        {renderLeftPane()}
        {renderTimelineWithActions()}
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
