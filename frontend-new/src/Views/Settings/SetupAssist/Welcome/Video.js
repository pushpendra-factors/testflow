import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import styles from './index.module.scss';

const Video = ({showModal, setShowModal}) => {

    const closeBTn = () => {
        setShowModal(false);
        let videos = document.querySelectorAll('iframe, video');
        Array.prototype.forEach.call(videos, function (video) {
            if (video.tagName.toLowerCase() === 'video') {
                video.pause();
            } else {
                let src = video.src;
                video.src = src;
            }
        });
    }

    return (
        <>
            <Modal
                title={null}
                visible={showModal}
                footer={null}
                centered={true}
                mask={true}
                maskClosable={false}
                maskStyle={{backgroundColor: 'rgb(0 0 0 / 70%)'}}
                closable={false}
                onCancel={()=> setShowModal(false)}
                className={'fa-modal--regular'}
                width={700}
            >
                <div className={'m-0 '}>
                    <Row className={'mt-2 mb-4'}>
                        <Col>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>See how Factors works</Text>
                            <Text type={'title'} level={7} weight={'regular'} extraClass={'m-0'} color={'grey'}>Here’s a quick display of the platform’s features and capabilities so you know exactly how you can drive revenue and pipeline with Factors</Text>
                        </Col>
                    </Row>
                    <div className={`${styles.video_responsive}`}>
                        <iframe width="560" height="315" src="https://www.youtube.com/embed/o6IU38NBwsQ" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
                    </div>
                    <Row justify='end' className={'pt-6 ml-4'}>
                        <Col>
                            <Button type={'default'} size={'large'} onClick={closeBTn}>Close</Button>
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
