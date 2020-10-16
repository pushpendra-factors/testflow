import React, { useState, useEffect } from 'react';
import {
  Row, Col, Modal, Input, Select
} from 'antd';
import { Text } from 'factorsComponents';
import { connect } from 'react-redux';
import { projectAgentInvite } from '../../../../reducers/agentActions';
const { Option } = Select;

function InviteUsers(props) {
  const [inviteCount, setInviteCount] = useState([1]);
  // const addInviteRow = () => {
  //   setInviteCount([...inviteCount, 'newElement']);
  // };
  useEffect(() => {
    props.projectAgentInvite(props.activeProjectID, 'baliga.vishnu+12@gmail.com');
  }, []);

  return (
    <>

      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={props.onCancel}
        className={'fa-modal--regular'}
        okText={'Invite'}
        onOk={props.onOk}
        confirmLoading={props.confirmLoading}
        centered={true}
        afterClose={() => setInviteCount([1])}
      >
        <div className={'p-4'}>
          <Row className={'mb-6'}>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Invite Users</Text>
            </Col>
          </Row>
          {inviteCount.map((item, index) => {
            return (
            <Row key={index} gutter={[24, 24]}>
                <Col span={16}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
                <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} />
                </Col>
                <Col span={8}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Role</Text>
                <Select disabled={props.confirmLoading} className={'fa-select w-full'} size={'large'} defaultValue="Admin">
                    <Option value="DD-MM">Admin</Option>
                    <Option value="MM-DD">Owner</Option>
                </Select>
                </Col>
            </Row>
            );
          })}

          {/* <Row className={'mt-6'}>
            <Col span={24}>
                <Button type="text" disabled={props.confirmLoading} onClick={() => addInviteRow(true)}>Add another user</Button>
            </Col>
          </Row> */}
        </div>

      </Modal>

    </>

  );
}
const mapStateToProps = (state) => ({
  activeProjectID: state.global.active_project.id
});
export default connect(mapStateToProps, { projectAgentInvite })(InviteUsers);
