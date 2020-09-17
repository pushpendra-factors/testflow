import React from 'react';
import {
  Row, Col, Modal, Input
} from 'antd';
import { Text } from 'factorsComponents';

function EditPassword(props) {
  return (
    <>

      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={props.onCancel}
        className={'fa-modal--regular'}
        okText={'Update Password'}
        onOk={props.onOk}
        confirmLoading={props.confirmLoading}
        centered={true}
      >
        <div className={'p-4'}>
          <Row>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Change Password</Text>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Old Password</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Old Password" />
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>New Password</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="New Password" />
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Confirm Password</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Confirm Password" />
            </Col>
          </Row>
        </div>

      </Modal>

    </>

  );
}

export default EditPassword;
