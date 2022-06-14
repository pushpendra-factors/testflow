import React, { useState } from 'react';
import {
  Row, Col, Modal, Input, Form, Button, message
} from 'antd';
import { Text } from 'factorsComponents';
import { updateAgentInfo, fetchAgentInfo } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import sanitizeInputString from 'Utils/sanitizeInputString';

function EditUserDetails(props) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);

  const onFinish = values => { 
    let sanitizedValues = {
      ...values,
      first_name: sanitizeInputString(values?.first_name),
      last_name: sanitizeInputString(values?.last_name),
    } 
    props.updateAgentInfo(sanitizedValues).then(() => {
      props.fetchAgentInfo().then(() => {
        message.success('Profile details updated!');
        props.onCancel();
      });
    }).catch((err) => {
      console.log('updateAgentInfo failed-->', err);
      seterrorInfo(err.error);
    });
  };

  const onReset = () => {
    seterrorInfo(null);
    props.onCancel();
    form.resetFields();
  };
  const onChange = () => {
    seterrorInfo(null);
  };
  const onModalCancel = () => {
    onReset();
  };

  const { agent } = props;

  return (
    <>
      <Modal
        visible={props.visible}
        zIndex={1020}
        onCancel={onReset}
        afterClose={onModalCancel}
        className={'fa-modal--regular fa-modal--slideInDown'}
        confirmLoading={props.confirmLoading}
        centered={true}
        footer={null}
        transitionName=""
        maskTransitionName=""
      >
        <div className={'p-4'}>

        <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          initialValues = {{
            first_name: agent?.first_name,
            last_name: agent?.last_name,
            phone: agent?.phone
          }}
          >

          <Row>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Edit Details</Text>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>First Name</Text>
              <Form.Item
                    name="first_name"
                    rules={[{ required: true, message: 'Please input your first name.' }]}
              >
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="First Name" />
                    </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Last Name</Text>
              <Form.Item
                    name="last_name"
              >
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Last Name" />
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Phone</Text>
              <Form.Item
                    name="phone"
              >
              <Input disabled={props.confirmLoading} size="large" className={'fa-input w-full'} placeholder="Phone" />
              </Form.Item>
            </Col>
          </Row>
          {errorInfo && <Col span={24}>
                <div className={'flex flex-col justify-center items-center mt-1'} >
                    <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                </div>
            </Col>
            }
            <Row className={'mt-6'}>
              <Col span={24}>
                <div className={'flex justify-end'}>
                  <Button size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>
                  <Button type="primary" size={'large'} htmlType="submit"> Update Details </Button>
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
  agent: state.agent.agent_details
});

export default connect(mapStateToProps, { updateAgentInfo, fetchAgentInfo })(EditUserDetails);
