import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo
} from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  Row,
  Col,
  Select,
  Button,
  Form,
  Input,
  message,
  notification,
  Checkbox,
  Modal,
  Popover,
  Tooltip,
  Avatar,
  Switch,
  Menu,
  Dropdown,
  Alert
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import {
  createEventAlert,
  deleteEventAlert,
  editEventAlert,
  testWebhhookUrl,
  fetchAllAlerts,
  testSlackAlert,
  testTeamsAlert,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels,
  updateEventAlertStatus,
  fetchSlackUsers
} from 'Reducers/global';
import {
  deleteGroupByForEvent,
  setGroupBy,
  delGroupBy,
  getUserPropertiesV2,
  resetGroupBy,
  getGroupProperties,
  getEventPropertiesV2,
  getGroups
} from 'Reducers/coreQuery/middleware';
import {
  QUERY_TYPE_EVENT,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA
} from 'Utils/constants';
import {
  DefaultDateRangeFormat,
  formatBreakdownsForQuery,
  formatFiltersForQuery,
  processBreakdownsFromQuery,
  processFiltersFromQuery
} from 'Views/CoreQuery/utils';
import TextArea from 'antd/lib/input/TextArea';
import useAutoFocus from 'hooks/useAutoFocus';
import GLobalFilter from 'Components/KPIComposer/GlobalFilter';
import _ from 'lodash';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import { setShowCriteria } from 'Reducers/analyticsQuery';
import {
  INITIALIZE_GROUPBY,
  setEventGroupBy,
  setGroupByActionList,
  setGroupByEventActionList
} from 'Reducers/coreQuery/actions';
import { ExclamationCircleOutlined, MoreOutlined } from '@ant-design/icons';
import { useHistory } from 'react-router-dom';
import { ScrollToTop } from 'Routes/feature';
import { ReactSortable } from 'react-sortablejs';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { selectSegments } from 'Reducers/timelines/selectors';
import { reorderDefaultDomainSegmentsToTop } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { getSavedSegments } from 'Reducers/timelines/middleware';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import cx from 'classnames';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import {
  getMsgPayloadMapping,
  dummyPayloadValue,
  convertObjectToKeyValuePairArray
} from '../utils';
import Teams from './Teams';
import Webhook from './Webhook';
import SelectChannels from '../SelectChannels';
import Slack from './Slack';
import QueryBlock from './QueryBlock';
import EventGroupBlock from '../../../../../components/QueryComposer/EventGroupBlock';

const { Option } = Select;

const SegmentIcon = (name) => defaultSegmentIconsMapping[name] || 'pieChart';

