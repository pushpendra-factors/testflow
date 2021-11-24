import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, message, Checkbox
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { signup } from 'Reducers/agentActions';

function SignUp({ signup }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [formData, setformData] = useState(null);

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const popBookDemo = () => {
    // eslint-disable-next-line
      if(Calendly){ Calendly.initPopupWidget({ url: 'https://calendly.com/factorsai/demo' }); }
  };

  const resendEmail = () => {
    console.log('resendEmail');
    signup(formData).then(() => {
      message.success('Email resent!');
    }).catch((err) => {
      console.log('Signup-resent email err-->', err);
      message.success('Email resent!');
    });
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
        history.push('/userdata');
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
                        { !formData &&

                        <Row>
                            <Col span={24}>

                        <Form
                        form={form}
                        name="login"
                        validateTrigger
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

                    }
                    {formData &&
                            <Row>
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center'}>

                                <div className={'mb-4'}>
                                    <svg width="98" height="92" viewBox="0 0 98 92" fill="none" xmlns="http://www.w3.org/2000/svg">
                                    <path d="M67.4139 88.9146C67.5194 88.9879 67.5664 89.0396 67.5867 89.0673C67.5664 89.095 67.5194 89.1467 67.4139 89.22C67.2089 89.3625 66.8743 89.5201 66.3955 89.6812C65.4444 90.0011 64.0412 90.2971 62.2752 90.5484C58.7506 91.0499 53.8649 91.362 48.457 91.362C43.0492 91.362 38.1635 91.0499 34.639 90.5484C32.8729 90.2971 31.4698 90.0011 30.5187 89.6812C30.0399 89.5201 29.7053 89.3625 29.5003 89.22C29.3948 89.1466 29.3478 89.095 29.3275 89.0673C29.3478 89.0396 29.3948 88.9879 29.5003 88.9146C29.7053 88.772 30.0399 88.6145 30.5187 88.4534C31.4698 88.1335 32.8729 87.8375 34.639 87.5863C38.1635 87.0848 43.0492 86.7728 48.457 86.7728C53.8649 86.7728 58.7506 87.0848 62.2752 87.5863C64.0413 87.8375 65.4444 88.1335 66.3955 88.4534C66.8743 88.6145 67.2089 88.772 67.4139 88.9146Z" fill="#EEEDF8" stroke="#EEEDF8"/>
                                    <path d="M88.6124 23.7197C88.6124 25.4406 87.2172 26.8357 85.4963 26.8357C83.7753 26.8357 82.3801 25.4406 82.3801 23.7197C82.3801 21.9987 83.7753 20.6036 85.4963 20.6036C87.2172 20.6036 88.6124 21.9987 88.6124 23.7197Z" fill="#EEEDF8" stroke="#EEEDF8" strokeWidth="17.102" strokeMiterlimit="10"/>
                                    <path d="M85.9966 54.7615H87.1661V62.1871H85.9966V54.7615Z" fill="#EEEDF8" stroke="#EEEDF8"/>
                                    <path d="M90.2941 59.0592H82.8685V57.8894H90.2941V59.0592Z" fill="#EEEDF8" stroke="#EEEDF8"/>
                                    <path d="M12.8534 71.5271C12.8534 72.5913 11.9908 73.4538 10.9267 73.4538C9.86266 73.4538 9 72.5913 9 71.5271C9 70.463 9.86266 69.6005 10.9267 69.6005C11.9908 69.6005 12.8534 70.463 12.8534 71.5271Z" fill="#EEEDF8" stroke="#EEEDF8" strokeWidth="17.102" strokeMiterlimit="10"/>
                                    <path d="M20.5769 10.0549C20.5769 10.0549 22.6662 7.68143 25.0396 10.0549C27.413 12.4283 29.1779 10.0549 29.1779 10.0549" fill="#EEEDF8"/>
                                    <path d="M20.5769 10.0549C20.5769 10.0549 22.6662 7.68143 25.0396 10.0549C27.413 12.4283 29.1779 10.0549 29.1779 10.0549" stroke="#EEEDF8" strokeWidth="17.102" strokeMiterlimit="10"/>
                                    <path d="M50.4284 69.0047C52.9432 68.4074 56.0043 67.4661 59.0868 65.9942C65.0125 63.1648 71.0513 58.3525 73.3956 50.2483V66.2667L70.6574 69.0047H50.4284Z" fill="#EEEDF8" stroke="#EEEDF8"/>
                                    <path d="M70.3526 29.9152H26.5618C23.7755 29.9152 21.5089 32.1817 21.5089 34.968V65.2848C21.5089 68.071 23.7755 70.3376 26.5618 70.3376H70.3526C73.139 70.3376 75.4054 68.071 75.4054 65.2848V34.968C75.4054 32.1817 73.139 29.9152 70.3526 29.9152ZM70.3526 33.2837C70.5815 33.2837 70.7991 33.3312 70.9981 33.4141L48.4571 52.9504L25.9162 33.4141C26.1151 33.3314 26.3328 33.2837 26.5617 33.2837H70.3526ZM70.3526 66.969H26.5618C25.6325 66.969 24.8775 66.2141 24.8775 65.2846V36.973L47.3534 56.4522C47.6709 56.7268 48.064 56.8634 48.4571 56.8634C48.8503 56.8634 49.2434 56.7269 49.5608 56.4522L72.0369 36.973V65.2848C72.0368 66.2141 71.2819 66.969 70.3526 66.969Z" fill="#5949BC"/>
                                    </svg>
                                </div>

                                    <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Congrats!! Time to see Factors in Action</Text>
                                    <Text type={'title'} level={7} color={'grey'} align="center" extraClass={'m-0 mt-4'}>We’ve sent a confirmation email to</Text>
                                    <Text type={'title'} level={7} align="center" extraClass={'m-0'}>{formData.email}</Text>
                                </div>
                            </Col>

                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-20'}>
                                    <Text type={'title'} level={7} color={'grey'} align="center" extraClass={'m-0'}>Our team would be happy to walk you through the product and answer any questions </Text>
                                    <Button size={'large'} className={'w-full mt-4'} style={{ maxWidth: '280px' }} onClick={() => popBookDemo()}>Schedule a demo</Button>
                                    <Text type={'title'} level={7} align="center" extraClass={'m-0 mt-6'}>Didn’t get an email? <a onClick={() => resendEmail()} >Click to resend</a></Text>
                                </div>
                            </Col>
                            </Row>
                    }

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

    </>

  );
}

export default connect(null, { signup })(SignUp);
