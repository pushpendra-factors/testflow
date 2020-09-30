import React, { useState } from 'react';
import {
  Row, Col, Button, Input
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';

function Login() {
  const [dataLoading, setDataLoading] = useState(false);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const CheckLogin = () => {
    setDataLoading(true);
    setTimeout(() => {
      setDataLoading(false);
      history.push('/');
    }, 1500);
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
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Login to Continue</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Input disabled={dataLoading} className={'fa-input'} size={'large'} placeholder="Username" />
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Input disabled={dataLoading} className={'fa-input'} size={'large'} placeholder="Password" />
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Button loading={dataLoading} onClick={() => CheckLogin()} type={'primary'} size={'large'}>LOG IN</Button>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Button disabled={dataLoading} type={'text'} size={'large'} onClick={() => routeChange('/forgotpassword')}>Forgot Password</Button>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Button disabled={dataLoading} type={'link'} size={'large'}>Don’t have an account? Sign Up</Button>
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

export default Login;
