import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { signup } from 'Reducers/agentActions';
import Congrats from './Congrats';
const axios = require('axios').default;

function UserData({ signup, data }) {
    const [form] = Form.useForm();
    const [dataLoading, setDataLoading] = useState(false);
    const [errorInfo, seterrorInfo] = useState(null);
    const [formData, setformData] = useState(null);

    const getOwner = () => {

        const ownersData = [
            {
                "value" : 116046946,
            },
            {
                "value" : 116047122,
            },
            {
                "value" : 116053799,
            }
        ]
        const index = Math.floor(Math.random()*3);
        const data = ownersData[index];
        return data;
    }

    const UserDataFn =() => {
        setDataLoading(true);
        form.validateFields().then((values) => {
            setDataLoading(true);

            const owner = getOwner();

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
                        "property": "website",
                        "value": values.website
                    },
                    {
                        "property": "phone",
                        "value": values.phone
                    },
                    {
                        "property": "monthly_tracked_users",
                        "value": values.monthly_tracked_users
                    },
                    {
                        "property": "team_size",
                        "value": values.team_size
                    },
                    {
                        "property": "hubspot_owner_id",
                        "value": owner.value
                    }                     
                ]
            }

            const APIKEY = '69137c15-00a5-4d12-91e7-9641797e9572';
            const url = `https://api.hubapi.com/contacts/v1/contact/createOrUpdate/email/${data.email}/?hapikey=${APIKEY}`;
            
            axios.post(url, jsonData)
            .then(function (response) {
                console.log(response);
                setDataLoading(false);
                setformData(values);
            })
            .catch(function (error) {
                console.log(error);
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
    {!formData &&
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
                                            name="phone"
                                            rules={[{ required: true, message: 'Please enter phone number' }]}
                                            >
                                            <Input className={'fa-input w-full'} disabled={dataLoading} size={'large'} placeholder="Phone Number" />
                                        </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="website"
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
                                            name="monthly_tracked_users"
                                            rules={[{ required: true, message: 'Please select Estimated monthly tracked users' }]}
                                            >
                                            <Select
                                                placeholder="Estimated monthly tracked users"
                                                allowClear
                                            >
                                                <Option value="Less than 5k">Less then 5k</Option>
                                                <Option value="5k - 50k">5k - 50k</Option>
                                                <Option value="50k - 200k">5k - 200k</Option>
                                                <Option value="Greater than 200k">Greater than 200k</Option>
                                            </Select>
                                            </Form.Item>
                                    </div>
                            </Col>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-5 w-full'} >
                                            <Form.Item label={null}
                                            name="team_size"
                                            rules={[{ required: true, message: 'Please select team size' }]}
                                            >
                                            <Select
                                                placeholder="Team Size"
                                                allowClear
                                            >
                                                <Option value="1-10 employees">1-10 employees</Option>
                                                <Option value="11-50 employees">11-50 employees</Option>
                                                <Option value="51-200 employees">51-200 employees</Option>
                                                <Option value="201-500 employees">201-500 employees</Option>
                                                <Option value="501-1000 employees">501-1000 employees</Option>
                                                <Option value="1001-5000 employees">1001-5000 employees</Option>
                                                <Option value="5001-10000 employees">5001-10000 employees</Option>
                                                <Option value="10001+ employees">10001+ employees</Option>
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
        }
        {formData &&
            <Congrats signup={signup} data = {data} />
        }
</>

  );
}

export default connect(null, { signup })(UserData);
