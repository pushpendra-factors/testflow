import React, { useState, useEffect } from 'react';
import { Row, Col, Modal, Button, Progress, message } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PlusOutlined, SlackOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import {
  addEventToTracked,
  addUserPropertyToTracked,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  delEventTracked,
  delUserPropertyTracked,
} from 'Reducers/factors';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
const { confirm } = Modal;

const ConfigureDP = (props) => {
  const {
    activeProject,
    tracked_events,
    tracked_user_property,
    userProperties,
    events,
    addEventToTracked,
    addUserPropertyToTracked,
    fetchFactorsTrackedEvents,
    fetchFactorsTrackedUserProperties,
    delEventTracked,
    delUserPropertyTracked,
  } = props;
  const [activeEventsTracked, setActiveEventsTracked] = useState(0);
  const [InQueueEventsEvents, setInQueueEventsEvents] = useState(0);
  const [activeUserProperties, setActiveUserProperties] = useState(0);
  const [InQueueUserProperties, setInQueueUserProperties] = useState(0);

  const [showDropDown, setShowDropDown] = useState(false);
  const [showDropDown1, setShowDropDown1] = useState(false);

  const [showInfo, setshowInfo] = useState(true);

  useEffect(() => {
    setActiveEventsTracked(0);
    setInQueueEventsEvents(0);
    setActiveUserProperties(0);
    setInQueueUserProperties(0);
    if (tracked_events) {
      let activeEvents = 0;
      let inQueueEvents = 0;
      tracked_events.map((event) => {
        if (
          event.is_active == true &&
          event.last_tracked_at == '0001-01-01T00:00:00Z'
        ) {
          activeEvents = activeEvents + 1;
        }
        if (
          event.is_active == true &&
          event.last_tracked_at !== '0001-01-01T00:00:00Z'
        ) {
          inQueueEvents = inQueueEvents + 1;
        }
      });
      setActiveEventsTracked(activeEvents);
      setInQueueEventsEvents(inQueueEvents);
    }

    if (tracked_user_property) {
      let activeUserProperties = 0;
      let inQueueUserProperties = 0;
      tracked_user_property.map((event) => {
        if (
          event.is_active == true &&
          event.last_tracked_at == '0001-01-01T00:00:00Z'
        ) {
          activeUserProperties = activeUserProperties + 1;
        }
        if (
          event.is_active == true &&
          event.last_tracked_at !== '0001-01-01T00:00:00Z'
        ) {
          inQueueUserProperties = inQueueUserProperties + 1;
        }
      });
      setActiveUserProperties(activeUserProperties);
      setInQueueUserProperties(inQueueUserProperties);
    }
  }, [activeProject, tracked_events, tracked_user_property]);

  useEffect(() => {
    fetchFactorsTrackedEvents(activeProject.id);
    fetchFactorsTrackedUserProperties(activeProject.id); 
  }, [activeProject]);

  const onChangeEventDD = (grp, value) => {
    // console.log('grp, value onChangeEventDD',grp, value);
    setShowDropDown(false);
    const EventData = {
      event_name: `${value[1] ? value[1] : value[0]}`,
    };
    addEventToTracked(activeProject.id, EventData)
      .then(() => {
        message.success('Event added successfully!');
        fetchFactorsTrackedEvents(activeProject.id);
      })
      .catch((err) => {
        const ErrMsg = err?.data.error
          ? err.data.error
          : `Oops! Something went wrong!`;
        message.error(ErrMsg);
      });
  };

  const onChangeUserPropertiesDD = (grp, value) => {
    console.log('grp, value onChangeUserPropertiesDD', grp, value);
    setShowDropDown1(false);
    const UserPropertyData = {
      user_property_name: `${value[1] ? value[1] : value[0]}`,
    };
    addUserPropertyToTracked(activeProject.id, UserPropertyData)
      .then(() => {
        message.success('User Property added successfully!');
        fetchFactorsTrackedUserProperties(activeProject.id);
      })
      .catch((err) => {
        const ErrMsg = err?.data.error
          ? err.data.error
          : `Oops! Something went wrong!`;
        message.error(ErrMsg);
      });
  };

  const DeleteEvent = (id) => {
    confirm({
      title: 'Do you want to remove this Event?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      zIndex: '1020',
      onOk() {
        delEventTracked(activeProject.id, { id: id })
          .then(() => {
            message.success('Event removed!');
            fetchFactorsTrackedEvents(activeProject.id);
          })
          .catch((err) => {
            const ErrMsg = err?.data?.error
              ? err.data.error
              : `Oops! Something went wrong!`;
            message.error(ErrMsg);
          });
      },
    });
  };

  const DeleteUserProperty = (id) => {
    confirm({
      title: 'Do you want to remove this User property?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      zIndex: '1020',
      onOk() {
        delUserPropertyTracked(activeProject.id, { id: id })
          .then(() => {
            message.success('User property removed!');
            fetchFactorsTrackedUserProperties(activeProject.id);
          })
          .catch((err) => {
            const ErrMsg = err?.data?.error
              ? err.data.error
              : `Oops! Something went wrong!`;
            message.error(ErrMsg);
          });
      },
    });
  };

  return (
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <Row className='w-full' justify={'center'}>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
                Top Events and Properties
              </Text>
              <Text
                type={'title'}
                level={7}
                color={'grey-2'}
                extraClass={'m-0 mb-8'}
              >
                The events and properties under this list will be considered as
                the top events and properties of the project and will be tracked
                frequently.{' '}
              </Text>
            </Col>
            {showInfo && (
              <Col span={24}>
                <div
                  className={
                    'border--thin-2 py-8 px-8 background-color--brand-color-1 border-radius--sm'
                  }
                >
                  <div className={'px-4'}>
                    <ul style={{ listStyle: 'disc' }}>
                      <li>
                        <Text type={'paragraph'} extraClass={'m-0 mb-2'}>
                          You can track upto 50 events and user properties while
                          analyzing Factors
                        </Text>
                      </li>
                      <li>
                        <Text type={'paragraph'} extraClass={'m-0 mb-2'}>
                          Your Goals should be comprised of these datapoints
                        </Text>
                      </li>
                      <li>
                        <Text type={'paragraph'} extraClass={'m-0 mb-2'}>
                          Please note that it may take a few days for us to
                          index newly added data points in your queries
                        </Text>
                      </li>
                      <li>
                        <Text type={'paragraph'} extraClass={'m-0 mb-2'}>
                          Apart from these, all your active campaigns will be
                          tracked.
                          {/* Learn more about <a>Configuring your Ad accounts</a> */}
                        </Text>
                      </li>
                    </ul>
                  </div>
                  <div className={'flex items-center mt-6'}>
                    <Button
                      size={'large'}
                      type={'primary'}
                      onClick={() => setshowInfo(false)}
                    >
                      Got it
                    </Button>
                    {/* <Button ghost size={'large'} className={'ml-4'}>Learn more</Button> */}
                  </div>
                </div>
              </Col>
            )}
            <Col span={24}>
              <Row justify={'center'}>
                <Col span={12}>
                  <div className={'pr-4'}>
                    <Row gutter={[24, 12]} justify={'center'}>
                      <Col span={24}>
                        <div className={'flex items-center mt-6'}>
                          <Text
                            type={'title'}
                            level={4}
                            weight={'bold'}
                            extraClass={'m-0'}
                          >
                            Events
                          </Text>
                          <Text
                            type={'title'}
                            level={4}
                            color={'grey'}
                            extraClass={'m-0 ml-2'}
                          >{`(${
                            activeEventsTracked + InQueueEventsEvents
                          })`}</Text>
                        </div>
                      </Col>
                      <Col span={24}>
                        <Progress
                          percent={
                            activeEventsTracked * 2 + InQueueEventsEvents * 2
                          }
                          strokeColor={'#ACA4DE'}
                          success={{
                            percent: activeEventsTracked * 2,
                            strokeColor: '#1E89FF',
                          }}
                          showInfo={false}
                        />
                      </Col>
                      <Col span={24}>
                        <div
                          className={
                            'flex w-full justify-between items-center border-bottom--thin-2 pb-4'
                          }
                        >
                          <div className={'flex items-center'}>
                            <div className={'flex flex-col'}>
                              <Text
                                type={'title'}
                                level={4}
                                weight={'bold'}
                                extraClass={'m-0'}
                              >
                                {activeEventsTracked}
                              </Text>
                              <Text
                                type={'title'}
                                level={7}
                                color={'grey'}
                                extraClass={'m-0'}
                              >
                                Tracked
                              </Text>
                            </div>
                            <div className={'flex flex-col ml-6'}>
                              <Text
                                type={'title'}
                                level={4}
                                weight={'bold'}
                                extraClass={'m-0'}
                              >
                                {InQueueEventsEvents}
                              </Text>
                              <Text
                                type={'title'}
                                level={7}
                                color={'grey'}
                                extraClass={'m-0'}
                              >
                                In Queue
                              </Text>
                            </div>
                          </div>
                          <div className={'relative'}>
                            {!showDropDown && (
                              <Button
                                onClick={() => setShowDropDown(true)}
                                size={'small'}
                                type='text'
                              >
                                <SVG name='plus' color={'grey'} />
                              </Button>
                            )}
                            {showDropDown && (
                              <>
                                <GroupSelect2
                                  extraClass={'right-0'}
                                  groupedProperties={events ? events : null}
                                  placeholder='Select Events'
                                  optionClick={(group, val) =>
                                    onChangeEventDD(group, val)
                                  }
                                  onClickOutside={() => setShowDropDown(false)}
                                />
                              </>
                            )}
                          </div>
                        </div>
                      </Col>
                      <Col span={24}>
                        {tracked_events &&
                          tracked_events.map((event, index) => {
                            if (event.is_active) {
                              return (
                                <div
                                  key={index}
                                  className={
                                    'flex items-center justify-between px-4 py-2 fa-cdp--item'
                                  }
                                >
                                  <div className={'flex items-center'}>
                                    <SVG
                                      size={16}
                                      name={'event'}
                                      color={'purple'}
                                      className={'mr-1'}
                                    />
                                    <Text
                                      type={'title'}
                                      level={7}
                                      weight={'thin'}
                                      extraClass={'m-0 ml-2'}
                                    >
                                      {event.name}
                                    </Text>
                                  </div>
                                  <Button
                                    onClick={() => DeleteEvent(event.id)}
                                    className={'fa-cdp--action fa-button-ghost'}
                                    size={'small'}
                                    type='text'
                                  >
                                    <SVG name='delete' color={'grey'} />
                                  </Button>
                                </div>
                              );
                            } else return null;
                          })}
                      </Col>
                    </Row>
                  </div>
                </Col>

                <Col span={12}>
                  <div className={'pl-4'}>
                    <Row gutter={[24, 12]} justify={'center'}>
                      <Col span={24}>
                        <div className={'flex items-center mt-6'}>
                          <Text
                            type={'title'}
                            level={4}
                            weight={'bold'}
                            extraClass={'m-0'}
                          >
                            User Properties
                          </Text>
                          <Text
                            type={'title'}
                            level={4}
                            color={'grey'}
                            extraClass={'m-0 ml-2'}
                          >{`(${
                            activeUserProperties + InQueueUserProperties
                          })`}</Text>
                        </div>
                      </Col>
                      <Col span={24}>
                        <Progress
                          percent={
                            InQueueUserProperties * 2 + activeUserProperties * 2
                          }
                          strokeColor={'#ACA4DE'}
                          success={{
                            percent: activeUserProperties * 2,
                            strokeColor: '#1E89FF',
                          }}
                          showInfo={false}
                        />
                      </Col>
                      <Col span={24}>
                        <div
                          className={
                            'flex w-full justify-between items-center border-bottom--thin-2 pb-4'
                          }
                        >
                          <div className={'flex items-center'}>
                            <div className={'flex flex-col'}>
                              <Text
                                type={'title'}
                                level={4}
                                weight={'bold'}
                                extraClass={'m-0'}
                              >
                                {activeUserProperties}
                              </Text>
                              <Text
                                type={'title'}
                                level={7}
                                color={'grey'}
                                extraClass={'m-0'}
                              >
                                Tracked
                              </Text>
                            </div>
                            <div className={'flex flex-col ml-6'}>
                              <Text
                                type={'title'}
                                level={4}
                                weight={'bold'}
                                extraClass={'m-0'}
                              >
                                {InQueueUserProperties}
                              </Text>
                              <Text
                                type={'title'}
                                level={7}
                                color={'grey'}
                                extraClass={'m-0'}
                              >
                                In Queue
                              </Text>
                            </div>
                          </div>
                          <div>
                            {!showDropDown1 && (
                              <Button
                                onClick={() => setShowDropDown1(true)}
                                size={'small'}
                                type='text'
                              >
                                <SVG name='plus' color={'grey'} />
                              </Button>
                            )}
                            {showDropDown1 && (
                              <>
                                <GroupSelect2
                                  extraClass={'right-0'}
                                  groupedProperties={
                                    userProperties
                                      ? [
                                          {
                                            label: 'User Properties',
                                            icon: 'fav',
                                            values: userProperties,
                                          },
                                        ]
                                      : null
                                  }
                                  placeholder='Select User Properties'
                                  optionClick={(group, val) =>
                                    onChangeUserPropertiesDD(group, val)
                                  }
                                  onClickOutside={() => setShowDropDown1(false)}
                                />
                              </>
                            )}
                          </div>
                        </div>
                      </Col>
                      <Col span={24}>
                        {tracked_user_property &&
                          tracked_user_property.map((event, index) => {
                            if (event.is_active) {
                              return (
                                <div
                                  key={index}
                                  className={
                                    'flex items-center justify-between px-4 py-2 fa-cdp--item'
                                  }
                                >
                                  <div className={'flex items-center'}>
                                    <SVG
                                      size={16}
                                      name={'user'}
                                      color={'purple'}
                                      className={'mr-1'}
                                    />
                                    <Text
                                      type={'title'}
                                      level={7}
                                      weight={'thin'}
                                      extraClass={'m-0 ml-2'}
                                    >
                                      {event.user_property_name}
                                    </Text>
                                  </div>
                                  <Button
                                    onClick={() => DeleteUserProperty(event.id)}
                                    className={'fa-cdp--action fa-button-ghost'}
                                    size={'small'}
                                    type='text'
                                  >
                                    <SVG name='delete' color={'grey'} />
                                  </Button>
                                </div>
                              );
                            } else return null;
                          })}
                      </Col>
                    </Row>
                  </div>
                </Col>
              </Row>
            </Col>
          </Row>
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    tracked_events: state.factors.tracked_events,
    tracked_user_property: state.factors.tracked_user_property,
    userProperties: state.coreQuery.userProperties,
    events: state.coreQuery.eventOptions,
  };
};

export default connect(mapStateToProps, {
  addEventToTracked,
  delEventTracked,
  delUserPropertyTracked,
  addUserPropertyToTracked,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
})(ConfigureDP);
