import React, { useState, useEffect, useCallback } from 'react';
import {
    Row, Col, Modal, Button, Form, Input, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import PropTypes from 'prop-types';
import AppModal from '../AppModal';
const { Option } = Select;

const ShareToSlackModal = ({visible, onSubmit, isLoading, channelOpts, setShowShareToSlackModal}) => {
    const [form] = Form.useForm();
    const [frequency, setFrequency] = useState('send_now');


    const resetModalState = useCallback(() => {
        form.resetFields();
        setFrequency('send_now');
        setShowShareToSlackModal(false);
    }, []);

    const handleCancel = useCallback(() => {
        if (!isLoading) {
          resetModalState();
        }
      }, [resetModalState, isLoading]);

    const handleSubmit = (data) => {
        onSubmit({
          data,
          frequency,
          onSuccess: () => {
            resetModalState();
          },
        });
    };


    const renderChannel = () => {
        return (
            <Select
                name="channel"
                placeholder="Select a channel"
                options={channelOpts}
                required
                showSearch
                filterOption={(input, option) => option.label.toLowerCase().includes(input.toLowerCase())}
            />
        );
    }

    return (
        <>
            <AppModal
                title={null}
                visible={visible}
                footer={null}
                centered={true}
                mask={true}
                maskClosable={false}
                maskStyle={{backgroundColor: 'rgb(0 0 0 / 70%)'}}
                closable={true}
                isLoading={isLoading}
                onCancel={handleCancel}
                className={`fa-modal--regular`}
                width={'640px'}
            >
                <div className={'m-0 mb-2'}>
                    <Row className={'m-0'}>
                        <Col>
                            <Text type={'title'} level={5} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Share this to Slack</Text>
                        </Col>
                    </Row>
                    <div className={'m-3'}>
                        <Row>
                            <Form
                            form={form}
                            name="slackShare"
                            validateTrigger
                            initialValues={{ remember: false }}
                            onFinish={handleSubmit}
                            // onChange={onChange}
                            >
                                <Row>
                                    <Col span={23}>
                                        <div className={'w-full mb-3'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Subject</Text>
                                            <Form.Item 
                                                label={null}
                                                name="subject"
                                                rules={[{ required: true, message: 'Please enter subject' }]}
                                                >
                                                <Input className={'fa-input w-full'}
                                                placeholder="Enter subject"
                                                />
                                            </Form.Item>
                                        </div>
                                    </Col>
                                    <Col span={23}>
                                        <div className={'w-full mb-3'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Select Channel</Text>
                                            <Form.Item 
                                                label={null}
                                                name="channel"
                                                rules={[{ required: true, message: 'Please select channel' }]}
                                                >
                                                {renderChannel()}
                                            </Form.Item>
                                        </div>
                                    </Col>
                                    <Col span={23}>
                                        <div className={'w-full mb-3'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Message</Text>
                                            <Form.Item 
                                                label={null}
                                                name="message"
                                                rules={[{ required: true, message: 'Please enter message' }]}
                                                >
                                                <Input className={'fa-input w-full'}
                                                placeholder="Your message"
                                                />
                                            </Form.Item>
                                        </div>
                                    </Col>
                                    <div className={'flex flex-col border-top--thin mb-4 mt-2 w-full'} />
                                    <Col span={24}>
                                        <Text type={'title'} level={7} extraClass={'m-0 mb-2'}>Frequency</Text>
                                    </Col>
                                    <Col span={23}>
                                        <Row className={'mb-3'}>
                                            <Col span={5}>
                                                <Form.Item 
                                                    label={null}
                                                    name="date_range"
                                                    >
                                                    <Select className={'fa-select'} defaultValue={frequency} value={frequency} onChange={(value) => setFrequency(value)}>
                                                        <Option value={'send_now'}>Send now</Option>
                                                        <Option value={'last_week'}>Weekly</Option>
                                                        {/* <Option value={'last_month'}>Monthly</Option>
                                                        <Option value={'last_quarter'}>Quarterly</Option> */}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                        </Row>
                                    </Col>
                                    <Col span={23}>
                                        <div className={'w-full'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Data is captured up to 2 hours before this email is sent. This is to make sure this email is delivered as on-time as possible</Text>
                                        </div>
                                    </Col>
                                    <div className={'flex flex-col border-top--thin my-4 w-full'} />
                                    <Col span={23}>
                                        <div className={'w-full mb-3'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Not getting our messages? <a href='#!'>Let us know</a></Text>
                                        </div>
                                    </Col>
                                    <Col span={24}>
                                        <Row justify='end' className={'w-full mb-1'}>
                                            <Col className={'mr-2'}>
                                                <Button type={'default'} onClick={handleCancel}>Cancel</Button>
                                            </Col>
                                            {frequency === 'send_now'?
                                            <Col className={'mr-2'}>
                                                <Button type={'primary'} htmlType='submit'>Send to Slack</Button>
                                            </Col>
                                            :
                                            <Col className={'mr-2'}>
                                                <Button type={'primary'} htmlType='submit'>Schedule</Button>
                                            </Col>
                                            }
                                        </Row>
                                    </Col>
                                </Row>
                            </Form>
                        </Row>
                    </div>
                </div>
            </AppModal> 
        </>
    );
};

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentAgent: state.agent.agent_details,
});

export default connect(mapStateToProps, null)(ShareToSlackModal);


ShareToSlackModal.propTypes = {
    visible: PropTypes.bool,
    isLoading: PropTypes.bool,
    onSubmit: PropTypes.func,
    setShowShareToSlackModal: PropTypes.func,
  };
  
  ShareToSlackModal.defaultProps = {
    visible: false,
    isLoading: false,
    onSubmit: _.noop,
    setShowShareToSlackModal: _.noop,
  };