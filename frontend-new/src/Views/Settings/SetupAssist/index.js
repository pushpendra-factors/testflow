import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button, Timeline
} from 'antd';
import { QrcodeOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import Website from './Website';
import AdPlatforms from './AdPlatforms';
import CRMS from './CRMS';
import OtherIntegrations from './OtherIntegrations';
import { useHistory } from 'react-router-dom';
import Lottie from 'react-lottie';
import setupAssistData from '../../../assets/lottie/Final Jan 3 Setupassist.json'
import { fetchProjectSettingsV1 } from 'Reducers/global';
const axios = require('axios').default;

const SetupAssist = ({currentAgent, integration, activeProject, fetchProjectSettingsV1}) => {
    const [current, setCurrent] = useState(0);
    const [showModal,setShowModal] = useState(false);
    const history = useHistory();

    const APIKEY = '69137c15-00a5-4d12-91e7-9641797e9572';
    let email = currentAgent.email;
    let ownerData;

    axios.get(`https://api.hubapi.com/contacts/v1/contact/email/${email}/profile?hapikey=${APIKEY}`)
    .then(response => response.json())
    .then( data => {
        ownerData = data['properties'].hubspot_owner_id.value;
    })

    const meetLink = ownerData === '116046946'? 'https://mails.factors.ai/meeting/factors/prajwalsrinivas0'
                    :ownerData === '116047122'? 'https://calendly.com/priyanka-267/30min'
                    :ownerData === '116053799'? 'https://factors1.us4.opv1.com/meeting/factors/ralitsa': null;

    const defaultOptions = {
        loop: true,
        autoplay: true,
        animationData: setupAssistData,
        rendererSettings: {
            preserveAspectRatio: "xMidYMid slice"
        }
    };

    useEffect(() => {
        fetchProjectSettingsV1(activeProject.id).then(() => {
            console.log('fetch project settings success');
        });
    }, [activeProject, current]);

    integration = integration?.project_settings || integration;

    const checkIntegration = integration?.int_segment || 
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit || integration?.int_completed;

    return (
        <>
            <div className={'fa-container'}>
                <Row gutter={[24, 24]} justify={'center'} className={'pt-24 pb-12 mt-0 '}>
                    <Col span={ checkIntegration ? 17 : 20}>
                        <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'}>Let's get started</Text>
                        <Text type={'title'} level={6} weight={'regular'} extraClass={'m-0'} color={'grey'}>The first step to get up and running with Factors is to get data into your project:</Text>
                        <img src='../../assets/images/Illustration=pop gift.png' style={{width: '100%',maxWidth: '80px', marginLeft:'610px',marginTop:'-80px'}}/>
                    </Col>
                    <Col>
                    { checkIntegration ?
                        <Button type="default" size={'large'} style={{borderColor:'#1E89FF', color:'#1E89FF', background:'#fff', marginTop:'1rem'}} onClick={() => setShowModal(true)}><SVG name={'dashboard'} size={20} color="blue"/>Go to Dashboards</Button>
                    : null}
                    </Col>
                </Row>
                <Row gutter={[24, 24]} justify={'center'}>
                    <Col span={5} style={{paddingRight: '20px'}}>
                        <Timeline>
                            <Timeline.Item color ={current === 0 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 0 ? 'brand-color': null} onClick={() => setCurrent(0)}>Connect with your website data</Text></Timeline.Item>
                            <Timeline.Item color ={current === 1 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 1 ? 'brand-color': null} onClick={() => setCurrent(1)}>Connect with your Ad platforms</Text></Timeline.Item>
                            <Timeline.Item color ={current === 2 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 2 ? 'brand-color': null} onClick={() => setCurrent(2)}>Connect with your CRMS</Text></Timeline.Item>
                            <Timeline.Item color ={current === 3 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 3 ? 'brand-color': null} onClick={() => setCurrent(3)}>Other integrations</Text></Timeline.Item>
                        </Timeline>
                        <Row className={'pt-16'}>
                            <Col>
                                <Text type={'title'} level={4} weight={'bold'} extraClass={'pb-4 m-0'}>Setup a call with a rep</Text>
                                <Text type={'title'} level={6} extraClass={'pb-6 m-0'}>We are always happy to assist you</Text>
                                <a href={meetLink} target='_blank' ><Button type={'primary'}>Setup Call</Button></a>
                                <img src='../../assets/images/character-1.png' style={{width: '100%',maxWidth: '80px',marginLeft:'110px', marginTop:'-30px'}}/>
                            </Col>
                        </Row>
                    </Col>
                    <Col span={16} style={{padding: '0px 0px 0px 30px'}}>
                        {current === 0 ? <Website />: current === 1 ? <AdPlatforms />: current === 2 ? <CRMS /> : <OtherIntegrations />}
                    </Col>
                </Row>
            </div>

            <Modal
                title={null}
                visible={showModal}
                footer={null}
                centered={true}
                mask={true}
                maskClosable={false}
                maskStyle={{backgroundColor: 'rgb(0 0 0 / 70%)'}}
                closable={true}
                onCancel={()=> setShowModal(false)}
                className={'fa-modal--regular'}
            >
                <div className={'fa-container'}>
                    <Row className={'mt-8'}>
                        <Col>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>You are leaving project setup assist now</Text>
                            <Text type={'title'} level={7} weight={'regular'} extraClass={'ml-5'} color={'grey'}>You can always access it from the bottom left corner</Text>
                        </Col>
                    </Row>
                    <Row style={{marginLeft:'80px'}}>
                        <Col>
                            <Lottie 
                            options={defaultOptions}
                            height={200}
                            width={200}
                            />
                        </Col>
                    </Row>
                    <Row style={{marginLeft: '120px'}} className={'pb-4'}>
                        <Col>
                            <Button type={'primary'} onClick={() => history.push('/')}>Got it, continue</Button>
                        </Col>
                    </Row>
                </div>
            </Modal>
        </>
    );
};

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentAgent: state.agent.agent_details,
    integration: state.global.currentProjectSettings
});

export default connect(mapStateToProps, { fetchProjectSettingsV1 })(SetupAssist);
