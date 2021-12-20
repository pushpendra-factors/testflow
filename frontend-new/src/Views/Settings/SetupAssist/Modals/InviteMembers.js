import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select
} from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { projectAgentBatchInvite, fetchProjectAgents } from 'Reducers/agentActions';
import Brand from './Brand';
const { Option } = Select;

function BasicDetails({handleCancel, fetchProjectAgents, projectAgentBatchInvite, activeProjectID}) {
  const [form] = Form.useForm();
  const [formData, setFormData] = useState(null);

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
      setFormData(finalData);
      message.success('Invitation sent successfully!');
    }).catch((err) => {
      console.log('invite error', err);
      form.resetFields();
      message.error(err);
    });
  }; 

  const onSkip = () => {
    form.resetFields();
    setFormData(true);
  };

  const RoleTypes =[
    {value: 2, label: 'Admin'},
    {value: 1, label: 'User'}
  ];

  const RoleTypeSelect = <Select options={RoleTypes} />;

  return (
    <>
    {!formData &&
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-20'}>
                        <Row className={'mb-20'}>
                            <Col span={24} >
                                <Text type={'title'} level={3} color={'grey-2'} weight={'bold'}>Invite Team Members</Text>
                                <Progress percent={66.66} strokeWidth={3} showInfo={false} />
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
                        <Col span={24}>
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
                            <Form.Item
                                label={null}
                                {...field}
                                name={[field.name, 'email']}
                                validateTrigger={['onChange', 'onBlur']}
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'}  addonAfter={<Form.Item name={[field.name, "role"]} noStyle initialValue={2}>{RoleTypeSelect}</Form.Item>} placeholder={'Enter email address'} />
                            </Form.Item>
                            {fields.length > 0 ? (
                              <MinusCircleOutlined
                                className="dynamic-delete-button"
                                onClick={() => remove(field.name)}
                              />
                            ) : null}
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
                            <div className={'mt-16 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type="primary" style={{width:'440px', height:'36px'}} className={'ml-2'} htmlType="submit">
                                    Invite and Continue
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'mt-4 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type={'text'} style={{width:'440px', height:'36px'}} htmlType="text" onClick={onSkip}>
                                    Skip now, I will invite later
                                    </Button>
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
      }
      {formData && <Brand handleCancel = {handleCancel}/>}
    </>

  );
}
const mapStateToProps = (state) => ({
    activeProjectID: state.global.active_project.id
});
export default connect(mapStateToProps, { projectAgentBatchInvite, fetchProjectAgents })(BasicDetails);
