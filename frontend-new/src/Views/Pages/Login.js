import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { login } from '../../reducers/agentActions';
import { EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import factorsai from 'factorsai';

function Login(props) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const CheckLogin = () => {
    setDataLoading(true);
    form.validateFields().then((value) => {
      setDataLoading(true);

      //Factors LOGIN tracking
      factorsai.track('LOGIN',{'username':value?.form_username});

      setTimeout(() => {
        props.login(value.form_username, value.form_password)
          .then(() => {
            setDataLoading(false);
            history.push('/');
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
                        <Row>
                            <Col span={24} >
                                <div className={'flex justify-center items-center mb-5'} >
                                    <SVG name={'BrandFull'} size={40} color="white"/>
                                </div>
                            </Col>
                        </Row>
                        <Row>
                            <Col span={24}>
                        <Form
                        form={form}
                        name="login"
                        validateTrigger
                        initialValues={{ remember: false }}
                        onFinish={CheckLogin}
                        onChange={onChange}
                        >
                        <Row>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mt-10'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Login to Continue</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-10 w-full'} >
                                        <Form.Item label={null}
                                            name="form_username"
                                            rules={[{ required: true, type: 'email', message: 'Please enter email' }]}
                                            >
                                            <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'} placeholder="Email" />
                                        </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="form_password"
                                            rules={[{ required: true, message: 'Please enter password' }]}
                                             >
                                            <Input.Password
                                                className={'fa-input w-full'}
                                                size={'large'}
                                                placeholder="Password"
                                                disabled={dataLoading}
                                                iconRender={visible => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                                            />
                                            </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Form.Item className={'m-0'} loading={dataLoading}>
                                        <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>LOG IN</Button>
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
                                    <a disabled={dataLoading} type={'text'} size={'large'} onClick={() => routeChange('/forgotpassword')}>Forgot Password</a>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                {/* <Text type={'paragraph'} mini color={'grey'}>Don’t have an account? <a disabled={dataLoading} onClick={() => routeChange('/signup')}> Sign Up</a></Text> */}
                                <Text type={'paragraph'} mini color={'grey'}>Want to try out Factors.AI? <a href={'https://www.factors.ai/schedule-a-demo'} target="_blank">Request A Demo</a></Text>
                                </div>
                            </Col>
                        </Row>
                        </Form>
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

export default connect(null, { login })(Login);
