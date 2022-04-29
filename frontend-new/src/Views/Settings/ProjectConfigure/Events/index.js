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
  message,
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import SmartEventsForm from './SmartEvents/SmartEventsForm';
import {
  fetchEventNames,
  getUserProperties,
} from 'Reducers/coreQuery/middleware';
import { MoreOutlined } from '@ant-design/icons';
import { removeSmartEvents, fetchSmartEvents } from 'Reducers/events';

const { TabPane } = Tabs;

function Events({
  smart_events,
  fetchEventNames,
  activeProject,
  removeSmartEvents,
  fetchSmartEvents,
}) {
  const [smartEvents, setsmartEvents] = useState(null);
  const [showSmartEventForm, setShowSmartEventForm] = useState(false);
  const [seletedEvent, setSeletedEvent] = useState(null);

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
      title: 'Diplay name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <span className={'capitalize'}>{text}</span>,
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      render: (text) => <span className={'capitalize'}>{text}</span>,
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (values) => (
        <Dropdown overlay={() => menu(values)} trigger={['hover']}>
          <Button type='text' icon={<MoreOutlined />} />
        </Dropdown>
      ),
    },
  ];

  const editEvent = (values) => {
    setSeletedEvent(values);
    setShowSmartEventForm(true);
  };

  const confirmRemove = (values) => {
    removeSmartEvents(values?.project_id, values?.id)
      .then(() => {
        message.success('Custom Event removed!');
        fetchSmartEvents(values?.project_id);
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
    fetchEventNames(activeProject.id);
    if (smart_events) {
      let smartEventsArray = [];
      smart_events?.map((item, index) => {
        smartEventsArray.push({
          key: index,
          name: item.name,
          source: item?.expr?.source,
          actions: item,
        });
      });
      setsmartEvents(smartEventsArray);
    }
  }, [smart_events]);

  return (
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className={'mb-10 pl-4'}>
            {!showSmartEventForm && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      Events
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className={'flex justify-end'}>
                      <Button
                        size={'large'}
                        onClick={() => {
                          setSeletedEvent(null);
                          setShowSmartEventForm(true);
                        }}
                      >
                        <SVG name={'plus'} extraClass={'mr-2'} size={16} />
                        New Event
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className={'mt-4'}>
                  <Col span={24}>
                    <div className={'mt-6'}>
                      <Tabs defaultActiveKey='1'>
                        <TabPane tab='Custom Events' key='1'>
                          <Table
                            className='fa-table--basic mt-4'
                            columns={columns}
                            dataSource={smartEvents}
                            pagination={false}
                          />
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
  activeProject: state.global.active_project,
});

export default connect(mapStateToProps, {
  fetchEventNames,
  removeSmartEvents,
  fetchSmartEvents,
})(Events);
