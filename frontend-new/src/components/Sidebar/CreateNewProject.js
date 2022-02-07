import React from 'react';
import {
  Row, Input, Button, Modal, Col, Form, message
} from 'antd';
import { Text } from 'factorsComponents';
import { createProject } from '../../reducers/global';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import factorsai from 'factorsai';

function CreateNewProject(props) {
  const [form] = Form.useForm();
  const history = useHistory();

  const onFinish = values => {

    //Factors CREATE_PROJECT tracking
    factorsai.track('CREATE_PROJECT',{'ProjectName':values?.projectName});

    props.createProject(values.projectName).then(() => {
      props.setCreateNewProjectModal(false);
      history.push('/');
      message.success('New Project Created!');
    }).catch((err) => {
      message.error('Oops! Something went wrong.');
      console.log('createProject Failed:', err);
    });
  };

  const onReset = () => {
    props.setCreateNewProjectModal(false);
    form.resetFields();
  };

  return (
        <Modal
        visible={props.visible}
        onCancel={onReset}
        zIndex={1020}
        className={'fa-modal--regular fa-modal--slideInDown'}
        footer={null}
        centered={true}
        afterClose={onReset}
        transitionName=""
        maskTransitionName=""
        >
          <div className={'p-4'}>
            <Row>
                <Col span={24}>
                    <Form
                    name="createNewProject"
                    initialValues={{ remember: false }}
                    onFinish={onFinish}
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
                                    <Button size={'large'} htmlType="button" onClick={onReset}>
                                    Cancel
                                    </Button>
                                    <Button size={'large'} type="primary" className={'ml-2'} htmlType="submit">
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
