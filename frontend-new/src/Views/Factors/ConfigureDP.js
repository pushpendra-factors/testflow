import React, { useState, useEffect } from 'react';
import {
  Row, Col, Modal, Button, Progress
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PlusOutlined, SlackOutlined } from '@ant-design/icons';
import { fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties } from 'Reducers/factors';
import { connect } from 'react-redux';

// const suggestionList = [
//   {
//     name: 'User_SignUp',
//     img: ''
//   },
//   {
//     name: 'chargebee.com...ebinar/Thank-you',
//     img: ''
//   },
//   {
//     name: 'chargebee.com/Plans',
//     img: ''
//   },
//   {
//     name: 'chargebee.com/Features',
//     img: ''
//   },
//   {
//     name: 'chargebee.com...ebinar/Thank-you',
//     img: ''
//   }
// ];

const ConfigureDP = (props) => {
  const {
    activeProject, fetchFactorsTrackedEvents, tracked_events, fetchFactorsTrackedUserProperties, tracked_user_property
  } = props;
  const [activeEventsTracked, setActiveEventsTracked] = useState(0);
  const [activeUserProperties, setActiveUserProperties] = useState(0);

  useEffect(() => {
    setActiveEventsTracked(0);
    setActiveUserProperties(0);
    if (!tracked_events || !tracked_user_property) {
      const getData = async () => {
        await fetchFactorsTrackedEvents(activeProject.id);
        await fetchFactorsTrackedUserProperties(activeProject.id);
      };
      getData();
    }
    if (tracked_events) {
      let activeEvents = 0;
      tracked_events.map((event) => {
        if (event.is_active) {
          activeEvents = activeEvents + 1;
        }
      });
      setActiveEventsTracked(activeEvents);
    };

    if (tracked_user_property) {
      let activeUserProperties = 0;
      tracked_user_property.map((event) => {
        if (event.is_active) {
          activeUserProperties = activeUserProperties + 1;
        }
      });
      setActiveUserProperties(activeUserProperties);
    }
  }, [tracked_events, tracked_user_property]);

  return (
    <Modal
    title={null}
    visible={props.visible}
    footer={null}
    centered={false}
    zIndex={1005}
    mask={false}
    closable={false}
    className={'fa-modal--full-width'}
  >

    <div className={'fa-modal--header'}>
      <div className={'fa-container'}>
        <Row justify={'space-between'} className={'py-4 m-0 '}>
          <Col>
            <SVG name={'brand'} size={40}/>
          </Col>
          <Col>
            <Button size={'large'} type="text" onClick={() => props.handleCancel()}><SVG name="times"></SVG></Button>
          </Col>
        </Row>
      </div>
    </div>

    <div className={'fa-container'}>
        <Row gutter={[24, 24]} justify={'center'}>
            <Col span={20}>
                <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 mt-8'} >What’s being tracked?</Text>
            </Col>
            <Col span={20}>
                <div className={'border--thin-2 py-8 px-8 background-color--brand-color-1 border-radius--sm'}>
                <div className={'px-4'}>
                    <ul style={{ listStyle: 'disc' }}>
                        <li><Text type={'paragraph'} extraClass={'m-0 mb-2'} >You can track upto 50 events and user properties while analyzing Factors</Text></li>
                        <li><Text type={'paragraph'} extraClass={'m-0 mb-2'} >Your Goals should be comprised of these datapoints</Text></li>
                        <li><Text type={'paragraph'} extraClass={'m-0 mb-2'} >Please note that it may take a few days for us to index newly added data points in your queries</Text></li>
                        <li><Text type={'paragraph'} extraClass={'m-0 mb-2'} >Apart from these, all your active campaigns will be tracked. Learn more about <a>Configuring your Ad accounts</a></Text></li>
                    </ul>
                </div>
                <div className={'flex items-center mt-6'}>
                <Button size={'large'} type={'primary'}>Got it</Button>
                <Button ghost size={'large'} className={'ml-4'}>learn more</Button>

                </div>
                </div>
            </Col>
            <Col span={20}>
                <Row gutter={[24, 24]} justify={'center'}>

                    <Col span={12}>
                        <div className={'pr-4'}>
                        <Row gutter={[24, 12]} justify={'center'}>
                            <Col span={24}>
                                <div className={'flex items-center mt-6'}>
                                    <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >Events</Text><Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'} >{tracked_events && `(${tracked_events.length})`}</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <Progress percent={60} strokeColor={'#ACA4DE'} success={{ percent: 30, strokeColor: '#5949BC' }} showInfo={false} />
                            </Col>
                            <Col span={24}>
                                <div className={'flex w-full justify-between items-center border-bottom--thin-2 pb-4'}>
                                    <div className={'flex items-center'}>
                                        <div className={'flex flex-col'}>
                                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >{activeEventsTracked}</Text>
                                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >Tracked</Text>
                                        </div>
                                        <div className={'flex flex-col ml-6'}>
                                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >{tracked_events && `${tracked_events.length - activeEventsTracked}`}</Text>
                                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >In Queue</Text>
                                        </div>
                                    </div>
                                    <div>
                                        <Button size={'small'} icon={<PlusOutlined />}></Button>
                                    </div>
                                </div>
                            </Col>
                            <Col span={24}>
                            {tracked_events && tracked_events.map((event, index) => {
                              return (
                                    <div key={index} className={'flex justify-between items-center mt-2'}>
                                        <Text type={'title'} level={7} weight={'thin'} extraClass={'m-0'} ><SlackOutlined className={'mr-1'} />{event.name}</Text>
                                    </div>
                              );
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
                                    <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >User Properties</Text><Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'} >{tracked_user_property && `(${tracked_user_property.length})`}</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <Progress percent={25} strokeColor={'#ACA4DE'} success={{ percent: 20, strokeColor: '#5949BC' }} showInfo={false} />
                            </Col>
                            <Col span={24}>
                                <div className={'flex w-full justify-between items-center border-bottom--thin-2 pb-4'}>
                                    <div className={'flex items-center'}>
                                        <div className={'flex flex-col'}>
                                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >{activeUserProperties}</Text>
                                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >Tracked</Text>
                                        </div>
                                        <div className={'flex flex-col ml-6'}>
                                        <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >{tracked_user_property && `${tracked_user_property.length - activeUserProperties}`}</Text>
                                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >In Queue</Text>
                                        </div>
                                    </div>
                                    <div>
                                        <Button size={'small'} icon={<PlusOutlined />}></Button>
                                    </div>
                                </div>
                            </Col>
                            <Col span={24}>
                            {tracked_user_property && tracked_user_property.map((event, index) => {
                              return (
                                      <div key={index} className={'flex justify-between items-center mt-2'}>
                                          <Text type={'title'} level={7} weight={'thin'} extraClass={'m-0'} ><SlackOutlined className={'mr-1'} />{event.user_property_name}</Text>
                                      </div>
                              );
                            })}
                            </Col>
                        </Row>
                        </div>
                    </Col>

                </Row>
            </Col>
        </Row>

    </div>

    </Modal>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    tracked_events: state.factors.tracked_events,
    tracked_user_property: state.factors.tracked_user_property
  };
};

export default connect(mapStateToProps, { fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties })(ConfigureDP);
