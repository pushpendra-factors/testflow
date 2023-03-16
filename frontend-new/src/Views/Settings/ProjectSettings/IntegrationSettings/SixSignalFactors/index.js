import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Row, Col, Modal, Input, Form, Button, message, Avatar } from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';
import ConnectedScreen from './ConnectedScreen';
import useAgentInfo from 'hooks/useAgentInfo';

function SixSignalFactorsIntegration({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent
}) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const { email: userEmail } = useAgentInfo();

  useEffect(() => {
    if (currentProjectSettings?.int_factors_six_signal_key) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const onFinish = (values) => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: '6Signal Factors',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      factors6_signal_key: values.api_key,
      int_factors_six_signal_key: true
    })
      .then(() => {
        setLoading(false);
        setShowForm(false);
        setTimeout(() => {
          message.success('6Signal integration successful');
        }, 500);
        setIsActive(true);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          '6Signal Factors'
        );
      })
      .catch((err) => {
        setShowForm(false);
        setLoading(false);
        seterrorInfo(err?.error);
        setIsActive(false);
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
          factors6_signal_key: '',
          int_factors_six_signal_key: false,
          six_signal_config: {}
        })
          .then(() => {
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('6Signal integration disconnected!');
            }, 500);
            setIsActive(false);
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

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with 6Signal Factors integrations' />
      }
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
        <div className='p-4'>
          <Form
            form={form}
            onFinish={onFinish}
            className='w-full'
            onChange={onChange}
          >
            <Row>
              <Col span={24}>
                <Avatar
                  size={40}
                  shape='square'
                  icon={<SVG name='Brand' size={40} color='purple' />}
                  style={{ backgroundColor: '#F5F6F8' }}
                />
              </Col>
            </Row>
            <Row>
              <Col span={24}>
                <Text
                  type='title'
                  level={6}
                  weight='bold'
                  extraClass='m-0 mt-2'
                >
                  Integrate with 6Signal by 6Sense
                </Text>
                <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
                  Add your Backend API key (i.e, 6Signal Secret Key) to connect
                  with your 6Signal account.
                </Text>
              </Col>
            </Row>
            <Row className='mt-6'>
              <Col span={24}>
                <Form.Item
                  name='api_key'
                  rules={[
                    {
                      required: true,
                      message: 'Please input your 6Signal API Key'
                    }
                  ]}
                >
                  <Input
                    size='large'
                    className='fa-input w-full'
                    placeholder='6Signal API Key'
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
                <div className='flex justify-end'>
                  {/* <Button disabled={loading} size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>  */}
                  <Button
                    loading={loading}
                    type='primary'
                    size='large'
                    htmlType='submit'
                  >
                    {' '}
                    Connect Now
                  </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </div>
      </Modal>

      {currentProjectSettings?.int_factors_six_signal_key &&
        userEmail === 'solutions@factors.ai' && (
          <ConnectedScreen
            apiKey={currentProjectSettings?.factors6_signal_key}
          />
        )}
      {currentProjectSettings?.int_factors_six_signal_key &&
        userEmail !== 'solutions@factors.ai' && (
          <div className='mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'>
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              Connected Account
            </Text>
            <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
              API Key
            </Text>
            <Input
              size='large'
              disabled
              placeholder='API Key'
              value={currentProjectSettings?.factors6_signal_key}
              style={{ width: '400px' }}
            />
          </div>
        )}
      <div className='mt-4 flex' data-tour='step-11'>
        {currentProjectSettings?.int_factors_six_signal_key ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disconnect
          </Button>
        ) : (
          <Button
            type='primary'
            loading={loading}
            onClick={() => setShowForm(!showForm)}
          >
            Connect Now
          </Button>
        )}
        {kbLink && (
          <a
            className='ant-btn ml-2 '
            target='_blank'
            href={kbLink}
            rel='noreferrer'
          >
            View documentation
          </a>
        )}
      </div>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings
})(SixSignalFactorsIntegration);
