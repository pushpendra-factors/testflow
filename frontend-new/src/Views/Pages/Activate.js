import React, { useEffect, useState } from 'react';
import {
  Row, Col, Button, Input, Form, message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import queryString from 'query-string';
import { activate } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import MoreAuthOptions from './MoreAuthOptions';
import { SSO_ACTIVATE_URL } from '../../utils/sso';

function Activate(props) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [dataLoading, setDataLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const checkError = () => {
    const url = new URL(window.location.href);
    const error = url.searchParams.get('error');
    if(error) {
        let str = error.replace("_", " ");
        let finalmsg = str.toLocaleLowerCase();
        message.error(finalmsg);
    }
  }

  useEffect(() => {
      checkError();
  },[]);

  const ResetPassword = () => {
    setDataLoading(true);
    const tokenFromUrl = queryString.parse(props.location.search)?.token;

    form.validateFields().then((value) => {
      setDataLoading(true);
      setTimeout(() => {
        props.activate(value.password, tokenFromUrl)
          .then(() => {
            setDataLoading(false);
            history.push('/');
            message.success('Account activated!');
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
            <Row justify={'center'} className={`${styles.start}`}>
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
                                  <SVG name={'BrandFull'} size={40} color="white"/>
                                </div>
                            </Col>
                        </Row>
                        <Row>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mt-10'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Activate your account </Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                <Form.Item
                                name="password"
                                rules={[
                                  {
                                    required: true,
                                    message: 'Please enter your password.'
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
                                <Input.Password disabled={dataLoading} size="large" className={'fa-input w-full'} placeholder="Enter Password" />
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
                                        message: 'Please confirm your password.'
                                      },
                                      ({ getFieldValue }) => ({
                                        validator(rule, value) {
                                          if (!value || getFieldValue('password') === value) {
                                            return Promise.resolve();
                                          }
                                          return Promise.reject(new Error('The password that you entered do not match!'));
                                        }
                                      })
                                    ]}

                                    >
                                    <Input.Password disabled={dataLoading} size="large" className={'fa-input w-full'} placeholder="Confirm Password" />
                                    </Form.Item>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Form.Item className={'m-0'} loading={dataLoading}>
                                            <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Activate</Button>
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
                              <div className={'flex justify-center items-center mt-6'} >
                                <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'} color={'grey'}>OR</Text>
                              </div>
                            </Col>

                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                  <Form.Item className={'m-0'} loading={dataLoading}>
                                    <a href={SSO_ACTIVATE_URL}><Button loading={dataLoading} type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'}><SVG name={'Google'} size={24} />Continue with Google</Button></a>
                                  </Form.Item>
                                </div>
                            </Col>

                            {/* <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                  <Form.Item className={'m-0'} loading={dataLoading}>
                                    <Button loading={dataLoading} type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'S_Key'} size={24} color={'#8692A3'} /> More SSO Options</Button>
                                  </Form.Item>
                                </div>
                            </Col> */}
                            {/* <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <a type={'text'} size={'large'} onClick={() => routeChange('/login')}>Go back to login</a>
                                </div>
                            </Col> */}
                            {/* <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Text type={'paragraph'} mini color={'grey'}>Don’t have an account? <a onClick={() => routeChange('/signup')}>Sign Up</a></Text>
                                    <Text type={'paragraph'} mini color={'grey'}>Want to try out Factors.AI? <a href={'https://www.factors.ai/schedule-a-demo'} target="_blank">Request A Demo</a></Text>
                                </div>
                            </Col> */}
                        </Row>
                        </Form>
                    </div>
                </Col>
            </Row>
            <div className={`${styles.hide}`}>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
            </div>
      </div>

      {/* <MoreAuthOptions showModal={showModal} setShowModal={setShowModal}/> */}

    </>

  );
}

export default connect(null, { activate })(Activate);
