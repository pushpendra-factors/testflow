import React, { useState, useEffect, useCallback, useRef } from 'react';
import { connect } from 'react-redux';
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
  Switch
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import {
  createEventAlert,
  fetchEventAlerts,
  deleteEventAlert,
  editEventAlert,
  testWebhhookUrl
} from 'Reducers/global';
import ConfirmationModal from 'Components/ConfirmationModal';
import QueryBlock from './QueryBlock';
import {
  deleteGroupByForEvent,
  setGroupBy,
  delGroupBy,
  getUserPropertiesV2,
  resetGroupBy,
  getGroupProperties,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { getEventsWithProperties, getStateFromFiltersEvent } from '../utils';
import {
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels,
  updateEventAlertStatus
} from 'Reducers/global';
import SelectChannels from '../SelectChannels';
import {
  QUERY_TYPE_EVENT,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE
} from 'Utils/constants';
import { DefaultDateRangeFormat } from '../../../../CoreQuery/utils';
import TextArea from 'antd/lib/input/TextArea';
import EventGroupBlock from '../../../../../components/QueryComposer/EventGroupBlock';
import useAutoFocus from 'hooks/useAutoFocus';
import GLobalFilter from 'Components/KPIComposer/GlobalFilter';
import _ from 'lodash';
import { fetchGroups } from 'Reducers/coreQuery/services';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import UpgradeButton from 'Components/GenericComponents/UpgradeButton';

const { Option } = Select;

const EventBasedAlert = ({
  activeProject,
  fetchEventAlerts,
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
  eventUserPropertiesV2,
  userPropNames,
  eventNames,
  getGroupProperties,
  getEventPropertiesV2,
  fetchGroups,
  groupOpts,
  testWebhhookUrl,
  teams,
  updateEventAlertStatus
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

  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const inputComponentRef = useAutoFocus();
  const [isAlertEnabled, setisAlertEnabled] = useState(false);
  const [enableWidgetModal, showEnableWidgetModal] = useState(false);

  // Webhook support
  const { isFeatureLocked: isWebHookFeatureLocked } = useFeatureLock(
    FEATURES.FEATURE_WEBHOOK
  );
  const [webhookUrl, setWebhookUrl] = useState('');
  const [finalWebhookUrl, setFinalWebhookUrl] = useState('');
  const [confirmBtn, setConfirmBtn] = useState(true);
  const [testMessageBtn, setTestMessageBtn] = useState(true);
  const [testMessageResponse, setTestMassageResponse] = useState('');
  const [confirmedMessageBtn, setConfirmedMessageBtn] = useState(false);
  const [showEditBtn, setShowEditBtn] = useState(false);
  const [disbleWebhookInput, setDisbleWebhookInput] = useState(false);
  const [hideTestMessageBtn, setHideTestMessageBtn] = useState(true);
  const webhookRef = useRef();

  const [form] = Form.useForm();

  // Event SELECTION
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });

  useEffect(() => {
    fetchGroups(activeProject.id);
  }, [activeProject]);

  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  const [breakdownOptions, setBreakdownOptions] = useState([]);
  const [EventPropertyDetails, setEventPropertyDetails] = useState({});

  useEffect(() => {
    let DDCategory = [];
    for (let property in eventPropertiesV2[queries[0]?.label]) {
      let nestedArrays = eventPropertiesV2[queries[0]?.label][property];
      DDCategory = _.union(nestedArrays, DDCategory);
    }
    if (groupOpts[queries[0]?.group]) {
      for (const key of Object.keys(groupProperties)) {
        if (key === queries[0]?.group) {
          DDCategory = _.union(
            DDCategory,
            groupProperties[groupOpts[queries[0]?.group]]
          );
        }
      }
    } else {
      for (let property in eventUserPropertiesV2) {
        let nestedArrays = eventUserPropertiesV2[property];
        DDCategory = _.union(DDCategory, nestedArrays);
      }
    }
    setBreakdownOptions(DDCategory);
    if (
      alertState?.state === 'edit' &&
      !(EventPropertyDetails?.name || EventPropertyDetails?.[0])
    ) {
      let property = DDCategory.filter(
        (data) =>
          data[1] ===
          viewAlertDetails?.event_alert?.breakdown_properties?.[0]?.pr
      );
      setEventPropertyDetails(property?.[0]);
    }
  }, [
    queries,
    eventPropertiesV2,
    groupProperties,
    eventUserPropertiesV2,
    viewAlertDetails,
    alertState
  ]);

  const matchEventName = (item) => {
    let findItem =
      eventPropNames?.[item] || userPropNames?.[item] || groupPropNames?.[item];
    return findItem ? findItem : item;
  };

  useEffect(() => {
    if (viewAlertDetails?.event_alert?.event) {
      getGroupProperties(
        activeProject.id,
        viewAlertDetails?.event_alert?.event
      );
    }
    if (viewAlertDetails?.event_alert?.event) {
      getEventPropertiesV2(
        activeProject.id,
        viewAlertDetails?.event_alert?.event
      );
    }
  }, [viewAlertDetails?.event_alert?.event]);

  const confirmRemove = (id) => {
    return deleteEventAlert(activeProject.id, id).then(
      (res) => {
        fetchEventAlerts(activeProject.id);
        notification.success({
          message: 'Success',
          description: 'Deleted Alert successfully ',
          duration: 5
        });
      },
      (err) => {
        notification.error({
          message: 'Error',
          description: err.data,
          duration: 5
        });
      }
    );
  };

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
      setAlertState({ state: 'list', index: 0 });
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [deleteWidgetModal]);

  const confirmPause = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      const status = 'paused';
      const id = viewAlertDetails?.id;
      return updateEventAlertStatus(activeProject?.id, id, status)
        .then((res) => {
          setDeleteApiCalled(false);
          showEnableWidgetModal(false);
          setisAlertEnabled(false);
          fetchEventAlerts(activeProject.id);
          message.success('Successfully paused/disabled alerts.');
        })
        .catch((err) => {
          setDeleteApiCalled(false);
          console.log('Oops! something went wrong-->', err);
          message.error('Oops! something went wrong. ' + err?.data?.error);
        });
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [enableWidgetModal]);

  useEffect(() => {
    if (viewAlertDetails?.event_alert?.filter) {
      const filter = getStateFromFiltersEvent(
        viewAlertDetails.event_alert.filter
      );
      setViewFilter(filter);
    }
    if (viewAlertDetails?.event_alert?.slack_channels) {
      setViewSelectedChannels(viewAlertDetails?.event_alert?.slack_channels);
      if (alertState?.state === 'edit') {
        setSlackEnabled(viewAlertDetails?.event_alert?.slack);
        setSaveSelectedChannel(viewAlertDetails?.event_alert?.slack_channels);
        setSelectedChannel(viewAlertDetails?.event_alert?.slack_channels);
      }
    }
    if (
      viewAlertDetails?.event_alert?.teams_channels_config?.team_channel_list
    ) {
      setTeamsViewSelectedChannels(
        viewAlertDetails?.event_alert?.teams_channels_config?.team_channel_list
      );
      if (alertState?.state === 'edit') {
        setTeamsEnabled(viewAlertDetails?.event_alert?.teams);
        setTeamsSaveSelectedChannel(
          viewAlertDetails?.event_alert?.teams_channels_config
            ?.team_channel_list
        );
        setTeamsSelectedChannel(
          viewAlertDetails?.event_alert?.teams_channels_config
            ?.team_channel_list
        );
        setSelectedWorkspace({
          name: viewAlertDetails?.event_alert?.teams_channels_config?.team_name,
          id: viewAlertDetails?.event_alert?.teams_channels_config?.team_id
        });
      }
    }
    if (alertState?.state === 'edit') {
      let queryData = [];
      queryData.push({
        alias: '',
        label: viewAlertDetails?.event_alert?.event,
        filters: getStateFromFiltersEvent(viewAlertDetails.event_alert.filter),
        group: ''
      });
      setQueries(queryData);
      setAlertLimit(viewAlertDetails?.event_alert?.alert_limit);
      setCoolDownTime(viewAlertDetails?.event_alert?.cool_down_time / 3600);
      setNotRepeat(viewAlertDetails?.event_alert?.repeat_alerts);
      setNotifications(viewAlertDetails?.event_alert?.notifications);
      const messageProperty = getGroupByFromState(
        viewAlertDetails?.event_alert?.message_property
      );
      messageProperty.forEach((property) => pushGroupBy(property));

      // webhook settings
      if (viewAlertDetails?.event_alert?.webhook) {
        setWebhookEnabled(viewAlertDetails?.event_alert?.webhook);
        setWebhookUrl(viewAlertDetails?.event_alert?.url);
        setFinalWebhookUrl(viewAlertDetails?.event_alert?.url);
        setConfirmBtn(false);
        setTestMessageBtn(true);
        setTestMassageResponse('');
        setConfirmedMessageBtn(false);
        setShowEditBtn(true);
        setDisbleWebhookInput(true);
        setHideTestMessageBtn(false);
      } else {
        setWebhookEnabled(viewAlertDetails?.event_alert?.webhook);
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
    }
  }, [viewAlertDetails, alertState]);

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
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={queryChange}
            groupAnalysis={queryOptions.group_analysis}
          />
        </div>
      );
    });

    if (queries.length < 1) {
      blockList.push(
        <div key='init'>
          <QueryBlock
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={queryChange}
            groupBy={queryOptions.groupBy}
            groupAnalysis={queryOptions.group_analysis}
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
        hideText={true}
      />
    ) : null;

  const groupByItems = () => {
    const groupByEvents = [];

    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
        .filter(
          (gbp) => gbp.eventName === queries?.[0].label && gbp.eventIndex === 1
        )
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
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
                hideText={true}
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

  const viewGroupByItems = (groupBy) => {
    const groupByEvents = [];

    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
        .filter((gbp) => gbp.eventName === viewAlertDetails?.event_alert?.event)
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
            <div key={gbpIndex} className='fa--query_block--filters'>
              <EventGroupBlock
                index={gbp.groupByIndex}
                grpIndex={gbpIndex}
                eventIndex={1}
                groupByEvent={orgGbp}
                event={viewAlertDetails?.event_alert?.event}
                delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                setGroupState={pushGroupBy}
                closeDropDown={() => setGroupByDDVisible(false)}
                hideText={true}
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

  const getGroupByFromProperties = (appliedGroupBy) => {
    return appliedGroupBy.map((opt) => {
      let gbpReq = {};
      if (opt.eventIndex) {
        gbpReq = {
          pr: opt.property,
          en: opt.prop_category === 'group' ? 'user' : opt.prop_category,
          pty: opt.prop_type,
          ena: opt.eventName,
          eni: opt.eventIndex
        };
      } else {
        gbpReq = {
          pr: opt.property,
          en: opt.prop_category === 'group' ? 'user' : opt.prop_category,
          pty: opt.prop_type,
          ena: opt.eventName
        };
      }
      if (opt.prop_type === 'datetime') {
        opt.grn ? (gbpReq.grn = opt.grn) : (gbpReq.grn = 'day');
      }
      if (opt.prop_type === 'numerical') {
        opt.gbty ? (gbpReq.gbty = opt.gbty) : (gbpReq.gbty = '');
      }
      return gbpReq;
    });
  };

  const getGroupByFromState = (appliedGroupBy) => {
    return appliedGroupBy.map((opt) => {
      let gbpReq = {};
      if (opt.eni) {
        gbpReq = {
          property: opt.pr,
          prop_category: opt.en === 'group' ? 'user' : opt.en,
          prop_type: opt.pty,
          eventName: opt.ena,
          eventIndex: opt.eni
        };
      } else {
        gbpReq = {
          property: opt.pr,
          prop_category: opt.en === 'group' ? 'user' : opt.en,
          prop_type: opt.pty,
          eventName: opt.ena
        };
      }
      if (opt.pty === 'datetime') {
        opt.grn ? (gbpReq.grn = opt.grn) : (gbpReq.grn = 'day');
      }
      if (opt.pty === 'numerical') {
        opt.gbty ? (gbpReq.gbty = opt.gbty) : (gbpReq.gbty = '');
      }
      return gbpReq;
    });
  };

  const onReset = () => {
    setQueries([]);
    setSlackEnabled(false);
    setAlertLimit(5);
    setNotRepeat(false);
    setNotifications(false);
    setSelectedChannel([]);
    setSaveSelectedChannel([]);
    form.resetFields();
    setAlertState({ state: 'list', index: 0 });
    resetGroupBy();
    setEventPropertyDetails({});
    setBreakdownOptions([]);
  };

  const onFinish = (data) => {
    setLoading(true);

    let breakDownProperties = [];
    if (
      queries.length > 0 &&
      (EventPropertyDetails?.name || EventPropertyDetails?.[1])
    ) {
      breakDownProperties = [
        {
          eventName: queries?.[0].label,
          property: EventPropertyDetails?.name || EventPropertyDetails?.[1],
          prop_type:
            EventPropertyDetails?.data_type || EventPropertyDetails?.[2],
          prop_category: 'event'
        }
      ];
    }

    if (
      queries.length > 0 &&
      (slackEnabled || webhookEnabled || teamsEnabled) &&
      (saveSelectedChannel.length > 0 ||
        finalWebhookUrl !== '' ||
        teamsSaveSelectedChannel.length > 0)
    ) {
      let payload = {
        title: data?.alert_name,
        event: queries[0]?.label,
        filter: getEventsWithProperties(queries),
        notifications: notifications,
        message: data?.message,
        message_property:
          groupBy && groupBy.length && groupBy[0] && groupBy[0].property
            ? getGroupByFromProperties(
                groupBy
                  .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
                  .filter(
                    (gbp) =>
                      gbp.eventName === queries[0]?.label &&
                      gbp.eventIndex === 1
                  )
              )
            : [],
        alert_limit: alertLimit,
        repeat_alerts: notRepeat,
        cool_down_time: coolDownTime * 3600,
        breakdown_properties: getGroupByFromProperties(breakDownProperties),
        slack: slackEnabled,
        slack_channels: saveSelectedChannel,
        webhook: webhookEnabled,
        url: finalWebhookUrl,
        teams: teamsEnabled,
        teams_channels_config: {
          team_id: selectedWorkspace?.id,
          team_name: selectedWorkspace?.name,
          team_channel_list: teamsSaveSelectedChannel
        }
      };

      if (alertState?.state === 'edit') {
        editEventAlert(activeProject.id, payload, viewAlertDetails?.id)
          .then((res) => {
            setLoading(false);
            fetchEventAlerts(activeProject.id);
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
            fetchEventAlerts(activeProject.id);
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

  const onConnectSlack = () => {
    enableSlackIntegration(activeProject.id)
      .then((r) => {
        if (r.status == 200) {
          window.open(r.data.redirectURL, '_blank');
        }
        if (r.status >= 400) {
          message.error('Error fetching slack redirect url');
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

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettings?.int_slack) {
      fetchSlackChannels(activeProject.id);
    }
  }, [activeProject, projectSettings?.int_slack, slackEnabled]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettings?.int_teams && teamsEnabled) {
      fetchTeamsWorkspace(activeProject.id)
        .then((res) => {
          if (res.ok) {
            let tempArr = [];
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

  useEffect(() => {
    if (slack?.length > 0) {
      let tempArr = [];
      for (let i = 0; i < slack.length; i++) {
        tempArr.push({
          name: slack[i].name,
          id: slack[i].id,
          is_private: slack[i].is_private
        });
      }
      setChannelOpts(tempArr);
    }
  }, [activeProject, agent_details, slack]);

  useEffect(() => {
    if (teams?.length > 0 && selectedWorkspace) {
      let tempArr = [];
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
      event: queries[0]?.label,
      message_property:
        groupBy && groupBy.length && groupBy[0] && groupBy[0].property
          ? getGroupByFromProperties(
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
      secret: ''
    };
    testWebhhookUrl(activeProject?.id, payload)
      .then((res) => {
        setTestMassageResponse(res?.data);
      })
      .catch((err) => {
        message.error(err?.data?.error);
      });
  };

  const handleClickConfirmBtn = () => {
    setConfirmedMessageBtn(true);
    setDisbleWebhookInput(true);
    setTimeout(() => {
      setConfirmedMessageBtn(false);
      setShowEditBtn(true);
      setTestMassageResponse('');
      setTestMessageBtn(true);
      setFinalWebhookUrl(webhookUrl);
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

  const toggleAlertEnabled = (checked) => {
    if (!checked) {
      showEnableWidgetModal(true);
    } else {
      const status = 'active';
      const id = viewAlertDetails?.id;
      updateEventAlertStatus(activeProject?.id, id, status)
        .then((res) => {
          setisAlertEnabled(true);
          fetchEventAlerts(activeProject.id);
          message.success('Successfully enabled alerts.');
        })
        .catch((err) => {
          console.log('Oops! something went wrong-->', err);
          message.error('Oops! something went wrong. ' + err?.data?.error);
        });
    }
  };

  const propOption = (item) => {
    return (
      <Tooltip title={item} placement={'right'}>
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
  };

  const renderEventForm = () => {
    return (
      <>
        <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          loading={loading}
        >
          <Row>
            <Col span={12}>
              <Text
                type={'title'}
                level={3}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Create new alert
              </Text>
            </Col>
            <Col span={12}>
              <div className={'flex justify-end'}>
                <Button
                  size={'large'}
                  disabled={loading}
                  onClick={() => {
                    onReset();
                  }}
                >
                  Cancel
                </Button>
                <Button
                  size={'large'}
                  disabled={loading}
                  loading={loading}
                  className={'ml-2'}
                  type={'primary'}
                  htmlType='submit'
                >
                  Save
                </Button>
              </div>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Alert name
              </Text>
            </Col>
          </Row>
          <Row className={'mt-2'}>
            <Col span={8} className={'m-0'}>
              <Form.Item
                name='alert_name'
                className={'m-0'}
                rules={[{ required: true, message: 'Please enter alert name' }]}
              >
                <Input
                  className={'fa-input'}
                  placeholder={'Enter name'}
                  onChange={(e) => setAlertName(e.target.value)}
                  ref={inputComponentRef}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Target Event
              </Text>
            </Col>
          </Row>
          <Row className={'m-0'}>
            <Col span={24}>
              <Form.Item name='event_name' className={'m-0'}>
                {queryList()}
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={16} className={'m-0'}>
              <Form.Item name='repeat_alerts' className={'m-0'}>
                <Checkbox
                  checked={notRepeat}
                  onChange={(e) => setNotRepeat(e.target.checked)}
                >
                  Do not repeat alerts more than once within
                </Checkbox>
                <div className='inline -ml-2'>
                  <Select
                    bordered={false}
                    size='small'
                    className='m-0 inline'
                    style={{
                      width: 110
                    }}
                    defaultValue={0.5}
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
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'m-0'}>
            <Col span={16}>
              <Form.Item name='event_property' className='m-0 inline'>
                <Text
                  type={'title'}
                  level={7}
                  color={'grey-2'}
                  extraClass={'m-0 inline ml-10'}
                >
                  for the same value of
                </Text>

                <div className='inline ml-2'>
                  <Select
                    className='inline fa-select'
                    style={{
                      width: 200
                    }}
                    // dropdownMatchSelectWidth={false}
                    disabled={!queries[0]?.label}
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
                    {breakdownOptions?.map((item) => {
                      return (
                        <Option
                          key={item[1]}
                          value={item[0]}
                          name={item[1]}
                          data_type={item[2]}
                        >
                          {propOption(item[0])}
                        </Option>
                      );
                    })}
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-2'}>
            <Col span={16} className={'m-0'}>
              <Form.Item name='notifications' className={'m-0'}>
                <Checkbox
                  checked={notifications}
                  onChange={(e) => setNotifications(e.target.checked)}
                >
                  Set limit for alerts per day to
                </Checkbox>
                <div className='inline -ml-2'>
                  <Select
                    bordered={false}
                    size='small'
                    className='m-0 inline'
                    style={{
                      width: 100
                    }}
                    defaultValue={5}
                    onChange={handleAlertLimit}
                  >
                    <Option value={5}>5 alerts</Option>
                    <Option value={10}>10 alerts</Option>
                    <Option value={15}>15 alerts</Option>
                    <Option value={20}>20 alerts</Option>
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Add a payload
              </Text>
            </Col>
          </Row>
          <Row className='m-0'>
            <Col span={24}>
              <Text
                type='paragraph'
                mini
                extraClass='m-0'
                color='grey'
                lineHeight='medium'
              >
                To send this event to other apps, choose what values and
                properties will be made available. Add a custom message if you
                need one.
              </Text>
            </Col>
          </Row>
          <Row className={'mt-2'}>
            <Col span={8} className={'ml-4'}>
              <div>
                <Text type={'title'} level={7} extraClass={'m-0 inline'}>
                  Add a message
                </Text>
                <Popover
                  placement='rightTop'
                  overlayInnerStyle={{ width: '340px' }}
                  title={null}
                  content={
                    <div className='m-0 m-2'>
                      <p className='m-0 text-gray-900 text-base font-bold'>
                        Your notification inside slack
                      </p>
                      <p className='m-0 mb-2 text-gray-700'>
                        As events across your marketing activities happen, get
                        alerts that motivate actions right inside Slack
                      </p>
                      <img
                        className='m-0'
                        src='../../../../../assets/icons/Slackmock.svg'
                      ></img>
                    </div>
                  }
                >
                  <div className='inline ml-1'>
                    <SVG
                      name='InfoCircle'
                      size={18}
                      color='#8692A3'
                      extraClass={'inline'}
                    />
                  </div>
                </Popover>
              </div>
              <Form.Item name='message' className={'m-0'}>
                <TextArea
                  className={'fa-input'}
                  placeholder={'Enter Message (max 300 characters)'}
                  onChange={(e) => setAlertMessage(e.target.value)}
                  maxLength={300}
                />
              </Form.Item>
            </Col>
          </Row>

          {queries.length > 0 && (
            <Row className={'mt-4'}>
              <Col span={12} className={'ml-4'}>
                <div>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0 inline mb-1 mr-1'}
                  >
                    Choose properties to use as a payload to other apps
                  </Text>
                  <Popover
                    placement='rightTop'
                    overlayInnerStyle={{ width: '300px' }}
                    title={null}
                    content={
                      <p className='m-0 m-2 text-gray-700'>
                        In Slack, you’ll get these values on your channel. With
                        a webhook, use these properties to power your own
                        workflows.
                      </p>
                    }
                  >
                    <div className='inline'>
                      <SVG
                        name='InfoCircle'
                        size={18}
                        color='#8692A3'
                        extraClass={'inline'}
                      />
                    </div>
                  </Popover>
                </div>
                <div className='fa--query_block_section borderless no-padding mt-0'>
                  {groupByItems()}
                </div>
                <Button
                  type='text'
                  style={{ color: '#8692A3' }}
                  icon={<SVG name='plus' color='#8692A3' />}
                  onClick={() => addGroupBy()}
                >
                  Add a Property
                </Button>
              </Col>
            </Row>
          )}

          <Row className={''}>
            <Col span={24}>
              <div className={'border-top--thin-2 pb-6 mt-6'} />
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Destinations
              </Text>
            </Col>
          </Row>
          <Row className='m-0'>
            <Col span={24}>
              <Text
                type='paragraph'
                mini
                extraClass='m-0'
                color='grey'
                lineHeight='medium'
              >
                Choose the destinations for your alerts. The properties you
                added above will be made available in these destinations.
              </Text>
            </Col>
          </Row>

          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'slack'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Slack
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        Post to slack when events you care about happen.
                        Motivate the right actions.
                      </Text>
                    </div>
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='slack_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          onChange={(checked) => setSlackEnabled(checked)}
                          checked={slackEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {slackEnabled && !projectSettings?.int_slack && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Slack is not integrated, Do you want to integrate with
                      your slack account now?
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button onClick={onConnectSlack}>
                      <SVG name={'Slack'} />
                      Connect to slack
                    </Button>
                  </Col>
                </Row>
              </div>
            )}
            {slackEnabled && projectSettings?.int_slack && (
              <div className='p-4'>
                {saveSelectedChannel.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {saveSelectedChannel.length > 1
                            ? 'Selected Channels'
                            : 'Selected Channel'}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {saveSelectedChannel.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
                {!saveSelectedChannel.length > 0 ? (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setShowSelectChannelsModal(true)}
                      >
                        Select Channel
                      </Button>
                    </Col>
                  </Row>
                ) : (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setShowSelectChannelsModal(true)}
                      >
                        {saveSelectedChannel.length > 1
                          ? 'Manage Channels'
                          : 'Manage Channel'}
                      </Button>
                    </Col>
                  </Row>
                )}
              </div>
            )}
          </div>

          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'Webhook'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Webhook
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        When this alert happens, send this information to other
                        apps to enable more flows.
                      </Text>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        <span className='font-bold'>Note:</span> Please add
                        payload to enable this option.
                      </Text>
                    </div>
                    {isWebHookFeatureLocked && (
                      <div className='p-2'>
                        <UpgradeButton />
                      </div>
                    )}
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='webhook_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          disabled={
                            !(
                              groupBy &&
                              groupBy.length &&
                              groupBy[0] &&
                              groupBy[0].property
                            ) || isWebHookFeatureLocked
                          }
                          onChange={(checked) => setWebhookEnabled(checked)}
                          checked={webhookEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {webhookEnabled && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={12} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={7}
                      weight='medium'
                      extraClass={'m-0'}
                    >
                      Paste your webhook URL here
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-1 ml-2'}>
                  <Col span={10}>
                    <Input
                      className='fa-input'
                      size='large'
                      placeholder='Webhook URL'
                      disabled={disbleWebhookInput}
                      ref={webhookRef}
                      value={webhookUrl}
                      onChange={(e) => {
                        setWebhookUrl(e.target.value);
                        setConfirmBtn(false);
                        setTestMessageBtn(false);
                      }}
                      onBlur={() => {
                        if (webhookUrl === '') {
                          setTestMessageBtn(true);
                          setConfirmBtn(true);
                        }
                        if (showEditBtn && webhookUrl === finalWebhookUrl) {
                          setHideTestMessageBtn(true);
                          setConfirmBtn(false);
                          setDisbleWebhookInput(true);
                        }
                      }}
                    ></Input>
                  </Col>
                  <Col span={6} className={'m-0 ml-2'}>
                    {!confirmedMessageBtn && !showEditBtn ? (
                      <Button
                        type='link'
                        disabled={confirmBtn}
                        onClick={() => handleClickConfirmBtn()}
                        size='large'
                      >
                        Confirm
                      </Button>
                    ) : confirmedMessageBtn && !showEditBtn ? (
                      <Button
                        type='link'
                        disabled
                        onClick={() => handleClickConfirmBtn()}
                        size='large'
                        icon={
                          <SVG
                            name={'Checkmark'}
                            size={16}
                            color={'#52C41A'}
                            extraClass={'m-0'}
                          />
                        }
                      >
                        Confirmed
                      </Button>
                    ) : (
                      <Button
                        type='link'
                        disabled={confirmBtn}
                        onClick={() => {
                          setDisbleWebhookInput(false);
                          setConfirmBtn(true);
                          setHideTestMessageBtn(true);
                          setTimeout(() => {
                            webhookRef.current.focus();
                          }, 200);
                        }}
                        size='large'
                      >
                        Edit
                      </Button>
                    )}
                  </Col>
                </Row>
                {hideTestMessageBtn && (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={24} className={'m-0'}>
                      {testMessageResponse ? (
                        <div>
                          <div className='inline'>
                            <SVG
                              name={'CheckCircle'}
                              size={16}
                              extraClass={'m-0 inline'}
                            />
                            <Text
                              type={'title'}
                              level={7}
                              extraClass={'m-0 ml-1 inline'}
                            >
                              We've sent a sample message to this endpoint.
                              Check and hit 'Confirm' if everything is alright!
                            </Text>
                          </div>
                          <div className='inline'>
                            <Button
                              type='link'
                              style={{
                                backgroundColor: 'white',
                                borderStyle: 'none'
                              }}
                              size='small'
                              disabled={testMessageBtn}
                              onClick={() => handleTestWebhook()}
                              icon={
                                <SVG
                                  name={'PaperPlane'}
                                  size={18}
                                  color={
                                    testMessageBtn ? '#00000040' : '#1e89ff'
                                  }
                                  extraClass={'-mt-1'}
                                />
                              }
                            >
                              Try Again
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <Button
                          type='link'
                          disabled={testMessageBtn}
                          style={{
                            backgroundColor: 'white',
                            borderStyle: 'none'
                          }}
                          size='small'
                          onClick={() => handleTestWebhook()}
                          icon={
                            <SVG
                              name={'PaperPlane'}
                              size={18}
                              color={testMessageBtn ? '#00000040' : '#1e89ff'}
                              extraClass={'-mt-1'}
                            />
                          }
                        >
                          Test this with a sample message
                        </Button>
                      )}
                    </Col>
                  </Row>
                )}
                <Row className='mt-3 ml-2'>
                  <Col>
                    <Text
                      type='paragraph'
                      mini
                      extraClass='m-0'
                      color='grey'
                      lineHeight='medium'
                    >
                      Note that if you edit this alert or its payload in the
                      future, you must reconfigure the flows to support these
                      changes
                    </Text>
                  </Col>
                </Row>
              </div>
            )}
          </div>

          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'MSTeam'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Teams
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        Post to teams when events you care about happen.
                        Motivate the right actions.
                      </Text>
                    </div>
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='teams_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          onChange={(checked) => setTeamsEnabled(checked)}
                          checked={teamsEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {teamsEnabled && !projectSettings?.int_teams && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Teams is not integrated, Do you want to integrate with
                      your Microsoft Teams account now?
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button onClick={onConnectMSTeams}>
                      <SVG name={'MSTeam'} size={20} />
                      Connect to Teams
                    </Button>
                  </Col>
                </Row>
              </div>
            )}
            {teamsEnabled && projectSettings?.int_teams && (
              <div className='p-4'>
                {teamsSaveSelectedChannel.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {teamsSaveSelectedChannel.length > 1
                            ? `Selected channels from the "${selectedWorkspace?.name}"`
                            : `Selected channels from the "${selectedWorkspace?.name}"`}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {teamsSaveSelectedChannel.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
                {!teamsSaveSelectedChannel.length > 0 ? (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setTeamsShowSelectChannelsModal(true)}
                      >
                        Select Channel
                      </Button>
                    </Col>
                  </Row>
                ) : (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setTeamsShowSelectChannelsModal(true)}
                      >
                        {teamsSaveSelectedChannel.length > 1
                          ? 'Manage Channels'
                          : 'Manage Channel'}
                      </Button>
                    </Col>
                  </Row>
                )}
              </div>
            )}
          </div>
        </Form>
      </>
    );
  };

  const renderEventEdit = () => {
    return (
      <>
        <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          loading={loading}
        >
          <Row>
            <Col span={12}>
              <Text
                type={'title'}
                level={3}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0 inline'}
              >
                Triggers alerts from an event
              </Text>
              <div className={'inline ml-3'}>
                <Switch
                  checkedChildren='On'
                  unCheckedChildren='OFF'
                  onChange={toggleAlertEnabled}
                  checked={isAlertEnabled}
                />
              </div>
            </Col>
            <Col span={12}>
              <div className={'flex justify-end'}>
                <Button
                  size={'large'}
                  disabled={loading}
                  onClick={() => {
                    onReset();
                  }}
                >
                  Cancel
                </Button>
                <Button
                  size={'large'}
                  disabled={loading}
                  loading={loading}
                  className={'ml-2'}
                  type={'primary'}
                  htmlType='submit'
                >
                  Save
                </Button>
              </div>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Alert name
              </Text>
            </Col>
          </Row>
          <Row className={'mt-2'}>
            <Col span={8} className={'m-0'}>
              <Form.Item
                name='alert_name'
                className={'m-0'}
                initialValue={viewAlertDetails?.title}
                rules={[{ required: true, message: 'Please enter alert name' }]}
              >
                <Input
                  className={'fa-input'}
                  placeholder={'Enter name'}
                  onChange={(e) => setAlertName(e.target.value)}
                  ref={inputComponentRef}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Target Event
              </Text>
            </Col>
          </Row>
          <Row className={'m-0'}>
            <Col span={24}>
              <Form.Item name='event_name' className={'m-0'}>
                {queryList()}
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={16} className={'m-0'}>
              <Form.Item name='repeat_alerts' className={'m-0'}>
                <Checkbox
                  checked={notRepeat}
                  onChange={(e) => setNotRepeat(e.target.checked)}
                >
                  Do not repeat alerts more than once within
                </Checkbox>
                <div className='inline -ml-2'>
                  <Select
                    bordered={false}
                    size='small'
                    className='m-0 inline'
                    style={{
                      width: 110
                    }}
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
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'m-0'}>
            <Col span={16}>
              <Form.Item name='event_property' className='m-0 inline'>
                <Text
                  type={'title'}
                  level={7}
                  color={'grey-2'}
                  extraClass={'m-0 inline ml-10'}
                >
                  for the same value of
                </Text>

                <div className='inline ml-2'>
                  <Select
                    className='inline fa-select'
                    style={{
                      width: 200
                    }}
                    dropdownMatchSelectWidth={false}
                    disabled={!queries[0]?.label}
                    value={EventPropertyDetails}
                    onChange={(value, details) => {
                      setEventPropertyDetails(details);
                      setNotRepeat(true);
                    }}
                    placeholder='Select Property'
                    showSearch
                    filterOption={(input, option) =>
                      option.children
                        .toLowerCase()
                        .indexOf(input.toLowerCase()) >= 0
                    }
                  >
                    {breakdownOptions?.map((item) => {
                      return (
                        <Option
                          key={item[1]}
                          value={item[0]}
                          name={item[1]}
                          data_type={item[2]}
                        >
                          {item[0]}
                        </Option>
                      );
                    })}
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-2'}>
            <Col span={16} className={'m-0'}>
              <Form.Item name='notifications' className={'m-0'}>
                <Checkbox
                  checked={notifications}
                  onChange={(e) => setNotifications(e.target.checked)}
                >
                  Set limit for alerts per day to
                </Checkbox>
                <div className='inline -ml-2'>
                  <Select
                    bordered={false}
                    size='small'
                    className='m-0 inline'
                    style={{
                      width: 100
                    }}
                    value={alertLimit}
                    onChange={handleAlertLimit}
                  >
                    <Option value={5}>5 alerts</Option>
                    <Option value={10}>10 alerts</Option>
                    <Option value={15}>15 alerts</Option>
                    <Option value={20}>20 alerts</Option>
                  </Select>
                </div>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Add a payload
              </Text>
            </Col>
          </Row>
          <Row className='m-0'>
            <Col span={24}>
              <Text
                type='paragraph'
                mini
                extraClass='m-0'
                color='grey'
                lineHeight='medium'
              >
                To send this event to other apps, choose what values and
                properties will be made available. Add a custom message if you
                need one.
              </Text>
            </Col>
          </Row>
          <Row className={'mt-2'}>
            <Col span={8} className={'ml-4'}>
              <div>
                <Text type={'title'} level={7} extraClass={'m-0 inline'}>
                  Add a message
                </Text>
                <Popover
                  placement='rightTop'
                  overlayInnerStyle={{ width: '340px' }}
                  title={null}
                  content={
                    <div className='m-0 m-2'>
                      <p className='m-0 text-gray-900 text-base font-bold'>
                        Your notification inside slack
                      </p>
                      <p className='m-0 mb-2 text-gray-700'>
                        As events across your marketing activities happen, get
                        alerts that motivate actions right inside Slack
                      </p>
                      <img
                        className='m-0'
                        src='../../../../../assets/icons/Slackmock.svg'
                      ></img>
                    </div>
                  }
                >
                  <div className='inline ml-1'>
                    <SVG
                      name='InfoCircle'
                      size={18}
                      color='#8692A3'
                      extraClass={'inline'}
                    />
                  </div>
                </Popover>
              </div>
              <Form.Item
                name='message'
                initialValue={viewAlertDetails?.event_alert?.message}
                className={'m-0'}
              >
                <TextArea
                  className={'fa-input'}
                  placeholder={'Enter Message (max 300 characters)'}
                  onChange={(e) => setAlertMessage(e.target.value)}
                  maxLength={300}
                />
              </Form.Item>
            </Col>
          </Row>

          {queries.length > 0 && (
            <Row className={'mt-4'}>
              <Col span={12} className={'ml-4'}>
                <div>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0 inline mb-1 mr-1'}
                  >
                    Choose properties to use as a payload to other apps
                  </Text>
                  <Popover
                    placement='rightTop'
                    overlayInnerStyle={{ width: '300px' }}
                    title={null}
                    content={
                      <p className='m-0 m-2 text-gray-700'>
                        In Slack, you’ll get these values on your channel. With
                        a webhook, use these properties to power your own
                        workflows.
                      </p>
                    }
                  >
                    <div className='inline'>
                      <SVG
                        name='InfoCircle'
                        size={18}
                        color='#8692A3'
                        extraClass={'inline'}
                      />
                    </div>
                  </Popover>
                </div>
                <div>{groupByItems()}</div>
                <Button
                  type='text'
                  style={{ color: '#8692A3' }}
                  icon={<SVG name='plus' color='#8692A3' />}
                  onClick={() => addGroupBy()}
                >
                  Add a Property
                </Button>
              </Col>
            </Row>
          )}

          <Row className={''}>
            <Col span={24}>
              <div className={'border-top--thin-2 pb-6 mt-6'} />
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Destinations
              </Text>
            </Col>
          </Row>
          <Row className='m-0'>
            <Col span={24}>
              <Text
                type='paragraph'
                mini
                extraClass='m-0'
                color='grey'
                lineHeight='medium'
              >
                Choose the destinations for your alerts. The properties you
                added above will be made available in these destinations.
              </Text>
            </Col>
          </Row>

          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'slack'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Slack
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        Post to slack when events you care about happen.
                        Motivate the right actions.
                      </Text>
                    </div>
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='slack_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          onChange={(checked) => setSlackEnabled(checked)}
                          checked={slackEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {slackEnabled && !projectSettings?.int_slack && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Slack is not integrated, Do you want to integrate with
                      your slack account now?
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button onClick={onConnectSlack}>
                      <SVG name={'Slack'} />
                      Connect to slack
                    </Button>
                  </Col>
                </Row>
              </div>
            )}
            {slackEnabled && projectSettings?.int_slack && (
              <div className='p-4'>
                {saveSelectedChannel.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {saveSelectedChannel.length > 1
                            ? 'Selected Channels'
                            : 'Selected Channel'}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {saveSelectedChannel.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
                {!saveSelectedChannel.length > 0 ? (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setShowSelectChannelsModal(true)}
                      >
                        Select Channel
                      </Button>
                    </Col>
                  </Row>
                ) : (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setShowSelectChannelsModal(true)}
                      >
                        {saveSelectedChannel.length > 1
                          ? 'Manage Channels'
                          : 'Manage Channel'}
                      </Button>
                    </Col>
                  </Row>
                )}
              </div>
            )}
          </div>

          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'Webhook'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Webhook
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        When this alert happens, send this information to other
                        apps to enable more flows.
                      </Text>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        <span className='font-bold'>Note:</span> Please add
                        payload to enable this option.
                      </Text>
                    </div>
                    {isWebHookFeatureLocked && (
                      <div className='p-2'>
                        <UpgradeButton />
                      </div>
                    )}
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='webhook_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          disabled={
                            !(
                              groupBy &&
                              groupBy.length &&
                              groupBy[0] &&
                              groupBy[0].property
                            )
                          }
                          onChange={(checked) => setWebhookEnabled(checked)}
                          checked={webhookEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {webhookEnabled && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={12} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={7}
                      weight='medium'
                      extraClass={'m-0'}
                    >
                      Paste your webhook URL here
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-1 ml-2'}>
                  <Col span={10}>
                    <Input
                      className='fa-input'
                      size='large'
                      placeholder='Webhook URL'
                      disabled={disbleWebhookInput}
                      ref={webhookRef}
                      value={webhookUrl}
                      onChange={(e) => {
                        setWebhookUrl(e.target.value);
                        setConfirmBtn(false);
                        setTestMessageBtn(false);
                      }}
                      onBlur={() => {
                        if (webhookUrl === '') {
                          setTestMessageBtn(true);
                          setConfirmBtn(true);
                        }
                        if (showEditBtn && webhookUrl === finalWebhookUrl) {
                          setHideTestMessageBtn(true);
                          setConfirmBtn(false);
                          setDisbleWebhookInput(true);
                        }
                      }}
                    ></Input>
                  </Col>
                  <Col span={6} className={'m-0 ml-2'}>
                    {!confirmedMessageBtn && !showEditBtn ? (
                      <Button
                        type='link'
                        disabled={confirmBtn}
                        onClick={() => handleClickConfirmBtn()}
                        size='large'
                      >
                        Confirm
                      </Button>
                    ) : confirmedMessageBtn && !showEditBtn ? (
                      <Button
                        type='link'
                        disabled
                        onClick={() => handleClickConfirmBtn()}
                        size='large'
                        icon={
                          <SVG
                            name={'Checkmark'}
                            size={16}
                            color={'#52C41A'}
                            extraClass={'m-0'}
                          />
                        }
                      >
                        Confirmed
                      </Button>
                    ) : (
                      <Button
                        type='link'
                        disabled={confirmBtn}
                        onClick={() => {
                          setDisbleWebhookInput(false);
                          setConfirmBtn(true);
                          setHideTestMessageBtn(true);
                          setTimeout(() => {
                            webhookRef.current.focus();
                          }, 200);
                        }}
                        size='large'
                      >
                        Edit
                      </Button>
                    )}
                  </Col>
                </Row>
                {hideTestMessageBtn && (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={24} className={'m-0'}>
                      {testMessageResponse ? (
                        <div>
                          <div className='inline'>
                            <SVG
                              name={'CheckCircle'}
                              size={16}
                              extraClass={'m-0 inline'}
                            />
                            <Text
                              type={'title'}
                              level={7}
                              extraClass={'m-0 ml-1 inline'}
                            >
                              We've sent a sample message to this endpoint.
                              Check and hit 'Confirm' if everything is alright!
                            </Text>
                          </div>
                          <div className='inline'>
                            <Button
                              type='link'
                              style={{
                                backgroundColor: 'white',
                                borderStyle: 'none'
                              }}
                              size='small'
                              disabled={testMessageBtn}
                              onClick={() => handleTestWebhook()}
                              icon={
                                <SVG
                                  name={'PaperPlane'}
                                  size={18}
                                  color={
                                    testMessageBtn ? '#00000040' : '#1e89ff'
                                  }
                                  extraClass={'-mt-1'}
                                />
                              }
                            >
                              Try Again
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <Button
                          type='link'
                          disabled={testMessageBtn}
                          style={{
                            backgroundColor: 'white',
                            borderStyle: 'none'
                          }}
                          size='small'
                          onClick={() => handleTestWebhook()}
                          icon={
                            <SVG
                              name={'PaperPlane'}
                              size={18}
                              color={testMessageBtn ? '#00000040' : '#1e89ff'}
                              extraClass={'-mt-1'}
                            />
                          }
                        >
                          Test this with a sample message
                        </Button>
                      )}
                    </Col>
                  </Row>
                )}
                <Row className='mt-3 ml-2'>
                  <Col>
                    <Text
                      type='paragraph'
                      mini
                      extraClass='m-0'
                      color='grey'
                      lineHeight='medium'
                    >
                      Note that if you edit this alert or its payload in the
                      future, you must reconfigure the flows to support these
                      changes
                    </Text>
                  </Col>
                </Row>
              </div>
            )}
          </div>
          <div className='border rounded mt-3'>
            <div style={{ backgroundColor: '#fafafa' }}>
              <Row className={'ml-2'}>
                <Col span={20}>
                  <div className='flex justify-between p-3'>
                    <div className='flex'>
                      <Avatar
                        size={40}
                        shape='square'
                        icon={<SVG name={'MSTeam'} size={40} color='purple' />}
                        style={{ backgroundColor: '#F5F6F8' }}
                      />
                    </div>
                    <div className='flex flex-col justify-start items-start ml-2 w-full'>
                      <div className='flex flex-row items-center justify-start'>
                        <Text
                          type='title'
                          level={7}
                          weight='medium'
                          extraClass='m-0'
                        >
                          Teams
                        </Text>
                      </div>
                      <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        Post to teams when events you care about happen.
                        Motivate the right actions.
                      </Text>
                    </div>
                  </div>
                </Col>
                <Col className={'m-0 mt-4'}>
                  <Form.Item name='teams_enabled' className={'m-0'}>
                    <div span={24} className={'flex flex-start items-center'}>
                      <Text
                        type={'title'}
                        level={7}
                        weight='medium'
                        extraClass={'m-0 mr-2'}
                      >
                        Enable
                      </Text>
                      <span style={{ width: '50px' }}>
                        <Switch
                          checkedChildren='On'
                          unCheckedChildren='OFF'
                          onChange={(checked) => setTeamsEnabled(checked)}
                          checked={teamsEnabled}
                        />
                      </span>{' '}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </div>
            {teamsEnabled && !projectSettings?.int_teams && (
              <div className='p-4'>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Teams is not integrated, Do you want to integrate with
                      your Microsoft Teams account now?
                    </Text>
                  </Col>
                </Row>
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button onClick={onConnectMSTeams}>
                      <SVG name={'MSTeam'} size={20} />
                      Connect to Teams
                    </Button>
                  </Col>
                </Row>
              </div>
            )}
            {teamsEnabled && projectSettings?.int_teams && (
              <div className='p-4'>
                {teamsSaveSelectedChannel.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {teamsSaveSelectedChannel.length > 1
                            ? `Selected channels from the "${selectedWorkspace?.name}"`
                            : `Selected channels from the "${selectedWorkspace?.name}"`}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {teamsSaveSelectedChannel.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
                {!teamsSaveSelectedChannel.length > 0 ? (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setTeamsShowSelectChannelsModal(true)}
                      >
                        Select Channel
                      </Button>
                    </Col>
                  </Row>
                ) : (
                  <Row className={'mt-2 ml-2'}>
                    <Col span={10} className={'m-0'}>
                      <Button
                        type={'link'}
                        onClick={() => setTeamsShowSelectChannelsModal(true)}
                      >
                        {teamsSaveSelectedChannel.length > 1
                          ? 'Manage Channels'
                          : 'Manage Channel'}
                      </Button>
                    </Col>
                  </Row>
                )}
              </div>
            )}
          </div>
        </Form>
      </>
    );
  };

  const renderEventView = () => {
    return (
      <>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
              View Alert
            </Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button
                size={'large'}
                disabled={loading}
                onClick={() => {
                  setAlertState({ state: 'list', index: 0 });
                }}
              >
                Back
              </Button>
            </div>
          </Col>
        </Row>

        <Row className={'mt-6'}>
          <Col span={18}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Alert name
            </Text>
          </Col>
        </Row>
        <Row className={'mt-4'}>
          <Col span={8} className={'m-0'}>
            <Input
              disabled={true}
              className={'fa-input'}
              value={viewAlertDetails?.title}
            />
          </Col>
        </Row>

        <Row className={'mt-4'}>
          <Col span={18}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Target Event
            </Text>
          </Col>
        </Row>
        <Row className={'m-0 mt-2'}>
          <Col>
            <Button className={`mr-2`} type='link' disabled={true}>
              {eventNames[viewAlertDetails?.event_alert?.event]
                ? eventNames[viewAlertDetails?.event_alert?.event]
                : viewAlertDetails?.event_alert?.event}
            </Button>
          </Col>
        </Row>
        {viewAlertDetails?.event_alert?.filter?.length > 0 && (
          <Row className={'mt-2'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0 my-1'}
              >
                Filters
              </Text>
              <GLobalFilter filters={viewFilter} delFilter={false} viewMode />
            </Col>
          </Row>
        )}
        <Row className={'mt-2'}>
          <Col span={16}>
            <Checkbox
              className='inline'
              disabled={true}
              checked={viewAlertDetails?.event_alert?.repeat_alerts}
            >
              Do not repeat alerts more than once within
            </Checkbox>
            <div className='inline ml-1'>
              <Input
                disabled={true}
                style={{
                  width: 110
                }}
                className={'inline fa-input'}
                value={
                  viewAlertDetails?.event_alert?.cool_down_time / 3600 +
                  ' hours'
                }
              />
            </div>
          </Col>
        </Row>
        <Row className={'m-0 my-2'}>
          <Col span={20}>
            <Text
              type={'title'}
              level={7}
              color={'grey-2'}
              extraClass={'inline m-0 ml-10'}
            >
              for the same value of
            </Text>
            <div className='inline ml-2'>
              <Select
                disabled={true}
                style={{
                  width: 200
                }}
                showArrow={false}
                className={'inline fa-select'}
                value={
                  viewAlertDetails?.event_alert?.breakdown_properties?.[0]?.pr
                }
              >
                <Option
                  value={
                    viewAlertDetails?.event_alert?.breakdown_properties?.[0]?.pr
                  }
                >
                  {propOption(
                    matchEventName(
                      viewAlertDetails?.event_alert?.breakdown_properties?.[0]
                        ?.pr
                    )
                  )}
                </Option>
              </Select>
            </div>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={16}>
            <Checkbox
              className='inline'
              disabled={true}
              checked={viewAlertDetails?.event_alert?.notifications}
            >
              Set limit for alerts per day to
            </Checkbox>
            <div className='inline ml-1'>
              <Input
                disabled={true}
                style={{
                  width: 100
                }}
                className={'inline fa-input'}
                value={viewAlertDetails?.event_alert?.alert_limit + ' alerts'}
              />
            </div>
          </Col>
        </Row>

        <Row className={'mt-4'}>
          <Col span={18}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Add a payload
            </Text>
          </Col>
        </Row>
        <Row className='m-0'>
          <Col span={24}>
            <Text
              type='paragraph'
              mini
              extraClass='m-0'
              color='grey'
              lineHeight='medium'
            >
              To send this event to other apps, choose what values and
              properties will be made available. Add a custom message if you
              need one.
            </Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={8} className={'ml-4'}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Add a message
            </Text>
            <TextArea
              disabled={true}
              className={'fa-input'}
              maxLength={300}
              value={viewAlertDetails?.event_alert?.message}
            />
          </Col>
        </Row>
        {viewAlertDetails?.event_alert?.message_property?.length > 0 && (
          <Row className={'mt-4'}>
            <Col span={12} className={'ml-4'}>
              <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>
                Choose properties to use as a payload to other apps
              </Text>
              <div className='fa--query_block_section borderless no-padding mt-0'>
                {viewGroupByItems(
                  viewAlertDetails?.event_alert?.message_property &&
                    viewAlertDetails?.event_alert?.message_property.length &&
                    viewAlertDetails?.event_alert?.message_property[0] &&
                    getGroupByFromState(
                      viewAlertDetails?.event_alert?.message_property
                        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
                        .filter(
                          (gbp) =>
                            gbp.ena === viewAlertDetails?.event_alert?.event &&
                            gbp.eni === 1
                        )
                    )
                )}
              </div>
            </Col>
          </Row>
        )}

        <Row className={''}>
          <Col span={24}>
            <div className={'border-top--thin-2 pb-6 mt-6'} />
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Destinations
            </Text>
          </Col>
        </Row>
        <Row className='m-0'>
          <Col span={24}>
            <Text
              type='paragraph'
              mini
              extraClass='m-0'
              color='grey'
              lineHeight='medium'
            >
              Choose the destinations for your alerts. The properties you added
              above will be made available in these destinations.
            </Text>
          </Col>
        </Row>

        <div className='border rounded mt-3'>
          <div style={{ backgroundColor: '#fafafa' }}>
            <Row className={'ml-2'}>
              <Col span={20}>
                <div className='flex justify-between p-3'>
                  <div className='flex'>
                    <Avatar
                      size={40}
                      shape='square'
                      icon={<SVG name={'slack'} size={40} color='purple' />}
                      style={{ backgroundColor: '#F5F6F8' }}
                    />
                  </div>
                  <div className='flex flex-col justify-start items-start ml-2 w-full'>
                    <div className='flex flex-row items-center justify-start'>
                      <Text
                        type='title'
                        level={7}
                        weight='medium'
                        extraClass='m-0'
                      >
                        Slack
                      </Text>
                    </div>
                    <Text
                      type='paragraph'
                      mini
                      extraClass='m-0'
                      color='grey'
                      lineHeight='medium'
                    >
                      Post to slack when events you care about happen. Motivate
                      the right actions.
                    </Text>
                  </div>
                </div>
              </Col>
              <Col className={'m-0 mt-4'}>
                <Form.Item name='slack_enabled' className={'m-0'}>
                  <div span={24} className={'flex flex-start items-center'}>
                    <Text
                      type={'title'}
                      level={7}
                      weight='medium'
                      extraClass={'m-0 mr-2'}
                    >
                      Enable
                    </Text>
                    <span style={{ width: '50px' }}>
                      <Switch
                        checkedChildren='On'
                        unCheckedChildren='OFF'
                        disabled
                        checked={viewAlertDetails?.event_alert?.slack}
                      />
                    </span>{' '}
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </div>

          {viewAlertDetails?.event_alert?.slack &&
            viewAlertDetails?.event_alert?.slack_channels.length > 0 && (
              <div className='p-4'>
                {viewSelectedChannels.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {viewSelectedChannels.length > 1
                            ? 'Selected Channels'
                            : 'Selected Channel'}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {viewSelectedChannels.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
              </div>
            )}
        </div>

        <div className='border rounded mt-3'>
          <div style={{ backgroundColor: '#fafafa' }}>
            <Row className={'ml-2'}>
              <Col span={20}>
                <div className='flex justify-between p-3'>
                  <div className='flex'>
                    <Avatar
                      size={40}
                      shape='square'
                      icon={<SVG name={'Webhook'} size={40} color='purple' />}
                      style={{ backgroundColor: '#F5F6F8' }}
                    />
                  </div>
                  <div className='flex flex-col justify-start items-start ml-2 w-full'>
                    <div className='flex flex-row items-center justify-start'>
                      <Text
                        type='title'
                        level={7}
                        weight='medium'
                        extraClass='m-0'
                      >
                        Webhook
                      </Text>
                    </div>
                    <Text
                      type='paragraph'
                      mini
                      extraClass='m-0'
                      color='grey'
                      lineHeight='medium'
                    >
                      When this alert happens, send this information to other
                      apps to enable more flows.
                    </Text>
                  </div>
                </div>
              </Col>
              <Col className={'m-0 mt-4'}>
                <Form.Item name='webhook_enabled' className={'m-0'}>
                  <div span={24} className={'flex flex-start items-center'}>
                    <Text
                      type={'title'}
                      level={7}
                      weight='medium'
                      extraClass={'m-0 mr-2'}
                    >
                      Enable
                    </Text>
                    <span style={{ width: '50px' }}>
                      <Switch
                        checkedChildren='On'
                        unCheckedChildren='OFF'
                        disabled
                        checked={viewAlertDetails?.event_alert?.webhook}
                      />
                    </span>{' '}
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </div>
          {viewAlertDetails?.event_alert?.webhook && (
            <div className='p-4'>
              <Row className={'mt-2 ml-2'}>
                <Col span={12} className={'m-0'}>
                  <Text
                    type={'title'}
                    level={7}
                    weight='medium'
                    extraClass={'m-0'}
                  >
                    Paste your webhook URL here
                  </Text>
                </Col>
              </Row>
              <Row className={'mt-1 ml-2'}>
                <Col span={10}>
                  <Input
                    className='fa-input'
                    size='large'
                    placeholder='Webhook URL'
                    disabled
                    value={viewAlertDetails?.event_alert?.url}
                  ></Input>
                </Col>
                <Col span={6} className={'m-0 ml-2'}>
                  <Button
                    type='link'
                    disabled
                    size='large'
                    icon={
                      <SVG
                        name={'Checkmark'}
                        size={16}
                        color={'#52C41A'}
                        extraClass={'m-0'}
                      />
                    }
                  >
                    Confirmed
                  </Button>
                </Col>
              </Row>
            </div>
          )}
        </div>
        <div className='border rounded mt-3'>
          <div style={{ backgroundColor: '#fafafa' }}>
            <Row className={'ml-2'}>
              <Col span={20}>
                <div className='flex justify-between p-3'>
                  <div className='flex'>
                    <Avatar
                      size={40}
                      shape='square'
                      icon={<SVG name={'MSTeam'} size={40} color='purple' />}
                      style={{ backgroundColor: '#F5F6F8' }}
                    />
                  </div>
                  <div className='flex flex-col justify-start items-start ml-2 w-full'>
                    <div className='flex flex-row items-center justify-start'>
                      <Text
                        type='title'
                        level={7}
                        weight='medium'
                        extraClass='m-0'
                      >
                        Teams
                      </Text>
                    </div>
                    <Text
                      type='paragraph'
                      mini
                      extraClass='m-0'
                      color='grey'
                      lineHeight='medium'
                    >
                      Post to teams when events you care about happen. Motivate
                      the right actions.
                    </Text>
                  </div>
                </div>
              </Col>
              <Col className={'m-0 mt-4'}>
                <Form.Item name='teams_enabled' className={'m-0'}>
                  <div span={24} className={'flex flex-start items-center'}>
                    <Text
                      type={'title'}
                      level={7}
                      weight='medium'
                      extraClass={'m-0 mr-2'}
                    >
                      Enable
                    </Text>
                    <span style={{ width: '50px' }}>
                      <Switch
                        checkedChildren='On'
                        unCheckedChildren='OFF'
                        disabled
                        checked={viewAlertDetails?.event_alert?.teams}
                      />
                    </span>{' '}
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </div>
          {viewAlertDetails?.event_alert?.teams &&
            viewAlertDetails?.event_alert?.teams_channels_config && (
              <div className='p-4'>
                {teamsViewSelectedChannels.length > 0 && (
                  <div>
                    <Row>
                      <Col>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'regular'}
                          extraClass={'m-0 mt-2 ml-2'}
                        >
                          {teamsViewSelectedChannels.length > 1
                            ? `Selected channels from the “${viewAlertDetails?.event_alert?.teams_channels_config?.team_name}”`
                            : `Selected channels from the “${viewAlertDetails?.event_alert?.teams_channels_config?.team_name}”`}
                        </Text>
                      </Col>
                    </Row>
                    <Row
                      className={'rounded border border-gray-200 ml-2 w-2/6'}
                    >
                      <Col className={'m-0'}>
                        {teamsViewSelectedChannels.map((channel, index) => (
                          <div key={index}>
                            <Text
                              type={'title'}
                              level={7}
                              color={'grey'}
                              extraClass={'m-0 ml-4 my-2'}
                            >
                              {'#' + channel.name}
                            </Text>
                          </div>
                        ))}
                      </Col>
                    </Row>
                  </div>
                )}
              </div>
            )}
        </div>

        <Row className={'mt-2'}>
          <Col span={24}>
            <div className={'border-top--thin-2 mt-2 mb-4'} />
            <Button
              type={'text'}
              size={'large'}
              style={{ color: '#EE3C3C' }}
              className={'m-0'}
              onClick={() => showDeleteWidgetModal(viewAlertDetails?.id)}
            >
              <SVG
                name={'Delete1'}
                extraClass={'-mt-1 -mr-1'}
                size={18}
                color={'#EE3C3C'}
              />
              Delete
            </Button>
          </Col>
        </Row>
      </>
    );
  };

  return (
    <div className={'fa-container'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className={'mb-10 pl-4'}>
            {alertState.state == 'add' && renderEventForm()}

            {alertState.state == 'view' && renderEventView()}

            {alertState.state == 'edit' && renderEventEdit()}

            <ConfirmationModal
              visible={deleteWidgetModal ? true : false}
              confirmationText='Do you really want to remove this alert?'
              onOk={confirmDelete}
              onCancel={showDeleteWidgetModal.bind(this, false)}
              title='Remove Alert'
              okText='Confirm'
              cancelText='Cancel'
              confirmLoading={deleteApiCalled}
            />
            <ConfirmationModal
              visible={enableWidgetModal ? true : false}
              confirmationText='Alerts and webhooks from this event will be paused. You can always turn this back on when needed.'
              onOk={confirmPause}
              onCancel={showEnableWidgetModal.bind(this, false)}
              title='Pause Alert?'
              okText='Yes'
              cancelText='Cancel'
              confirmLoading={deleteApiCalled}
            />
          </div>
        </Col>
      </Row>

      <Modal
        title={null}
        visible={showSelectChannelsModal}
        centered={true}
        zIndex={1005}
        width={700}
        onCancel={handleCancel}
        onOk={handleOk}
        className={'fa-modal--regular p-4 fa-modal--slideInDown'}
        closable={true}
        okText={'Save'}
        cancelText={'Close'}
        transitionName=''
        maskTransitionName=''
        okButtonProps={{ size: 'large' }}
        cancelButtonProps={{ size: 'large' }}
      >
        <div>
          <Row>
            <Col span={24}>
              <Text
                type={'title'}
                level={4}
                weight={'bold'}
                size={'grey'}
                extraClass={'m-0'}
              >
                Select slack channels
              </Text>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
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
        centered={true}
        zIndex={1005}
        width={700}
        onCancel={handleCancelTeams}
        onOk={handleOkTeams}
        className={'fa-modal--regular p-4 fa-modal--slideInDown'}
        closable={true}
        okText={'Save'}
        cancelText={'Close'}
        transitionName=''
        maskTransitionName=''
        okButtonProps={{ size: 'large' }}
        cancelButtonProps={{ size: 'large' }}
      >
        <div>
          <Row>
            <Col span={24}>
              <Text
                type={'title'}
                level={4}
                weight={'bold'}
                size={'grey'}
                extraClass={'m-0'}
              >
                Select Teams channels
              </Text>
            </Col>
          </Row>
          <Row className='my-3'>
            <Col span={24}>
              <Text
                type={'title'}
                level={6}
                color={'grey-2'}
                extraClass={'m-0 inline mr-2'}
              >
                Workspace
              </Text>
              <Select
                className={'fa-select inline'}
                options={teamsWorkspaceOpts}
                placeholder='Select Workspace'
                showSearch
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
              ></Select>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
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
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  userPropNames: state.coreQuery.userPropNames,
  eventNames: state.coreQuery.eventNames,
  groupOpts: state.groups.data
});

export default connect(mapStateToProps, {
  fetchEventAlerts,
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
  fetchGroups,
  testWebhhookUrl,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels,
  updateEventAlertStatus
})(EventBasedAlert);
