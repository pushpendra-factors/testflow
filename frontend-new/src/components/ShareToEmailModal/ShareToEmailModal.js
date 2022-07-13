import React, { useState, useEffect, useCallback } from 'react';
import {
    Row, Col, Modal, Button, Form, Input, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { PlusOutlined } from '@ant-design/icons';
import PropTypes from 'prop-types';
import AppModal from '../AppModal';

const ShareToEmailModal = ({visible, onSubmit, isLoading, setShowShareToEmailModal}) => {
    const [form] = Form.useForm();
    const [frequency, setFrequency] = useState('send_now');

    const resetModalState = useCallback(() => {
        form.resetFields();
        setFrequency('send_now');
        setShowShareToEmailModal(false);
    }, [setShowShareToEmailModal]);

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
                            <Text type={'title'} level={5} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Email this dashboard</Text>
                        </Col>
                    </Row>
                    <div className={'m-3'}>
                        <Row>
                            <Form
                            form={form}
                            name="emailShare"
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
                                    <Col span={24}>
                                        <div className={'w-full mb-3'}>
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Recipients</Text>
                                            <Row className={'mt-1'}>
                                                <Col span={23}>
                                                    <Form.Item
                                                        label={null}
                                                        name={'email'}
                                                        validateTrigger={['onChange', 'onBlur']}
                                                        rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                                                    >
                                                    <Input className={'fa-input'} placeholder={'yourmail@gmail.com'} />
                                                    </Form.Item>
                                                </Col>
                                                <Form.List
                                                name="emails"
                                                >
                                                {(fields, { add, remove }) => (
                                                    <>
                                                    {fields.map((field, index) => (
                                                    <Col span={24}>
                                                    <Form.Item
                                                        required={false}
                                                        key={field.key}
                                                    >
                                                    <Row className={'mt-2'}>
                                                        <Col span={23}>
                                                            <Form.Item
                                                                label={null}
                                                                {...field}
                                                                name={[field.name, 'email']}
                                                                validateTrigger={['onChange', 'onBlur']}
                                                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                                                            >
                                                            <Input className={'fa-input'} placeholder={'yourmail@gmail.com'} />
                                                            </Form.Item>
                                                        </Col>
                                                        {fields.length > 0 ? (
                                                            <Col span={1} >
                                                                <Button style={{backgroundColor:'white'}} className={'mt-0.5 ml-2'} onClick={() => remove(field.name)}>
                                                                <SVG
                                                                name={'Trash'}
                                                                size={20}
                                                                color='gray'
                                                                /></Button>
                                                            </Col>
                                                        ) : null}
                                                    </Row>
                                                    </Form.Item>
                                                    </Col>
                                                    ))}
                                                    <Col span={20} className={'mt-2'}>
                                                    {fields.length === 4 ? null: <Button type={'text'} icon={<PlusOutlined style={{color:'gray', fontSize:'18px'}} />} onClick={() => add()}>Add Email</Button>}
                                                    </Col>
                                                    </>
                                                    )}
                                                </Form.List>
                                            </Row>
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
                                                    <Select className={'fa-select'} defaultValue={frequency} onChange={(value) => setFrequency(value)}>
                                                        <Option value={'send_now'}>Send now</Option>
                                                        <Option value={'last_week'}>Weekly</Option>
                                                        {/* <Option value={'last_month'}>Monthly</Option>
                                                        <Option value={'last_quarter'}>Quarterly</Option> */}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                            {/* {frequency !== 'send_now'?<>
                                            <Col span={2} className={'ml-6 mr-0.5 mt-1'}>
                                                <Text type={'title'} level={7} extraClass={'m-0'}>Every</Text>
                                            </Col>
                                            {frequency === 'weekly'?
                                            <Col span={8} className={'m-0'}>
                                                <Form.Item 
                                                    label={null}
                                                    name="day"
                                                    rules={[{ required: true, message: 'Please select day' }]}
                                                    >
                                                    <Select className={'fa-select'} defaultValue={'sunday'}>
                                                        <Option value={'sunday'}>Sunday</Option>
                                                        <Option value={'monday'}>Monday</Option>
                                                    </Select>
                                                </Form.Item>
                                            </Col>:null}
                                            {frequency === 'monthly'?
                                            <Col span={8} className={'m-0'}>
                                                <Form.Item 
                                                    label={null}
                                                    name="month"
                                                    rules={[{ required: true, message: 'Please select day' }]}
                                                    >
                                                    <Select disabled={true} defaultValue={'3rd_of_every_month'} className={'fa-select'}>
                                                        <Option value={'3rd_of_every_month'}>3rd of every month</Option>
                                                    </Select>
                                                </Form.Item>
                                            </Col>:null}
                                            {frequency === 'quarterly'?
                                            <Col span={8} className={'m-0'}>
                                                <Form.Item 
                                                    label={null}
                                                    name="quarter"
                                                    rules={[{ required: true, message: 'Please select day' }]}
                                                    >
                                                    <Select disabled={true} defaultValue={'3rd_of_every_quarter'} className={'fa-select'}>
                                                        <Option value={'3rd_of_every_quarter'}>3rd of every quarter</Option>
                                                    </Select>
                                                </Form.Item>
                                            </Col>:null}
                                            {frequency === 'weekly'?<>
                                            <Col span={1} className={'ml-6 mt-1'}>
                                                <Text type={'title'} level={7} extraClass={'m-0'}>At</Text>
                                            </Col>
                                            <Col span={4} className={'m-0'}>
                                                <Form.Item 
                                                    label={null}
                                                    name="time"
                                                    rules={[{ required: true, message: 'Please select day' }]}
                                                    >
                                                    <Select className={'fa-select'} defaultValue={'9AM'}>
                                                        <Option value={'9AM'}>9 AM</Option>
                                                        <Option value={'10AM'}>10 AM</Option>
                                                        <Option value={'11AM'}>11 AM</Option>
                                                        <Option value={'12PM'}>12 PM</Option>
                                                        <Option value={'1PM'}>1 PM</Option>
                                                    </Select>
                                                </Form.Item>
                                            </Col></>:null}
                                            </>:null} */}
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
                                            <Text type={'title'} level={7} extraClass={'m-0 mb-1'}>Not getting our emails? <a href='#!'>Let us now</a></Text>
                                        </div>
                                    </Col>
                                    <Col span={24}>
                                        <Row justify='end' className={'w-full mb-1'}>
                                            <Col className={'mr-2'}>
                                                <Button type={'default'} onClick={handleCancel}>Cancel</Button>
                                            </Col>
                                            {frequency === 'send_now'?
                                            <Col className={'mr-2'}>
                                                <Button type={'primary'} htmlType='submit'>Send Email</Button>
                                            </Col>
                                            :
                                            <Col className={'mr-2'}>
                                                <Button type={'primary'} htmlType='submit'>Schedule Email</Button>
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

export default connect(mapStateToProps, null)(ShareToEmailModal);


ShareToEmailModal.propTypes = {
    visible: PropTypes.bool,
    isLoading: PropTypes.bool,
    onSubmit: PropTypes.func,
    setShowShareToEmailModal: PropTypes.func,
  };
  
  ShareToEmailModal.defaultProps = {
    visible: false,
    isLoading: false,
    onSubmit: _.noop,
    setShowShareToEmailModal: _.noop,
  };