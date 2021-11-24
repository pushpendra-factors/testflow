import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { userdata } from '../../reducers/agentActions';

function UserData(props) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const history = useHistory();

  const UserDataFn = () => {
    setDataLoading(true);
    form.validateFields().then((value) => {
      setDataLoading(true);
      setTimeout(() => {
        props.userdata(value.form_phone, value.form_website,form_estimateusers,form_teamsize)
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
                                <div className={'flex justify-center items-center'} >
                                    <SVG name={'BrandFull'} width={250} height={90} color="white"/>
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
                        onFinish={UserDataFn}
                        onChange={onChange}
                        >
                        <Row>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mb-5'} >
                                    <Text type={'title'} level={4} extraClass={'m-0'} weight={'bold'}>You are almost there</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-10 w-full'} >
                                        <Form.Item label={null}
                                            name="form_phone"
                                            rules={[{ required: true, message: 'Please enter phone number' }]}
                                            >
                                            <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'} placeholder="Phone Number" />
                                        </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="form_website"
                                            rules={[{ required: true, message: 'Please enter company website' }]}
                                            >
                                            <Input
                                                className={'fa-input w-full'}
                                                size={'large'}
                                                placeholder="Company Website"
                                                disabled={dataLoading}
                                            />
                                            </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="form_estimateusers"
                                            rules={[{ required: true, message: 'Please select estimated users' }]}
                                            >
                                            <Select
                                                placeholder="Estimated monthly tracked users"
                                                // onChange={this.onGenderChange}
                                                allowClear
                                            >
                                                <Option value="10">Less then 10</Option>
                                                <Option value="50">11-50</Option>
                                                <Option value="100">51-100</Option>
                                                <Option value="500">101-500</Option>
                                            </Select>
                                            </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="form_teamsize"
                                            rules={[{ required: true, message: 'Please select team size' }]}
                                            >
                                            <Select
                                                placeholder="Team Size"
                                                // onChange={this.onGenderChange}
                                                allowClear
                                            >
                                                <Option value="10">Less then 10</Option>
                                                <Option value="50">11-50</Option>
                                                <Option value="100">51-100</Option>
                                                <Option value="500">101-500</Option>
                                            </Select>
                                            </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-20'} >
                                    <Form.Item className={'m-0'} loading={dataLoading}>
                                        <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Done</Button>
                                    </Form.Item>
                                </div>
                            </Col>
                            {errorInfo && <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-1'} >
                                    <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                                </div>
                            </Col>
                            }
                            {/* <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <a disabled={dataLoading} type={'text'} size={'large'} onClick={() => routeChange('/forgotpassword')}>Forgot Password</a>
                                </div>
                            </Col> */}
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

export default connect(null, { userdata })(UserData);
