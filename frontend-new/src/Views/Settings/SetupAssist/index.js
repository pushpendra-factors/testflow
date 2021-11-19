import React, { useState } from 'react';
import {
    Row, Col, Modal, Button, Timeline
} from 'antd';
import { QrcodeOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import Website from './Website';

const SetupAssist = (props) => {
    const [current, setCurrent] = useState(0);

    return (
        <>
            <Modal title={null}
            visible={props.visible}
            footer={null}
            centered={false}
            mask={false}
            closable={false}
            className={'fa-modal--full-width'}
            >
                <div className={'fa-container'}>
                <Row gutter={[24, 24]} justify={'center'} className={'pt-12 pb-8 mt-0 '}>
                    <Col span={17}>
                    <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'}>Congratulations, Let's get started</Text>
                    <Text type={'title'} level={6} weight={'regular'} extraClass={'m-0'} color={'grey'}>The first step to get up and running with Factors is to get data into your project:</Text>
                    </Col>
                    <Col>
                        <Button type="default" size={'large'} style={{borderColor:'#1E89FF', color:'#1E89FF', background:'#fff'}} onClick={() => props.handleCancel()}><QrcodeOutlined style={{color:'#1E89FF'}} />Go to Dashboards</Button>
                    </Col>
                </Row>
                <Row gutter={[24, 24]} justify={'center'}>
                    <Col span={5}>
                        <Timeline>
                            <Timeline.Item><Text type={'title'} level={7} weight={'bold'} value={current} color ={current === 0 ? 'blue': 'grey'} onClick={() => setCurrent(0)}>Connect with your website data</Text></Timeline.Item>
                            <Timeline.Item><Text type={'title'} level={7} weight={'bold'} value={current} color ={current === 1 ? 'blue': 'grey'} onClick={() => setCurrent(1)}>Connect with your Ad platforms</Text></Timeline.Item>
                            <Timeline.Item><Text type={'title'} level={7} weight={'bold'} value={current} color ={current === 2 ? 'blue': 'grey'} onClick={() => setCurrent(2)}>Connect with your CRMS</Text></Timeline.Item>
                            <Timeline.Item><Text type={'title'} level={7} weight={'bold'} value={current} color ={current === 3 ? 'blue': 'grey'} onClick={() => setCurrent(3)}>Other integrations</Text></Timeline.Item>
                        </Timeline>
                    </Col>
                    <Col span={15}>
                        <Website />
                    </Col>
                </Row>
                </div>
            </Modal>
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
