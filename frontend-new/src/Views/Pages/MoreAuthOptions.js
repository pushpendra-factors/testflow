import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';

const Video = ({showModal, setShowModal}) => {

    const history = useHistory();
    const routeChange = (url) => {
        history.push(url);
    };

    return (
        <>
              <Modal
                title={null}
                visible={showModal}
                footer={
                    <Row className={'mb-2 mt-2'}>
                      <Col span={24}>
                          <div className={'flex flex-col justify-center items-center'} >
                          <Text type={'paragraph'} mini weight={'bold'} color={'grey-2'}>Don’t have an account? <a onClick={() => routeChange('/signup')}> Sign Up</a></Text>
                          </div>
                      </Col>
                    </Row>
                }
                centered={true}
                mask={true}
                maskClosable={false}
                maskStyle={{backgroundColor: 'rgb(0 0 0 / 70%)'}}
                closable={true}
                onCancel={()=> setShowModal(false)}
                className={`fa-modal--regular`}
                width={368}
            >
                <div className={'m-0 mb-2'}>
                    <Row justify={'center'} className={'mt-2'}>
                        <Col>
                            <Text type={'title'} level={5} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Other Login Options</Text>
                        </Col>
                    </Row>
                    <Row>
                        <Col span={24}>
                            <div className={'flex flex-col justify-center items-center mt-5'} >
                                <Button type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'Salesforce_ads'} size={24} /> Continue with Salesforce</Button>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'flex flex-col justify-center items-center mt-5'} >
                                <Button type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'Microsoft'} size={24} /> Continue with Microsoft</Button>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'flex flex-col justify-center items-center mt-5'} >
                                <Button type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'Hubspot_ads'} size={24} /> Continue with Hubspot</Button>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'flex flex-col justify-center items-center mt-5'} >
                                <Button type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'Facebook_ads'} size={24} /> Continue with Facebook</Button>
                            </div>
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
});

export default connect(mapStateToProps, null)(Video);
