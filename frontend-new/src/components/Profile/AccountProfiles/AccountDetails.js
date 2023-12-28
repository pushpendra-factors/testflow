import React, { useCallback, useState, useEffect, useMemo } from 'react';
import { Button, Dropdown, Menu, notification, Popover, Tabs } from 'antd';
import styles from './index.module.scss';
import { bindActionCreators } from 'redux';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Text, SVG } from '../../factorsComponents';
import AccountTimelineBirdView from './AccountTimelineBirdView';
import useKey from 'hooks/useKey';
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
import {
  getAccountOverview,
  getProfileAccountDetails
} from '../../../reducers/timelines/middleware';
import {
  addEnabledFlagToActivities,
  formatUserPropertiesToCheckList
} from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../MyComponents/LeftPanePropBlock';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped,
  processProperties
} from 'Utils/dataFormatter';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from '../../../routes/pathUrls';
import UpgradeModal from '../UpgradeModal';
import { FEATURES, PLANS, PLANS_V0 } from 'Constants/plans.constants';
import {
  getGroupProperties,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import AccountOverview from './AccountOverview';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { featureLock } from '../../../routes/feature';
import getGroupIcon from 'Utils/getGroupIcon';
import useFeatureLock from 'hooks/useFeatureLock';
import { getGroups } from 'Reducers/coreQuery/middleware';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { ArrowLeftOutlined } from '@ant-design/icons';
import AccountTimelineTableView from './AccountTimelineTableView';
import { isValidURL } from 'Utils/truncateURL';

function AccountDetails({
  accounts,
  accountDetails,
  accountOverview,
  activeProject,
  getGroups,
  groups,
  getGroupProperties,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileAccountDetails,
  getAccountOverview,
  userPropertiesV2,
  groupProperties,
  eventNamesMap,
  eventPropertiesV2,
  getEventPropertiesV2
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
  const [timelineViewMode, setTimelineViewMode] = useState('birdview');
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [openPopover, setOpenPopover] = useState(false);
  const [requestedEvents, setRequestedEvents] = useState({});
  const [eventPropertiesType, setEventPropertiesType] = useState({});

  const handleOpenPopoverChange = (value) => {
    setOpenPopover(value);
  };

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const { plan } = useSelector((state) => state.featureConfig);
  const isFreePlan =
    plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0.PLAN_FREE;
  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;

  useEffect(() => {
    if (featureLock(activeAgent)) {
      setTimelineViewMode('overview');
    }
  }, [activeAgent]);

  const uniqueEventNames = useMemo(() => {
    const accountEvents = accountDetails.data?.account_events || [];

    const eventsArray = accountEvents
      .filter(
        (event) =>
          !isValidURL(event.event_name) &&
          Object.keys(event?.properties || {}).length
      )
      .map((event) => event.event_name);

    const pageViewEvent = accountEvents.find(
      (event) => event?.properties?.['$is_page_view']
    );

    if (pageViewEvent) {
      eventsArray.push(pageViewEvent.event_name);
    }

    return Array.from(new Set(eventsArray));
  }, [accountDetails.data?.account_events]);

  useEffect(() => {
    const fetchData = async () => {
      const promises = uniqueEventNames.map(async (eventName) => {
        if (!requestedEvents[eventName]) {
          setRequestedEvents((prevRequestedEvents) => ({
            ...prevRequestedEvents,
            [eventName]: true
          }));
          if (!eventPropertiesV2[eventName])
            await getEventPropertiesV2(activeProject?.id, eventName);
        }
      });

      await Promise.all(promises);

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

    fetchData();
  }, [uniqueEventNames, requestedEvents, activeProject?.id, eventPropertiesV2]);

  const titleIcon = useMemo(() => {
    if (Boolean(location?.state?.activeSegment?.id)) {
      return defaultSegmentIconsMapping[location?.state?.activeSegment?.name]
        ? defaultSegmentIconsMapping[location?.state?.activeSegment?.name]
        : 'pieChart';
    }
    return 'buildings';
  }, [location]);

  const pageTitle = useMemo(() => {
    if (location?.state?.activeSegment?.name) {
      return location?.state?.activeSegment?.name;
    }
    return 'All Accounts';
  }, [location]);

  useEffect(() => {
    if (!groups || Object.keys(groups).length === 0) {
      getGroups(activeProject?.id);
    }
  }, [activeProject?.id, groups]);

  useEffect(() => {
    return () => {
      setGranularity('Daily');
      setCollapseAll(true);
      setPropSelectOpen(false);
    };
  }, []);

  const [activeId, activeGroup, activeView] = useMemo(() => {
    const urlSearchParams = new URLSearchParams(location.search);
    const params = Object.fromEntries(urlSearchParams.entries());
    const id = atob(location.pathname.split('/').pop());
    const group = params.group ? params.group : GROUP_NAME_DOMAINS;
    const view = params.view ? params.view : timelineViewMode;
    document.title = 'Accounts - FactorsAI';
    return [id, group, view];
  }, [location, timelineViewMode]);

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
    Object.keys(groups?.account_groups || {}).forEach((group) => {
      if (!groupProperties[group]) {
        getGroupProperties(activeProject?.id, group);
      }
    });
  }, [activeProject.id, groups]);

  useEffect(() => {
    const mergedProps = [];
    const filterProps = {};

    for (const group of Object.keys(groups?.account_groups || {})) {
      const values = groupProperties?.[group] || [];
      mergedProps.push(...values);
      filterProps[group] = values;
    }

    const groupProps = Object.entries(filterProps)
      .map(([group, values]) => ({
        label: groups?.account_groups?.[group] || PropTextFormat(group),
        iconName: group,
        values: processProperties(values)
      }))
      .map((opt) => ({
        iconName: getGroupIcon(opt.iconName),
        label: opt.label,
        values: opt.values
      }));

    setListProperties(mergedProps);
    setFilterProperties(groupProps);
  }, [groupProperties, groups]);

  useEffect(() => {
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userPropertiesModified,
      [currentProjectSettings.timelines_config?.account_config?.user_prop]
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [currentProjectSettings, userPropertiesV2]);

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
    handleOpenPopoverChange(false);
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
    handleOpenPopoverChange(false);
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
    handleOpenPopoverChange(false);
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
    if (!timelinesConfig.account_config.table_props) {
      timelinesConfig.account_config.table_props = [];
    }

    if (!timelinesConfig.account_config.table_props.includes(option?.value)) {
      timelinesConfig.account_config.table_props.push(option?.value);
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

  const handleOptionBackClick = useCallback(() => {
    history.replace(PathUrls.ProfileAccounts, {
      activeSegment: location.state?.activeSegment,
      fromDetails: true,
      accountPayload: location.state?.accountPayload,
      currentPage: location.state?.currentPage,
      currentPageSize: location.state?.currentPageSize,
      activeSorter: location.state?.activeSorter
    });
  }, []);

  const onDelete = (option) => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.account_config.table_props.splice(
      timelinesConfig.account_config.table_props.indexOf(option),
      1
    );
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const renderModalHeader = () => {
    const accountName = accountDetails?.data?.name;
    return (
      <div className='fa-timeline--header'>
        <div className='flex items-center'>
          <div
            className='flex items-center cursor-pointer'
            onClick={handleOptionBackClick}
          >
            <div className='flex items-center rounded justify-center mr-1'>
              <SVG name={titleIcon} size={32} color='#FF4D4F' />
            </div>
            <Text
              type='title'
              level={6}
              weight='bold'
              extraClass='m-0 underline'
            >
              {pageTitle}
            </Text>
          </div>
          {accountName && (
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              {'\u00A0/ ' + accountName}
            </Text>
          )}
        </div>
        <Button size='large' onClick={handleOptionBackClick}>
          Close
        </Button>
      </div>
    );
  };

  const listLeftPaneProps = (props = {}) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.account_config?.table_props?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      ) || [];
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
    !currentProjectSettings?.timelines_config?.account_config?.table_props ||
    currentProjectSettings?.timelines_config?.account_config?.table_props
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
          <a
            className='flex items-center'
            href={`https://${encodeURIComponent(accountDetails?.data?.name)}`}
            target='_blank'
            rel='noopener noreferrer'
          >
            <Text
              type='title'
              level={6}
              extraClass='m-0 mr-1 py-2'
              weight='bold'
            >
              {accountDetails?.data?.name}
            </Text>
            <SVG name='ArrowUpRightSquare' />
          </a>
        </div>
      </div>

      <div className='props'>
        {listLeftPaneProps(accountDetails.data.leftpane_props)}
      </div>
      <div className='add-prop-btn with-attr'>{renderAddNewProp()}</div>
      <div className='logo_attr'>
        <a
          className='font-size--small'
          href='https://clearbit.com'
          target='_blank'
          rel='noreferrer'
        >
          Brand Logo provided by Clearbit
        </a>
      </div>
    </div>
  );

  const renderOverview = () => (
    <AccountOverview
      overview={accountOverview?.data || {}}
      loading={accountOverview?.isLoading}
    />
  );

  const renderSingleTimelineView = () => (
    <AccountTimelineTableView
      timelineEvents={getFilteredEvents(
        activities
          ?.filter((activity) => activity.enabled === true)
          .slice(0, 1000) || []
      )}
      timelineUsers={getTimelineUsers()}
      loading={accountDetails?.isLoading}
      eventPropsType={eventPropertiesType}
    />
  );

  const getFilteredEvents = (events) => {
    if (isFreePlan) {
      return events.filter((activity) => !activity.isGroupEvent);
    }
    return events;
  };

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
        timelineEvents={getFilteredEvents(
          activities?.filter((activity) => activity.enabled === true) || []
        )}
        timelineUsers={getTimelineUsers()}
        collapseAll={collapseAll}
        setCollapseAll={setCollapseAll}
        granularity={granularity}
        loading={accountDetails?.isLoading}
        propertiesType={eventPropertiesType}
        eventNamesMap={eventNamesMap}
      />
    </div>
  );

  useEffect(() => {
    if (
      timelineViewMode === 'overview' &&
      activeId !== accountOverview?.data?.id
    ) {
      getAccountOverview(activeProject.id, activeGroup, activeId);
    }
  }, [timelineViewMode, activeId]);

  const handleTabChange = (val) => {
    insertUrlParam(window.history, 'view', val);
    setTimelineViewMode(val);
    setGranularity(granularity);
  };

  const renderTabPane = ({ key, tabName, content }) => (
    <TabPane
      tab={<span className='fa-activity-filter--tabname'>{tabName}</span>}
      key={key}
    >
      {content}
    </TabPane>
  );

  const renderTimelineView = () => {
    return (
      <div className='timeline-view'>
        <Tabs
          className='timeline-view--tabs'
          defaultActiveKey='birdview'
          size='small'
          activeKey={timelineViewMode}
          onChange={handleTabChange}
        >
          {!isScoringLocked &&
            renderTabPane({
              key: 'overview',
              tabName: 'Overview',
              content: renderOverview()
            })}
          {renderTabPane({
            key: 'timeline',
            tabName: 'Timeline',
            content: renderSingleTimelineView()
          })}
          {renderTabPane({
            key: 'birdview',
            tabName: 'Birdview',
            content: renderBirdviewWithActions()
          })}
        </Tabs>
      </div>
    );
  };

  useKey(['Escape'], handleOptionBackClick);

  return (
    <div>
      <div className='fa-timeline'>
        {renderModalHeader()}
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
  groups: state.coreQuery.groups,
  accounts: state.timelines.accounts,
  accountDetails: state.timelines.accountDetails,
  accountOverview: state.timelines.accountOverview,
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  groupProperties: state.coreQuery.groupProperties,
  eventNamesMap: state.coreQuery.eventNamesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getGroups,
      getGroupProperties,
      getAccountOverview,
      getEventPropertiesV2,
      getProfileAccountDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountDetails);
