import React from 'react';
import {
  Row, Col, Button, Input
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';

function ForgotPassword() {
  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  return (
    <>
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={12} >
                    <div className={'flex flex-col justify-center items-center login-container'}>
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
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Forget password?</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mt-10'} >
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 desc-text'}>Please enter the email address. We will send an email with a reset link and instructions to reset your password</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Input className={'fa-input fa-input-50'} size={'large'} placeholder="Enter your email" />
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Button type={'primary'} className={'fa-button-50'} size={'large'} onClick={() => routeChange('/resetpassword')}>Send Reset Link</Button>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <a type={'text'} size={'large'} onClick={() => routeChange('/login')}>Go back to login</a>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Text type={'paragraph'} mini color={'grey'}>Don’t have an account? <a>Sign Up</a></Text>
                                </div>
                            </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
      </div>

    </>

  );
}

export default ForgotPassword;
