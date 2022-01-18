import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Checkbox
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { signup } from 'Reducers/agentActions';
import UserData from './UserData';

function SignUp({ signup }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [formData, setformData] = useState(null);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const SignUpFn = () => {
    setDataLoading(true);
    form.validateFields().then((values) => {
        setDataLoading(true);
        const filteredValues = Object.fromEntries(
        Object.entries(values).filter(([key, value]) => key !== 'terms_and_conditions') );
        
        signup(filteredValues).then(() => {
            setDataLoading(false);
            setformData(filteredValues);
        }).catch((err) => {
            setDataLoading(false);
            form.resetFields();
            seterrorInfo(err);
        });
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
    { !formData &&
      <div className={'fa-container h-screen w-full'}>

            <Row justify={'space-between'} className={'py-4 m-0 '}>
              <Col>
                <div className={'flex items-center'}>
                    <SVG name={'BrandFull'} size={40} color="white"/>
                </div>
              </Col>
              <Col>
                <Button size={'large'} onClick={() => routeChange('/login')} >Sign In</Button>
              </Col>
            </Row>

            {/* //parent container starts here */}
            <Row className={' signup-container w-full'}>

                    <div className={'flex items-center '}>
                    {/* //left side content starts here */}
                <Col span={12} >
                    <Row align="center">
                            <Col span={14}>
                        

                        <Row>
                            <Col span={24}>

                        <Form
                        form={form}
                        name="login"
                        // validateTrigger
                        initialValues={{ remember: false }}
                        onFinish={SignUpFn}
                        onChange={onChange}
                        >
                            <Row>
                                <Col span={24}>
                                    <div className={'flex mb-2'} >
                                        <Text type={'title'} level={5} extraClass={'m-0'} weight={'bold'}>Create your Factors account</Text>
                                    </div>
                                </Col>
                            </Row>

                            <Row gutter={[24, 0]}>
                                <Col span={12}>
                                        <div className={'flex flex-col mt-5'} >
                                        <Text type={'title'} level={7} extraClass={'m-0'}>First Name</Text>
                                            <Form.Item label={null}
                                                name="first_name"
                                                rules={[{ required: true, message: 'Please enter first name' }]}
                                                className={'w-full'}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                // placeholder="First Name"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col>
                                <Col span={12}>
                                        <div className={'flex flex-col mt-5'} >
                                        <Text type={'title'} level={7} extraClass={'m-0'}>Last Name</Text>
                                            <Form.Item label={null}
                                                name="last_name"
                                                rules={[{ required: true, message: 'Please enter last name' }]}
                                                className={'w-full'}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                // placeholder="Last Name"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col>
                            </Row>

                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5 w-full'} >
                                            <Text type={'title'} level={7} extraClass={'m-0'}>Work Email</Text>
                                            <Form.Item label={null}
                                                name="email"
                                                rules={[{ required: true, type: 'email', message: 'Please enter work email' }]}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                // placeholder="Work Email"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col>
                            </Row>

                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5 w-full'} > 
                                            <Form.Item label={null}
                                                name="subscribe_newsletter" valuePropName={'checked'}                                 
                                                >
                                                <div className='flex items-center'>
                                                    <Checkbox disabled={dataLoading} ></Checkbox>
                                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 ml-4'} >Please keep me up to date on Factors, including news, new products, and services.</Text>
                                                </div>
                                            </Form.Item>
                                        </div>
                                </Col>
                            </Row>
                            
                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5 w-full'} >
                                            <Form.Item label={null} 
                                                name='terms_and_conditions' valuePropName={'checked'}
                                                rules={[{ required: true, transform: value => (value || undefined), type: 'boolean', message: 'Please agree to the terms and conditions' }]}
                                                >
                                                <div className='flex items-center' >
                                                    <Checkbox disabled={dataLoading} ></Checkbox>
                                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 ml-4'} >By signing up, I agree to the <a href='https://www.factors.ai/terms-of-use' target='_blank'>terms of service</a> and <a href='https://www.factors.ai/privacy-policy' target='_blank'>privacy policy</a> of factors.ai</Text>
                                                </div>
                                            </Form.Item>
                                        </div>
                                </Col>
                            </Row>

                            <Row>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5'} >
                                        <Form.Item className={'m-0 w-full'} loading={dataLoading}>
                                            <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Get Started</Button>
                                        </Form.Item>
                                    </div>
                                </Col>
                            </Row>

                            <Row>
                                {errorInfo && <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-1'} >
                                        <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                                    </div>
                                </Col>
                                }
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-10'} >
                                    <Text type={'paragraph'} mini color={'grey'}>Already have an account?<a disabled={dataLoading} onClick={() => routeChange('/login')}> Sign In</a></Text>
                                    </div>
                                </Col>
                            </Row>

                        </Form>
                        </Col>

                        </Row>

                        </Col>
                        </Row>
                </Col>
                        {/* //left side content ends here */}

                        {/* //right side content starts here */}

                <Col span={12}>
                    <Row align="center">
                            <Col span={14}>
                                <Row>
                                        <Col span={24}>
                                            <img src="assets/images/illustration.png" className={'mb-10'} style={{ marginLeft: '-80px' }}/>
                                        </Col>
                                </Row>
                                <Row>
                                        <Col span={24}>
                                        <Text type={'title'} level={3} extraClass={'m-0'} weight={'bold'}>Marketing Decisioning made Radically Smarter</Text>
                                        <Text type={'title'} color={'grey'} level={7} extraClass={'m-0'} >An end-to-end marketing analytics platform that integrates across data silos to deliver focussed AI-fueled actionable insights.</Text>
                                        </Col>
                                </Row>
                            </Col>
                    </Row>
                </Col>
                        {/* //right side content ends here */}
            </div>
            </Row>
                        {/* //parent container ends here */}
      </div>
        }
        {formData &&
            <UserData data = {formData} />
        }
    </>

  );
}

export default connect(null, { signup })(SignUp);
