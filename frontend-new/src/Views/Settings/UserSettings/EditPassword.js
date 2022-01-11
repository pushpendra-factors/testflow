import React, { useState } from 'react';
import {
  Row, Col, Modal, Input, Form, Button, notification
} from 'antd';
import { Text } from 'factorsComponents';
import { updateAgentPassword, signout } from 'Reducers/agentActions';
import { connect } from 'react-redux';

function EditPassword(props) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [success, setSuccess] = useState(false);
  const onFinish = values => {
    props.updateAgentPassword(values).then(() => {
      props.onCancel();
      setSuccess(true);
    }).catch((err) => {
      setSuccess(false);
      console.log('change password failed-->', err);
      seterrorInfo(err.error);
    });
  };

  const onReset = () => {
    seterrorInfo(null);
    props.onCancel();
    form.resetFields();
  };
  const onChange = () => {
    seterrorInfo(null);
  };
  const onModalCancel = () => {
    onReset();
    if (success) {
      notification.success({
        message: 'Password Changed!',
        description: 'Please Login again to continue.',
        duration: 10
      });
      setTimeout(() => {
        props.signout();
      }, 3000);
    }
  };
  return (
    <>

      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={onReset}
        afterClose={onModalCancel}
        className={'fa-modal--regular fa-modal--slideInDown'}
        okText={'Update Password'}
        onOk={props.onOk}
        confirmLoading={props.confirmLoading}
        centered={true}
        footer={null}
        transitionName=""
        maskTransitionName=""
      >
        <div className={'p-4'}>
          <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          >
            <Row>
              <Col span={24}>
                <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Change Password</Text>
              </Col>
            </Row>
            <Row className={'mt-6'}>
              <Col span={24}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Old Password</Text>
                <Form.Item
                    name="current_password"
                    rules={[
                      {
                        required: true,
                        message: 'Please input your old password.'
                      },
                    ]}

                    >
                      <Input.Password disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Old Password" />
                      </Form.Item>
              </Col>
            </Row>
            <Row className={'mt-6'}>
              <Col span={24}>
                <Text type={'title'} level={7} extraClass={'m-0'}>New Password</Text>
                <Form.Item
                    name="new_password"
                    rules={[
                      {
                        required: true,
                        message: 'Please input your new password.'
                      },
                      ({ getFieldValue }) => ({
                        validator(rule, value) { 
                          if (!value || value.match(/^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*?[#?!@$%^&*-]).{8,}$/)) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('Your password must contain at least eight characters, at least one number and both lower and uppercase letters and special characters.'));
                        }
                      }) 
                    ]}

                    >
                      <Input.Password disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="New Password" />
                      </Form.Item>
              </Col>
            </Row>
            <Row className={'mt-6'}>
              <Col span={24}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Confirm Password</Text>
                <Form.Item
                    name="confirm_password"
                    dependencies={['new_password']}
                    rules={[
                      {
                        required: true,
                        message: 'Please confirm your new password.'
                      },
                      ({ getFieldValue }) => ({
                        validator(rule, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('The new password that you entered do not match!'));
                        }
                      })
                    ]}

                    >
                      <Input.Password disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Confirm Password" />
                      </Form.Item>
              </Col>
            </Row>
            {errorInfo && <Col span={24}>
                <div className={'flex flex-col justify-center items-center mt-1'} >
                    <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                </div>
            </Col>
            }
            <Row className={'mt-6'}>
              <Col span={24}>
                <div className={'flex justify-end'}>
                  <Button size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>
                  <Button type="primary" size={'large'} htmlType="submit"> Update Password </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </div>

      </Modal>

    </>

  );
}

export default connect(null, { updateAgentPassword, signout })(EditPassword);
