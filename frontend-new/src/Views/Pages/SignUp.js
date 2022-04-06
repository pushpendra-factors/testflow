import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Checkbox, Modal, message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { signup } from 'Reducers/agentActions';
import factorsai from 'factorsai';
import Congrats from './Congrats';
import { createHubspotContact, getHubspotContact } from '../../reducers/global';
import { getOwner } from '../../utils/hubspot';
import { URL1, URL2 } from '../../utils/mailmodo';
import MoreAuthOptions from './MoreAuthOptions';
import { SSO_SIGNUP_URL } from '../../utils/sso';

function SignUp({ signup, createHubspotContact, getHubspotContact }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [formData, setformData] = useState(null);
  const [ownerID, setownerID] = useState();
  const [showModal, setShowModal] = useState(false);
  const [Showform, setShowform] = useState(false);

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

  const startMailModo = (email) => {
    let data = {
            "email": email,
            "data": {} 
        }
    let params = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            "mmapikey": "TJ5JF61-44NMRN5-GAEA2WH-8Z99P4H"
        },
        body: JSON.stringify(data)
    }

    fetch(URL1, params)
    .then((response) => response.json())
    .then((response) => {
      console.log(response);
    })
    .catch((err) => {
      console.log('err',err);
    });

    fetch(URL2, params)
    .then((response) => response.json())
    .then((response) => {
      console.log(response);
    })
    .catch((err) => {
      console.log('err',err);
    });
  }

  const hubspotCall = (data) => {
        const owner = getOwner();

        getHubspotContact(data.email).then((res) => {
            setownerID(res.data.hubspot_owner_id)
        }).catch((err) => {
            console.log(err.data.error)
        });


        const jsonData = {
            "properties": [
                {
                    "property": "email",
                    "value": data.email
                },
                {
                    "property": "firstname",
                    "value": data.first_name
                },
                {
                    "property": "lastname",
                    "value": data.last_name
                },
                {
                    "property": "phone",
                    "value": data?.phone
                },
                {
                    "property": "hubspot_owner_id",
                    "value": ownerID ? ownerID: owner.value
                },
                {
                    "property": "signup_method",
                    "value": "Self-Serve Onboarding"
                }                     
            ]
        }
        
        createHubspotContact(data.email, jsonData)
        .then((response) => {
            console.log(response);
        })
        .catch((err) => {
            console.log(err);
        });
  };

  const sendSlackNotification = (user) => {
    let webhookURL = 'https://hooks.slack.com/services/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';
    let data = {
        "text": `User ${user.first_name} with email ${user.email} just signed up`,
        "username" : "Signup User Actions",
        "icon_emoji" : ":golf:"
    }
    let params = {
        method: 'POST',
        body: JSON.stringify(data)
    }

    fetch(webhookURL, params)
    .then((response) => response.json())
    .then((response) => {
        console.log(response);
    })
    .catch((err) => {
        console.log('err',err);
    });
}


  const SignUpFn = () => {
    setDataLoading(true);
    form.validateFields().then((values) => {
        setDataLoading(true);

        //Factors SIGNUP tracking
        factorsai.track('SIGNUP',{'first_name':values?.first_name,'email':values?.email});
        
        let data = {...values, 'last_name': ''};
        const filteredValues = Object.fromEntries(
        Object.entries(data).filter(([key, value]) => key !== 'terms_and_conditions') );
        
        signup(filteredValues).then(() => {
            setDataLoading(false);
            setformData(filteredValues);
            startMailModo(filteredValues.email);
            hubspotCall(filteredValues);
            sendSlackNotification(filteredValues);
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
      <div className={'fa-content-container.no-sidebar h-screen w-full h-full'}>

            {/* //parent container starts here */}

            <div className={'flex items-center w-full h-full '}>
                {/* //left side content starts here */}
                <Col xs={{ span: 0}} sm={{ span: 12}} style={{background: '#E6F7FF'}} className={'w-full h-full'}>
                    <Row align="center" className={'my-40'}>
                            <Col span={14}>
                                <Row>
                                    <Col span={24}>
                                        <Text type={'title'} level={3} extraClass={'m-0'} weight={'bold'}>Marketing decisioning made radically smarter</Text>
                                        <Text type={'title'} color={'grey'} level={7} extraClass={'m-0'} >Measure and validate every marketing initiative. Understand the entire buyer journey. Then drive pipeline and revenue like never before, all inside Factors.</Text>
                                    </Col>
                                </Row>
                                <Row>
                                    <Col span={24}>
                                        <img src="https://s3.amazonaws.com/www.factors.ai/assets/img/product/review.svg" className={'m-0 mt-4 -ml-2'}/>
                                    </Col>
                                </Row>
                                <Row>
                                    <Col span={24}>
                                    <img src="https://s3.amazonaws.com/www.factors.ai/assets/img/product/marketing-teams.svg" className={'m-0 -ml-2'}/>
                                    </Col>
                                </Row>
                            </Col>
                    </Row>
                </Col>
                
                {/* //left side content ends here */}

                {/* //right side content starts here */}

                <Col xs={{ span: 24}} sm={{ span: 12}} >
                    <Row align="center">
                        <Col span={14}>
                        

                        <Row>
                        <Col span={20}>

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
                                    <div style={{marginLeft: '5.5vw'}}>
                                        <SVG name={'BrandFull'} size={40} color="white"/>
                                    </div>
                                </Col>
                            </Row>
                            <Row>
                                <Col span={24}>
                                    <div>
                                        <Text type={'title'} level={6} align={'center'} color={'grey-2'} extraClass={'m-0 ml-1 mt-2'} weight={'bold'}>Sign Up to continue</Text>
                                    </div>
                                </Col>
                            </Row>
                            {Showform? 
                            <>
                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5'} >
                                        {/* <Text type={'title'} level={7} extraClass={'m-0'}>First Name</Text> */}
                                            <Form.Item label={null}
                                                name="first_name"
                                                rules={[{ required: true, message: 'Please enter first name' }]}
                                                className={'w-full'}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                placeholder="First Name"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col>
                                {/* <Col span={12}>
                                        <div className={'flex flex-col mt-5'} >
                                        <Text type={'title'} level={7} extraClass={'m-0'}>Last Name</Text>
                                            <Form.Item label={null}
                                                name="last_name"
                                                rules={[{ required: true, message: 'Please enter last name' }]}
                                                className={'w-full'}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                placeholder="Last Name"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col> */}
                            </Row>

                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5 w-full'} >
                                            {/* <Text type={'title'} level={7} extraClass={'m-0'}>Work Email</Text> */}
                                            <Form.Item label={null}
                                                name="email"
                                                rules={[
                                                    { 
                                                        required: true, message: 'Please enter your business email address.' 
                                                    },
                                                    ({ getFieldValue }) => ({
                                                        validator(rule, value) { 
                                                          if (!value || value.match(/^([a-z0-9!'#$%&*+\/=?^_`{|}~-]+(?:\.[a-z0-9!'#$%&*+\/=?^_`{|}~-]+)*@(?!gmail.com)(?!yahoo.com)(?!hotmail.com)(?!yahoo.co.in)(?!hey.com)(?!icloud.com)(?!me.com)(?!mac.com)(?!aol.com)(?!abc.com)(?!xyz.com)(?!pqr.com)(?!rediffmail.com)(?!live.com)(?!outlook.com)(?!msn.com)(?!ymail.com)([\w-]+\.)+[\w-]{2,})?$/)) {
                                                            return Promise.resolve();
                                                          }
                                                          return Promise.reject(new Error('Please enter your business email address.'));
                                                        }
                                                    })
                                                ]}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                placeholder="Work Email"
                                                 />
                                            </Form.Item>
                                        </div>
                                </Col>
                            </Row>

                            <Row>
                                <Col span={24}>
                                        <div className={'flex flex-col mt-5 w-full'} >
                                            {/* <Text type={'title'} level={7} extraClass={'m-0'}>Phone Number</Text> */}
                                            <Form.Item label={null}
                                                name="phone"
                                                rules={[
                                                    ({ getFieldValue }) => ({
                                                        validator(rule, value) { 
                                                          if (!value || value.match(/^[0-9\b]+$/)) {
                                                            return Promise.resolve();
                                                          }
                                                          return Promise.reject(new Error('Please enter valid phone number.'));
                                                        }
                                                    })
                                                ]}
                                                >
                                                <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'}
                                                placeholder="Phone Number (Optional)"
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
                                            <Button htmlType="submit" loading={dataLoading} type={'primary'} size={'large'} className={'w-full'}>Signup</Button>
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
                            </Row>

                            <Row>
                                <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-6'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'} color={'grey'}>OR</Text>
                                </div>
                                </Col>
                            </Row>
                            </>
                            : 
                            <Row>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5'} >
                                        <Form.Item className={'m-0 w-full'} loading={dataLoading}>
                                            <Button loading={dataLoading} type={'primary'} size={'large'} className={'w-full'} onClick={() => setShowform(true)}>Signup with Email</Button>
                                        </Form.Item>
                                    </div>
                                </Col>
                            </Row>}

                            <Row>   
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Form.Item className={'m-0 w-full'}>
                                        <a href={SSO_SIGNUP_URL}><Button type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'}><SVG name={'Google'} size={24} />Continue with Google</Button></a>
                                    </Form.Item>
                                    </div>
                                </Col>
                            </Row>

                            {/* <Row>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5'} >
                                    <Form.Item className={'m-0 w-full'} loading={dataLoading}>
                                        <Button loading={dataLoading} type={'default'} size={'large'} style={{background:'#fff', boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'}} className={'w-full'} onClick={() => setShowModal(true)}><SVG name={'S_Key'} size={24} color={'#8692A3'} /> More SSO Options</Button>
                                    </Form.Item>
                                    </div>
                                </Col>
                            </Row> */}

                            <Row>
                                <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-6'} >
                                    <Text type={'paragraph'} mini color={'grey'}>Already have an account?<a disabled={dataLoading} onClick={() => routeChange('/login')}> Log In</a></Text>
                                    </div>
                                </Col>
                            </Row>
                        </Form>
                        </Col>

                        </Row>

                        </Col>
                    </Row>
                </Col>
                {/* //right side content ends here */}
            </div>
            {/* //parent container ends here */}
      </div>
        }
        {formData &&
            <Congrats data = {formData} />
        }

        {/* <MoreAuthOptions showModal={showModal} setShowModal={setShowModal}/>   */}
    </>

  );
}

export default connect(null, { signup, createHubspotContact, getHubspotContact })(SignUp);
