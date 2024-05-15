import React, { useCallback, useState, useEffect, useMemo } from 'react';
import {
  Button,
  Dropdown,
  Menu,
  message,
  notification,
  Popover,
  Tabs
} from 'antd';
import { bindActionCreators } from 'redux';
import { connect, useSelector } from 'react-redux';
import { insertUrlParam } from 'Utils/global';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped,
  processProperties
} from 'Utils/dataFormatter';
import { useHistory, useLocation } from 'react-router-dom';
import { FEATURES, PLANS, PLANS_V0 } from 'Constants/plans.constants';
import {
  getGroupProperties,
  getEventPropertiesV2,
  getGroups,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import useFeatureLock from 'hooks/useFeatureLock';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import logger from 'Utils/logger';
import EmptyScreen from 'Components/EmptyScreen';
import AccountOverview from './AccountOverview';
import UpgradeModal from '../UpgradeModal';
import { PathUrls } from '../../../routes/pathUrls';
import LeftPanePropBlock from '../MyComponents/LeftPanePropBlock';
import SearchCheckList from '../../SearchCheckList';
import {
  addEnabledFlagToActivities,
  formatUserPropertiesToCheckList
} from '../../../reducers/timelines/utils';
import {
  getAccountOverview,
  getProfileAccountDetails,
  setActivePageviewEvent
} from '../../../reducers/timelines/middleware';
import { udpateProjectSettings } from '../../../reducers/global';
import {
  eventsFormattedForGranularity,
  flattenObjects,
  getHost,
  getPropType
} from '../utils';
import AccountTimelineBirdView from './AccountTimelineBirdView';
import { Text, SVG } from '../../factorsComponents';
import styles from './index.module.scss';
import AccountTimelineTableView from './AccountTimelineTableView';
import { GranularityOptions } from '../constants';
import AccountsOverviewUpgrade from '../../../assets/images/illustrations/AccountsOverviewUpgrade.png';

function AccountDetails({
  getGroups,
  getGroupProperties,
  udpateProjectSettings,
  getProfileAccountDetails,
  getAccountOverview,
  getEventPropertiesV2,
  setActivePageviewEvent,
  getUserPropertiesV2
}) {
  const { TabPane } = Tabs;

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
  const [tlConfig, setTLConfig] = useState({});
  const [timelineViewMode, setTimelineViewMode] = useState('');
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [openPopover, setOpenPopover] = useState(false);
  const [requestedEvents, setRequestedEvents] = useState({});
  const [eventPropertiesType, setEventPropertiesType] = useState({});
  const [userPropertiesType, setUserPropertiesType] = useState({});
  const [eventDrawerVisible, setEventDrawerVisible] = useState(false);
  const [birdviewFormatEvents, setBirdviewFormatEvents] = useState({});

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const yourParam = params.get('view');
    setTimelineViewMode(yourParam || 'timeline');
  }, [location]);

  const handleOptionBackClick = useCallback(() => {
    const path = location.state?.path || PathUrls.ProfileAccounts;
    history.replace(path, {
      fromDetails: true,
      accountPayload: location.state?.accountPayload,
      currentPage: location.state?.currentPage,
      currentPageSize: location.state?.currentPageSize,
      activeSorter: location.state?.activeSorter,
      appliedFilters: location.state?.appliedFilters,
      accountsTableRow: location.state?.accountsTableRow
    });
  }, []);

  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === 'Escape' && eventDrawerVisible) {
        setEventDrawerVisible(false);
      } else if (event.key === 'Escape' && !eventDrawerVisible) {
        handleOptionBackClick();
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [eventDrawerVisible]);

  const handleOpenPopoverChange = (value) => {
    setOpenPopover(value);
  };

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const { accountDetails, accountOverview } = useSelector(
    (state) => state.timelines
  );
  const { active_project: activeProject, currentProjectSettings } = useSelector(
    (state) => state.global
  );
  const {
    groups,
    groupProperties,
    userPropertiesV2,
    eventPropertiesV2,
    groupPropNames,
    eventNamesMap
  } = useSelector((state) => state.coreQuery);

  const { plan } = useSelector((state) => state.featureConfig);
  const isFreePlan =
    plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0.PLAN_FREE;

  const uniqueEventNames = useMemo(() => {
    const accountEvents = accountDetails.data?.events || [];
    const eventsArray = accountEvents
      .filter((event) => event.display_name !== 'Page View')
      .map((event) => event.name);

    const pageViewEvent = accountEvents.find(
      (event) => event.display_name === 'Page View'
    );

    if (pageViewEvent) {
      eventsArray.push(pageViewEvent.name);
      setActivePageviewEvent(pageViewEvent.name);
    }

    return Array.from(new Set(eventsArray));
  }, [accountDetails.data?.events]);

  const fetchEventPropertiesWithType = async () => {
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

    await Promise.allSettled(promises);

    const typeMap = {};
    Object.values(eventPropertiesV2).forEach((propertyGroup) => {
      Object.values(propertyGroup || {}).forEach((arr) => {
        arr.forEach((property) => {
          const [, propName, category] = property;
          typeMap[propName] = category;
        });
      });
    });
    setEventPropertiesType(typeMap);
  };

  useEffect(() => {
    fetchEventPropertiesWithType();
  }, [uniqueEventNames, requestedEvents, activeProject?.id, eventPropertiesV2]);

  useEffect(() => {
    if (!userPropertiesV2) {
      getUserPropertiesV2(activeProject?.id);
    } else {
      const typeMap = {};
      Object.values(userPropertiesV2).forEach((arr) => {
        arr.forEach(([, propName, category]) => {
          typeMap[propName] = category;
        });
      });
      setUserPropertiesType(typeMap);
    }
  }, [userPropertiesV2, activeProject?.id]);

  const titleIcon = useMemo(() => {
    if (location?.state?.accountPayload?.segment?.id) {
      return defaultSegmentIconsMapping[
        location?.state?.accountPayload?.segment?.name
      ]
        ? defaultSegmentIconsMapping[
            location?.state?.accountPayload?.segment?.name
          ]
        : 'pieChart';
    }
    return 'buildings';
  }, [location]);

  const pageTitle = useMemo(() => {
    if (location?.state?.accountPayload?.segment?.name) {
      return location?.state?.accountPayload?.segment?.name;
    }
    return 'All Accounts';
  }, [location]);

  const activeId = useMemo(() => {
    const id = atob(location.pathname.split('/').pop());
    document.title = 'Accounts - FactorsAI';
    return id;
  }, [location]);

  useEffect(
    () => () => {
      setGranularity('Daily');
      setCollapseAll(true);
      setPropSelectOpen(false);
    },
    []
  );

  useEffect(() => {
    if (timelineViewMode) {
      insertUrlParam(window.history, 'view', timelineViewMode);
    }
  }, [timelineViewMode]);

  const fetchGroups = async () => {
    if (!groups || Object.keys(groups).length === 0) {
      await getGroups(activeProject?.id);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, [activeProject?.id, groups]);

  const getAccountDetails = async () => {
    await getProfileAccountDetails(
      activeProject.id,
      activeId,
      GROUP_NAME_DOMAINS
    );
  };

  useEffect(() => {
    const shouldGetDetails = activeProject?.id && activeId && activeId !== '';
    if (shouldGetDetails) {
      getAccountDetails();
    }
  }, [activeProject.id, activeId]);

  useEffect(() => {
    if (
      timelineViewMode === 'overview' &&
      activeId !== accountOverview?.data?.id
    ) {
      if (!isScoringLocked)
        getAccountOverview(activeProject.id, GROUP_NAME_DOMAINS, activeId);
    }
  }, [timelineViewMode, activeId]);

  useEffect(() => {
    if (!currentProjectSettings?.timelines_config) return;
    setTLConfig(currentProjectSettings.timelines_config);
  }, [currentProjectSettings?.timelines_config]);

  useEffect(() => {
    const listActivities = addEnabledFlagToActivities(
      accountDetails.data?.events,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, accountDetails]);

  const fetchGroupProperties = async () => {
    const missingGroups = Object.keys(groups?.account_groups || {}).filter(
      (group) => !groupProperties[group]
    );

    if (missingGroups.length > 0) {
      await Promise.allSettled(
        missingGroups.forEach((group) =>
          getGroupProperties(activeProject?.id, group)
        )
      );
    }
  };

  useEffect(() => {
    fetchGroupProperties();
  }, [activeProject?.id, groups, groupProperties]);

  useEffect(() => {
    const mergedProps = [];
    const filterProps = {};

    Object.keys(groups?.account_groups || {}).forEach((group) => {
      const values = groupProperties?.[group] || [];
      mergedProps.push(...values);
      filterProps[group] = values;
    });

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

  const handleEventsChange = async (option) => {
    const { timelines_config } = currentProjectSettings;
    const { display_name, enabled } = option;
    const disabledEvents = [...timelines_config.disabled_events];

    if (enabled) {
      disabledEvents.push(display_name);
    } else {
      const indexToRemove = disabledEvents.indexOf(display_name);
      if (indexToRemove !== -1) {
        disabledEvents.splice(indexToRemove, 1);
      }
    }

    const updatedTimelinesConfig = {
      ...timelines_config,
      disabled_events: disabledEvents
    };
    setOpenPopover(false);
    await udpateProjectSettings(activeProject.id, {
      timelines_config: updatedTimelinesConfig
    });
  };

  const handlePropChange = async (option) => {
    const { timelines_config } = currentProjectSettings;

    if (option.prop_name !== timelines_config.account_config.user_prop) {
      const updatedTimelinesConfig = {
        ...timelines_config,
        account_config: {
          ...timelines_config.account_config,
          user_prop: option.prop_name
        }
      };

      try {
        setOpenPopover(false);
        await udpateProjectSettings(activeProject.id, {
          timelines_config: updatedTimelinesConfig
        });
      } catch (error) {
        logger.error(error);
      }
    }
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

  const applyMilestones = async () => {
    const { account_config } = tlConfig;
    const selectedMilestones = checkListMilestones
      .filter((item) => item.enabled === true)
      .map((item) => item?.prop_name);

    const updatedTimelinesConfig = {
      ...tlConfig,
      account_config: {
        ...account_config,
        milestones: selectedMilestones
      }
    };
    try {
      setOpenPopover(false);
      await udpateProjectSettings(activeProject.id, {
        timelines_config: updatedTimelinesConfig
      });
    } catch (error) {
      logger.error(error);
    }
  };

  const controlsPopoverContent = (
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
      {GranularityOptions.map((option) => (
        <Menu.Item key={option} onClick={(key) => setGranularity(key.key)}>
          <div className='flex items-center'>
            <span className='mr-3'>{option}</span>
          </div>
        </Menu.Item>
      ))}
    </Menu>
  );

  const renderHeader = () => {
    const accountName =
      accountDetails?.data?.name || accountDetails?.data?.domain;
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
              {`\u00A0/ ${accountName}`}
            </Text>
          )}
        </div>
        <Button size='large' onClick={handleOptionBackClick}>
          Close
        </Button>
      </div>
    );
  };

  const onLeftpanePropSelect = async (option, group) => {
    const updatedTimelinesConfig = { ...tlConfig };
    if (!updatedTimelinesConfig.account_config.table_props) {
      tlConfig.account_config.table_props = [];
    }

    if (
      !updatedTimelinesConfig.account_config.table_props.includes(option?.value)
    ) {
      updatedTimelinesConfig.account_config.table_props.push(option?.value);
      try {
        setPropSelectOpen(false);
        await udpateProjectSettings(activeProject.id, {
          timelines_config: { ...updatedTimelinesConfig }
        });
      } catch (error) {
        logger.error(error);
      }
    } else {
      message.error('Property Already Exists');
    }
  };

  const onLeftpanePropDelete = async (option) => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.account_config.table_props.splice(
      timelinesConfig.account_config.table_props.indexOf(option),
      1
    );
    await udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const listLeftPaneProps = (props = {}) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.account_config?.table_props?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      ) || [];
    showProps.forEach((prop) => {
      const propType = getPropType(listProperties, prop);
      const mergedGroupPropNames = flattenObjects(groupPropNames);
      const propDisplayName = mergedGroupPropNames[prop]
        ? mergedGroupPropNames[prop]
        : PropTextFormat(prop);
      const value = props[prop];
      propsList.push(
        <LeftPanePropBlock
          property={prop}
          type={propType}
          displayName={propDisplayName}
          value={value}
          onDelete={onLeftpanePropDelete}
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
          optionClickCallback={onLeftpanePropSelect}
          onClickOutside={() => setPropSelectOpen(false)}
          allowSearchTextSelection={false}
          extraClass={styles.account_profiles__event_selector__select}
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
          <img
            src={`https://logo.clearbit.com/${getHost(
              accountDetails?.data?.domain
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
            href={`https://${encodeURIComponent(accountDetails?.data?.domain)}`}
            target='_blank'
            rel='noopener noreferrer'
          >
            <Text
              type='title'
              level={6}
              extraClass='m-0 mr-1 py-2'
              weight='bold'
            >
              {accountDetails?.data?.name || accountDetails?.data?.domain}
            </Text>
            <SVG name='ArrowUpRightSquare' size={16} />
          </a>
        </div>
      </div>

      <div className='props scroll-shadows'>
        {listLeftPaneProps(accountDetails.data.leftpane_props)}
      </div>
      {!currentProjectSettings?.timelines_config?.account_config?.table_props ||
      currentProjectSettings?.timelines_config?.account_config?.table_props
        ?.length < 12 ? (
        <div className='add-prop-btn with-attr'>{renderAddNewProp()}</div>
      ) : null}
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

  const getTimelineUsers = (view = 'timeline') => {
    const timelineUsers = accountDetails.data?.users || [];
    const filteredUsers = [];

    if (view === 'birdview') {
      const knownUsers = timelineUsers.filter(
        (user) => !user?.isAnonymous && user?.name !== 'group_user'
      );
      const anonymousUsers = timelineUsers.filter((user) => user?.isAnonymous);
      const groupedAnonymousUser = anonymousUsers.length
        ? [
            {
              name: 'Anonymous Users',
              extraProp: `${
                anonymousUsers.length === 1
                  ? '1 Anonymous User'
                  : `${anonymousUsers.length}${
                      timelineUsers.length > 25 ? '+' : ''
                    } Anonymous Users`
              }`,
              id: 'new_user',
              isAnonymous: true
            }
          ]
        : [];
      const accountUser = timelineUsers.filter(
        (user) => user?.name === 'group_user'
      );

      filteredUsers.push(
        ...knownUsers,
        ...groupedAnonymousUser,
        ...accountUser
      );
    } else {
      filteredUsers.push(...timelineUsers);
    }

    if (isFreePlan) {
      return filteredUsers.filter((user) => user?.username !== 'group_user');
    }

    return filteredUsers;
  };

  const getFilteredEvents = (events) => {
    if (isFreePlan) {
      return events.filter((activity) => !activity.is_group_user);
    }
    return events;
  };

  const renderOverview = () => (
    <AccountOverview
      overview={accountOverview?.data || {}}
      loading={accountOverview?.isLoading}
      top_engagement_signals={
        accountDetails?.data?.leftpane_props?.$top_enagagement_signals
      }
    />
  );

  const renderTimelineTableView = () => (
    <>
      <div className='h-6' />
      <AccountTimelineTableView
        timelineEvents={getFilteredEvents(
          activities
            ?.filter((activity) => activity.enabled === true)
            ?.slice(0, 1000) || []
        )}
        loading={accountDetails?.isLoading}
        eventPropsType={eventPropertiesType}
        userPropsType={userPropertiesType}
        eventDrawerVisible={eventDrawerVisible}
        setEventDrawerVisible={setEventDrawerVisible}
      />
    </>
  );

  useEffect(() => {
    if (!activities) return;
    const events = getFilteredEvents(
      activities?.filter((activity) => activity.enabled === true) || []
    );
    const data = eventsFormattedForGranularity(
      events,
      granularity,
      collapseAll
    );
    document.title = 'Accounts - FactorsAI';
    setBirdviewFormatEvents(data);
  }, [activities, granularity]);

  useEffect(() => {
    const data = Object.keys(birdviewFormatEvents).reduce((acc, key) => {
      acc[key] = Object.keys(birdviewFormatEvents[key]).reduce(
        (userAcc, username) => {
          userAcc[username] = {
            ...birdviewFormatEvents[key][username],
            collapsed:
              collapseAll === undefined
                ? birdviewFormatEvents[key][username].collapsed
                : collapseAll
          };
          return userAcc;
        },
        {}
      );
      return acc;
    }, {});
    setBirdviewFormatEvents(data);
  }, [collapseAll]);

  const renderBirdviewWithActions = () => (
    <div className='flex flex-col'>
      <div className={isFreePlan ? 'flex justify-between items-center' : ''}>
        {isFreePlan && (
          <div className='flex items-baseline flex-wrap'>
            <Text
              type='paragraph'
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
                <Text type='paragraph' mini color='brand-color-6'>
                  Upgrade plan
                </Text>
              </span>
            </Text>
          </div>
        )}

        <div className='tl-actions-row'>
          <div className='collapse-btns'>
            <Button
              className='collapse-btns--btn'
              onClick={() => setCollapseAll(false)}
            >
              <SVG name='line_height' size={22} />
            </Button>
            <Button
              className='collapse-btns--btn'
              onClick={() => setCollapseAll(true)}
            >
              <SVG name='grip_lines' size={22} />
            </Button>
          </div>
          <Popover
            overlayClassName='fa-activity--filter'
            placement='bottomLeft'
            trigger='click'
            content={controlsPopoverContent}
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
            <Button type='text'>
              {granularity}
              <SVG name='caretDown' size={16} extraClass='ml-1' />
            </Button>
          </Dropdown>
        </div>
      </div>
      <AccountTimelineBirdView
        events={birdviewFormatEvents || {}}
        setEvents={setBirdviewFormatEvents}
        timelineUsers={getTimelineUsers('birdview')}
        setCollapseAll={setCollapseAll}
        loading={accountDetails?.isLoading}
        propertiesType={eventPropertiesType}
        eventNamesMap={eventNamesMap}
      />
    </div>
  );

  const handleTabChange = (val) => {
    setTimelineViewMode(val);
  };

  const renderTabPane = ({ key, tabName, content }) => (
    <TabPane
      tab={<span className='fa-activity-filter--tabname'>{tabName}</span>}
      key={key}
    >
      {key === 'overview' && isScoringLocked ? (
        <div className='overview-container'>
          <EmptyScreen
            upgradeScreen
            image={AccountsOverviewUpgrade}
            imageStyle={{ width: '600px', height: '450px' }}
            title={
              <div>
                <Text type='title' level={3} weight='bold' extraClass='m-0'>
                  Your plan doesn’t have this feature
                </Text>
                <Text type='title' level={7} extraClass='m-0'>
                  This feature is not included in your current plan. Please
                  upgrade to use this feature
                </Text>
              </div>
            }
            learnMore='https://www.youtube.com/watch?v=sbgrCYaAnwQ'
            ActionButton={{
              onClick: () => {
                history.push('/settings/pricing?activeTab=upgrade');
              },
              text: 'Upgrade Now',
              icon: null
            }}
          />
        </div>
      ) : (
        content
      )}
    </TabPane>
  );

  const renderTimelineView = () => (
    <div className='timeline-view'>
      <Tabs
        className='timeline-view--tabs'
        defaultActiveKey={timelineViewMode}
        size='small'
        activeKey={timelineViewMode}
        onChange={handleTabChange}
      >
        {renderTabPane({
          key: 'overview',
          tabName: 'Overview',
          content: renderOverview()
        })}
        {renderTabPane({
          key: 'timeline',
          tabName: 'Timeline',
          content: renderTimelineTableView()
        })}
        {renderTabPane({
          key: 'birdview',
          tabName: 'Birdview',
          content: renderBirdviewWithActions()
        })}
      </Tabs>
    </div>
  );

  return (
    <>
      <div className='fa-timeline'>
        {renderHeader()}
        <div className='fa-timeline--content'>
          {renderLeftPane()}
          {renderTimelineView()}
        </div>
      </div>
      <UpgradeModal
        visible={isUpgradeModalVisible}
        variant='timeline'
        onCancel={() => setIsUpgradeModalVisible(false)}
      />
    </>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getGroups,
      getGroupProperties,
      getAccountOverview,
      getEventPropertiesV2,
      getUserPropertiesV2,
      getProfileAccountDetails,
      udpateProjectSettings,
      setActivePageviewEvent
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AccountDetails);
