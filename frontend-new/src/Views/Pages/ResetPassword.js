import React from 'react';
import {
  Row, Col, Button, Input
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';

function ResetPassword() {
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
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Reset your Password</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Input className={'fa-input'} size={'large'} placeholder="Enter Your New Password" />
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Input className={'fa-input'} size={'large'} placeholder="Confirm New Password" />
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Button type={'primary'} size={'large'} onClick={() => routeChange('/login')}>Reset Password</Button>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Button type={'text'} size={'large'} onClick={() => routeChange('/login')}>Go back to login</Button>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Button type={'link'} size={'large'} onClick={() => routeChange('/signup')} >Don’t have an account? Sign Up</Button>
                                </div>
                            </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
      </div>

    </>

  );
}

export default ResetPassword;
