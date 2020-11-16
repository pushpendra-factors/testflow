import React, { useState } from 'react';
import {
  Row, Col, Button, Input, Form, message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import queryString from 'query-string';
import { setPassword } from 'Reducers/agentActions';
import { connect } from 'react-redux';

function ResetPassword(props) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [dataLoading, setDataLoading] = useState(false);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const ResetPassword = () => {
    setDataLoading(true);
    const tokenFromUrl = queryString.parse(props.location.search)?.token;

    form.validateFields().then((value) => {
      setDataLoading(true);
      setTimeout(() => {
        props.setPassword(value.password, tokenFromUrl)
          .then(() => {
            setDataLoading(false);
            history.push('/');
            message.success('Password Changed!');
          }).catch((err) => {
            setDataLoading(false);
            form.resetFields();
            seterrorInfo(err);
          });
      }, 200);
    }).catch((info) => {
      setDataLoading(false);
      form.resetFields();
      seterrorInfo(info);
    });
  };

  const onChange = () => {
    seterrorInfo(null);
  };

  return (
    <>
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={12} >
                    <div className={'flex flex-col justify-center items-center login-container'}>
                    <Form
                        form={form}
                        onFinish={ResetPassword}
                        className={'w-full'}
                        onChange={onChange}
                        >
                        <Row>
                            <Col span={24} >
                                <div className={'flex justify-center items-center mb-5'} >
                                    <SVG name={'brand'} size={40} color="white"/><Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>FACTORS.AI</Text>
                                </div>
                            </Col>
                        </Row>
                        <Row>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mt-10'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Reset your Password</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                <Form.Item
                                name="password"
                                rules={[
                                  {
                                    required: true,
                                    message: 'Please input your old password.'
                                  }
                                ]}

                                >
                                <Input.Password disabled={dataLoading} size="large" className={'fa-input w-full'} placeholder="Enter New Password" />
                                </Form.Item>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                <Form.Item
                                    name="confirm_password"
                                    dependencies={['password']}
                                    rules={[
                                      {
                                        required: true,
                                        message: 'Please confirm your new password.'
                                      },
                                      ({ getFieldValue }) => ({
                                        validator(rule, value) {
                                          if (!value || getFieldValue('password') === value) {
                                            return Promise.resolve();
                                          }
                                          return Promise.reject(new Error('The new password that you entered do not match!'));
                                        }
                                      })
                                    ]}

                                    >
                                    <Input.Password disabled={dataLoading} size="large" className={'fa-input w-full'} placeholder="Confirm New Password" />
                                    </Form.Item>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Form.Item className={'m-0'} loading={dataLoading}>
                                            <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Reset Password</Button>
                                        </Form.Item>
                                </div>
                            </Col>
                            {errorInfo && <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-1'} >
                                    <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                                </div>
                            </Col>
                            }
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <a type={'text'} size={'large'} onClick={() => routeChange('/login')}>Go back to login</a>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Text type={'paragraph'} mini color={'grey'}>Don’t have an account? <a onClick={() => routeChange('/signup')}>Sign Up</a></Text>
                                </div>
                            </Col>
                        </Row>
                        </Form>
                    </div>
                </Col>
            </Row>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
      </div>

    </>

  );
}

export default connect(null, { setPassword })(ResetPassword);
