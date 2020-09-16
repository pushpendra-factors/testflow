import React, { useState, useEffect } from 'react';
import {
  Row, Col, Modal, Input
} from 'antd';
import { Text } from 'factorsComponents';

function EditUserDetails(props) {
  return (
    <>
      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={props.onCancel}
        className={'fa-modal--regular'}
        okText={'Update Details'}
        onOk={props.onOk}
        confirmLoading={props.confirmLoading}
        centered={true}
      >
        <div className={'p-4'}>
          <Row>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Edit Details</Text>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Name</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Name" />
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Email" />
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Phone</Text>
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Phone" />
            </Col>
          </Row>
        </div>

      </Modal>

    </>

  );
}

export default EditUserDetails;
