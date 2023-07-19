import React, { useState, useEffect, useMemo } from 'react';
import { Button, Dropdown, Menu, notification, Popover, Tabs } from 'antd';
import styles from './index.module.scss';
import { bindActionCreators } from 'redux';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Text, SVG } from '../../factorsComponents';
import AccountTimelineBirdView from './AccountTimelineBirdView';
import {
  DEFAULT_TIMELINE_CONFIG,
  getHost,
  getPropType,
  granularityOptions,
  hoverEvents,
  TIMELINE_VIEW_OPTIONS
} from '../utils';
import { insertUrlParam } from 'Utils/global';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileAccountDetails } from '../../../reducers/timelines/middleware';
import {
  addEnabledFlagToActivities,
  formatUserPropertiesToCheckList
} from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../MyComponents/LeftPanePropBlock';
import AccountTimelineSingleView from './AccountTimelineSingleView';
import { PropTextFormat } from 'Utils/dataFormatter';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from '../../../routes/pathUrls';
import { fetchGroups } from 'Reducers/coreQuery/services';
import UpgradeModal from '../UpgradeModal';
import { PLANS } from 'Constants/plans.constants';
import {
  getGroupProperties,
  getEventProperties
} from 'Reducers/coreQuery/middleware';
import AccountOverview from './AccountOverview';
import GroupSelect from 'Components/GenericComponents/GroupSelect';