const EventBasedAlert = ({
  activeProject,
  deleteEventAlert,
  createEventAlert,
  editEventAlert,
  agent_details,
  slack,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  projectSettings,
  enableSlackIntegration,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels,
  viewAlertDetails,
  alertState,
  setAlertState,
  setGroupBy,
  delGroupBy,
  getUserPropertiesV2,
  groupBy,
  resetGroupBy,
  eventPropertiesV2,
  eventPropNames,
  groupProperties,
  groupPropNames,
  userPropertiesV2,
  userPropNames,
  eventNamesSpecial,
  getGroupProperties,
  getEventPropertiesV2,
  getGroups,
  groups,
  testWebhhookUrl,
  teams,
  updateEventAlertStatus,
  setShowCriteria,
  fetchAllAlerts,
  fetchSlackUsers,
  slack_users,
  testSlackAlert,
  testTeamsAlert,
  getSavedSegments
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [alertName, setAlertName] = useState('');
  const [alertMessage, setAlertMessage] = useState('');
  const [webhookEnabled, setWebhookEnabled] = useState(false);
  const [slackEnabled, setSlackEnabled] = useState(false);
  const [teamsEnabled, setTeamsEnabled] = useState(false);
  const [notRepeat, setNotRepeat] = useState(false);
  const [notifications, setNotifications] = useState(false);
  const [isHyperLinkEnabled, setIsHyperLinkEnabled] = useState(true);
  const [alertLimit, setAlertLimit] = useState(5);
  const [coolDownTime, setCoolDownTime] = useState(0.5);
  const [viewFilter, setViewFilter] = useState([]);
  const [channelOpts, setChannelOpts] = useState([]);
  const [selectedChannel, setSelectedChannel] = useState([]);
  const [saveSelectedChannel, setSaveSelectedChannel] = useState([]);
  const [showSelectChannelsModal, setShowSelectChannelsModal] = useState(false);
  const [viewSelectedChannels, setViewSelectedChannels] = useState([]);

  const [teamsWorkspaceOpts, setTeamsWorkspaceOpts] = useState([]);
  const [selectedWorkspace, setSelectedWorkspace] = useState(null);
  const [teamsChannelOpts, setTeamsChannelOpts] = useState([]);
  const [teamsSelectedChannel, setTeamsSelectedChannel] = useState([]);
  const [teamsSaveSelectedChannel, setTeamsSaveSelectedChannel] = useState([]);
  const [teamsShowSelectChannelsModal, setTeamsShowSelectChannelsModal] =
    useState(false);
  const [teamsViewSelectedChannels, setTeamsViewSelectedChannels] = useState(
    []
  );

  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  // const inputComponentRef = useAutoFocus();
  const [isAlertEnabled, setisAlertEnabled] = useState(false);
  const [enableWidgetModal, showEnableWidgetModal] = useState(false);

  const [webhookUrl, setWebhookUrl] = useState('');
  const [finalWebhookUrl, setFinalWebhookUrl] = useState('');
  const [confirmBtn, setConfirmBtn] = useState(true);
  const [testMessageBtn, setTestMessageBtn] = useState(true);
  const [testMessageResponse, setTestMassageResponse] = useState('');
  const [confirmedMessageBtn, setConfirmedMessageBtn] = useState(false);
  const [showEditBtn, setShowEditBtn] = useState(false);
  const [disbleWebhookInput, setDisbleWebhookInput] = useState(false);
  const [hideTestMessageBtn, setHideTestMessageBtn] = useState(true);
  const [showAdvSettings, setShowAdvSettings] = useState(false);
  const [selectedMentions, setSelectedMentions] = useState([]);

  const [showSlackInt, setShowSlackInt] = useState(false);
  const [showTeamInt, setShowTeamInt] = useState(false);
  const [showWHInt, setShowWHInt] = useState(false);

  const [slackTestMsgLoading, setSlackTestMsgLoading] = useState(false);
  const [slackTestMsgTxt, setSlackTestMsgTxt] = useState(false);
  const [slackMentionLoading, setSlackMentionLoading] = useState(false);

  const [teamsTestMsgLoading, setTeamsTestMsgLoading] = useState(false);
  const [teamsTestMsgTxt, setTeamsTestMsgTxt] = useState(false);

  const [WHTestMsgLoading, setWHTestMsgLoading] = useState(false);
  const [WHTestMsgTxt, setWHTestMsgTxt] = useState(false);
  const [factorsURLinWebhook, setFactorsURLinWebhook] = useState(true);

  const webhookRef = useRef();
  const [form] = Form.useForm();
  const dispatch = useDispatch();
  const { confirm } = Modal;

  // Segment Support
  const [segmentType, setSegmentType] = useState('action_event');
  const [selectedSegment, setSelectedSegment] = useState('');
  const [segmentOptions, setSegmentOptions] = useState([]);
  const segments = useSelector(selectSegments);
  const segmentsList = useMemo(
    () => reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [],
    [segments]
  );

  // Event SELECTION
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });

  const [activeGrpBtn, setActiveGrpBtn] = useState(QUERY_TYPE_EVENT);

  // Webhook support
  const { isFeatureLocked: isWebHookFeatureLocked } = useFeatureLock(
    FEATURES.FEATURE_WEBHOOK
  );

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const getGroupPropsFromAPI = useCallback(
    async (group) => {
      if (!groupProperties[group]) {
        await getGroupProperties(activeProject.id, group);
      }
    },
    [activeProject.id, groupProperties]
  );

  const fetchGroupProperties = async () => {
    // separate call for $domain = All account group.
    getGroupPropsFromAPI(GROUP_NAME_DOMAINS);

    const missingGroups = Object.keys(groups?.all_groups || {}).filter(
      (group) => !groupProperties[group]
    );
    if (missingGroups && missingGroups?.length > 0) {
      await Promise.allSettled(
        missingGroups?.map((group) =>
          getGroupProperties(activeProject?.id, group)
        )
      );
    }
  };

  useEffect(() => {
    fetchGroupProperties();
  }, [activeProject?.id, groups, groupProperties]);

  const fetchGroups = async () => {
    if (!groups || Object.keys(groups).length === 0) {
      await getGroups(activeProject?.id);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, [activeProject?.id, groups]);

  // fetch segments and on Change functions
  useEffect(() => {
    getSavedSegments(activeProject?.id);
  }, [activeProject?.id]);

  const renderOptions = (segment) => {
    const iconColor = getSegmentColorCode(segment?.name);
    const icon = SegmentIcon(segment?.name);
    return (
      <div className={cx('flex col-gap-1 items-center w-full')}>
        <ControlledComponent controller={icon != null}>
          <SVG name={icon} size={20} color={iconColor} />
        </ControlledComponent>
        <div className='m-0 ml-1 truncate'>{segment?.name}</div>
      </div>
    );
  };

  useEffect(() => {
    const segmentListWithLabels = segmentsList.map((segment) => ({
      value: segment?.id,
      label: renderOptions(segment)
    }));
    setSegmentOptions(segmentListWithLabels);
  }, [segmentsList]);

  const getSegmentNameFromId = (Id) => {
    const segmentName = segmentsList.find((segment) => segment?.id === Id);
    if (segmentName) return segmentName?.name;
    return '';
  };

  const onChangeSegmentType = (value) => {
    setSegmentType(value);
  };

  const onChangeSegment = (segment) => {
    setSelectedSegment(segment?.value);
  };

  // useEffect(() => {
  //   if (groups && Object.keys(groups).length != 0) {
  //     Object.keys(groups?.all_groups).forEach((item) => {
  //       getGroupProperties(activeProject.id, item)
  //     });
  //   }
  // }, [activeProject.id, groups]);

  const groupsList = useMemo(() => {
    const listGroups = [];
    Object.entries(groups?.all_groups || {}).forEach(
      ([group_name, display_name]) => {
        listGroups.push([display_name, group_name]);
      }
    );
    return listGroups;
  }, [groups]);

  const setGroupAnalysis = (group) => {
    setActiveGrpBtn(group);

    if (!['users', 'events'].includes(group)) {
      getGroupProperties(activeProject.id, group);
    }

    const criteria =
      group === 'events' ? TOTAL_EVENTS_CRITERIA : TOTAL_USERS_CRITERIA;
    setShowCriteria(criteria);

    const opts = {
      ...queryOptions,
      group_analysis: group,
      globalFilters: []
    };

    dispatch({
      type: INITIALIZE_GROUPBY,
      payload: {
        global: [],
        event: []
      }
    });

    setQueries([]);
    setQueryOptions(opts);
    setSegmentType('action_event');
    setSelectedSegment('');
  };

  const confirmGroupSwitch = (group) => {
    if (queries.length > 0 || segmentType !== 'action_event') {
      Modal.confirm({
        title: 'Are you sure?',
        content:
          'Switching between "Account and People" will lose your current configured data',
        okText: 'Yes, proceed',
        cancelText: 'No, go back',
        onOk: () => {
          setGroupAnalysis(group);
        }
      });
    } else {
      setGroupAnalysis(group);
    }
  };

  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  const [breakdownOptions, setBreakdownOptions] = useState([]);
  const [EventPropertyDetails, setEventPropertyDetails] = useState({});

  useEffect(() => {
    let DDCategory = [];
    for (const property in eventPropertiesV2[queries[0]?.label]) {
      const nestedArrays = eventPropertiesV2[queries[0]?.label][property];
      DDCategory = _.union(nestedArrays, DDCategory);
    }
    if (groups?.all_groups?.[queries[0]?.group]) {
      for (const key of Object.keys(groupProperties)) {
        if (key === queries[0]?.group) {
          DDCategory = _.union(
            DDCategory,
            groupProperties[groups?.all_groups?.pts[queries[0]?.group]]
          );
        }
      }
    } else {
      for (const property in userPropertiesV2) {
        const nestedArrays = userPropertiesV2[property];
        DDCategory = _.union(DDCategory, nestedArrays);
      }
    }
    setBreakdownOptions(DDCategory);
    if (
      alertState?.state === 'edit' &&
      !(EventPropertyDetails?.name || EventPropertyDetails?.[0])
    ) {
      const property = DDCategory.filter(
        (data) =>
          data[1] === viewAlertDetails?.alert?.breakdown_properties?.[0]?.pr
      );
      setEventPropertyDetails(property?.[0]);
    }
  }, [
    queries,
    eventPropertiesV2,
    groupProperties,
    userPropertiesV2,
    viewAlertDetails,
    alertState
  ]);

  const matchEventName = (item) => {
    const findItem =
      eventPropNames?.[item] ||
      userPropNames?.[item] ||
      groupPropNames?.[item] ||
      eventNamesSpecial?.[item];
    return findItem || item;
  };

  useEffect(() => {
    if (viewAlertDetails?.alert?.event) {
      getGroupProperties(activeProject.id, viewAlertDetails?.alert?.event);
    }
    if (viewAlertDetails?.alert?.event) {
      getEventPropertiesV2(activeProject.id, viewAlertDetails?.alert?.event);
    }
  }, [viewAlertDetails?.alert?.event]);

  useEffect(() => {
    if (viewAlertDetails?.alert?.filter) {
      const filter = processFiltersFromQuery(viewAlertDetails?.alert?.filter);
      setViewFilter(filter);
    }
    if (viewAlertDetails?.alert?.slack_channels) {
      setViewSelectedChannels(viewAlertDetails?.alert?.slack_channels);
      if (alertState?.state === 'edit') {
        setSlackEnabled(viewAlertDetails?.alert?.slack);
        setSaveSelectedChannel(viewAlertDetails?.alert?.slack_channels);
        setSelectedChannel(viewAlertDetails?.alert?.slack_channels);
      }
    }
    if (viewAlertDetails?.alert?.teams_channels_config?.team_channel_list) {
      setTeamsViewSelectedChannels(
        viewAlertDetails?.alert?.teams_channels_config?.team_channel_list
      );
      if (alertState?.state === 'edit') {
        setTeamsEnabled(viewAlertDetails?.alert?.teams);
        setTeamsSaveSelectedChannel(
          viewAlertDetails?.alert?.teams_channels_config?.team_channel_list
        );
        setTeamsSelectedChannel(
          viewAlertDetails?.alert?.teams_channels_config?.team_channel_list
        );
        setSelectedWorkspace({
          name: viewAlertDetails?.alert?.teams_channels_config?.team_name,
          id: viewAlertDetails?.alert?.teams_channels_config?.team_id
        });
      }
    }
    if (alertState?.state === 'edit') {
      const queryData = [];
      queryData.push({
        alias: '',
        label: viewAlertDetails?.alert?.event,
        filters: processFiltersFromQuery(viewAlertDetails?.alert?.filter),
        group: ''
      });
      setActiveGrpBtn(
        viewAlertDetails?.alert?.event_level === 'account' ? 'events' : 'users'
      );
      setQueries(queryData);

      if (
        viewAlertDetails?.alert?.action_performed === 'action_segment_entry' ||
        viewAlertDetails?.alert?.action_performed === 'action_segment_exit'
      ) {
        setSegmentType(viewAlertDetails?.alert?.action_performed);
        setSelectedSegment(viewAlertDetails?.alert?.event);
        setQueries([]);
      } else {
        setSegmentType('action_event');
        setSelectedSegment('');
      }

      setAlertName(viewAlertDetails?.alert?.title);
      setAlertMessage(viewAlertDetails?.alert?.message);
      setAlertLimit(viewAlertDetails?.alert?.alert_limit);
      setCoolDownTime(viewAlertDetails?.alert?.cool_down_time / 3600);
      setNotRepeat(viewAlertDetails?.alert?.repeat_alerts);
      setNotifications(viewAlertDetails?.alert?.notifications);
      setIsHyperLinkEnabled(!viewAlertDetails?.alert?.is_hyperlink_disabled);

      const isWebHookFactorsUrlEnabled = viewAlertDetails?.alert
        ?.is_factors_url_in_payload
        ? viewAlertDetails?.alert?.is_factors_url_in_payload
        : false;
      setFactorsURLinWebhook(isWebHookFactorsUrlEnabled);

      const messageProperty = processBreakdownsFromQuery(
        viewAlertDetails?.alert?.message_property
      );
      messageProperty.forEach((property) => pushGroupBy(property));

      // open advanced settings by default
      if (
        viewAlertDetails?.alert?.repeat_alerts ||
        !viewAlertDetails?.alert?.is_hyperlink_disabled
      ) {
        setShowAdvSettings(true);
      }

      if (viewAlertDetails?.alert?.slack_mentions) {
        const selectedUser = viewAlertDetails?.alert?.slack_mentions?.map(
          (item) => item?.name
        );
        setSelectedMentions(selectedUser);
      }

      // webhook settings
      if (viewAlertDetails?.alert?.webhook) {
        setWebhookEnabled(viewAlertDetails?.alert?.webhook);
        setWebhookUrl(viewAlertDetails?.alert?.url);
        setFinalWebhookUrl(viewAlertDetails?.alert?.url);
        setConfirmBtn(false);
        setTestMessageBtn(true);
        setTestMassageResponse('');
        setConfirmedMessageBtn(false);
        setShowEditBtn(true);
        setDisbleWebhookInput(true);
        setHideTestMessageBtn(false);
      } else {
        setWebhookEnabled(viewAlertDetails?.alert?.webhook);
        setWebhookUrl('');
        setFinalWebhookUrl('');
        setConfirmBtn(true);
        setTestMessageBtn(true);
        setTestMassageResponse('');
        setConfirmedMessageBtn(false);
        setShowEditBtn(false);
        setDisbleWebhookInput(false);
        setHideTestMessageBtn(true);
      }
    } else if (alertState?.state === 'add' && viewAlertDetails) {
      setAlertName(viewAlertDetails?.alert?.title);
      setAlertMessage(viewAlertDetails?.alert?.message);

      setQueries(viewAlertDetails?.alert?.currentQuery);
      const messageProperty = viewAlertDetails?.alert?.message_property;

      messageProperty.forEach((property) => pushGroupBy(property));
    }
    return () => {
      // reset form values on unmount
      onReset();
    };
  }, [viewAlertDetails, alertState]);

  const menu = () => (
    <Menu style={{ width: '140px' }}>
      <Menu.Item key='1' onClick={() => createDuplicateAlert(viewAlertDetails)}>
        <div className='flex items-center'>
          <SVG name='Pluscopy' size={16} color='grey' extraClass='mr-1' />
          <Text type='title' level={7} color='grey-2' extraClass='m-0 ml-1'>
            Create copy
          </Text>
        </div>
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key='2' onClick={() => confirmDeleteAlert(viewAlertDetails)}>
        <div className='flex items-center'>
          <SVG name='Delete1' size={16} color='red' extraClass='mr-1' />
          <Text type='title' level={7} color='red' extraClass='m-0 ml-1'>
            Delete
          </Text>
        </div>
      </Menu.Item>
    </Menu>
  );

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
            resetGroupBy();
            setEventPropertyDetails({});
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          resetGroupBy();
          queryupdated.splice(index, 1);
          setEventPropertyDetails({});
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setQueries(queryupdated);
    },
    [queries]
  );

  const queryList = () => {
    const blockList = [];
    queries.forEach((event, index) => {
      blockList.push(
        <div key={index}>
          <QueryBlock
            availableGroups={groupsList}
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={queryChange}
            groupAnalysis={activeGrpBtn}
          />
        </div>
      );
    });

    if (queries.length < 1) {
      blockList.push(
        <div key='init'>
          <QueryBlock
            availableGroups={groupsList}
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={queryChange}
            groupBy={queryOptions.groupBy}
            groupAnalysis={activeGrpBtn}
          />
        </div>
      );
    }

    return blockList;
  };

  useEffect(() => {
    getUserPropertiesV2(activeProject.id, queryType);
  }, [queries]);

  const addGroupBy = () => {
    setGroupByDDVisible(true);
  };

  const deleteGroupBy = (groupState, id, type = 'event') => {
    delGroupBy(type, groupState, id);
  };

  const pushGroupBy = (groupState, ind) => {
    const i = ind >= 0 ? ind : groupBy.length;
    setGroupBy('event', groupState, i);
  };
  const selectGroupByEvent = () =>
    isGroupByDDVisible ? (
      <EventGroupBlock
        eventIndex={1}
        event={queries?.[0]}
        setGroupState={pushGroupBy}
        closeDropDown={() => setGroupByDDVisible(false)}
        hideText
        noMargin
        eventGroup={
          groupsList?.filter(
            (item) => item?.[0] == queries?.[0]?.group
          )?.[0]?.[1]
        }
        groupAnalysis={activeGrpBtn}
      />
    ) : null;

  const groupByItems = () => {
    const groupByEvents = [];
    let results;
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      const sortableList = groupBy
        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
        .filter(
          (gbp) => gbp.eventName === queries?.[0]?.label && gbp.eventIndex === 1
        );

      results = (
        <ReactSortable
          list={sortableList}
          setList={(listItems) => {
            dispatch(setGroupByEventActionList(listItems));
          }}
        >
          {sortableList.map((gbp, gbpIndex) => {
            const { groupByIndex, ...orgGbp } = gbp;
            return (
              <div key={gbpIndex} className='fa--query_block--filters'>
                <EventGroupBlock
                  index={gbp.groupByIndex}
                  grpIndex={gbpIndex}
                  eventIndex={1}
                  groupByEvent={orgGbp}
                  event={queries?.[0]}
                  delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                  setGroupState={pushGroupBy}
                  closeDropDown={() => setGroupByDDVisible(false)}
                  hideText
                  noMargin
                  eventGroup={
                    groupsList?.filter(
                      (item) => item?.[0] == queries?.[0]?.group
                    )?.[0]?.[1]
                  }
                  groupAnalysis={activeGrpBtn}
                />
              </div>
            );
          })}
        </ReactSortable>
      );
    }

    if (isGroupByDDVisible) {
      groupByEvents.push(
        <div key='init' className='fa--query_block--filters'>
          {selectGroupByEvent()}
        </div>
      );
    }

    results = (
      <>
        {results} {groupByEvents}
      </>
    );
    return results;
  };

  const viewGroupByItems = (groupBy) => {
    const groupByEvents = [];

    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
        .filter((gbp) => gbp?.eventName === viewAlertDetails?.alert?.event)
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
            <div key={gbpIndex} className='fa--query_block--filters'>
              <EventGroupBlock
                index={gbp.groupByIndex}
                grpIndex={gbpIndex}
                eventIndex={1}
                groupByEvent={orgGbp}
                event={viewAlertDetails?.alert?.event}
                delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                setGroupState={pushGroupBy}
                closeDropDown={() => setGroupByDDVisible(false)}
                hideText
                noMargin
                eventGroup={
                  groupsList?.filter(
                    (item) => item?.[0] == queries?.[0]?.group
                  )?.[0]?.[1]
                }
                groupAnalysis={activeGrpBtn}
              />
            </div>
          );
        });
    }

    if (isGroupByDDVisible) {
      groupByEvents.push(
        <div key='init' className='fa--query_block--filters'>
          {selectGroupByEvent()}
        </div>
      );
    }

    return groupByEvents;
  };

  const onReset = () => {
    setQueries([]);
    setSlackEnabled(false);
    setAlertLimit(5);
    setNotRepeat(false);
    setNotifications(false);
    setIsHyperLinkEnabled(false);
    setSelectedChannel([]);
    setSaveSelectedChannel([]);
    form.resetFields();
    setAlertState({ ...alertState, state: 'list', index: 0 });
    resetGroupBy();
    setEventPropertyDetails({});
    setBreakdownOptions([]);
    setSegmentType('action_event');
    setSelectedSegment('');
    setSegmentOptions([]);
  };

  const confirmDeleteAlert = (item) => {
    confirm({
      title: 'Do you want to delete this alert?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      onOk() {
        deleteEventAlert(activeProject?.id, item?.id)
          .then(() => {
            message.success('Deleted Alert successfully!');
            setAlertState({ ...alertState, state: 'list', index: 0 });
            fetchAllAlerts(activeProject.id);
          })
          .catch((err) => {
            message.error(err);
          });
      }
    });
  };

  const getSlackProfileDetails = (users) => {
    const slackUserList = users?.map((user) =>
      slack_users?.find((item) => item?.name == user)
    );
    return slackUserList;
  };

  const updatepayloadDisplayNames = (payload) => {
    if (payload) {
      const newObj = {};
      Object?.keys(payload)?.map((item) => {
        const newKey = matchEventName(item);
        const val = dummyPayloadValue[item] || payload[item];
        newObj[newKey] = val;
      });
      return newObj;
    }
    return {};
  };

  const sendTestSlackMessage = () => {
    const payload = {
      title: alertName,
      event_level: activeGrpBtn === 'events' ? 'account' : 'user',
      // action_performed: segmentType,
      event:
        segmentType === 'action_event' ? queries[0]?.label : selectedSegment,
      message: alertMessage,
      message_property: convertObjectToKeyValuePairArray(
        updatepayloadDisplayNames(getMsgPayloadMapping(groupBy))
      ),
      slack: slackEnabled,
      slack_channels: saveSelectedChannel,
      slack_mentions: getSlackProfileDetails(selectedMentions),
      is_hyperlink_disabled: !isHyperLinkEnabled
    };
    setSlackTestMsgLoading(true);
    testSlackAlert(activeProject?.id, payload)
      .then((res) => {
        setSlackTestMsgLoading(false);
        setSlackTestMsgTxt(true);
        setTimeout(() => {
          setSlackTestMsgTxt(false);
        }, 5000);
      })
      .catch((err) => {
        console.log('testSlackAlert failed! -->', err);
        setSlackTestMsgLoading(false);
      });
  };
  const sendTestTeamsMessage = () => {
    const payload = {
      title: alertName,
      event_level: activeGrpBtn === 'events' ? 'account' : 'user',
      // action_performed: segmentType,
      event:
        segmentType === 'action_event' ? queries[0]?.label : selectedSegment,
      message: alertMessage,
      message_property: convertObjectToKeyValuePairArray(
        updatepayloadDisplayNames(getMsgPayloadMapping(groupBy))
      ),
      teams: teamsEnabled,
      teams: teamsEnabled,
      teams_channels_config: {
        team_id: selectedWorkspace?.id,
        team_name: selectedWorkspace?.name,
        team_channel_list: teamsSaveSelectedChannel
      }
    };

    setTeamsTestMsgLoading(true);
    testTeamsAlert(activeProject?.id, payload)
      .then((res) => {
        setTeamsTestMsgLoading(false);
        setTeamsTestMsgTxt(true);
        setTimeout(() => {
          setTeamsTestMsgTxt(false);
        }, 5000);
      })
      .catch((err) => {
        setTeamsTestMsgLoading(false);
        console.log('testTeamsAlert failed! -->', err);
      });
  };

  const onFinish = (data) => {
    setLoading(true);

    let breakDownProperties = [];
    if (
      (queries.length > 0 || selectedSegment) &&
      (EventPropertyDetails?.name || EventPropertyDetails?.[1])
    ) {
      let category;

      for (const property in eventPropertiesV2[queries[0]?.label]) {
        const nestedArrays = eventPropertiesV2[queries[0]?.label][property];
        category = nestedArrays.filter(
          (prop) =>
            prop[1] ===
            (EventPropertyDetails?.name || EventPropertyDetails?.[1])
        );
      }

      breakDownProperties = [
        {
          eventName: queries?.[0]?.label || selectedSegment,
          property: EventPropertyDetails?.name || EventPropertyDetails?.[1],
          prop_type:
            EventPropertyDetails?.data_type || EventPropertyDetails?.[2],
          prop_category: category?.length > 0 ? 'event' : 'user'
        }
      ];
    }

    if (
      (queries.length > 0 || selectedSegment) &&
      (slackEnabled || webhookEnabled || teamsEnabled) &&
      (saveSelectedChannel.length > 0 ||
        finalWebhookUrl !== '' ||
        teamsSaveSelectedChannel.length > 0)
    ) {
      const payload = {
        title: data?.alert_name,
        event_level: activeGrpBtn === 'events' ? 'account' : 'user',
        action_performed: segmentType,
        event:
          segmentType === 'action_event' ? queries[0]?.label : selectedSegment,
        filter: formatFiltersForQuery(queries?.[0]?.filters),
        notifications,
        is_hyperlink_disabled: !isHyperLinkEnabled,
        message: data?.message,
        message_property:
          groupBy && groupBy.length && groupBy[0] && groupBy[0].property
            ? formatBreakdownsForQuery(
                groupBy
                  .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
                  .filter(
                    (gbp) =>
                      gbp?.eventName === queries[0]?.label &&
                      gbp.eventIndex === 1
                  )
              )
            : [],
        alert_limit: alertLimit,
        repeat_alerts: notRepeat,
        cool_down_time: coolDownTime * 3600,
        breakdown_properties: formatBreakdownsForQuery(breakDownProperties),
        slack: slackEnabled,
        slack_team_id: saveSelectedChannel?.[0]?.team_id,
        slack_channels: saveSelectedChannel.map(({ name, id, is_private }) => ({
          name,
          id,
          is_private
        })),
        webhook: webhookEnabled,
        url: finalWebhookUrl,
        teams: teamsEnabled,
        teams_channels_config: {
          team_id: selectedWorkspace?.id,
          team_name: selectedWorkspace?.name,
          team_channel_list: teamsSaveSelectedChannel
        },
        slack_mentions: getSlackProfileDetails(selectedMentions),
        is_factors_url_in_payload: factorsURLinWebhook
      };

      if (alertState?.state === 'edit') {
        editEventAlert(activeProject.id, payload, viewAlertDetails?.id)
          .then((res) => {
            setLoading(false);
            fetchAllAlerts(activeProject.id);
            notification.success({
              message: 'Alerts Saved',
              description: 'Alerts is edited and saved successfully.'
            });
            onReset();
          })
          .catch((err) => {
            setLoading(false);
            notification.error({
              message: 'Error',
              description: err?.data?.error
            });
          });
      } else {
        createEventAlert(activeProject.id, payload)
          .then((res) => {
            setLoading(false);
            fetchAllAlerts(activeProject.id);
            notification.success({
              message: 'Alerts Saved',
              description: 'New Alerts is created and saved successfully.'
            });
            onReset();
          })
          .catch((err) => {
            setLoading(false);
            notification.error({
              message: 'Error',
              description: err?.data?.error
            });
          });
      }
    } else {
      setLoading(false);
      if (queries.length === 0) {
        notification.error({
          message: 'Error',
          description: 'Please select Event to send alert.'
        });
      }
      if (!slackEnabled && !webhookEnabled && !teamsEnabled) {
        notification.error({
          message: 'Error',
          description:
            'Please select atleast one delivery option to send alert.'
        });
      }
      if (slackEnabled && saveSelectedChannel.length === 0) {
        notification.error({
          message: 'Error',
          description: 'Empty Slack Channel List'
        });
      }
      if (webhookEnabled && finalWebhookUrl === '') {
        notification.error({
          message: 'Error',
          description: 'Empty Webhook Url'
        });
      }
      if (teamsEnabled && teamsSaveSelectedChannel.length === 0) {
        notification.error({
          message: 'Error',
          description: 'Empty Teams Channel List'
        });
      }
    }
  };

  const createDuplicateAlert = (item) => {
    const payload = {
      ...item?.alert,
      title: `Copy of ${item?.alert?.title}`
    };
    createEventAlert(activeProject?.id, payload)
      .then((res) => {
        setLoading(false);
        fetchAllAlerts(activeProject?.id);
        onReset();
        notification.success({
          message: 'Alert Created',
          description: 'Copy of alert is created and saved successfully.'
        });
      })
      .catch((err) => {
        setLoading(false);
        notification.error({
          message: 'Error',
          description: err?.data?.error
        });
      });
  };

  const onConnectSlack = () => {
    enableSlackIntegration(activeProject.id)
      .then((r) => {
        if (r.status == 200) {
          window.open(r.data.redirectURL, '_blank');
        }
        if (r.status >= 400) {
          message.error('Error fetching Slack redirect url');
        }
      })
      .catch((err) => {
        console.log('Slack error-->', err);
      });
  };

  const onConnectMSTeams = () => {
    enableTeamsIntegration(activeProject.id)
      .then((r) => {
        if (r.status == 200) {
          window.open(r.data.redirectURL, '_blank');
        }
        if (r.status >= 400) {
          message.error('Error fetching teams redirect url');
        }
      })
      .catch((err) => {
        console.log('Teams error-->', err);
      });
  };

  const onChange = () => {
    seterrorInfo(null);
  };

  const handleAlertLimit = (value) => {
    setAlertLimit(value);
    setNotifications(true);
  };

  const handleCoolDownTimeChange = (value) => {
    setCoolDownTime(value);
    setNotRepeat(true);
  };

  const fetchSlackDetails = () => {
    fetchProjectSettingsV1(activeProject.id);
    if (slackEnabled) {
      setSlackMentionLoading(true);
      fetchSlackChannels(activeProject.id);
      fetchSlackUsers(activeProject.id)
        .then(() => {
          setSlackMentionLoading(false);
        })
        .catch(() => {
          setSlackMentionLoading(false);
        });
    }
  };

  useEffect(() => {
    fetchSlackDetails();
  }, [activeProject, projectSettings?.int_slack, slackEnabled]);

  useEffect(() => {
    queries.forEach((ev) => {
      if (!eventPropertiesV2[ev.label]) {
        getEventPropertiesV2(activeProject.id, ev.label);
      }
    });
  }, [activeProject?.id, eventPropertiesV2, getEventPropertiesV2, queries]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettings?.int_teams && teamsEnabled) {
      fetchTeamsWorkspace(activeProject.id)
        .then((res) => {
          if (res.ok) {
            const tempArr = [];
            for (let i = 0; i < res?.data?.length; i++) {
              tempArr.push({
                label: res?.data[i]?.displayName,
                value: res?.data[i]?.id
              });
            }
            setTeamsWorkspaceOpts(tempArr);
          }
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
  }, [activeProject, projectSettings?.int_teams, teamsEnabled]);

  useEffect(() => {
    if (projectSettings?.int_teams && selectedWorkspace) {
      fetchTeamsChannels(activeProject.id, selectedWorkspace?.id);
    }
  }, [
    activeProject,
    projectSettings?.int_teams,
    teamsEnabled,
    selectedWorkspace
  ]);

  const fetchTeamsDetails = () => {
    fetchProjectSettingsV1(activeProject.id);
    if (teamsEnabled) {
      fetchTeamsWorkspace(activeProject.id);
    }
    if (projectSettings?.int_teams && selectedWorkspace) {
      fetchTeamsChannels(activeProject.id, selectedWorkspace?.id);
    }
  };

  useEffect(() => {
    if (slack?.length > 0) {
      const tempArr = [];
      for (let i = 0; i < slack.length; i++) {
        tempArr.push({
          name: slack[i].name,
          id: slack[i].id,
          is_private: slack[i].is_private,
          team_id: slack[i].context_team_id
        });
      }
      setChannelOpts(tempArr);
    }
  }, [activeProject, agent_details, slack]);

  useEffect(() => {
    if (teams?.length > 0 && selectedWorkspace) {
      const tempArr = [];
      for (let i = 0; i < teams?.length; i++) {
        tempArr.push({
          name: teams?.[i]?.displayName,
          id: teams?.[i]?.id
        });
      }
      setTeamsChannelOpts(tempArr);
    } else {
      setTeamsChannelOpts([]);
    }
  }, [activeProject, agent_details, teams]);

  const handleOk = () => {
    setSaveSelectedChannel(selectedChannel);
    setShowSelectChannelsModal(false);
  };

  const handleCancel = () => {
    setSelectedChannel(saveSelectedChannel);
    setShowSelectChannelsModal(false);
  };

  const handleOkTeams = () => {
    setTeamsSaveSelectedChannel(teamsSelectedChannel);
    setTeamsShowSelectChannelsModal(false);
  };

  const handleCancelTeams = () => {
    setTeamsSelectedChannel(teamsSaveSelectedChannel);
    setTeamsShowSelectChannelsModal(false);
  };

  // Webhook settings
  const handleTestWebhook = () => {
    const payload = {
      title: alertName,
      // action_performed: segmentType,
      event:
        segmentType === 'action_event' ? queries[0]?.label : selectedSegment,
      message_property:
        groupBy && groupBy.length && groupBy[0] && groupBy[0].property
          ? formatBreakdownsForQuery(
              groupBy
                .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
                .filter(
                  (gbp) =>
                    gbp.eventName === queries[0]?.label && gbp.eventIndex === 1
                )
            )
          : [],
      message: alertMessage,
      url: webhookUrl,
      secret: '',
      event_level: activeGrpBtn == 'events' ? 'account' : 'user',
      is_factors_url_in_payload: factorsURLinWebhook
    };
    setWHTestMsgLoading(true);
    testWebhhookUrl(activeProject?.id, payload)
      .then((res) => {
        setTestMassageResponse(res?.data);
        setWHTestMsgLoading(false);
        setWHTestMsgTxt(true);
        setTimeout(() => {
          setWHTestMsgTxt(false);
        }, 5000);
      })
      .catch((err) => {
        setWHTestMsgLoading(false);
        message.error(err?.data?.error);
      });
  };

  const handleClickConfirmBtn = () => {
    setConfirmedMessageBtn(true);
    setDisbleWebhookInput(true);
    setFinalWebhookUrl(webhookUrl);
    setTimeout(() => {
      setConfirmedMessageBtn(false);
      setShowEditBtn(true);
      setTestMassageResponse('');
      setTestMessageBtn(true);
      setHideTestMessageBtn(false);
    }, 2000);
  };

  useEffect(() => {
    if (showEditBtn && webhookUrl !== finalWebhookUrl) {
      setShowEditBtn(false);
      setConfirmedMessageBtn(false);
      setConfirmBtn(false);
      setTestMassageResponse('');
      setTestMessageBtn(false);
    }
  }, [webhookUrl, finalWebhookUrl]);

  useEffect(() => {
    if (viewAlertDetails?.status === 'active') {
      setisAlertEnabled(true);
    }
  }, [viewAlertDetails]);

  const confirmAlertPause = (item) => {
    const status = 'paused';
    confirm({
      title: 'Pause Alert?',
      icon: <ExclamationCircleOutlined />,
      content:
        'Alerts and webhooks from this event will be paused. You can always turn this back on when needed.',
      onOk() {
        setLoading(true);
        updateEventAlertStatus(activeProject?.id, item?.id, status)
          .then(() => {
            message.success('Successfully paused/disabled alerts.');
            setisAlertEnabled(false);
            onReset();
            fetchAllAlerts(activeProject.id);
            setLoading(false);
          })
          .catch((err) => {
            message.error(err);
            setLoading(false);
          });
      }
    });
  };

  const toggleAlertEnabled = (checked) => {
    if (!checked) {
      confirmAlertPause(viewAlertDetails);
    } else {
      const status = 'active';
      const id = viewAlertDetails?.id;
      setLoading(true);
      updateEventAlertStatus(activeProject?.id, id, status)
        .then((res) => {
          setisAlertEnabled(true);
          fetchAllAlerts(activeProject.id);
          message.success('Successfully enabled alerts.');
          setLoading(false);
        })
        .catch((err) => {
          console.log('Oops! something went wrong-->', err);
          message.error(`Oops! something went wrong. ${err?.data?.error}`);
          setLoading(false);
        });
    }
  };

  const propOption = (item) => (
    <Tooltip title={item} placement='right'>
      <div style={{ width: '210px' }}>
        <div
          style={{
            maxWidth: '200px',
            overflow: 'hidden',
            whiteSpace: 'nowrap',
            textOverflow: 'ellipsis'
          }}
        >
          {item}
        </div>
      </div>{' '}
    </Tooltip>
  );

  const renderEventForm = () => (
    <Form
      form={form}
      onFinish={onFinish}
      className='w-full'
      onChange={onChange}
      loading={loading}
    >
      <Row>
        {alertState.state == 'edit' ? (
          <>
            {viewAlertDetails?.last_fail_details &&
              !viewAlertDetails?.last_fail_details?.is_paused_automatically && (
                <Col span={24} className='mb-4'>
                  <Alert
                    message='We are unable to send this alert to the destinations you selected. Please check the destination settings below to continue receiving alerts'
                    type='error'
                    showIcon
                  />
                </Col>
              )}
            {viewAlertDetails?.last_fail_details &&
              viewAlertDetails?.last_fail_details?.is_paused_automatically && (
                <Col span={24} className='mb-4'>
                  <Alert
                    message='Alert paused due to unresolved issues with selected destinations. Please check the errors in the destinations to resume getting alerts.'
                    type='info'
                    showIcon
                  />
                </Col>
              )}
            <Col span={18}>
              <div className='flex items-center'>
                <div className='flex items-baseline'>
                  <Text
                    type='title'
                    level={3}
                    weight='bold'
                    extraClass='m-0'
                    truncate
                    charLimit={50}
                  >
                    {`${viewAlertDetails?.title}`}
                  </Text>
                </div>
                <div className='ml-4'>
                  <Switch
                    checkedChildren='On'
                    unCheckedChildren='OFF'
                    onChange={toggleAlertEnabled}
                    checked={isAlertEnabled}
                    size='large'
                    loading={loading}
                  />
                </div>
              </div>
            </Col>
            <Col span={6}>
              <div className='flex justify-end items-center'>
                <Dropdown
                  trigger={['click']}
                  overlay={menu}
                  placement='bottomRight'
                  className='mr-2'
                >
                  <Button
                    type='text'
                    icon={
                      <MoreOutlined
                        rotate={90}
                        style={{ color: 'gray', fontSize: '18px' }}
                      />
                    }
                  />
                </Dropdown>

                <Button
                  size='large'
                  disabled={loading}
                  onClick={() => {
                    onReset();
                  }}
                >
                  Cancel
                </Button>
                <Button
                  size='large'
                  disabled={loading}
                  loading={loading}
                  className='ml-2'
                  type='primary'
                  htmlType='submit'
                >
                  Save
                </Button>
              </div>
            </Col>
          </>
        ) : (
          <>
            <Col span={12}>
              <Text type='title' level={3} weight='bold' extraClass='m-0'>
                Create new alert
              </Text>
            </Col>
            <Col span={12}>
              <div className='flex justify-end'>
                <Button
                  size='large'
                  disabled={loading}
                  onClick={() => {
                    onReset();
                  }}
                >
                  Cancel
                </Button>
                <Button
                  size='large'
                  disabled={loading}
                  loading={loading}
                  className='ml-2'
                  type='primary'
                  htmlType='submit'
                >
                  Save
                </Button>
              </div>
            </Col>
          </>
        )}
      </Row>

      <Row className='mt-6 border-top--thin-2 pt-6'>
        <Col span={18}>
          <Text type='title' level={7} weight='bold' extraClass='m-0'>
            When to trigger alert
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0'>
            Choose the event you wish to be alerted for. You can choose events
            at an account level or at a people level
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={18}>
          <Text type='title' level={8} extraClass='m-0'>
            When
          </Text>
        </Col>
      </Row>
      <Row className='mt-1 mb-4'>
        <Col span={12}>
          <div className='flex items-center justify-start btn-custom--radio-container'>
            <Button
              type='default'
              className={`${activeGrpBtn == 'events' ? 'active' : 'no-border'}`}
              onClick={() => confirmGroupSwitch('events')}
            >
              Accounts
            </Button>
            <Button
              type='default'
              className={`${activeGrpBtn == 'users' ? 'active' : 'no-border'}`}
              onClick={() => confirmGroupSwitch('users')}
            >
              People
            </Button>
          </div>
        </Col>
      </Row>
      <Row className='mt-4 mb-1'>
        <Col span={18}>
          <Text type='title' level={7} extraClass='m-0'>
            Do this
          </Text>
        </Col>
      </Row>
      <Row>
        <Col span={22}>
          <Select
            showSearch
            style={{ minWidth: 350 }}
            className='fa-select'
            placeholder='Select segment type'
            optionFilterProp='children'
            onChange={onChangeSegmentType}
            filterOption={(input, option) =>
              option.props.children
                .toLowerCase()
                .indexOf(input.toLowerCase()) >= 0
            }
            value={segmentType}
          >
            {activeGrpBtn === 'users' ? (
              <Option value='action_event'>Performs an event</Option>
            ) : (
              <>
                <Option value='action_event'>Performs an event</Option>
                <Option value='action_segment_entry'>Enter the segment</Option>
                <Option value='action_segment_exit'>Exit the segment</Option>
              </>
            )}
          </Select>
        </Col>
      </Row>
      {segmentType !== 'action_event' ? (
        <>
          <Row className='mt-4'>
            <Col span={18}>
              <Text type='title' level={7} extraClass='m-0'>
                Segment name
              </Text>
            </Col>
          </Row>
          <Row className='mt-2 mb-4 border-bottom--thin-2 pb-6'>
            <Col span={18}>
              <Select
                showSearch
                style={{
                  width: 'fix-content',
                  minWidth: 350
                }}
                className='fa-select'
                placeholder='Select or search segment'
                labelInValue
                value={selectedSegment}
                onChange={onChangeSegment}
                filterOption={(input, option) =>
                  (option?.value
                    ? getSegmentNameFromId(option?.value).toLowerCase()
                    : ''
                  ).includes(input.toLowerCase())
                }
                options={segmentOptions}
              />
            </Col>
          </Row>
        </>
      ) : (
        <>
          <Row className='mt-4'>
            <Col span={18}>
              <Text type='title' level={7} extraClass='m-0'>
                Event details
              </Text>
            </Col>
          </Row>
          <Row className='mt-2 mb-4 border-bottom--thin-2 pb-6'>
            <Col span={22}>
              <div className='border--thin-2 px-4 py-2 border-radius--sm'>
                <Form.Item name='event_name' className='m-0'>
                  {queryList()}
                </Form.Item>
              </div>
            </Col>
          </Row>
        </>
      )}

      <Row className='mt-6'>
        <Col span={18}>
          <Text type='title' level={7} weight='bold' extraClass='m-0'>
            What to include in the alert
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0'>
            Choose the information you wish to see in the alerts
          </Text>
        </Col>
      </Row>
      <Row className='mt-2'>
        <Col span={18}>
          <Text type='title' level={7} extraClass='m-0 mt-4'>
            Alert Name
          </Text>
        </Col>

        <Col span={10} className='m-0'>
          <Form.Item
            name='alert_name'
            className='m-0'
            initialValue={viewAlertDetails?.title}
            rules={[{ required: true, message: 'Please enter alert name' }]}
          >
            <Input
              className='fa-input'
              placeholder='Enter name'
              onChange={(e) => setAlertName(e.target.value)}
              // ref={inputComponentRef}
            />
          </Form.Item>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={10}>
          <div>
            <Text type='title' level={7} extraClass='m-0 inline'>
              Add a message
            </Text>
            <Popover
              placement='right'
              overlayInnerStyle={{ width: '340px' }}
              title={null}
              content={
                <div className='m-0 m-2'>
                  <p className='m-0 text-gray-900 text-base font-bold'>
                    Your notification inside Slack
                  </p>
                  <p className='m-0 mb-2 text-gray-700'>
                    As events across your marketing activities happen, get
                    alerts that motivate actions right inside Slack
                  </p>
                  <img
                    className='m-0'
                    src='../../../../../assets/icons/Slackmock.svg'
                  />
                </div>
              }
            >
              <div className='inline ml-1'>
                <SVG
                  name='InfoCircle'
                  size={16}
                  color='#8692A3'
                  extraClass='inline'
                />
              </div>
            </Popover>
          </div>
          <Form.Item
            name='message'
            initialValue={viewAlertDetails?.alert?.message}
            className='m-0'
          >
            <TextArea
              className='fa-input'
              placeholder='Enter Message (max 300 characters)'
              onChange={(e) => setAlertMessage(e.target.value)}
              maxLength={300}
            />
          </Form.Item>
        </Col>
      </Row>

      {(queries.length > 0 || selectedSegment) && (
        <Row className='mt-4'>
          <Col span={12}>
            <div>
              <Text type='title' level={7} extraClass='m-0 inline mb-1 mr-1'>
                Add properties to show
              </Text>
              <Popover
                placement='rightTop'
                overlayInnerStyle={{ width: '300px' }}
                title={null}
                content={
                  <p className='m-0 m-2 text-gray-700'>
                    In Slack, you’ll get these values on your channel. With a
                    webhook, use these properties to power your own workflows.
                  </p>
                }
              >
                <div className='inline'>
                  <SVG
                    name='InfoCircle'
                    size={18}
                    color='#8692A3'
                    extraClass='inline'
                  />
                </div>
              </Popover>
            </div>
            <div
              className='fa--query_block_section borderless no-padding mt-0'
              style={{ marginLeft: '-20px' }}
            >
              {groupByItems()}
            </div>
            <Button
              type='text'
              style={{ color: '#8692A3', margin: '2px auto' }}
              icon={<SVG name='plus' color='#8692A3' />}
              onClick={() => addGroupBy()}
            >
              Add a Property
            </Button>
          </Col>
        </Row>
      )}

      <Row className=''>
        <Col span={24}>
          <div className='border-top--thin-2 pb-6 mt-6' />
          <Text type='title' level={7} weight='bold' extraClass='m-0'>
            {' '}
            Where to get the alert{' '}
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0'>
            {' '}
            Choose where you wish to get the alert. You can select multiple
            destinations as well{' '}
          </Text>
        </Col>
      </Row>

      {/* {showSlackInt && <Slack */}
      <Slack
        viewAlertDetails={viewAlertDetails}
        slackEnabled={slackEnabled}
        setSlackEnabled={setSlackEnabled}
        projectSettings={projectSettings}
        onConnectSlack={onConnectSlack}
        saveSelectedChannel={saveSelectedChannel}
        setSaveSelectedChannel={setSaveSelectedChannel}
        setShowSelectChannelsModal={setShowSelectChannelsModal}
        selectedMentions={selectedMentions}
        setSelectedMentions={setSelectedMentions}
        slack_users={slack_users}
        sendTestSlackMessage={sendTestSlackMessage}
        alertMessage={alertMessage}
        alertName={alertName}
        groupBy={groupBy}
        fetchSlackDetails={fetchSlackDetails}
        matchEventName={matchEventName}
        slackTestMsgLoading={slackTestMsgLoading}
        slackTestMsgTxt={slackTestMsgTxt}
        slackMentionLoading={slackMentionLoading}
      />

      {/* {showTeamInt && <Teams */}
      <Teams
        viewAlertDetails={viewAlertDetails}
        setTeamsEnabled={setTeamsEnabled}
        teamsEnabled={teamsEnabled}
        projectSettings={projectSettings}
        onConnectMSTeams={onConnectMSTeams}
        teamsSaveSelectedChannel={teamsSaveSelectedChannel}
        selectedWorkspace={selectedWorkspace}
        setTeamsShowSelectChannelsModal={setTeamsShowSelectChannelsModal}
        alertMessage={alertMessage}
        alertName={alertName}
        groupBy={groupBy}
        sendTestTeamsMessage={sendTestTeamsMessage}
        matchEventName={matchEventName}
        teamsTestMsgTxt={teamsTestMsgTxt}
        teamsTestMsgLoading={teamsTestMsgLoading}
        fetchTeamsDetails={fetchTeamsDetails}
      />

      {/* {showWHInt && <Webhook */}
      <Webhook
        viewAlertDetails={viewAlertDetails}
        groupBy={groupBy}
        webhookEnabled={webhookEnabled}
        setWebhookEnabled={setWebhookEnabled}
        disbleWebhookInput={disbleWebhookInput}
        webhookRef={webhookRef}
        webhookUrl={webhookUrl}
        setWebhookUrl={setWebhookUrl}
        setConfirmBtn={setConfirmBtn}
        setTestMessageBtn={setTestMessageBtn}
        showEditBtn={showEditBtn}
        finalWebhookUrl={finalWebhookUrl}
        setHideTestMessageBtn={setHideTestMessageBtn}
        setDisbleWebhookInput={setDisbleWebhookInput}
        confirmedMessageBtn={confirmedMessageBtn}
        handleClickConfirmBtn={handleClickConfirmBtn}
        testMessageResponse={testMessageResponse}
        testMessageBtn={testMessageBtn}
        handleTestWebhook={handleTestWebhook}
        confirmBtn={confirmBtn}
        hideTestMessageBtn={hideTestMessageBtn}
        alertMessage={alertMessage}
        alertName={alertName}
        WHTestMsgTxt={WHTestMsgTxt}
        WHTestMsgLoading={WHTestMsgLoading}
        selectedEvent={queries?.length ? matchEventName(queries[0]?.label) : ''}
        matchEventName={matchEventName}
        factorsURLinWebhook={factorsURLinWebhook}
        setFactorsURLinWebhook={setFactorsURLinWebhook}
        activeGrpBtn={activeGrpBtn}
      />

      {/* 
          <div className='mt-4 mb-2'>
            <Button disabled={showSlackInt} className='ml-2' onClick={() => { setShowSlackInt(true); setSlackEnabled(true) }}><SVG name={'slack'} size={18} color='purple' />Add Slack</Button>
            <Button disabled={showTeamInt} className='ml-2' onClick={() => { setShowTeamInt(true); setTeamsEnabled(true) }}><SVG name={'MSTeam'} size={18} color='purple' />Add Teams</Button>
            <Button disabled={
              !(
                groupBy &&
                groupBy.length &&
                groupBy[0] &&
                groupBy[0].property
              ) || isWebHookFeatureLocked
            } className='ml-2' onClick={() => { setShowWHInt(true); setWebhookEnabled(true) }}><SVG name={'Webhook'} size={18} color='purple' />Setup Webhook</Button>
          </div> */}

      <Row className='border-top--thin-2 mt-6 pt-6'>
        {showAdvSettings && (
          <>
            <Col span={24}>
              <Text
                type='title'
                level={7}
                weight='bold'
                color='grey-2'
                extraClass='m-0'
              >
                {' '}
                Advanced settings
              </Text>
            </Col>
            <Col span={16} className='m-0 mt-4'>
              <Form.Item name='repeat_alerts' className='m-0'>
                <Checkbox
                  checked={notRepeat}
                  onChange={(e) => setNotRepeat(e.target.checked)}
                >
                  Limit alerts
                </Checkbox>
              </Form.Item>
            </Col>
            <Col span={20}>
              <Form.Item name='event_property' className='m-0 inline'>
                <Text
                  type='title'
                  level={7}
                  color='grey-2'
                  extraClass='m-0 inline'
                >
                  For the same value of
                </Text>

                <div className='inline ml-2'>
                  <Select
                    className='inline fa-select'
                    style={{
                      width: 250
                    }}
                    // dropdownMatchSelectWidth={false}
                    value={EventPropertyDetails}
                    // disabled={!queries[0]?.label}
                    onChange={(value, details) => {
                      setEventPropertyDetails(details);
                      setNotRepeat(true);
                    }}
                    placeholder='Select Property'
                    showSearch
                    filterOption={(input, option) =>
                      option.value.toLowerCase().indexOf(input.toLowerCase()) >=
                      0
                    }
                  >
                    {breakdownOptions?.map((item) => (
                      <Option
                        key={item[1]}
                        value={item[0]}
                        name={item[1]}
                        data_type={item[2]}
                      >
                        {propOption(item[0])}
                      </Option>
                    ))}
                  </Select>
                </div>
                <Text
                  type='title'
                  level={7}
                  color='grey-2'
                  extraClass='m-0 inline ml-2 mr-2'
                >
                  show alert every
                </Text>
                <div className='inline ml-2'>
                  <Select
                    className='inline fa-select'
                    style={{
                      width: 110
                    }}
                    defaultValue={0.5}
                    value={coolDownTime}
                    onChange={handleCoolDownTimeChange}
                  >
                    <Option value={0.5}>0.5 hours</Option>
                    <Option value={1}>1 hours</Option>
                    <Option value={2}>2 hours</Option>
                    <Option value={4}>4 hours</Option>
                    <Option value={6}>6 hours</Option>
                    <Option value={8}>8 hours</Option>
                    <Option value={12}>12 hours</Option>
                    <Option value={24}>24 hours</Option>
                    {/* convert days into hours */}
                    <Option value={7 * 24}>7 days</Option>
                    <Option value={14 * 24}>14 days</Option>
                    <Option value={21 * 24}>21 days</Option>
                    <Option value={28 * 24}>28 days</Option>
                  </Select>
                </div>
              </Form.Item>
            </Col>
            <Col span={16} className='m-0 mt-2'>
              <Form.Item name='is_hyperlink_enabled' className='m-0'>
                <Checkbox
                  checked={isHyperLinkEnabled}
                  onChange={(e) => setIsHyperLinkEnabled(e.target.checked)}
                >
                  Show buttons and hyperlinks in alerts
                </Checkbox>
              </Form.Item>
            </Col>{' '}
          </>
        )}

        <Col span={16} className='m-0 mt-4'>
          <a
            type='link'
            onClick={() => setShowAdvSettings(!showAdvSettings)}
          >{`${
            showAdvSettings ? 'Hide advanced options' : 'Show advanced options'
          }`}</a>
        </Col>
      </Row>

      {alertState.state == 'edit' ? (
        <Row className='border-top--thin-2 mt-6 pt-6'>
          <Col span={12}>
            {/* <a type={'link'} className={'mr-2'} onClick={() => createDuplicateAlert(viewAlertDetails)}>{'Create copy'}</a>
                <a type={'link'} color={'red'} onClick={() => confirmDeleteAlert(viewAlertDetails)}>{`Delete`}</a> */}

            <Button
              type='text'
              color='red'
              onClick={() => createDuplicateAlert(viewAlertDetails)}
            >
              <div className='flex items-center'>
                <SVG name='Pluscopy' size={16} color='grey' extraClass='mr-1' />
                <Text type='title' level={7} extraClass='m-0'>
                  Create copy{' '}
                </Text>
              </div>
            </Button>
            <Button
              type='text'
              color='red'
              onClick={() => confirmDeleteAlert(viewAlertDetails)}
            >
              <div className='flex items-center'>
                <SVG name='Delete1' size={16} color='red' extraClass='mr-1' />
                <Text type='title' level={7} color='red' extraClass='m-0'>
                  Delete{' '}
                </Text>
              </div>
            </Button>
          </Col>
          <Col span={12}>
            <div className='flex justify-end'>
              <Button
                size='large'
                disabled={loading}
                onClick={() => {
                  onReset();
                }}
              >
                Cancel
              </Button>
              <Button
                size='large'
                disabled={loading}
                loading={loading}
                className='ml-2'
                type='primary'
                htmlType='submit'
              >
                Save
              </Button>
            </div>
          </Col>
        </Row>
      ) : (
        <Row className='border-top--thin-2 mt-6 pt-6'>
          <Col span={12} />
          <Col span={12}>
            <div className='flex justify-end'>
              <Button
                size='large'
                disabled={loading}
                onClick={() => {
                  onReset();
                }}
              >
                Cancel
              </Button>
              <Button
                size='large'
                disabled={loading}
                loading={loading}
                className='ml-2'
                type='primary'
                htmlType='submit'
              >
                Save
              </Button>
            </div>
          </Col>
        </Row>
      )}
    </Form>
  );

  return (
    <div className='fa-container'>
      <ScrollToTop />
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          <div className='mb-10 pl-4'>{renderEventForm()}</div>
        </Col>
      </Row>
      <Modal
        title={null}
        visible={showSelectChannelsModal}
        centered
        zIndex={1005}
        width={700}
        onCancel={handleCancel}
        onOk={handleOk}
        className='fa-modal--regular p-4 fa-modal--slideInDown'
        closable
        okText='Save'
        cancelText='Close'
        transitionName=''
        maskTransitionName=''
        okButtonProps={{ size: 'large' }}
        cancelButtonProps={{ size: 'large' }}
      >
        <div>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={22}>
              <Text
                type='title'
                level={4}
                weight='bold'
                size='grey'
                extraClass='m-0'
              >
                Select Slack channels
              </Text>
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={22}>
              <SelectChannels
                channelOpts={channelOpts}
                selectedChannel={selectedChannel}
                setSelectedChannel={setSelectedChannel}
              />
            </Col>
          </Row>
        </div>
      </Modal>
      <Modal
        title={null}
        visible={teamsShowSelectChannelsModal}
        centered
        zIndex={1005}
        width={700}
        onCancel={handleCancelTeams}
        onOk={handleOkTeams}
        className='fa-modal--regular p-4 fa-modal--slideInDown'
        closable
        okText='Save'
        cancelText='Close'
        transitionName=''
        maskTransitionName=''
        okButtonProps={{ size: 'large' }}
        cancelButtonProps={{ size: 'large' }}
      >
        <div>
          <Row>
            <Col span={22}>
              <Text
                type='title'
                level={4}
                weight='bold'
                size='grey'
                extraClass='m-0'
              >
                Select Teams channels
              </Text>
            </Col>
          </Row>
          <Row className='my-3'>
            <Col span={24}>
              <Text
                type='title'
                level={6}
                color='grey-2'
                extraClass='m-0 inline mr-2'
              >
                Workspace
              </Text>
              <Select
                className='fa-select inline'
                options={teamsWorkspaceOpts}
                placeholder='Select Workspace'
                showSearch
                style={{ minWidth: '250px' }}
                value={
                  selectedWorkspace
                    ? {
                        label: selectedWorkspace?.name,
                        value: selectedWorkspace?.id
                      }
                    : null
                }
                onChange={(value, op) => {
                  setSelectedWorkspace({ name: op?.label, id: value });
                  setTeamsSaveSelectedChannel([]);
                  setTeamsSelectedChannel([]);
                }}
              />
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={22}>
              <SelectChannels
                channelOpts={teamsChannelOpts}
                selectedChannel={teamsSelectedChannel}
                setSelectedChannel={setTeamsSelectedChannel}
              />
            </Col>
          </Row>
        </div>
      </Modal>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  savedEventAlerts: state.global.eventAlerts,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  teams: state.global.teams,
  projectSettings: state.global.projectSettingsV1,
  groupBy: state.coreQuery.groupBy.event,
  groupByMagic: state.coreQuery.groupBy,
  groupProperties: state.coreQuery.groupProperties,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventPropNames: state.coreQuery.eventPropNames,
  groupPropNames: state.coreQuery.groupPropNames,
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  userPropNames: state.coreQuery.userPropNames,
  eventNamesSpecial: state.coreQuery.eventNamesSpecial,
  groups: state.coreQuery.groups,
  slack_users: state.global.slack_users
});

export default connect(mapStateToProps, {
  deleteEventAlert,
  createEventAlert,
  editEventAlert,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  setGroupBy,
  delGroupBy,
  getUserPropertiesV2,
  resetGroupBy,
  getGroupProperties,
  getEventPropertiesV2,
  getGroups,
  testWebhhookUrl,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels,
  updateEventAlertStatus,
  setShowCriteria,
  deleteEventAlert,
  fetchAllAlerts,
  fetchSlackUsers,
  testSlackAlert,
  testTeamsAlert,
  getSavedSegments
})(EventBasedAlert);
