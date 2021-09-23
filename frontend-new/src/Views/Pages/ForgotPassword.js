import React, { useState } from 'react';
import {
  Row, Col, Button, Input, Form
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { connect } from 'react-redux';
import { forgotPassword } from 'Reducers/agentActions';

function ForgotPassword({ forgotPassword }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [userEmail, setUserEmail] = useState(null);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const SendResetLink = () => {
    setDataLoading(true);
    form.validateFields().then((value) => {
      setDataLoading(true);
      setTimeout(() => {
        forgotPassword(value.email)
          .then(() => {
            setUserEmail(value.email);
            // history.push('/');
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
                        name="login"
                        validateTrigger
                        initialValues={{ remember: false }}
                        onFinish={SendResetLink}
                        onChange={onChange}
                        >
                            <Row>
                                <Col span={24} >
                                    <div className={'flex justify-center items-center mb-5'} >
                                        <SVG name={'BrandFull'} size={40} color="white"/>
                                    </div>
                                </Col>
                            </Row>
                            <Row>
                                <Col span={24}>
                                    <div className={'flex justify-center items-center mt-10'} >
                                        <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Forgot password?</Text>
                                    </div>
                                </Col>
                                {!userEmail && <>
                                <Col span={24}>
                                    <div className={'flex justify-center items-center mt-10'} >
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 desc-text'}>Please enter the email address. We will send an email with a reset link and instructions to reset your password</Text>
                                    </div>
                                </Col>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Form.Item label={null}
                                            name="email"
                                            rules={[{ required: true, type: 'email', message: 'Please enter your email' }]}
                                            >
                                            <Input className={'fa-input w-full'} loading={dataLoading} size={'large'} placeholder="Enter your email" />
                                        </Form.Item>

                                    </div>
                                </Col>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5'} >
                                        {/* <Button type={'primary'} className={'fa-button-50'} size={'large'} onClick={() => routeChange('/resetpassword')}>Send Reset Link</Button> */}
                                        <Form.Item className={'m-0'} loading={dataLoading}>
                                            <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Send Reset Link</Button>
                                        </Form.Item>
                                    </div>
                                </Col>
                                </>
                                }
                                {userEmail && <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center my-6'} >
                                        <Text type={'title'} size={'6'} color={'grey'} align={'center'} extraClass={'m-0 desc-text'}>{`An email has been sent to ${userEmail}. Please follow the link in the email to reset your password.`}</Text>
                                    </div>
                                </Col>
                                }
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
                                        {/* <Text type={'paragraph'} mini color={'grey'}>Don’t have an account? <a onClick={() => routeChange('/signup')}>Sign Up</a></Text> */}
                                        <Text type={'paragraph'} mini color={'grey'}>Want to try out Factors.AI? <a href={'https://www.factors.ai/schedule-a-demo'} target="_blank">Request A Demo</a></Text>
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

export default connect(null, { forgotPassword })(ForgotPassword);
