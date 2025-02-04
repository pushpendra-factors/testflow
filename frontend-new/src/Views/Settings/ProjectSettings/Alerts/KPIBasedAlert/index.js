import React, { useState, useEffect, useCallback } from 'react';
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
  Modal,
  Switch,
  Avatar,
  Popover
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PlusOutlined } from '@ant-design/icons';
import _ from 'lodash';
import {
  createAlert,
  fetchAlerts,
  deleteAlert,
  editAlert,
  fetchAllAlerts,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels
} from 'Reducers/global';
import ConfirmationModal from 'Components/ConfirmationModal';
import { deleteGroupByForEvent } from 'Reducers/coreQuery/middleware';
import useAutoFocus from 'hooks/useAutoFocus';
import GLobalFilter from 'Components/KPIComposer/GlobalFilter';
import {
  getEventsWithPropertiesKPI,
  getStateFromKPIFilters
} from 'Views/CoreQuery/utils';
import SelectChannels from '../SelectChannels';
import QueryBlock from './QueryBlock';

const { Option } = Select;

const KPIBasedAlert = ({
  activeProject,
  kpi,
  createAlert,
  fetchAllAlerts,
  deleteAlert,
  editAlert,
  savedAlerts,
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
  teams
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [operatorState, setOperatorState] = useState(null);
  const [Value, setValue] = useState(null);
  const [emailEnabled, setEmailEnabled] = useState(false);
  const [slackEnabled, setSlackEnabled] = useState(false);
  const [teamsEnabled, setTeamsEnabled] = useState(false);
  const [showCompareField, setShowCompareField] = useState(false);
  const [alertType, setAlertType] = useState(1);
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
  const [form] = Form.useForm();

  const alertDetails = viewAlertDetails?.alert;

  // KPI SELECTION
  const [queryType, setQueryType] = useState('kpi');
  const [queries, setQueries] = useState([]);
  const [selectedMainCategory, setSelectedMainCategory] = useState(false);
  const [KPIConfigProps, setKPIConfigProps] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    group_analysis: 'users',
    groupBy: [
      {
        prop_category: '', // user / event
        property: '', // user/eventproperty
        prop_type: '', // categorical  /numberical
        eventValue: '', // event name (funnel only)
        eventName: '', // eventName $present for global user breakdown
        eventIndex: 0
      }
    ],
    globalFilters: []
  });

  const confirmRemove = (id) =>
    deleteAlert(activeProject.id, id).then(
      (res) => {
        fetchAllAlerts(activeProject.id);
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

  // if operatorstate is include 'by more than' in string then show compare field
  useEffect(() => {
    if (operatorState && operatorState.includes('by_more_than')) {
      setShowCompareField(true);
      setAlertType(2);
    } else {
      setShowCompareField(false);
      setAlertType(1);
    }
  }, [operatorState]);

  useEffect(() => {
    if (alertDetails?.alert_description?.query?.fil) {
      const filter = getStateFromKPIFilters(
        alertDetails?.alert_description?.query?.fil
      );
      setViewFilter(filter);
    }
    if (alertDetails?.alert_configuration?.slack_channels_and_user_groups) {
      const obj =
        alertDetails?.alert_configuration?.slack_channels_and_user_groups;
      for (const key in obj) {
        if (obj[key]?.length > 0) {
          setViewSelectedChannels(obj[key]);
          if (alertState.state === 'edit') {
            setSaveSelectedChannel(obj[key]);
            setSelectedChannel(obj[key]);
          }
        }
      }
    }

    if (
      alertDetails?.alert_configuration?.teams_channel_config?.team_channel_list
    ) {
      setTeamsViewSelectedChannels(
        alertDetails?.alert_configuration?.teams_channel_config
          ?.team_channel_list
      );
      if (alertState.state === 'edit') {
        setTeamsSaveSelectedChannel(
          alertDetails?.alert_configuration?.teams_channel_config
            ?.team_channel_list
        );
        setTeamsSelectedChannel(
          alertDetails?.alert_configuration?.teams_channel_config
            ?.team_channel_list
        );
        setSelectedWorkspace({
          name: alertDetails?.alert_configuration?.teams_channel_config
            ?.team_name,
          id: alertDetails?.alert_configuration?.teams_channel_config?.team_id
        });
      }
    }

    if (alertState?.state === 'edit') {
      const queryData = [];
      queryData.push({
        alias: '',
        label: _.startCase(alertDetails?.alert_description?.name),
        filters: getStateFromKPIFilters(
          alertDetails?.alert_description?.query?.fil
            ? alertDetails?.alert_description?.query?.fil
            : []
        ),
        group: alertDetails?.alert_description?.query?.dc,
        metric: alertDetails?.alert_description?.name,
        metricType: '',
        qt: alertDetails?.alert_description?.qt,
        pageViewVal: alertDetails?.alert_description?.queries?.pgUrl,
        category: alertDetails?.alert_description?.query?.ca
      });
      setQueries(queryData);
      setAlertType(alertDetails?.alert_type);
      setOperatorState(alertDetails?.alert_description?.operator);
      setValue(alertDetails?.alert_description?.value);
      setEmailEnabled(alertDetails?.alert_configuration?.email_enabled);
      setSlackEnabled(alertDetails?.alert_configuration?.slack_enabled);
      setTeamsEnabled(alertDetails?.alert_configuration?.teams_enabled);
    }
  }, [alertState?.state, alertDetails]);

  const queryChange = (newEvent, index, changeType = 'add', flag = null) => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
          deleteGroupByForEvent(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else if (changeType === 'filters_updated') {
        // dont remove group by if filter is changed
        queryupdated[index] = newEvent;
      } else {
        deleteGroupByForEvent(newEvent, index);
        queryupdated.splice(index, 1);
      }
    } else {
      if (flag) {
        Object.assign(newEvent, { pageViewVal: flag });
      }
      queryupdated.push(newEvent);
    }
    setQueries(queryupdated);
  };

  useEffect(() => {
    setSelectedMainCategory(queries[0]);
  }, [queries]);

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
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
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
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    }

    return blockList;
  };

  const onReset = () => {
    setOperatorState('');
    setValue('');
    setQueries([]);
    setShowCompareField(false);
    setEmailEnabled(false);
    setSlackEnabled(false);
    setSelectedChannel([]);
    setSaveSelectedChannel([]);
    form.resetFields();
    setAlertState({ state: 'list', index: 0 });
  };

  const onFinish = (data) => {
    setLoading(true);
    // Putting All emails into single array
    let emails = [];
    if (emailEnabled) {
      if (data.emails) {
        emails = data.emails.map((item) => item.email);
      }
      if (data.email) {
        emails.push(data.email);
      }
    }

    let slackChannels = {};
    if (slackEnabled) {
      const map = new Map();
      map.set(agent_details.uuid, saveSelectedChannel);
      for (const [key, value] of map) {
        slackChannels = { ...slackChannels, [key]: value };
      }
    }

    const arr = slackChannels[agent_details?.uuid]; // access array with key
    const size = arr?.length;
    if (
      queries.length > 0 &&
      (emailEnabled || slackEnabled || teamsEnabled) &&
      (emails.length > 0 || size > 0 || teamsSaveSelectedChannel.length > 0)
    ) {
      const payload = {
        alert_name: data?.alert_name,
        alert_type: alertType,
        alert_description: {
          name: queries[0]?.metric,
          query: {
            ca: queries[0]?.category,
            dc: queries[0]?.group,
            fil: getEventsWithPropertiesKPI(
              queries[0]?.filters,
              queries[0]?.category
            ),
            me: [queries[0]?.metric],
            pgUrl: queries[0]?.pageViewVal ? queries[0]?.pageViewVal : '',
            tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
            qt: queries[0]?.qt
          },
          query_type: 'kpi',
          operator: operatorState,
          value: Value,
          date_range: data?.date_range,
          compared_to: data?.compared_to,
          message: ''
        },
        alert_configuration: {
          email_enabled: emailEnabled,
          slack_enabled: slackEnabled,
          emails,
          slack_channels_and_user_groups: slackChannels,
          teams_enabled: teamsEnabled,
          teams_channel_config: {
            team_id: selectedWorkspace?.id,
            team_name: selectedWorkspace?.name,
            team_channel_list: teamsSaveSelectedChannel
          }
        }
      };

      if (alertState?.state === 'edit') {
        editAlert(activeProject.id, payload, alertDetails?.id)
          .then((res) => {
            setLoading(false);
            fetchAllAlerts(activeProject.id);
            notification.success({
              message: 'Alerts Saved',
              description: 'Alerts is saved successfully.'
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
        createAlert(activeProject.id, payload, 0)
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
          description: 'Please select KPI to send alert.'
        });
      }

      if (!emailEnabled && !slackEnabled && !teamsEnabled) {
        notification.error({
          message: 'Error',
          description:
            'Please select atleast one delivery option to send alert.'
        });
      }

      if (slackEnabled && size === 0) {
        notification.error({
          message: 'Error',
          description: 'Empty Slack Channel List'
        });
      }

      if (emailEnabled && emails.length === 0) {
        notification.error({
          message: 'Error',
          description: 'Empty Email List'
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

  const emailView = () => {
    if (alertDetails.alert_configuration.emails) {
      return alertDetails.alert_configuration.emails.map((item, index) => (
        <div className='mb-3'>
          <Input
            disabled
            key={index}
            value={item}
            className='fa-input'
            placeholder='yourmail@gmail.com'
          />
        </div>
      ));
    }
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

  const DateRangeTypes = [
    { value: 'last_week', label: 'Weekly' },
    { value: 'last_month', label: 'Monthly' },
    { value: 'last_quarter', label: 'Quarterly' }
  ];

  const DateRangeTypeSelect = (
    <Select
      className='fa-select w-full'
      options={DateRangeTypes}
      value={alertDetails?.alert_description?.date_range}
      placeholder='Date range'
      showSearch
    />
  );

  const operatorOpts = [
    { label: 'is less than', value: 'is_less_than' },
    { label: 'is greater than', value: 'is_greater_than' },
    { label: 'decreased by more than', value: 'decreased_by_more_than' },
    { label: 'increased by more than', value: 'increased_by_more_than' },
    {
      label: 'increased or decreased by more than',
      value: 'increased_or_decreased_by_more_than'
    },
    {
      label: '% has decreased by more than',
      value: '%_has_decreased_by_more_than'
    },
    {
      label: '% has increased by more than',
      value: '%_has_increased_by_more_than'
    },
    {
      label: '% has increased or decreased by more than',
      value: '%_has_increased_or_decreased_by_more_than'
    }
  ];

  const selectOperator = (
    <Select
      className='fa-select w-full'
      options={operatorOpts}
      placeholder='Operator'
      showSearch
      value={alertDetails?.alert_description?.operator}
      onChange={(value) => {
        setOperatorState(value);
      }}
    />
  );

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

  const renderKPIForm = () => (
    <Form
      form={form}
      onFinish={onFinish}
      className='w-full'
      onChange={onChange}
      loading={loading}
    >
      <Row>
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
      </Row>
      <Row className='mt-6'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Alert name
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Form.Item
            name='alert_name'
            className='m-0'
            rules={[{ required: true, message: 'Please enter alert name' }]}
          >
            <Input
              className='fa-input'
              placeholder='Enter name'
              ref={inputComponentRef}
            />
          </Form.Item>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Notify me when
          </Text>
        </Col>
      </Row>
      <Row className='m-0'>
        <Col span={24}>
          <Form.Item name='query_type' className='m-0'>
            {queryList()}
          </Form.Item>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Operator
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Form.Item
            name='operator'
            className='m-0'
            rules={[{ required: true, message: 'Please select Operator' }]}
          >
            {selectOperator}
          </Form.Item>
        </Col>
        <Col span={8} className='ml-4'>
          <Form.Item
            name='value'
            className='m-0'
            rules={[{ required: true, message: 'Please enter value' }]}
          >
            <Input
              className='fa-input'
              type='number'
              placeholder='Qualifier'
              onChange={(e) => setValue(e.target.value)}
            />
          </Form.Item>
        </Col>
      </Row>

      <Row className='mt-4'>
        <Col span={8}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0 mb-1'
          >
            In the period of
          </Text>
          <Form.Item
            name='date_range'
            className='m-0'
            rules={[
              {
                required: true,
                message: 'Please select Date range'
              }
            ]}
          >
            {DateRangeTypeSelect}
          </Form.Item>
        </Col>
        {showCompareField && (
          <Col span={8} className='ml-4'>
            <Text
              type='title'
              level={7}
              weight='bold'
              color='grey-2'
              extraClass='m-0 mb-1'
            >
              Compared to
            </Text>
            <Form.Item
              name='compared_to'
              className='m-0'
              initialValue='previous_period'
              rules={[{ required: true, message: 'Please select Compare' }]}
            >
              <Select
                className='fa-select w-full'
                placeholder='Compare'
                showSearch
                disabled
              >
                <Option value='previous_period'>Previous period</Option>
              </Select>
            </Form.Item>
          </Col>
        )}
      </Row>

      <Row className=''>
        <Col span={24}>
          <div className='border-top--thin-2 pb-6 mt-6' />
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Delivery options
          </Text>
        </Col>
      </Row>

      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='slack' size={40} color='purple' />}
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
                    Post to Slack when events you care about happen. Motivate
                    the right actions.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='slack_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
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
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Text type='title' level={6} color='grey' extraClass='m-0'>
                  Slack is not integrated, Do you want to integrate with your
                  Slack account now?
                </Text>
              </Col>
            </Row>
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Button onClick={onConnectSlack}>
                  <SVG name='Slack' />
                  Connect to Slack
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
                      type='title'
                      level={7}
                      weight='regular'
                      extraClass='m-0 mt-2 ml-2'
                    >
                      {saveSelectedChannel.length > 1
                        ? 'Selected Channels'
                        : 'Selected Channel'}
                    </Text>
                  </Col>
                </Row>
                <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                  <Col className='m-0'>
                    {saveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type='title'
                          level={7}
                          color='grey'
                          extraClass='m-0 ml-4 my-2'
                        >
                          {channel?.is_private ? (
                            <>
                              <SVG
                                name='Lock'
                                color='gray'
                                extraClass='inline'
                              />{' '}
                              {channel?.name}
                            </>
                          ) : (
                            `#${channel?.name}`
                          )}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              </div>
            )}
            {!saveSelectedChannel.length > 0 ? (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
                    onClick={() => setShowSelectChannelsModal(true)}
                  >
                    Select Channel
                  </Button>
                </Col>
              </Row>
            ) : (
              <Row className='mt-1 ml-2'>
                <Col span={10} className='m-0'>
                  <div className='mt-1 mb-2 flex' style={{ width: '375px' }}>
                    <Popover
                      placement='right'
                      overlayInnerStyle={{ width: '340px' }}
                      title={null}
                      content={
                        <div className='m-0 m-2'>
                          <p className='m-0 text-gray-900 text-base font-bold'>
                            Preview of the alert in Slack
                          </p>
                          <p className='m-0 mb-2 text-gray-700'>
                            The message will be sent from your name
                          </p>
                          <img
                            className='m-0'
                            src='../../../../../assets/icons/privateAlertPreview.png'
                          />
                        </div>
                      }
                    >
                      <div className='inline mr-1'>
                        <SVG
                          name='InfoCircle'
                          size={16}
                          color='grey'
                          extraClass='inline'
                        />
                      </div>
                    </Popover>
                    <Text
                      type='title'
                      level={7}
                      color='grey'
                      extraClass='m-0 inline'
                    >
                      If you select a private channel, the alert message will be
                      sent through your account.
                    </Text>
                  </div>
                  <Button
                    type='link'
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
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='Email' size={38} color='purple' />}
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
                      Email
                    </Text>
                  </div>
                  <Text
                    type='paragraph'
                    mini
                    extraClass='m-0'
                    color='grey'
                    lineHeight='medium'
                  >
                    When this alert happens, send this information to the
                    selected mails.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='email_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
                  >
                    Enable
                  </Text>
                  <span style={{ width: '50px' }}>
                    <Switch
                      checkedChildren='On'
                      unCheckedChildren='OFF'
                      onChange={(checked) => setEmailEnabled(checked)}
                      checked={emailEnabled}
                    />
                  </span>{' '}
                </div>
              </Form.Item>
            </Col>
          </Row>
        </div>
        {emailEnabled && (
          <Row className='p-4 mt-2 ml-2'>
            <Col span={8}>
              <Form.Item
                label={null}
                name='email'
                validateTrigger={['onChange', 'onBlur']}
                rules={[
                  {
                    type: 'email',
                    message: 'Please enter a valid e-mail'
                  },
                  { required: true, message: 'Please enter email' }
                ]}
                className='m-0'
              >
                <Input className='fa-input' placeholder='yourmail@gmail.com' />
              </Form.Item>
            </Col>
            <Form.List name='emails'>
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field, index) => (
                    <Col span={21}>
                      <Form.Item required={false} key={field.key}>
                        <Row className='mt-4'>
                          <Col span={9}>
                            <Form.Item
                              label={null}
                              {...field}
                              name={[field.name, 'email']}
                              validateTrigger={['onChange', 'onBlur']}
                              rules={[
                                {
                                  type: 'email',
                                  message: 'Please enter a valid e-mail'
                                },
                                {
                                  required: true,
                                  message: 'Please enter email'
                                }
                              ]}
                              className='m-0'
                            >
                              <Input
                                className='fa-input'
                                placeholder='yourmail@gmail.com'
                              />
                            </Form.Item>
                          </Col>
                          {fields.length > 0 ? (
                            <Col span={1}>
                              <Button
                                style={{ backgroundColor: 'white' }}
                                className='mt-0.5 ml-2'
                                onClick={() => remove(field.name)}
                              >
                                <SVG name='Trash' size={20} color='gray' />
                              </Button>
                            </Col>
                          ) : null}
                        </Row>
                      </Form.Item>
                    </Col>
                  ))}
                  <Col span={20} className='mt-3'>
                    {fields.length === 4 ? null : (
                      <Button
                        type='text'
                        icon={
                          <PlusOutlined
                            style={{
                              color: 'gray',
                              fontSize: '18px'
                            }}
                          />
                        }
                        onClick={() => add()}
                      >
                        Add Email
                      </Button>
                    )}
                  </Col>
                </>
              )}
            </Form.List>
          </Row>
        )}
      </div>
      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='MSTeam' size={40} color='purple' />}
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
            <Col className='m-0 mt-4'>
              <Form.Item name='teams_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
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
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Text type='title' level={6} color='grey' extraClass='m-0'>
                  Teams is not integrated, Do you want to integrate with your
                  Microsoft Teams account now?
                </Text>
              </Col>
            </Row>
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Button onClick={onConnectMSTeams}>
                  <SVG name='MSTeam' size={20} />
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
                      type='title'
                      level={7}
                      weight='regular'
                      extraClass='m-0 mt-2 ml-2'
                    >
                      {teamsSaveSelectedChannel.length > 1
                        ? `Selected channels from the "${selectedWorkspace?.name}"`
                        : `Selected channels from the "${selectedWorkspace?.name}"`}
                    </Text>
                  </Col>
                </Row>
                <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                  <Col className='m-0'>
                    {teamsSaveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type='title'
                          level={7}
                          color='grey'
                          extraClass='m-0 ml-4 my-2'
                        >
                          {`#${channel.name}`}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              </div>
            )}
            {!teamsSaveSelectedChannel.length > 0 ? (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
                    onClick={() => setTeamsShowSelectChannelsModal(true)}
                  >
                    Select Channel
                  </Button>
                </Col>
              </Row>
            ) : (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
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
  );

  const renderKPIEdit = () => (
    <Form
      form={form}
      onFinish={onFinish}
      className='w-full'
      onChange={onChange}
      loading={loading}
    >
      <Row>
        <Col span={12}>
          <Text
            type='title'
            level={3}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Edit alert
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
      </Row>
      <Row className='mt-6'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Alert name
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Form.Item
            name='alert_name'
            className='m-0'
            initialValue={alertDetails?.alert_name}
            rules={[{ required: true, message: 'Please enter alert name' }]}
          >
            <Input
              className='fa-input'
              placeholder='Enter name'
              ref={inputComponentRef}
            />
          </Form.Item>
        </Col>
      </Row>

      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Notify me when
          </Text>
        </Col>
      </Row>
      <Row className='m-0'>
        <Col span={24}>
          <Form.Item name='query_type' className='m-0'>
            {queryList()}
          </Form.Item>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Operator
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Form.Item
            name='operator'
            className='m-0'
            initialValue={alertDetails?.alert_description?.operator}
            rules={[{ required: true, message: 'Please select Operator' }]}
          >
            {selectOperator}
          </Form.Item>
        </Col>
        <Col span={8} className='ml-4'>
          <Form.Item
            name='value'
            className='m-0'
            initialValue={alertDetails?.alert_description?.value}
            rules={[{ required: true, message: 'Please enter value' }]}
          >
            <Input
              className='fa-input'
              type='number'
              placeholder='Qualifier'
              value={alertDetails?.alert_description?.value}
              onChange={(e) => setValue(e.target.value)}
            />
          </Form.Item>
        </Col>
      </Row>

      <Row className='mt-4'>
        <Col span={8}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0 mb-1'
          >
            In the period of
          </Text>
          <Form.Item
            name='date_range'
            className='m-0'
            initialValue={alertDetails?.alert_description?.date_range}
            rules={[
              {
                required: true,
                message: 'Please select Date range'
              }
            ]}
          >
            {DateRangeTypeSelect}
          </Form.Item>
        </Col>
        {showCompareField && (
          <Col span={8} className='ml-4'>
            <Text
              type='title'
              level={7}
              weight='bold'
              color='grey-2'
              extraClass='m-0 mb-1'
            >
              Compared to
            </Text>
            <Form.Item
              name='compared_to'
              className='m-0'
              initialValue={
                alertDetails?.alert_description?.compared_to ||
                'previous_period'
              }
              rules={[{ required: true, message: 'Please select Compare' }]}
            >
              <Select
                className='fa-select w-full'
                placeholder='Compare'
                showSearch
                value={
                  alertDetails?.alert_description?.compared_to ||
                  'previous_period'
                }
                disabled
              >
                <Option value='previous_period'>Previous period</Option>
              </Select>
            </Form.Item>
          </Col>
        )}
      </Row>

      <Row className=''>
        <Col span={24}>
          <div className='border-top--thin-2 pb-6 mt-6' />
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Delivery options
          </Text>
        </Col>
      </Row>

      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='slack' size={40} color='purple' />}
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
                    Post to Slack when events you care about happen. Motivate
                    the right actions.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='slack_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
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
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Text type='title' level={6} color='grey' extraClass='m-0'>
                  Slack is not integrated, Do you want to integrate with your
                  Slack account now?
                </Text>
              </Col>
            </Row>
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Button onClick={onConnectSlack}>
                  <SVG name='Slack' />
                  Connect to Slack
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
                      type='title'
                      level={7}
                      weight='regular'
                      extraClass='m-0 mt-2 ml-2'
                    >
                      {saveSelectedChannel.length > 1
                        ? 'Selected Channels'
                        : 'Selected Channel'}
                    </Text>
                  </Col>
                </Row>
                <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                  <Col className='m-0'>
                    {saveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type='title'
                          level={7}
                          color='grey'
                          extraClass='m-0 ml-4 my-2'
                        >
                          {channel?.is_private ? (
                            <>
                              <SVG
                                name='Lock'
                                color='gray'
                                extraClass='inline'
                              />{' '}
                              {channel?.name}
                            </>
                          ) : (
                            `#${channel?.name}`
                          )}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              </div>
            )}
            {!saveSelectedChannel.length > 0 ? (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
                    onClick={() => setShowSelectChannelsModal(true)}
                  >
                    Select Channel
                  </Button>
                </Col>
              </Row>
            ) : (
              <Row className='mt-1 ml-2'>
                <Col span={10} className='m-0'>
                  <div className='mt-1 mb-2 flex' style={{ width: '375px' }}>
                    <Popover
                      placement='right'
                      overlayInnerStyle={{ width: '340px' }}
                      title={null}
                      content={
                        <div className='m-0 m-2'>
                          <p className='m-0 text-gray-900 text-base font-bold'>
                            Preview of the alert in Slack
                          </p>
                          <p className='m-0 mb-2 text-gray-700'>
                            The message will be sent from your name
                          </p>
                          <img
                            className='m-0'
                            src='../../../../../assets/icons/privateAlertPreview.png'
                          />
                        </div>
                      }
                    >
                      <div className='inline mr-1'>
                        <SVG
                          name='InfoCircle'
                          size={16}
                          color='grey'
                          extraClass='inline'
                        />
                      </div>
                    </Popover>
                    <Text
                      type='title'
                      level={7}
                      color='grey'
                      extraClass='m-0 inline'
                    >
                      If you select a private channel, the alert message will be
                      sent through your account.
                    </Text>
                  </div>
                  <Button
                    type='link'
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
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='Email' size={38} color='purple' />}
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
                      Email
                    </Text>
                  </div>
                  <Text
                    type='paragraph'
                    mini
                    extraClass='m-0'
                    color='grey'
                    lineHeight='medium'
                  >
                    When this alert happens, send this information to the
                    selected mails.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='email_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
                  >
                    Enable
                  </Text>
                  <span style={{ width: '50px' }}>
                    <Switch
                      checkedChildren='On'
                      unCheckedChildren='OFF'
                      onChange={(checked) => setEmailEnabled(checked)}
                      checked={emailEnabled}
                    />
                  </span>{' '}
                </div>
              </Form.Item>
            </Col>
          </Row>
        </div>
        {emailEnabled && (
          <Row className='p-4 ml-2'>
            <Form.List
              name='emails'
              initialValue={alertDetails?.alert_configuration?.emails}
            >
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field, index) => (
                    <Col span={21}>
                      <Form.Item required={false} key={field.key}>
                        <Row className='mt-2'>
                          <Col span={9}>
                            <Form.Item
                              label={null}
                              initialValue={
                                alertDetails?.alert_configuration?.emails[
                                  field.name
                                ]
                              }
                              {...field}
                              name={[field.name, 'email']}
                              validateTrigger={['onChange', 'onBlur']}
                              rules={[
                                {
                                  type: 'email',
                                  message: 'Please enter a valid e-mail'
                                },
                                {
                                  required: true,
                                  message: 'Please enter email'
                                }
                              ]}
                              className='m-0'
                            >
                              <Input
                                className='fa-input'
                                placeholder='yourmail@gmail.com'
                              />
                            </Form.Item>
                          </Col>
                          {fields.length > 0 ? (
                            <Col span={1}>
                              <Button
                                style={{ backgroundColor: 'white' }}
                                className='mt-0.5 ml-2'
                                onClick={() => remove(field.name)}
                              >
                                <SVG name='Trash' size={20} color='gray' />
                              </Button>
                            </Col>
                          ) : null}
                        </Row>
                      </Form.Item>
                    </Col>
                  ))}
                  <Col span={20} className='mt-3'>
                    {fields.length >= 5 ? null : (
                      <Button
                        type='text'
                        icon={
                          <PlusOutlined
                            style={{
                              color: 'gray',
                              fontSize: '18px'
                            }}
                          />
                        }
                        onClick={() => add()}
                      >
                        Add Email
                      </Button>
                    )}
                  </Col>
                </>
              )}
            </Form.List>
          </Row>
        )}
      </div>
      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='MSTeam' size={40} color='purple' />}
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
            <Col className='m-0 mt-4'>
              <Form.Item name='teams_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
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
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Text type='title' level={6} color='grey' extraClass='m-0'>
                  Teams is not integrated, Do you want to integrate with your
                  Microsoft Teams account now?
                </Text>
              </Col>
            </Row>
            <Row className='mt-2 ml-2'>
              <Col span={10} className='m-0'>
                <Button onClick={onConnectMSTeams}>
                  <SVG name='MSTeam' size={20} />
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
                      type='title'
                      level={7}
                      weight='regular'
                      extraClass='m-0 mt-2 ml-2'
                    >
                      {teamsSaveSelectedChannel.length > 1
                        ? `Selected channels from the "${selectedWorkspace?.name}"`
                        : `Selected channels from the "${selectedWorkspace?.name}"`}
                    </Text>
                  </Col>
                </Row>
                <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                  <Col className='m-0'>
                    {teamsSaveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type='title'
                          level={7}
                          color='grey'
                          extraClass='m-0 ml-4 my-2'
                        >
                          {`#${channel.name}`}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              </div>
            )}
            {!teamsSaveSelectedChannel.length > 0 ? (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
                    onClick={() => setTeamsShowSelectChannelsModal(true)}
                  >
                    Select Channel
                  </Button>
                </Col>
              </Row>
            ) : (
              <Row className='mt-2 ml-2'>
                <Col span={10} className='m-0'>
                  <Button
                    type='link'
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
  );

  const renderKPIView = () => (
    <>
      <Row>
        <Col span={12}>
          <Text type='title' level={3} weight='bold' extraClass='m-0'>
            View Alert
          </Text>
        </Col>
        <Col span={12}>
          <div className='flex justify-end'>
            <Button
              size='large'
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

      <Row className='mt-6'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Alert name
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Input
            disabled
            className='fa-input'
            value={alertDetails?.alert_name}
          />
        </Col>
      </Row>

      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Notify me when
          </Text>
        </Col>
      </Row>
      <Row className='m-0 mt-2'>
        <Col>
          <Button className='mr-2' type='link' disabled>
            {`${_.startCase(
              alertDetails?.alert_description?.name
            )} [ ${_.startCase(alertDetails?.alert_description?.query?.dc)} ]`}
          </Button>
        </Col>
        <Col>
          {alertDetails?.alert_description?.query?.pgUrl && (
            <div>
              <span className='mr-2'>from</span>
              <Button className='mr-2' type='link' disabled>
                {alertDetails?.alert_description?.query?.pgUrl}
              </Button>
            </div>
          )}
        </Col>
      </Row>
      {alertDetails?.alert_description?.query?.fil?.length > 0 && (
        <Row className='mt-2'>
          <Col span={18}>
            <Text
              type='title'
              level={7}
              weight='bold'
              color='grey-2'
              extraClass='m-0 my-1'
            >
              Filters
            </Text>
            <GLobalFilter filters={viewFilter} delFilter={false} viewMode />
          </Col>
        </Row>
      )}
      <Row className='mt-4'>
        <Col span={18}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Operator
          </Text>
        </Col>
      </Row>
      <Row className='mt-4'>
        <Col span={8} className='m-0'>
          <Input
            disabled
            className='fa-input w-full'
            value={(alertDetails?.alert_description?.operator).replace(
              /_/g,
              ' '
            )}
          />
        </Col>
        <Col span={8} className='ml-4 w-24'>
          <Input
            disabled
            className='fa-input'
            type='number'
            value={alertDetails?.alert_description?.value}
          />
        </Col>
      </Row>

      <Row className='mt-4'>
        <Col span={8}>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0 mb-1'
          >
            In the period of
          </Text>
          <Input
            disabled
            className='fa-input w-full'
            value={_.startCase(alertDetails?.alert_description?.date_range)}
          />
        </Col>
        {alertDetails?.alert_description?.compared_to && (
          <Col span={8} className='ml-4'>
            <Text
              type='title'
              level={7}
              weight='bold'
              color='grey-2'
              extraClass='m-0 mb-1'
            >
              Compared to
            </Text>
            <Input
              disabled
              className='fa-input w-full'
              value={_.startCase(alertDetails?.alert_description?.compared_to)}
            />
          </Col>
        )}
      </Row>

      <Row className=''>
        <Col span={24}>
          <div className='border-top--thin-2 pb-6 mt-6' />
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey-2'
            extraClass='m-0'
          >
            Delivery options
          </Text>
        </Col>
      </Row>

      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='slack' size={40} color='purple' />}
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
                    Post to Slack when events you care about happen. Motivate
                    the right actions.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='slack_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
                  >
                    Enable
                  </Text>
                  <span style={{ width: '50px' }}>
                    <Switch
                      checkedChildren='On'
                      unCheckedChildren='OFF'
                      disabled
                      checked={alertDetails?.alert_configuration?.slack_enabled}
                    />
                  </span>{' '}
                </div>
              </Form.Item>
            </Col>
          </Row>
        </div>

        {alertDetails?.alert_configuration?.slack_enabled &&
          alertDetails?.alert_configuration?.slack_channels_and_user_groups && (
            <div className='p-4'>
              {viewSelectedChannels.length > 0 && (
                <div>
                  <Row>
                    <Col>
                      <Text
                        type='title'
                        level={7}
                        weight='regular'
                        extraClass='m-0 mt-2 ml-2'
                      >
                        {viewSelectedChannels.length > 1
                          ? 'Selected Channels'
                          : 'Selected Channel'}
                      </Text>
                    </Col>
                  </Row>
                  <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                    <Col className='m-0'>
                      {viewSelectedChannels.map((channel, index) => (
                        <div key={index}>
                          <Text
                            type='title'
                            level={7}
                            color='grey'
                            extraClass='m-0 ml-4 my-2'
                          >
                            {`#${channel.name}`}
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
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='Email' size={38} color='purple' />}
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
                      Email
                    </Text>
                  </div>
                  <Text
                    type='paragraph'
                    mini
                    extraClass='m-0'
                    color='grey'
                    lineHeight='medium'
                  >
                    When this alert happens, send this information to the
                    selected mails.
                  </Text>
                </div>
              </div>
            </Col>
            <Col className='m-0 mt-4'>
              <Form.Item name='email_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
                  >
                    Enable
                  </Text>
                  <span style={{ width: '50px' }}>
                    <Switch
                      checkedChildren='On'
                      unCheckedChildren='OFF'
                      disabled
                      checked={alertDetails?.alert_configuration?.email_enabled}
                    />
                  </span>{' '}
                </div>
              </Form.Item>
            </Col>
          </Row>
        </div>
        <Row className='p-4 ml-2 mt-3'>
          <Col span={8}>{emailView()}</Col>
        </Row>
      </div>
      <div className='border rounded mt-3'>
        <div style={{ backgroundColor: '#fafafa' }}>
          <Row className='ml-2'>
            <Col span={20}>
              <div className='flex justify-between p-3'>
                <div className='flex'>
                  <Avatar
                    size={40}
                    shape='square'
                    icon={<SVG name='MSTeam' size={40} color='purple' />}
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
            <Col className='m-0 mt-4'>
              <Form.Item name='teams_enabled' className='m-0'>
                <div span={24} className='flex flex-start items-center'>
                  <Text
                    type='title'
                    level={7}
                    weight='medium'
                    extraClass='m-0 mr-2'
                  >
                    Enable
                  </Text>
                  <span style={{ width: '50px' }}>
                    <Switch
                      checkedChildren='On'
                      unCheckedChildren='OFF'
                      disabled
                      checked={alertDetails?.alert_configuration?.teams_enabled}
                    />
                  </span>{' '}
                </div>
              </Form.Item>
            </Col>
          </Row>
        </div>
        {alertDetails?.alert_configuration?.teams_enabled &&
          alertDetails?.alert_configuration?.teams_channel_config && (
            <div className='p-4'>
              {teamsViewSelectedChannels.length > 0 && (
                <div>
                  <Row>
                    <Col>
                      <Text
                        type='title'
                        level={7}
                        weight='regular'
                        extraClass='m-0 mt-2 ml-2'
                      >
                        {teamsViewSelectedChannels.length > 1
                          ? `Selected channels from the “${alertDetails?.alert_configuration?.teams_channel_config?.team_name}”`
                          : `Selected channels from the “${alertDetails?.alert_configuration?.teams_channel_config?.team_name}”`}
                      </Text>
                    </Col>
                  </Row>
                  <Row className='rounded border border-gray-200 ml-2 w-2/6'>
                    <Col className='m-0'>
                      {teamsViewSelectedChannels.map((channel, index) => (
                        <div key={index}>
                          <Text
                            type='title'
                            level={7}
                            color='grey'
                            extraClass='m-0 ml-4 my-2'
                          >
                            {`#${channel.name}`}
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

      <Row className='mt-2'>
        <Col span={24}>
          <div className='border-top--thin-2 mt-2 mb-4' />
          <Button
            type='text'
            size='large'
            style={{ color: '#EE3C3C' }}
            className='m-0'
            onClick={() => showDeleteWidgetModal(alertDetails?.id)}
          >
            <SVG
              name='Delete1'
              extraClass='-mt-1 -mr-1'
              size={18}
              color='#EE3C3C'
            />
            Delete
          </Button>
        </Col>
      </Row>
    </>
  );

  return (
    <div className='fa-container '>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          <div className='mb-10 pl-4'>
            {alertState.state == 'add' && renderKPIForm()}

            {alertState.state == 'view' && renderKPIView()}

            {alertState.state == 'edit' && renderKPIEdit()}

            <ConfirmationModal
              visible={!!deleteWidgetModal}
              confirmationText='Do you really want to remove this alert?'
              onOk={confirmDelete}
              onCancel={showDeleteWidgetModal.bind(this, false)}
              title='Remove Alert'
              okText='Confirm'
              cancelText='Cancel'
              confirmLoading={deleteApiCalled}
            />
          </div>
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
              <Text type='title' level={4} weight='bold' extraClass='m-0'>
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
            <Col span={24}>
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
  savedAlerts: state.global.Alerts,
  kpi: state?.kpi,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  teams: state.global.teams,
  projectSettings: state.global.projectSettingsV1
});

export default connect(mapStateToProps, {
  createAlert,
  fetchAllAlerts,
  deleteAlert,
  editAlert,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  enableTeamsIntegration,
  fetchTeamsWorkspace,
  fetchTeamsChannels
})(KPIBasedAlert);
