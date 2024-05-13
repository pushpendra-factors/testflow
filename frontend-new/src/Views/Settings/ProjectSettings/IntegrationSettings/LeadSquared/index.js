import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import {
  fetchProjectSettings,
  udpateProjectSettings,
  enableLeadSquaredIntegration,
  disableLeadSquaredIntegration
} from 'Reducers/global';
import { Row, Col, Modal, Input, Form, Button, message } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const LeadSquaredIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  currentAgent,
  enableLeadSquaredIntegration,
  disableLeadSquaredIntegration
}) => {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);

  const onFinish = (values) => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'leadSquared',
      activeProjectID: activeProject.id
    });

    enableLeadSquaredIntegration(activeProject.id, {
      access_key: values.access_key,
      secret_key: values.secret_key,
      host: values.host
    })
      .then(() => {
        setLoading(false);
        fetchProjectSettings(activeProject.id);
        setTimeout(() => {
          message.success('LeadSquared integration successful');
        }, 500);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          'Leadsquared'
        );
      })
      .catch((err) => {
        setLoading(false);
        seterrorInfo(err.error);
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
        disableLeadSquaredIntegration(activeProject.id)
          .then(() => {
            setLoading(false);
            fetchProjectSettings(activeProject.id);
            setTimeout(() => {
              message.success('LeadSquared integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const onChange = () => {
    seterrorInfo(null);
  };

  useEffect(() => {
    if (currentProjectSettings?.lead_squared_config !== null) {
      // settings form initial values
      form.setFieldsValue({
        access_key:
          currentProjectSettings?.lead_squared_config?.access_key || '',
        secret_key:
          currentProjectSettings?.lead_squared_config?.secret_key || '',
        host: currentProjectSettings?.lead_squared_config?.host || ''
      });
    }
  }, [currentProjectSettings]);

  const isLeadSquaredEnabled =
    currentProjectSettings?.lead_squared_config !== null;

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with LeadSquared integrations' />
      }
      onError={FaErrorLog}
    >
      <Form
        form={form}
        onFinish={onFinish}
        className='w-full'
        onChange={onChange}
        style={{ width: 320 }}
      >
        <Row className='mt-4'>
          <Col span={24}>
            <Text
              type='title'
              level={7}
              color='character-primary'
              extraClass='m-0 mb-1'
            >
              Access Key
            </Text>
            <Form.Item
              name='access_key'
              rules={[
                {
                  required: true,
                  message: 'Please input your Leadsquared Access Key'
                }
              ]}
            >
              <Input
                disabled={isLeadSquaredEnabled}
                className='fa-input'
                placeholder='Access Key'
                style={{ background: isLeadSquaredEnabled ? '#fff' : '' }}
              />
            </Form.Item>
          </Col>
          <Col span={24} className='mt-2'>
            <Text
              type='title'
              level={7}
              color='character-primary'
              extraClass='m-0 mb-1'
            >
              Secret Key
            </Text>
            <Form.Item
              name='secret_key'
              rules={[
                {
                  required: true,
                  message: 'Please input your Leadsquared Secret Key'
                }
              ]}
            >
              <Input
                disabled={isLeadSquaredEnabled}
                className='fa-input w-full'
                placeholder='Secret Key'
                style={{ background: isLeadSquaredEnabled ? '#fff' : '' }}
              />
            </Form.Item>
          </Col>
          <Col span={24} className='mt-2'>
            <Text
              type='title'
              level={7}
              color='character-primary'
              extraClass='m-0 mb-1'
            >
              Host
            </Text>
            <Form.Item
              name='host'
              rules={[
                {
                  required: true,
                  message: 'Please input your Leadsquared Host'
                }
              ]}
            >
              <Input
                disabled={isLeadSquaredEnabled}
                className='fa-input w-full'
                placeholder='Host'
                style={{ background: isLeadSquaredEnabled ? '#fff' : '' }}
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
        <Row className='mt-6'>
          <Col span={24}>
            {isLeadSquaredEnabled ? (
              <Button loading={loading} onClick={() => onDisconnect()}>
                Disconnect
              </Button>
            ) : (
              <Button loading={loading} type='primary' htmlType='submit'>
                {' '}
                Connect Now{' '}
              </Button>
            )}
          </Col>
        </Row>
      </Form>
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
  udpateProjectSettings,
  enableLeadSquaredIntegration,
  disableLeadSquaredIntegration
})(LeadSquaredIntegration);
