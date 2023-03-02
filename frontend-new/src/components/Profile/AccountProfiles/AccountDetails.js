import React, { useState, useEffect } from 'react';
import { Button, Dropdown, Menu, notification, Popover, Tabs } from 'antd';
import { bindActionCreators } from 'redux';
import { connect, useSelector } from 'react-redux';
import { Text, SVG } from '../../factorsComponents';
import AccountTimelineBirdView from './AccountTimelineBirdView';
import {
  DEFAULT_TIMELINE_CONFIG,
  getHost,
  getPropType,
  granularityOptions
} from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileAccountDetails } from '../../../reducers/timelines/middleware';
import {
  formatUserPropertiesToCheckList,
  getActivitiesWithEnableKeyConfig
} from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../MyComponents/LeftPanePropBlock';
import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import AccountTimelineSingleView from './AccountTimelineSingleView';
import { PropTextFormat } from 'Utils/dataFormatter';

function AccountDetails({
  accountId,
  onCancel,
  accountDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileAccountDetails,
  userProperties,
  groupProperties,
  eventNamesMap
}) {
  const [granularity, setGranularity] = useState('Daily');
  const [collapseAll, setCollapseAll] = useState(true);
  const [activities, setActivities] = useState([]);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [listProperties, setListProperties] = useState([]);
  const [checkListMilestones, setCheckListMilestones] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const { TabPane } = Tabs;

  const { groupPropNames } = useSelector((state) => state.coreQuery);

  useEffect(() => {
    const lsitDatetimeProperties = listProperties.filter(
      (item) => item[2] === 'datetime'
    );
    const propsWithEnableKey = formatUserPropertiesToCheckList(
      lsitDatetimeProperties,
      currentProjectSettings.timelines_config?.account_config?.milestones
    );
    setCheckListMilestones(propsWithEnableKey);
  }, [currentProjectSettings, listProperties]);

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

  useEffect(() => {
    const listActivities = getActivitiesWithEnableKeyConfig(
      accountDetails.data?.account_events,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, accountDetails]);

  useEffect(() => {
    const mergeGroupedProps = [
      ...(groupProperties.$hubspot_company
        ? groupProperties.$hubspot_company
        : []),
      ...(groupProperties.$salesforce_account
        ? groupProperties.$salesforce_account
        : [])
    ];
    setListProperties(mergeGroupedProps);
  }, [groupProperties]);

  useEffect(() => {
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userProperties,
      [currentProjectSettings.timelines_config?.account_config?.user_prop]
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [currentProjectSettings, userProperties]);

  const handlePropChange = (option) => {
    if (
      option.prop_name !==
      currentProjectSettings.timelines_config.account_config.user_prop
    ) {
      const timelinesConfig = { ...currentProjectSettings.timelines_config };
      timelinesConfig.account_config.user_prop = option.prop_name;
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      }).then(() =>
        getProfileAccountDetails(
          activeProject.id,
          accountId,
          currentProjectSettings?.timelines_config
        )
      );
    }
  };

  const handleEventsChange = (option) => {
    const timelinesConfig = { ...currentProjectSettings.timelines_config };
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
    timelinesConfig.account_config.milestones = checkListMilestones
      .filter((item) => item.enabled === true)
      .map((item) => item?.prop_name);
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    }).then(() =>
      getProfileAccountDetails(
        activeProject.id,
        accountId,
        currentProjectSettings?.timelines_config
      )
    );
  };

  const controlsPopover = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <TabPane
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
      </TabPane>
      <TabPane
        tab={<span className='fa-activity-filter--tabname'>Properties</span>}
        key='properties'
      >
        <SearchCheckList
          placeholder='Search Properties'
          mapArray={checkListUserProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
        />
      </TabPane>
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

  const handleOptionClick = (group, value) => {
    const timelinesConfig = { ...tlConfig };
    if (!timelinesConfig.account_config.leftpane_props.includes(value[1])) {
      timelinesConfig.account_config.leftpane_props.push(value[1]);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      }).then(() =>
        getProfileAccountDetails(
          activeProject.id,
          accountId,
          currentProjectSettings?.timelines_config
        )
      );
    }
    setPropSelectOpen(false);
  };

  const onDelete = (option) => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.account_config.leftpane_props.splice(
      timelinesConfig.account_config.leftpane_props.indexOf(option),
      1
    );
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

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
            setGranularity('Daily');
            setCollapseAll(true);
            setPropSelectOpen(false);
          }}
        />
        <Text type='title' level={4} weight='bold' extraClass='m-0'>
          Account Details
        </Text>
      </div>
      <Button
        size='large'
        type='text'
        onClick={() => {
          onCancel();
          setGranularity('Daily');
          setCollapseAll(true);
          setPropSelectOpen(false);
        }}
        icon={<SVG name='times' />}
      />
    </div>
  );

  const listLeftPaneProps = (props = []) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.account_config
        ?.leftpane_props || [];
    showProps.forEach((prop, index) => {
      const propType = getPropType(listProperties, prop);
      const propDisplayName = groupPropNames[prop]
        ? groupPropNames[prop]
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

  const generateAccountProps = () => {
    const groupProps = [
      { label: 'Account Properties', icon: 'users', values: [] }
    ];
    groupProps[0].values = listProperties;
    return groupProps;
  };

  const selectProps = () =>
    propSelectOpen && (
      <div className='relative'>
        <GroupSelect2
          groupedProperties={generateAccountProps()}
          placeholder='Select Property'
          optionClick={handleOptionClick}
          onClickOutside={() => setPropSelectOpen(false)}
        />
      </div>
    );

  const renderAddNewProp = () =>
    !currentProjectSettings?.timelines_config?.account_config?.leftpane_props ||
    currentProjectSettings?.timelines_config?.account_config?.leftpane_props
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
        <img
          src={`https://logo.clearbit.com/${getHost(
            accountDetails?.data?.host
          )}`}
          onError={(e) => {
            if (
              e.target.src !==
              'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg'
            ) {
              e.target.src =
                'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg';
            }
          }}
          alt=''
          height={96}
          width={96}
        />
        <Text type='title' level={6} extraClass='m-0 py-2' weight='bold'>
          {accountDetails?.data?.name}
        </Text>
      </div>
      <div className='leftpane__props'>
        {listLeftPaneProps(accountDetails.data.left_pane_props)}
      </div>
      <div className='px-8 pb-8 pt-2'>{renderAddNewProp()}</div>
      <div className='leftpane__logo_attr'>
        <a
          className='font-size--small'
          href='https://clearbit.com'
          target='_blank'
        >
          Brand Logo provided by Clearbit
        </a>
      </div>
    </div>
  );

  const renderSingleTimelineView = () => (
    <AccountTimelineSingleView
      timelineEvents={
        activities?.filter((activity) => activity.enabled === true) || []
      }
      timelineUsers={accountDetails.data?.account_users || []}
      milestones={accountDetails.data?.milestones}
      loading={accountDetails?.isLoading}
      eventNamesMap={eventNamesMap}
      listProperties={[...listProperties, ...userProperties]}
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
              onClick={() => setCollapseAll(false)}
            >
              <SVG name='line_height' size={22} />
            </Button>
            <Button
              className='collapse-btn collapse-btn--right'
              type='text'
              onClick={() => setCollapseAll(true)}
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
      <AccountTimelineBirdView
        timelineEvents={
          activities?.filter((activity) => activity.enabled === true) || []
        }
        timelineUsers={accountDetails.data?.account_users || []}
        milestones={accountDetails.data?.milestones}
        collapseAll={collapseAll}
        setCollapseAll={setCollapseAll}
        granularity={granularity}
        loading={accountDetails?.isLoading}
        eventNamesMap={eventNamesMap}
        listProperties={[...listProperties, ...userProperties]}
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
  accountDetails: state.timelines.accountDetails,
  userProperties: state.coreQuery.userProperties,
  groupProperties: state.coreQuery.groupProperties,
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccountDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountDetails);