function AccountDetails({
  accounts,
  accountDetails,
  activeProject,
  fetchGroups,
  groupOpts,
  getGroupProperties,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileAccountDetails,
  userProperties,
  groupProperties,
  eventNamesMap,
  eventProperties,
  getEventProperties
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const [granularity, setGranularity] = useState('Daily');
  const [collapseAll, setCollapseAll] = useState(true);
  const [activities, setActivities] = useState([]);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [listProperties, setListProperties] = useState([]);
  const [checkListMilestones, setCheckListMilestones] = useState([]);
  const [filterProperties, setFilterProperties] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const { TabPane } = Tabs;
  const [timelineViewMode, setTimelineViewMode] = useState('overview');
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [openPopover, setOpenPopover] = useState(false);
  const handleOpenPopoverChange = (value) => {
    setOpenPopover(value);
  };

  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const { plan } = useSelector((state) => state.featureConfig);
  const isFreePlan = plan?.name === PLANS.PLAN_FREE;

  useEffect(() => {
    fetchGroups(activeProject?.id, true);
  }, [activeProject?.id]);

  useEffect(() => {
    return () => {
      setGranularity('Daily');
      setCollapseAll(true);
      setPropSelectOpen(false);
    };
  }, []);

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  const [activeId, activeGroup, activeView] = useMemo(() => {
    const urlSearchParams = new URLSearchParams(location.search);
    const params = Object.fromEntries(urlSearchParams.entries());
    const id = atob(location.pathname.split('/').pop());
    const group = params.group ? params.group : 'All';
    const view = params.view ? params.view : 'birdview';
    document.title = 'Accounts - FactorsAI';
    return [id, group, view];
  }, [location]);

  useEffect(() => {
    if (activeId && activeId !== '')
      getProfileAccountDetails(
        activeProject.id,
        activeId,
        activeGroup,
        currentProjectSettings?.timelines_config
      );
    if (activeView && TIMELINE_VIEW_OPTIONS.includes(activeView)) {
      setTimelineViewMode(activeView);
    }
  }, [
    activeProject.id,
    activeId,
    activeGroup,
    activeView,
    currentProjectSettings?.timelines_config
  ]);

  useEffect(() => {
    const listDatetimeProperties = listProperties.filter(
      (item) => item[2] === 'datetime'
    );
    const propsWithEnableKey = formatUserPropertiesToCheckList(
      listDatetimeProperties,
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
    const listActivities = addEnabledFlagToActivities(
      accountDetails.data?.account_events,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, accountDetails]);

  useEffect(() => {
    hoverEvents.forEach((event) => {
      if (
        !eventProperties[event] &&
        accountDetails.data?.account_events?.some(
          (activity) => activity?.event_name === event
        )
      ) {
        getEventProperties(activeProject?.id, event);
      }
    });
  }, [activeProject?.id, eventProperties, accountDetails.data?.account_events]);

  useEffect(() => {
    Object.keys(groupOpts || {}).forEach((group) =>
      getGroupProperties(activeProject.id, group)
    );
  }, [activeProject.id, groupOpts]);

  useEffect(() => {
    const mergedProps = [];
    const filterProps = {};

    Object.keys(groupOpts || {}).forEach((group) => {
      mergedProps.push(...(groupProperties?.[group] || []));
      filterProps[group] = groupProperties?.[group];
    });

    let groupProps = Object.entries(filterProps).map(([group, values]) => ({
      label: `${PropTextFormat(group)} Properties`,
      iconName: group,
      values
    }));
    groupProps = groupProps?.map((opt) => {
      return {
        iconName: opt?.iconName,
        label: opt?.label,
        values: opt?.values?.map((op) => {
          return {
            value: op?.[1],
            label: op?.[0],
            extraProps: {
              valueType: op?.[2]
            }
          };
        })
      };
    });
    setListProperties(mergedProps);
    setFilterProperties(groupProps);
  }, [groupProperties, groupOpts]);

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
          activeId,
          activeGroup,
          currentProjectSettings?.timelines_config
        )
      );
    }
    setOpenPopover(false);
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
    timelinesConfig.account_config.milestones = checkListMilestones
      .filter((item) => item.enabled === true)
      .map((item) => item?.prop_name);
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    }).then(() =>
      getProfileAccountDetails(
        activeProject.id,
        activeId,
        activeGroup,
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
          placeholder='Select Events to Show'
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
          placeholder='Select a User Property'
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
          placeholder='Select Up To 5 Milestones'
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

  const handleOptionClick = (option, group) => {
    const timelinesConfig = { ...tlConfig };
    if (
      !timelinesConfig.account_config.leftpane_props.includes(option?.value)
    ) {
      timelinesConfig.account_config.leftpane_props.push(option?.value);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      }).then(() =>
        getProfileAccountDetails(
          activeProject.id,
          activeId,
          activeGroup,
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
            history.replace(PathUrls.ProfileAccounts, {
              activeSegment: location.state?.activeSegment,
              fromDetails: location.state?.fromDetails,
              accountPayload: location.state?.accountPayload,
              currentPage: location.state?.currentPage
            });
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
          history.replace(PathUrls.ProfileAccounts, {
            activeSegment: location.state?.activeSegment,
            fromDetails: location.state?.fromDetails,
            accountPayload: location.state?.accountPayload,
            currentPage: location.state?.currentPage
          });
        }}
        icon={<SVG name='times' />}
      />
    </div>
  );

  const listLeftPaneProps = (props = {}) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.account_config
        ?.leftpane_props || [];
    showProps.forEach((prop, index) => {
      const propType = getPropType(listProperties, prop);
      const propDisplayName = groupPropNames[prop]
        ? groupPropNames[prop]
        : PropTextFormat(prop);
      const value = props[prop];
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
      <div className={styles.account_profiles__event_selector}>
        <GroupSelect
          options={filterProperties}
          searchPlaceHolder='Select Property'
          optionClickCallback={handleOptionClick}
          onClickOutside={() => setPropSelectOpen(false)}
          allowSearchTextSelection={false}
          extraClass={styles.account_profiles__event_selector__select}
          allowSearch={true}
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
      <div className='header'>
        <div className='user'>
          <img
            src={`https://logo.uplead.com/${getHost(
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
      </div>

      <div className='props'>
        {listLeftPaneProps(accountDetails.data.left_pane_props)}
        <div className='px-8 pb-8 pt-2'>{renderAddNewProp()}</div>
      </div>
      <div className='logo_attr'>
        <a
          className='font-size--small'
          href='https://www.uplead.com'
          target='_blank'
        >
          Brand Logo provided by UpLead
        </a>
      </div>
    </div>
  );

  // temp hack for engagement
  const formatOverview = useMemo(() => {
    const account = accounts?.data?.find((item) => item?.identity === activeId);
    const { data: { overview } = {} } = accountDetails;
    const formattedOverview = { ...overview, engagement: account?.engagement };
    return formattedOverview;
  }, [accounts, accountDetails, activeId]);

  const renderOverview = () => (
    <AccountOverview
      overview={formatOverview || {}}
      loading={accountDetails?.isLoading}
    />
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

  const getTimelineUsers = () => {
    const timelineUsers = accountDetails.data?.account_users || [];
    if (isFreePlan) {
      return timelineUsers.filter(
        (userConfig) => userConfig?.title !== 'group_user'
      );
    }
    return timelineUsers;
  };

  const renderBirdviewWithActions = () => (
    <div className='flex flex-col'>
      <div
        className={`timeline-actions ${
          isFreePlan ? 'justify-between' : 'flex-row-reverse'
        } `}
      >
        {isFreePlan && (
          <div className='flex items-baseline flex-wrap'>
            <Text
              type={'paragraph'}
              mini
              color='character-primary'
              extraClass='inline-block'
            >
              LinkedIn ads engagement and G2 intent data is not available in
              free plan. To unlock,
              <span
                className='inline-block cursor-pointer ml-1'
                onClick={() => setIsUpgradeModalVisible(true)}
              >
                <Text type={'paragraph'} mini color='brand-color-6'>
                  {'  '} Upgrade plan
                </Text>
              </span>
            </Text>
          </div>
        )}
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
      <AccountTimelineBirdView
        timelineEvents={
          activities?.filter((activity) => activity.enabled === true) || []
        }
        timelineUsers={getTimelineUsers()}
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
          defaultActiveKey='overview'
          size='small'
          activeKey={timelineViewMode}
          onChange={(val) => {
            insertUrlParam(window.history, 'view', val);
            setTimelineViewMode(val);
            setGranularity(granularity);
          }}
        >
          <TabPane
            tab={<span className='fa-activity-filter--tabname'>Overview</span>}
            key='overview'
          >
            {renderOverview()}
          </TabPane>
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
      <UpgradeModal
        visible={isUpgradeModalVisible}
        variant='timeline'
        onCancel={() => setIsUpgradeModalVisible(false)}
      />
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  groupOpts: state.groups.data,
  accounts: state.timelines.accounts,
  accountDetails: state.timelines.accountDetails,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  groupProperties: state.coreQuery.groupProperties,
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchGroups,
      getGroupProperties,
      getEventProperties,
      getProfileAccountDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountDetails);
