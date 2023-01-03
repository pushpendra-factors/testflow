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
  Checkbox,
  Modal,
  Tabs
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PlusOutlined } from '@ant-design/icons';
import _ from 'lodash';
import { createAlert, fetchAlerts, deleteAlert, editAlert } from 'Reducers/global';
import ConfirmationModal from 'Components/ConfirmationModal';
import QueryBlock from './QueryBlock';
import { deleteGroupByForEvent } from 'Reducers/coreQuery/middleware';
import { getEventsWithPropertiesKPI, getStateFromFilters } from '../utils';
import {
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
} from 'Reducers/global';
import SelectChannels from '../SelectChannels';
import FAFilterSelect from 'Components/FaFilterSelect';

const { Option } = Select;

const KPIBasedAlert = ({
  activeProject,
  kpi,
  createAlert,
  fetchAlerts,
  deleteAlert,
  editAlert,
  savedAlerts,
  agent_details,
  slack,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  projectSettings,
  enableSlackIntegration,
  viewAlertDetails,
  alertState,
  setAlertState,
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [operatorState, setOperatorState] = useState(null);
  const [Value, setValue] = useState(null);
  const [emailEnabled, setEmailEnabled] = useState(false);
  const [slackEnabled, setSlackEnabled] = useState(false);
  const [showCompareField, setShowCompareField] = useState(false);
  const [alertType, setAlertType] = useState(1);
  const [viewFilter, setViewFilter] = useState([]);
  const [channelOpts, setChannelOpts] = useState([]);
  const [selectedChannel, setSelectedChannel] = useState([]);
  const [saveSelectedChannel, setSaveSelectedChannel] = useState([]);
  const [showSelectChannelsModal, setShowSelectChannelsModal] = useState(false);
  const [viewSelectedChannels, setViewSelectedChannels] = useState([]);

  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);

  const [form] = Form.useForm();

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

  const confirmRemove = (id) => {
    return deleteAlert(activeProject.id, id).then(
      (res) => {
        fetchAlerts(activeProject.id);
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
    if (viewAlertDetails?.alert_description?.query?.fil) {
      const filter = getStateFromFilters(
        viewAlertDetails.alert_description.query.fil
      );
      setViewFilter(filter);
    }
    if (viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups) {
      let obj =
        viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups;
      for (let key in obj) {
        if (obj[key].length > 0) {
          setViewSelectedChannels(obj[key]);
          if(alertState.state === 'edit') {
            setSaveSelectedChannel(obj[key]);
            setSelectedChannel(obj[key]);
          }
        }
      }
    }

    if(alertState.state === 'edit') {
      setEmailEnabled(viewAlertDetails?.alert_configuration?.email_enabled);
      setSlackEnabled(viewAlertDetails?.alert_configuration?.slack_enabled);
    }
  }, [viewAlertDetails]);

  const queryChange = (newEvent, index, changeType = 'add', flag = null) => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
          deleteGroupByForEvent(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else {
        if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
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
        <div key={'init'}>
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
        emails = data.emails.map((item) => {
          return item.email;
        });
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
    if (
      queries.length > 0 &&
      (emails.length > 0 || Object.keys(slackChannels).length > 0)
    ) {
      let payload = {
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
          emails: emails,
          slack_channels_and_user_groups: slackChannels
        }
      };

      createAlert(activeProject.id, payload, 0)
        .then((res) => {
          setLoading(false);
          fetchAlerts(activeProject.id);
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
          console.log('create alerts error->', err);
        });
    } else {
      setLoading(false);
      notification.error({
        message: 'Error',
        description:
          'Please select KPI and atleast one delivery option to send alert.'
      });
    }
  };

  const onEdit = (data) => {
    setLoading(true);
    // Putting All emails into single array
    let emails = [];
    if (emailEnabled) {
      if (data.emails) {
        emails = data.emails.map((item) => {
          return item.email;
        });
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

    if (
      (emails.length > 0 || Object.keys(slackChannels).length > 0)
    ) {
      let payload = {
        alert_name: data?.alert_name,
        alert_configuration: {
          email_enabled: emailEnabled,
          slack_enabled: slackEnabled,
          emails: emails,
          slack_channels_and_user_groups: slackChannels
        }
      };

      editAlert(activeProject.id, payload, viewAlertDetails?.id)
        .then((res) => {
          setLoading(false);
          fetchAlerts(activeProject.id);
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
      setLoading(false);
      notification.error({
        message: 'Error',
        description:
          'Please select atleast one delivery option to send alert.'
      });
    }
  };

  const emailView = () => {
    if (viewAlertDetails.alert_configuration.emails) {
      return viewAlertDetails.alert_configuration.emails.map((item, index) => {
        return (
          <div className={'mb-3'}>
            <Input
              disabled={true}
              key={index}
              value={item}
              className={'fa-input'}
              placeholder={'yourmail@gmail.com'}
            />
          </div>
        );
      });
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
      className={'fa-select w-full'}
      options={DateRangeTypes}
      placeholder='Date range'
      showSearch
    ></Select>
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
      className={'fa-select w-full'}
      options={operatorOpts}
      placeholder='Operator'
      showSearch
      onChange={(value) => {
        setOperatorState(value);
      }}
    ></Select>
  );

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettings?.int_slack) {
      fetchSlackChannels(activeProject.id);
    }
  }, [activeProject, projectSettings?.int_slack, slackEnabled]);

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

  const handleOk = () => {
    setSaveSelectedChannel(selectedChannel);
    setShowSelectChannelsModal(false);
  };

  const handleCancel = () => {
    setSelectedChannel(saveSelectedChannel);
    setShowSelectChannelsModal(false);
  };

  const renderKPIForm = () => {
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
          <Row className={'mt-4'}>
            <Col span={8} className={'m-0'}>
              <Form.Item
                name='alert_name'
                className={'m-0'}
                rules={[{ required: true, message: 'Please enter alert name' }]}
              >
                <Input className={'fa-input'} placeholder={'Enter name'} />
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
                Notify me when
              </Text>
            </Col>
          </Row>
          <Row className={'m-0'}>
            <Col span={18}>
              <Form.Item name='query_type' className={'m-0'}>
                {queryList()}
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
                Operator
              </Text>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={8} className={'m-0'}>
              <Form.Item
                name='operator'
                className={'m-0'}
                rules={[{ required: true, message: 'Please select Operator' }]}
              >
                {selectOperator}
              </Form.Item>
            </Col>
            <Col span={8} className={'ml-4'}>
              <Form.Item
                name='value'
                className={'m-0'}
                rules={[{ required: true, message: 'Please enter value' }]}
              >
                <Input
                  className={'fa-input'}
                  type={'number'}
                  placeholder={'Qualifier'}
                  onChange={(e) => setValue(e.target.value)}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={8}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0 mb-1'}
              >
                In the period of
              </Text>
              <Form.Item
                name='date_range'
                className={'m-0'}
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
              <Col span={8} className={'ml-4'}>
                <Text
                  type={'title'}
                  level={7}
                  weight={'bold'}
                  color={'grey-2'}
                  extraClass={'m-0 mb-1'}
                >
                  Compared to
                </Text>
                <Form.Item
                  name='compared_to'
                  className={'m-0'}
                  initialValue={'previous_period'}
                  rules={[{ required: true, message: 'Please select Compare' }]}
                >
                  <Select
                    className={'fa-select w-full'}
                    placeholder='Compare'
                    showSearch
                    disabled={true}
                  >
                    <Option value='previous_period'>Previous period</Option>
                  </Select>
                </Form.Item>
              </Col>
            )}
          </Row>

          <Row className={'mt-2'}>
            <Col span={24}>
              <div className={'border-top--thin-2 pt-2 mt-2'} />
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Delivery options
              </Text>
            </Col>
          </Row>

          <Row className={'mt-2 ml-2'}>
            <Col span={4}>
              <Form.Item name='email_enabled' className={'m-0'}>
                <Checkbox
                  defaultChecked={emailEnabled}
                  onChange={(e) => setEmailEnabled(e.target.checked)}
                >
                  Email
                </Checkbox>
              </Form.Item>
            </Col>
          </Row>
          {emailEnabled && (
            <Row className={'mt-4'}>
              <Col span={8}>
                <Form.Item
                  label={null}
                  name={'email'}
                  validateTrigger={['onChange', 'onBlur']}
                  rules={[
                    {
                      type: 'email',
                      message: 'Please enter a valid e-mail'
                    },
                    { required: true, message: 'Please enter email' }
                  ]}
                  className={'m-0'}
                >
                  <Input
                    className={'fa-input'}
                    placeholder={'yourmail@gmail.com'}
                  />
                </Form.Item>
              </Col>
              <Form.List name='emails'>
                {(fields, { add, remove }) => (
                  <>
                    {fields.map((field, index) => (
                      <Col span={21}>
                        <Form.Item required={false} key={field.key}>
                          <Row className={'mt-4'}>
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
                                className={'m-0'}
                              >
                                <Input
                                  className={'fa-input'}
                                  placeholder={'yourmail@gmail.com'}
                                />
                              </Form.Item>
                            </Col>
                            {fields.length > 0 ? (
                              <Col span={1}>
                                <Button
                                  style={{ backgroundColor: 'white' }}
                                  className={'mt-0.5 ml-2'}
                                  onClick={() => remove(field.name)}
                                >
                                  <SVG name={'Trash'} size={20} color='gray' />
                                </Button>
                              </Col>
                            ) : null}
                          </Row>
                        </Form.Item>
                      </Col>
                    ))}
                    <Col span={20} className={'mt-3'}>
                      {fields.length === 4 ? null : (
                        <Button
                          type={'text'}
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
          <Row className={'mt-2 ml-2'}>
            <Col className={'m-0'}>
              <Form.Item name='slack_enabled' className={'m-0'}>
                <Checkbox
                  defaultChecked={slackEnabled}
                  onChange={(e) => setSlackEnabled(e.target.checked)}
                >
                  Slack
                </Checkbox>
              </Form.Item>
            </Col>
          </Row>
          {slackEnabled && !projectSettings?.int_slack && (
            <>
              <Row className={'mt-2 ml-2'}>
                <Col span={10} className={'m-0'}>
                  <Text
                    type={'title'}
                    level={6}
                    color={'grey'}
                    extraClass={'m-0'}
                  >
                    Slack is not integrated, Do you want to integrate with your
                    slack account now?
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
            </>
          )}
          {slackEnabled && projectSettings?.int_slack && (
            <>
              {saveSelectedChannel.length > 0 && (
                <Row
                  className={'rounded-lg border-2 border-gray-200 mt-2 w-2/6'}
                >
                  <Col className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey-2'}
                      extraClass={'m-0 mt-2 ml-2'}
                    >
                      Selected Channels
                    </Text>
                    {saveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type={'title'}
                          level={7}
                          color={'grey'}
                          extraClass={'m-0 ml-2 my-1'}
                        >
                          {'#' + channel.name}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              )}
              {!saveSelectedChannel.length > 0 ? (
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button
                      type={'link'}
                      onClick={() => setShowSelectChannelsModal(true)}
                    >
                      Select Channels
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
                      Manage Channels
                    </Button>
                  </Col>
                </Row>
              )}
            </>
          )}
        </Form>
      </>
    );
  };

  const renderKPIEdit = () => {
    return (
      <>
        <Form
          form={form}
          onFinish={onEdit}
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
                Edit alert
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
          <Row className={'mt-4'}>
            <Col span={8} className={'m-0'}>
              <Form.Item
                name='alert_name'
                className={'m-0'}
                initialValue={viewAlertDetails?.alert_name}
                rules={[{ required: true, message: 'Please enter alert name' }]}
              >
                <Input className={'fa-input'} placeholder={'Enter name'} />
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
                Notify me when
              </Text>
            </Col>
          </Row>
          <Row className={'m-0 mt-2'}>
            <Col>
              <Button className={`mr-2`} type='link' disabled={true}>
                {_.startCase(viewAlertDetails?.alert_description?.name)}
              </Button>
            </Col>
            <Col>
              {viewAlertDetails?.alert_description?.query?.pgUrl && (
                <div>
                  <span className={'mr-2'}>from</span>
                  <Button className={`mr-2`} type='link' disabled={true}>
                    {viewAlertDetails?.alert_description?.query?.pgUrl}
                  </Button>
                </div>
              )}
            </Col>
          </Row>
          {viewAlertDetails?.alert_description?.query?.fil?.length > 0 && (
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
                {viewFilter.map((filt, index) => (
                  <div key={index} className={'mt-2'}>
                    <FAFilterSelect
                      filter={filt}
                      disabled={true}
                      applyFilter={() => {}}
                    ></FAFilterSelect>
                  </div>
                ))}
              </Col>
            </Row>
          )}
          <Row className={'mt-4'}>
            <Col span={18}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Operator
              </Text>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={8} className={'m-0'}>
              <Input
                disabled={true}
                className={'fa-input w-full'}
                value={_.startCase(viewAlertDetails?.alert_description?.operator)}
              />
            </Col>
            <Col span={8} className={'ml-4 w-24'}>
              <Input
                disabled={true}
                className={'fa-input'}
                type={'number'}
                value={viewAlertDetails?.alert_description?.value}
              />
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={8}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0 mb-1'}
              >
                In the period of
              </Text>
              <Input
                disabled={true}
                className={'fa-input w-full'}
                value={_.startCase(viewAlertDetails?.alert_description?.date_range)}
              />
            </Col>
            {viewAlertDetails?.alert_description?.compared_to && (
              <Col span={8} className={'ml-4'}>
                <Text
                  type={'title'}
                  level={7}
                  weight={'bold'}
                  color={'grey-2'}
                  extraClass={'m-0 mb-1'}
                >
                  Compared to
                </Text>
                <Input
                  disabled={true}
                  className={'fa-input w-full'}
                  value={_.startCase(viewAlertDetails?.alert_description?.compared_to)}
                />
              </Col>
            )}
          </Row>

          <Row className={'mt-2'}>
            <Col span={24}>
              <div className={'border-top--thin-2 pt-2 mt-2'} />
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Delivery options
              </Text>
            </Col>
          </Row>

          <Row className={'mt-2 ml-2'}>
            <Col span={4}>
              <Form.Item name='email_enabled' className={'m-0'}>
                <Checkbox
                  checked={emailEnabled}
                  onChange={(e) => setEmailEnabled(e.target.checked)}
                >
                  Email
                </Checkbox>
              </Form.Item>
            </Col>
          </Row>
          {emailEnabled && (
            <Row className={'mt-4'}>
              <Col span={8}>
                <Form.Item
                  label={null}
                  name={'email'}
                  initialValue={viewAlertDetails?.alert_configuration?.emails[0]}
                  validateTrigger={['onChange', 'onBlur']}
                  rules={[
                    {
                      type: 'email',
                      message: 'Please enter a valid e-mail'
                    },
                    { required: true, message: 'Please enter email' }
                  ]}
                  className={'m-0'}
                >
                  <Input
                    className={'fa-input'}
                    placeholder={'yourmail@gmail.com'}
                  />
                </Form.Item>
              </Col>
              <Form.List name='emails' initialValue={viewAlertDetails?.alert_configuration?.emails}>
                {(fields, { add, remove }) => (
                  <>
                    {fields.map((field, index) => (
                      <Col span={21}>
                        <Form.Item required={false} key={field.key}>
                          <Row className={'mt-4'}>
                            <Col span={9}>
                              <Form.Item
                                label={null}
                                initialValue={viewAlertDetails?.alert_configuration?.emails[field.name + 1]}
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
                                className={'m-0'}
                              >
                                <Input
                                  className={'fa-input'}
                                  placeholder={'yourmail@gmail.com'}
                                />
                              </Form.Item>
                            </Col>
                            {fields.length > 0 ? (
                              <Col span={1}>
                                <Button
                                  style={{ backgroundColor: 'white' }}
                                  className={'mt-0.5 ml-2'}
                                  onClick={() => remove(field.name)}
                                >
                                  <SVG name={'Trash'} size={20} color='gray' />
                                </Button>
                              </Col>
                            ) : null}
                          </Row>
                        </Form.Item>
                      </Col>
                    ))}
                    <Col span={20} className={'mt-3'}>
                      {fields.length >= 4 ? null : (
                        <Button
                          type={'text'}
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
          <Row className={'mt-2 ml-2'}>
            <Col className={'m-0'}>
              <Form.Item name='slack_enabled' className={'m-0'}>
                <Checkbox
                  checked={slackEnabled}
                  onChange={(e) => setSlackEnabled(e.target.checked)}
                >
                  Slack
                </Checkbox>
              </Form.Item>
            </Col>
          </Row>
          {slackEnabled && !projectSettings?.int_slack && (
            <>
              <Row className={'mt-2 ml-2'}>
                <Col span={10} className={'m-0'}>
                  <Text
                    type={'title'}
                    level={6}
                    color={'grey'}
                    extraClass={'m-0'}
                  >
                    Slack is not integrated, Do you want to integrate with your
                    slack account now?
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
            </>
          )}
          {slackEnabled && projectSettings?.int_slack && (
            <>
              {saveSelectedChannel.length > 0 && (
                <Row
                  className={'rounded-lg border-2 border-gray-200 mt-2 w-2/6'}
                >
                  <Col className={'m-0'}>
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey-2'}
                      extraClass={'m-0 mt-2 ml-2'}
                    >
                      Selected Channels
                    </Text>
                    {saveSelectedChannel.map((channel, index) => (
                      <div key={index}>
                        <Text
                          type={'title'}
                          level={7}
                          color={'grey'}
                          extraClass={'m-0 ml-2 my-1'}
                        >
                          {'#' + channel.name}
                        </Text>
                      </div>
                    ))}
                  </Col>
                </Row>
              )}
              {!saveSelectedChannel.length > 0 ? (
                <Row className={'mt-2 ml-2'}>
                  <Col span={10} className={'m-0'}>
                    <Button
                      type={'link'}
                      onClick={() => setShowSelectChannelsModal(true)}
                    >
                      Select Channels
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
                      Manage Channels
                    </Button>
                  </Col>
                </Row>
              )}
            </>
          )}
        </Form>
      </>
    );
  };

  const renderKPIView = () => {
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
              value={viewAlertDetails?.alert_name}
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
              Notify me when
            </Text>
          </Col>
        </Row>
        <Row className={'m-0 mt-2'}>
          <Col>
            <Button className={`mr-2`} type='link' disabled={true}>
              {_.startCase(viewAlertDetails?.alert_description?.name)}
            </Button>
          </Col>
          <Col>
            {viewAlertDetails?.alert_description?.query?.pgUrl && (
              <div>
                <span className={'mr-2'}>from</span>
                <Button className={`mr-2`} type='link' disabled={true}>
                  {viewAlertDetails?.alert_description?.query?.pgUrl}
                </Button>
              </div>
            )}
          </Col>
        </Row>
        {viewAlertDetails?.alert_description?.query?.fil?.length > 0 && (
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
              {viewFilter.map((filt, index) => (
                <div key={index} className={'mt-2'}>
                  <FAFilterSelect
                    filter={filt}
                    disabled={true}
                    applyFilter={() => {}}
                  ></FAFilterSelect>
                </div>
              ))}
            </Col>
          </Row>
        )}
        <Row className={'mt-4'}>
          <Col span={18}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Operator
            </Text>
          </Col>
        </Row>
        <Row className={'mt-4'}>
          <Col span={8} className={'m-0'}>
            <Input
              disabled={true}
              className={'fa-input w-full'}
            value={(viewAlertDetails?.alert_description?.operator).replace(
              /_/g,
              ' '
            )}
          />
        </Col>
        <Col span={8} className={'ml-4 w-24'}>
          <Input
            disabled={true}
            className={'fa-input'}
            type={'number'}
            value={viewAlertDetails?.alert_description?.value}
          />
        </Col>
      </Row>

        <Row className={'mt-4'}>
          <Col span={8}>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0 mb-1'}
            >
              In the period of
            </Text>
            <Input
              disabled={true}
              className={'fa-input w-full'}
              value={_.startCase(viewAlertDetails?.alert_description?.date_range)}
            />
          </Col>
          {viewAlertDetails?.alert_description?.compared_to && (
            <Col span={8} className={'ml-4'}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                color={'grey-2'}
                extraClass={'m-0 mb-1'}
              >
                Compared to
              </Text>
              <Input
                disabled={true}
                className={'fa-input w-full'}
                value={_.startCase(viewAlertDetails?.alert_description?.compared_to)}
              />
            </Col>
          )}
        </Row>

        <Row className={'mt-2'}>
          <Col span={24}>
            <div className={'border-top--thin-2 pt-2 mt-2'} />
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'grey-2'}
              extraClass={'m-0'}
            >
              Delivery options
            </Text>
          </Col>
        </Row>

        <Row className={'mt-2 ml-2'}>
          <Col span={4}>
            <Checkbox
              disabled={true}
              checked={viewAlertDetails?.alert_configuration?.email_enabled}
            >
              Email
            </Checkbox>
          </Col>
        </Row>
        <Row className={'mt-4'}>
          <Col span={8}>{emailView()}</Col>
        </Row>
        <Row className={'mt-2 ml-2'}>
          <Col span={4}>
            <Checkbox
              disabled={true}
              checked={viewAlertDetails?.alert_configuration?.slack_enabled}
            >
              Slack
            </Checkbox>
          </Col>
        </Row>
        {viewAlertDetails?.alert_configuration?.slack_enabled &&
          viewAlertDetails?.alert_configuration
            ?.slack_channels_and_user_groups && (
            <Row className={'rounded-lg border-2 border-gray-200 mt-2 w-2/6'}>
              <Col className={'m-0'}>
                <Text
                  type={'title'}
                  level={6}
                  color={'grey-2'}
                  extraClass={'m-0 mt-2 ml-2'}
                >
                  Selected Channels
                </Text>
                {viewSelectedChannels.map((channel, index) => (
                  <div key={index}>
                    <Text
                      type={'title'}
                      level={7}
                      color={'grey'}
                      extraClass={'m-0 ml-2 my-1'}
                    >
                      {'#' + channel.name}
                    </Text>
                  </div>
                ))}
              </Col>
            </Row>
          )}
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
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className={'mb-10 pl-4'}>
            {alertState.state == 'add' && renderKPIForm()}

            {alertState.state == 'view' && renderKPIView()}

            {alertState.state == 'edit' && renderKPIEdit()}

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
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  savedAlerts: state.global.Alerts,
  kpi: state?.kpi,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  projectSettings: state.global.projectSettingsV1
});

export default connect(mapStateToProps, {
  createAlert,
  fetchAlerts,
  deleteAlert,
  editAlert,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
})(KPIBasedAlert);
