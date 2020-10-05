import React from 'react';
import {
  Row, Input, Button, Modal, Col, Form
} from 'antd';
import { Text } from 'factorsComponents';
import { createProject } from '../../reducers/global';
import { connect } from 'react-redux';

function CreateNewProject(props) {
  const [form] = Form.useForm();

  const onFinish = values => {
    console.log('Success:', values);
    const submitForm = props.createProject(values.projectName);
    submitForm.then(() => {
      props.setCreateNewProjectModal(false);
    }).catch((err) => {
      console.log('createProject Failed:', err);
    });
  };

  const onFinishFailed = errorInfo => {
    console.log('Failed:', errorInfo);
  };
  const onReset = () => {
    form.resetFields();
  };

  return (
        <Modal
        visible={props.visible}
        onCancel={() => props.setCreateNewProjectModal(false)}
        zIndex={1020}
        className={'fa-modal--regular'}
        footer={null}
        centered={true}
        >
          <div className={'p-4'}>
            <Row>
                <Col span={24}>
                    <Form
                    name="createNewProject"
                    initialValues={{ remember: false }}
                    onFinish={onFinish}
                    onFinishFailed={onFinishFailed}
                    form={form}
                    >
                    <Row>
                        <Col span={24}>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 mb-2'}>Create new project</Text>
                            <Form.Item
                                label={null}
                                name="projectName"
                                rules={[{ required: true, message: 'Please input your Project Name!' }]}
                            >
                            <Input className={'fa-input'} size={'large'} />
                            </Form.Item>
                        </Col>
                        <Col span={24}>
                            <div className={'mt-2 flex justify-end'}>
                                <Form.Item className={'m-0'} noStyle={true}>
                                    <Button htmlType="button" onClick={onReset}>
                                    Reset
                                    </Button>
                                    <Button type="primary" className={'ml-2'} htmlType="submit">
                                    Submit
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                    </Row>
                    </Form>
                </Col>
            </Row>
          </div>

        </Modal>
  );
}

export default connect(null, { createProject })(CreateNewProject);
