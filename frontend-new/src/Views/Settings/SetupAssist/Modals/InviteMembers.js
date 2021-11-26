import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { projectAgentInvite, fetchProjectAgents } from 'Reducers/agentActions';
const { Option } = Select;

function BasicDetails(props) {
  const [errorInfo, seterrorInfo] = useState(null);
  const [form] = Form.useForm();
  const [role, setRole]= useState('admin');

  const inviteUser = (payload) => {
    // console.log('Success! payload values:', payload);
    seterrorInfo(null);
    const data = {...payload, 'role':role}
    props.projectAgentInvite(props.activeProjectID, data).then(() => {
      props.fetchProjectAgents(props.activeProjectID);
      props.onCancel();
      message.success('Invitation sent successfully!');
    }).catch((err) => {
      console.log('invite error', err);
      form.resetFields();
      seterrorInfo(err); 
    });
  }; 
  const onChange = () => {
    seterrorInfo(null);
  };
  const onReset = () => {
    props.onCancel();
    seterrorInfo(null);
    form.resetFields();
  };

  const selectAfter = (
    <Select defaultValue="admin" onChange={(value) => setRole(value)} className="select-after">
      <Option value="admin">Admin</Option>
      <Option value="user">User</Option>
    </Select>
  );

  return (
    <>
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
                    onChange={onChange}
                    >
                    <Row>
                        <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mb-2'}>Email</Text>
                            <Form.Item
                                label={null}
                                name="email"
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'} addonAfter={selectAfter} placeholder={'Enter email address'} />
                            </Form.Item>
                        </Col>
                        {/* <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-6 mb-2'}>Email</Text>
                            <Form.Item
                                label={null}
                                name="email1"
                                rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                            >
                            <Input className={'fa-input'} size={'large'} addonAfter={selectAfter} placeholder={'Enter email address'} />
                            </Form.Item>
                        </Col> */}
                        <Col span={24}>
                            <div className={'mt-20 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type="primary" style={{width:'280px', height:'36px'}} className={'ml-2'} htmlType="submit">
                                    Invite and Continue
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                        <Col span={24}>
                            <div className={'mt-4 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type={'text'} style={{width:'280px', height:'36px'}} htmlType="text" onClick={onReset}>
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

    </>

  );
}
const mapStateToProps = (state) => ({
    activeProjectID: state.global.active_project.id
});
export default connect(mapStateToProps, { projectAgentInvite, fetchProjectAgents })(BasicDetails);
