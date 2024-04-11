import React, { useState, useEffect } from 'react';
import {
  Row,
  Col,
  Switch,
  Menu,
  Dropdown,
  Button,
  Tabs,
  Table,
  Tag,
  Space,
  message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import SmartEventsForm from './SmartEvents/SmartEventsForm';
import {
  fetchEventNames,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { MoreOutlined } from '@ant-design/icons';
import { removeSmartEvents, fetchSmartEvents } from 'Reducers/events';
import EmptyScreen from 'Components/EmptyScreen';

const { TabPane } = Tabs;

function Events({
  smart_events,
  fetchEventNames,
  activeProject,
  removeSmartEvents,
  fetchSmartEvents
}) {
  const [smartEvents, setsmartEvents] = useState(null);
  const [showSmartEventForm, setShowSmartEventForm] = useState(false);
  const [seletedEvent, setSeletedEvent] = useState(null);
  const [loading, setLoading] = useState(false);

  const menu = (values) => {
    return (
      <Menu>
        <Menu.Item key='0' onClick={() => confirmRemove(values)}>
          <a>Remove Event</a>
        </Menu.Item>
        <Menu.Item key='0' onClick={() => viewEvent(values)}>
          <a>View Event</a>
        </Menu.Item>
      </Menu>
    );
  };

  const columns = [
    {
      title: 'Display name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (values) => (
        <Dropdown overlay={() => menu(values)} trigger={['hover']}>
          <Button type='text' icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  const editEvent = (values) => {
    setSeletedEvent(values);
    setShowSmartEventForm(true);
  };

  const confirmRemove = (values) => {
    removeSmartEvents(activeProject?.id, values?.id)
      .then(() => {
        message.success('Custom Event removed!');
        fetchSmartEvents(activeProject?.id);
      })
      .catch((err) => {
        message.error(err?.data?.error);
        console.log('error in removing Smartevent:', err);
      });
    return false;
  };

  const viewEvent = (values) => {
    setSeletedEvent(values);
    setShowSmartEventForm(true);
  };

  useEffect(() => {
    setLoading(true);
    fetchSmartEvents(activeProject?.id)
      .then(() => {
        setLoading(false);
      })
      .catch((err) => {
        console.log('Fetch SmartEvents catch', err);
        setLoading(false);
      });
  }, [activeProject]);

  useEffect(() => {
    fetchEventNames(activeProject.id);
    if (smart_events) {
      let smartEventsArray = [];
      smart_events?.map((item, index) => {
        smartEventsArray.push({
          key: index,
          name: item.name,
          source: item?.expr?.source,
          actions: item
        });
      });
      setsmartEvents(smartEventsArray);
    }
  }, [smart_events]);
  const handleNewEvent = () => {
    setSeletedEvent(null);
    setShowSmartEventForm(true);
  };
  return (
    <div className={'fa-container'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          <div className={'mb-10'}>
            {!showSmartEventForm && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                      id={'fa-at-text--page-title'}
                    >
                      Events
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className={'flex justify-end'}>
                      <Button
                        type='primary'
                        onClick={handleNewEvent}
                        icon={<SVG name={'plus'} color={'white'} size={16} />}
                      >
                        New Event
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className={'mt-4'}>
                  <Col span={24}>
                    <div className={'mt-6'}>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0'}
                      >
                        Track and analyze user interactions in a way that aligns
                        perfectly with your business objectives.
                      </Text>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0 mt-2'}
                      >
                        Define and capture custom events that matter most to
                        your business, such as clicks, form submissions,
                        lifecycle stage changes, or other specific actions.{' '}
                        <a
                          href='https://help.factors.ai/en/articles/7284092-custom-events'
                          target='_blank'
                        >
                          Learn more
                        </a>
                      </Text>

                      <Tabs defaultActiveKey='1'>
                        <TabPane tab='Custom Events' key='1'>
                          {smartEvents && smartEvents.length > 0 ? (
                            <Table
                              className='fa-table--basic mt-4'
                              columns={columns}
                              dataSource={smartEvents}
                              pagination={false}
                              loading={loading}
                            />
                          ) : (
                            <EmptyScreen
                              title={`Set up unique touch points on your website that track user interactions that go beyond what's automatically tracked by Factors. For example, you can track signups, form submissions, lifecycle stage changes, or other specific actions.`}
                              learnMore={'https://help.factors.ai/'}
                            />
                          )}
                        </TabPane>
                      </Tabs>
                    </div>
                  </Col>
                </Row>
              </>
            )}
            {showSmartEventForm && (
              <>
                <SmartEventsForm
                  seletedEvent={seletedEvent}
                  setShowSmartEventForm={setShowSmartEventForm}
                />
              </>
            )}
          </div>
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  smart_events: state.events.smart_events,
  activeProject: state.global.active_project
});

export default connect(mapStateToProps, {
  fetchEventNames,
  removeSmartEvents,
  fetchSmartEvents
})(Events);
