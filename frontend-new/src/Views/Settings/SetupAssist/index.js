import React, { useState } from 'react';
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

const SetupAssist = () => {
    const [current, setCurrent] = useState(0);
    const history = useHistory();

    return (
        <>
            <div className={'fa-container'}>
                <Row gutter={[24, 24]} justify={'center'} className={'pt-24 pb-12 mt-0 '}>
                    <Col span={17}>
                        <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'}>Congratulations, Let's get started</Text>
                        <Text type={'title'} level={6} weight={'regular'} extraClass={'m-0'} color={'grey'}>The first step to get up and running with Factors is to get data into your project:</Text>
                        <img src='../../assets/images/Illustration=pop gift.png' style={{width: '100%',maxWidth: '80px', marginLeft:'650px',marginTop:'-80px'}}/>
                    </Col>
                    <Col>
                        <Button type="default" size={'large'} style={{borderColor:'#1E89FF', color:'#1E89FF', background:'#fff'}} onClick={() => history.push('/')}><QrcodeOutlined style={{color:'#1E89FF'}} />Go to Dashboards</Button>
                    </Col>
                </Row>
                <Row gutter={[24, 24]} justify={'center'}>
                    <Col span={5}>
                        <Timeline>
                            <Timeline.Item color ={current === 0 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 0 ? 'blue': null} onClick={() => setCurrent(0)}>Connect with your website data</Text></Timeline.Item>
                            <Timeline.Item color ={current === 1 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 1 ? 'blue': null} onClick={() => setCurrent(1)}>Connect with your Ad platforms</Text></Timeline.Item>
                            <Timeline.Item color ={current === 2 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 2 ? 'blue': null} onClick={() => setCurrent(2)}>Connect with your CRMS</Text></Timeline.Item>
                            <Timeline.Item color ={current === 3 ? 'blue': 'grey'}><Text type={'title'} level={6} style={{paddingBottom:'20px', cursor: 'pointer'}} color ={current === 3 ? 'blue': null} onClick={() => setCurrent(3)}>Other integrations</Text></Timeline.Item>
                        </Timeline>
                        <Row style={{width:'120vh'}} className={'pt-20'}>
                            <Col span={5}>
                                <Text type={'title'} level={4} weight={'bold'} extraClass={'pb-4 m-0'}>Setup a call with a rep</Text>
                                <Text type={'title'} level={6} extraClass={'pb-6 m-0'}>We are always happy to assist you</Text>
                                <Button type={'primary'}>Setup Call</Button>
                                <img src='../../assets/images/character-1.png' style={{width: '100%',maxWidth: '80px',marginLeft:'100px'}}/>
                            </Col>
                        </Row>
                    </Col>
                    <Col span={15} style={{padding: '0'}}>
                        {current === 0 ? <Website />: current === 1 ? <AdPlatforms />: current === 2 ? <CRMS /> : <OtherIntegrations />}
                    </Col>
                </Row>
            </div>
        </>
    )
}

const mapStateToProps = (state) => {
    return ({
        agent: state.agent.agent_details
    }
    );
};

export default connect(mapStateToProps)(SetupAssist);
