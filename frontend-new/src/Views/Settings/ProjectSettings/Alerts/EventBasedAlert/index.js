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
  Popover
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import {
  createEventAlert,
  fetchEventAlerts,
  deleteEventAlert
} from 'Reducers/global';
import ConfirmationModal from 'Components/ConfirmationModal';
import QueryBlock from './QueryBlock';
import {
  deleteGroupByForEvent,
  setGroupBy,
  delGroupBy,
  getUserProperties
} from 'Reducers/coreQuery/middleware';
import { getEventsWithPropertiesKPI, getStateFromFilters } from '../utils';
import {
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
} from 'Reducers/global';
import SelectChannels from '../SelectChannels';
import FAFilterSelect from 'Components/FaFilterSelect';
import {
  QUERY_TYPE_EVENT,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE
} from 'Utils/constants';
import { DefaultDateRangeFormat } from '../../../../CoreQuery/utils';
import TextArea from 'antd/lib/input/TextArea';
import EventGroupBlock from '../../../../../components/QueryComposer/EventGroupBlock';
import useAutoFocus from 'hooks/useAutoFocus';

const { Option } = Select;

const EventBasedAlert = ({
  activeProject,
  fetchEventAlerts,
  deleteEventAlert,
  createEventAlert,
  agent_details,
  slack,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  projectSettings,
  enableSlackIntegration,
  viewAlertDetails,
  alertState,
  setAlertState,
  setGroupBy,
  delGroupBy,
  getUserProperties,
  groupBy
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [webhookEnabled, setWebhookEnabled] = useState(false);
  const [slackEnabled, setSlackEnabled] = useState(false);
  const [notRepeat, setNotRepeat] = useState(false);
  const [notifications, setNotifications] = useState(false);
  const [alertLimit, setAlertLimit] = useState('5');
  const [viewFilter, setViewFilter] = useState([]);
  const [channelOpts, setChannelOpts] = useState([]);
  const [selectedChannel, setSelectedChannel] = useState([]);
  const [saveSelectedChannel, setSaveSelectedChannel] = useState([]);
  const [showSelectChannelsModal, setShowSelectChannelsModal] = useState(false);
  const [viewSelectedChannels, setViewSelectedChannels] = useState([]);

  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const inputComponentRef = useAutoFocus();

  const [form] = Form.useForm();

  // Event SELECTION
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });

  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

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

  useEffect(() => {
    if (viewAlertDetails?.filter) {
      const filter = getStateFromFilters(viewAlertDetails.filter);
      setViewFilter(filter);
    }
    if (viewAlertDetails?.slack_channels) {
      let obj = viewAlertDetails?.slack_channels;
      for (let key in obj) {
        if (obj[key].length > 0) {
          setViewSelectedChannels(obj[key]);
        }
      }
    }
  }, [viewAlertDetails]);

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
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
    getUserProperties(activeProject.id, queryType);
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

  const onReset = () => {
    setQueries([]);
    setSlackEnabled(false);
    setAlertLimit('5');
    setNotRepeat(false);
    setNotifications(false);
    setSelectedChannel([]);
    setSaveSelectedChannel([]);
    form.resetFields();
    setAlertState({ state: 'list', index: 0 });
  };

  const onFinish = (data) => {
    setLoading(true);

    let slackChannels = [];
    if (slackEnabled) {
      slackChannels = saveSelectedChannel;
    }

    if (queries.length > 0 && slackChannels.length > 0) {
      let payload = {
        title: data?.alert_name,
        event: queries[0]?.label,
        filter: getEventsWithPropertiesKPI(queries[0]?.filters, 'event'),
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
        slack: slackEnabled,
        slack_channels: slackChannels
      };

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
            <Col span={18}>
              <Form.Item name='event_name' className={'m-0'}>
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
                Configure your notification
              </Text>
            </Col>
          </Row>
          <Row className={'mt-4'}>
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
                        This is how it shown in the slacks
                      </p>
                      <p className='m-0 mb-2 text-gray-700'>
                        The slack message will be like this
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
                  maxLength={300}
                />
              </Form.Item>
            </Col>
          </Row>

          {queries.length > 0 && (
            <Row className={'mt-4'}>
              <Col span={12} className={'ml-4'}>
                <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>
                  Attach properties with this message
                </Text>
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

          <Row className={'mt-4'}>
            <Col span={8} className={'ml-4'}>
              <Form.Item name='repeat_alerts' className={'m-0'}>
                <Checkbox
                  defaultChecked={notRepeat}
                  onChange={(e) => setNotRepeat(e.target.checked)}
                >
                  Do not repeat an alert more than once
                </Checkbox>
                <Popover
                  placement='rightTop'
                  overlayInnerStyle={{ width: '300px' }}
                  title={null}
                  content={
                    <p className='m-0 m-2 text-gray-700'>
                      Check this box if you only want a single alert when
                      multiple events with the same filters occur
                    </p>
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
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={9} className={'ml-4'}>
              <Form.Item name='notifications' className={'m-0'}>
                <Checkbox
                  defaultChecked={notifications}
                  onChange={(e) => setNotifications(e.target.checked)}
                >
                  Set limit for alerts per day to
                </Checkbox>
                <div className='inline'>
                  <Select
                    bordered={false}
                    size='small'
                    className='m-0 inline'
                    style={{
                      width: 100
                    }}
                    defaultValue='5'
                    onChange={() => setAlertLimit(value)}
                  >
                    <Option value='5'>5 alerts</Option>
                    <Option value='10'>10 alerts</Option>
                    <Option value='15'>15 alerts</Option>
                    <Option value='15'>20 alerts</Option>
                  </Select>
                </div>
              </Form.Item>
            </Col>
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

          {/* <Row className={'mt-2 ml-2'}>
            <Col className={'m-0'}>
              <Form.Item name='webhook_enabled' className={'m-0'}>
                <Checkbox
                  defaultChecked={webhookEnabled}
                  onChange={(e) => setWebhookEnabled(e.target.checked)}
                >
                  Webhook
                </Checkbox>
              </Form.Item>
            </Col>
          </Row> */}
          {/* {webhookEnabled && (
            <>
              <Row className={'mt-2 ml-2'}>
                <Col span={12} className={'m-0'}>
                  <Text
                    type={'title'}
                    level={7}
                    color={'grey'}
                    extraClass={'m-0'}
                  >
                    Share an endpoint to receive alert notifications and trigger
                    more flows
                  </Text>
                </Col>
              </Row>
              <Row className={'mt-2 ml-2'}>
                <Col span={7}>
                  <Input className='fa-input' placeholder='Webhook URL'></Input>
                </Col>
                <Col span={6} className={'m-0 ml-2'}>
                  <Button type='primary'>Confirm</Button>
                </Col>
              </Row>
            </>
          )} */}
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
              {viewAlertDetails?.event}
            </Button>
          </Col>
        </Row>
        {viewAlertDetails?.filter?.length > 0 && (
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
              Configure your notification
            </Text>
          </Col>
        </Row>
        <Row className={'mt-4'}>
          <Col span={8} className={'ml-4'}>
            <Input
              disabled={true}
              className={'fa-input'}
              type={'number'}
              value={viewAlertDetails?.message}
            />
          </Col>
        </Row>
        <Row className={'mt-2 ml-4'}>
          <Col span={8}>
            <Checkbox disabled={true} checked={viewAlertDetails?.repeat_alerts}>
              Do not repeat an alert more than once
            </Checkbox>
          </Col>
        </Row>
        <Row className={'mt-2 ml-4'}>
          <Col span={8}>
            <Checkbox disabled={true} checked={viewAlertDetails?.alert_limit}>
              Set limit for alerts per day to
            </Checkbox>
          </Col>
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
            <Checkbox disabled={true} checked={viewAlertDetails?.slack}>
              Slack
            </Checkbox>
          </Col>
        </Row>
        {viewAlertDetails?.slack &&
          viewAlertDetails?.slack_channels.length > 0 && (
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
            {alertState.state == 'add' && renderEventForm()}

            {alertState.state == 'view' && renderEventView()}

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
  savedEventAlerts: state.global.eventAlerts,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  projectSettings: state.global.projectSettingsV1,
  groupBy: state.coreQuery.groupBy.event,
  groupByMagic: state.coreQuery.groupBy
});

export default connect(mapStateToProps, {
  fetchEventAlerts,
  deleteEventAlert,
  createEventAlert,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration,
  setGroupBy,
  delGroupBy,
  getUserProperties
})(EventBasedAlert);
