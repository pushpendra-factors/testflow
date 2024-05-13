import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Row, Col, Modal, Input, Form, Button, message, Avatar } from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const G2Intergration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  currentAgent
}) => {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);

  const onFinish = (values) => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'G2',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      int_g2_api_key: values.api_key,
      int_g2: true
    })
      .then(() => {
        setLoading(false);
        setShowForm(false);
        setTimeout(() => {
          message.success('G2 integration successful');
        }, 500);
        sendSlackNotification(currentAgent.email, activeProject.name, 'G2');
      })
      .catch((err) => {
        setShowForm(false);
        setLoading(false);
        seterrorInfo(err?.error);
      });
  };

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        udpateProjectSettings(activeProject.id, {
          int_g2_api_key: '',
          int_g2: false
        })
          .then(() => {
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('G2 integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setShowForm(false);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const onReset = () => {
    seterrorInfo(null);
    setShowForm(false);
    form.resetFields();
  };
  const onChange = () => {
    seterrorInfo(null);
  };

  useEffect(() => {
    if (currentProjectSettings?.int_g2) {
      form.setFieldsValue({
        api_key: currentProjectSettings.int_g2_api_key || ''
      });
    }
  }, [currentProjectSettings]);

  const isG2Enabled = Boolean(currentProjectSettings?.int_g2);

  return (
    <ErrorBoundary
      fallback={<FaErrorComp subtitle='Facing issues with G2 integrations' />}
      onError={FaErrorLog}
    >
      <Modal
        visible={showForm}
        zIndex={1020}
        onCancel={onReset}
        afterClose={() => setShowForm(false)}
        className='fa-modal--regular fa-modal--slideInDown'
        centered
        footer={null}
        closable={false}
        transitionName=''
        maskTransitionName=''
      >
        <div className='p-4' />
      </Modal>
      <div className='mt-4'>
        <Form
          form={form}
          onFinish={onFinish}
          onChange={onChange}
          style={{ width: 320 }}
        >
          <Row>
            <Col span={24}>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0 mb-4'
              >
                G2 API Key
              </Text>
              <Form.Item
                name='api_key'
                rules={[
                  {
                    required: true,
                    message: 'Please input your G2 API Key'
                  }
                ]}
              >
                <Input
                  disabled={isG2Enabled}
                  className='fa-input w-full'
                  placeholder='G2 API Key'
                  style={{ background: isG2Enabled ? '#fff' : '' }}
                />
              </Form.Item>
            </Col>
            {errorInfo && (
              <Col span={24}>
                <div className='flex flex-col justify-center items-center mt-1'>
                  <Text type='title' color='red' size='7' className='m-0'>
                    {errorInfo}
                  </Text>
                </div>
              </Col>
            )}
          </Row>
          <Row className='mt-4'>
            <Col span={24}>
              {isG2Enabled ? (
                <Button loading={loading} onClick={() => onDisconnect()}>
                  Disconnect
                </Button>
              ) : (
                <Button loading={loading} type='primary' htmlType='submit'>
                  Connect Now
                </Button>
              )}
            </Col>
          </Row>
        </Form>
      </div>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings
})(G2Intergration);
