import React, { useState } from 'react';
import {
  Row, Col, Modal, Input, Select, Form, Button, message
} from 'antd';
import { Text } from 'factorsComponents';
import { connect } from 'react-redux';
import { projectAgentInvite, fetchProjectAgents } from 'Reducers/agentActions';
import factorsai from 'factorsai';
const { Option } = Select;

function InviteUsers(props) {
  const [errorInfo, seterrorInfo] = useState(null);
  const [form] = Form.useForm();

  const inviteUser = (payload) => {
    // console.log('Success! payload values:', payload);

    //Factors INVITE tracking
    factorsai.track('INVITE',{'email':payload?.email,'role': payload?.role, 'activeProjectID': props?.activeProjectID});

    seterrorInfo(null);
    props.projectAgentInvite(props.activeProjectID, payload).then(() => {
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

  return (
    <>

      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={props.onCancel}
        className={'fa-modal--regular fa-modal--slideInDown'}
        footer={false}
        confirmLoading={props.confirmLoading}
        centered={true}
        maskClosable={false}
        afterClose={onReset}
        transitionName=""
        maskTransitionName=""
      >
        <div className={'p-4'}>
          <Row className={'mb-6'}>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Invite Users</Text>
            </Col>
          </Row>
         <Form
            form={form}
            name="inviteUser"
            onFinish={inviteUser}
            onChange={onChange}
            className={'w-full'}
          >
            <Row gutter={[24, 24]}>

                <Col span={16}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
                  <Form.Item name="email" rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'} >
                    <Input size="large" className={'fa-input w-full'} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Role</Text>
                  <Form.Item name="role" rules={[{ required: true, message: 'Please choose user role' }]} className={'m-0'} >
                    <Select className={'fa-select w-full'} size={'large'}>
                        <Option value={2}>Admin</Option>
                        <Option value={1}>User</Option>
                    </Select>
                  </Form.Item>
                </Col>
                {errorInfo && <Col span={24}>
                    <div className={'flex flex-col justify-center items-center mt-1'} >
                        <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                    </div>
                </Col>
                }
                <Col span={24}>
                  <div className={'flex justify-end'}>
                    <Button size={'large'} onClick={onReset} className={'mr-2'}>Cancel</Button>
                    <Button size={'large'} type="primary" htmlType="submit">Invite</Button>
                  </div>
                </Col>

            </Row>
                </Form>

        </div>

      </Modal>

    </>

  );
}
const mapStateToProps = (state) => ({
  activeProjectID: state.global.active_project.id
});
export default connect(mapStateToProps, { projectAgentInvite, fetchProjectAgents })(InviteUsers);
