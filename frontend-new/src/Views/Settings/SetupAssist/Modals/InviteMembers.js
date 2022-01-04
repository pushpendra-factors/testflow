import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select, Space
} from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import { Text, SVG } from 'factorsComponents';
import { projectAgentBatchInvite, fetchProjectAgents } from 'Reducers/agentActions';
import { useHistory } from 'react-router-dom';
const { Option } = Select;

function BasicDetails({handleCancel, fetchProjectAgents, projectAgentBatchInvite, activeProjectID}) {
  const [form] = Form.useForm();
  const history = useHistory();

  const inviteUser = (payload) => {

    const filteredData = Object.fromEntries(
      Object.entries(payload).filter(([key, value]) => key !== 'emails') );

    const emailData = {};
    let i = 1;

    if (payload['emails']) {
      payload['emails'].forEach ((item) => {
        emailData[i] = item; 
        i++;
      })
    }

    const data = {...filteredData, ...emailData}
    
    let finalData = [];
    for(let val in data) {
      finalData[val] = data[val];
    }

    projectAgentBatchInvite(activeProjectID, finalData).then(() => {
      fetchProjectAgents(activeProjectID);
        message.success('Invitation sent successfully!');
        handleCancel();
        history.push('/project-setup');
    }).catch((err) => {
      console.log('invite error', err);
      form.resetFields();
      message.error(err);
    });
  }; 

  const onSkip = () => {
    form.resetFields();
    handleCancel();
    history.push('/project-setup');
  };

  const RoleTypes =[
    {value: 2, label: 'Admin'},
    {value: 1, label: 'User'}
  ];

  const RoleTypeSelect = <Select options={RoleTypes} />;

  return (
    <>
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-16'}>
                        <Row className={'m-0'}>
                            <Col span={24} >
                                <Text type={'title'} level={3} color={'grey-2'} align={'center'} weight={'bold'}>Invite Team Members</Text>
                                {/* <Progress percent={66.66} strokeWidth={3} showInfo={false} /> */}
                            </Col>
                        </Row>
                        <Row className={'mb-2 -mt-2'}>
                            <Col span={24} >
                                <Text type={'title'} size={10} color={'grey'} extraClass={'max-w-md'}>Invite people into your new project for better collaboration and planning. You can always invite more under User Settings</Text>
                            </Col>
                        </Row>
                        <Row>
                    <Col span={24}>
                    <Form
                    form={form}
                    name="inviteUser"
                    onFinish={inviteUser}
                    >
                    <Row>
                        <Col span={23}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mb-2'}>Email</Text>
                            <Form.Item
                                label={null}
                                name={[0, "email"]}
                                validateTrigger={['onChange', 'onBlur']}
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'} addonAfter={<Form.Item name={[0, "role"]} noStyle initialValue={2}>{RoleTypeSelect}</Form.Item>} placeholder={'Enter email address'} />
                            </Form.Item>
                        </Col>
                        {/* <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2 mb-2'}>Email</Text>
                            <Form.Item
                                label={null}
                                name={[1, 'email']}
                                validateTrigger={['onChange', 'onBlur']}
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'} addonAfter={<Form.Item name={[1, "role"]} noStyle initialValue="admin">{RoleTypeSelect}</Form.Item>} placeholder={'Enter email address'} />
                            </Form.Item>
                        </Col> */}
                    <Form.List
                      name="emails"
                      rules={[
                        {
                          validator: async (_, names) => {
                            if (!names || names.length < 1) {
                              return Promise.reject(new Error('At least 1 users'));
                            }
                          },
                        },
                      ]}
                    >
                      {(fields, { add, remove }) => (
                        <>
                          {fields.map((field, index) => (
                        <Col span={24}>
                          <Form.Item
                            required={false}
                            key={field.key}
                          >
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2 mb-2'}>Email</Text>
                          <Row className={`${styles.show}`}>
                          <Col span={23}>
                            <Form.Item
                                label={null}
                                {...field}
                                name={[field.name, 'email']}
                                validateTrigger={['onChange', 'onBlur']}
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'}  addonAfter={<Form.Item name={[field.name, "role"]} noStyle initialValue={2}>{RoleTypeSelect}</Form.Item>} placeholder={'Enter email address'} />
                            </Form.Item>
                            </Col>
                            {fields.length > 0 ? (
                            <Col span={1} className={`${styles.hide}`}>
                              <Button style={{backgroundColor:'white'}} className={'mt-1'} onClick={() => remove(field.name)}>
                                <SVG
                                  name={'Trash'}
                                  size={20}
                                  color='gray'
                                /></Button>
                            </Col>
                                ) : null}
                            </Row>
                          </Form.Item>
                        </Col>
                        ))}
                        
                        <Col span={24} className={'mt-6 ml-2'}>
                          {fields.length === 4 ? null: <Button type={'text'} icon={<PlusOutlined style={{color:'gray', fontSize:'18px'}} />} onClick={() => add()}>Add another user</Button>}
                        </Col>
                        </>
                        )}
                        </Form.List>
                        <Col span={24}>
                            <div className={'mt-8 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type="primary" style={{width:'28vw', height:'36px'}} className={'ml-2'} htmlType="submit">
                                    Invite and Continue
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'mt-4 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                <Button size={'large'} type={'link'} style={{width:'27vw', height:'36px', backgroundColor:'white'}} className={'m-0'} onClick={onSkip}>Skip and continue</Button>
                                </Form.Item>
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
const mapStateToProps = (state) => ({
    activeProjectID: state.global.active_project.id
});
export default connect(mapStateToProps, { projectAgentBatchInvite, fetchProjectAgents })(BasicDetails);
