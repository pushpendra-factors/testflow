import React, { useState } from 'react';
import { connect, useDispatch } from 'react-redux';
import { udpateProjectSettings } from 'Reducers/global';
import { Row, Col, Modal, Input, Form, Button, message, Avatar } from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { createDashboardFromTemplate } from 'Reducers/dashboard_templates/services';
import { fetchDashboards } from 'Reducers/dashboard/services';
import { fetchQueries } from 'Reducers/coreQuery/services';
import logger from 'Utils/logger';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';
import { getDefaultTimelineConfigForSixSignal } from '../util';
import { sendSlackNotification } from '../../../../../utils/slack';

function SixSignalIntegration({
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  kbLink = false,
  currentAgent
}) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const dispatch = useDispatch();

  // activating visitor identification template when 6 signal keys are added
  const activateVisitorIdentificationTemplate = async () => {
    try {
      if (!activeProject?.id) return;
      const res = await createDashboardFromTemplate(
        activeProject.id,
        // eslint-disable-next-line no-undef
        BUILD_CONFIG.firstTimeDashboardTemplates?.websitevisitoridentification
      );
      if (res) {
        dispatch(fetchDashboards(activeProject.id));
        dispatch(fetchQueries(activeProject.id));
      }
    } catch (error) {
      logger.error('Error in activating visitor identification', error);
    }
  };

  const onFinish = (values) => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: '6Signal',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      client6_signal_key: values.api_key,
      int_client_six_signal_key: true,
      // updating table user and account table config when six signal is activated
      timelines_config: getDefaultTimelineConfigForSixSignal(
        currentProjectSettings
      )
    })
      .then(() => {
        setLoading(false);
        setShowForm(false);
        activateVisitorIdentificationTemplate();
        setTimeout(() => {
          message.success('6Signal integration successful');
        }, 500);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          '6Signal'
        );
        dispatch(fetchFeatureConfig(activeProject.id));
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
          client6_signal_key: '',
          int_client_six_signal_key: false
        })
          .then(() => {
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('6Signal integration disconnected!');
            }, 500);
            dispatch(fetchFeatureConfig(activeProject.id));
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
        <FaErrorComp subtitle='Facing issues with 6Signal integrations' />
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
                  icon={<SVG name='SixSignalLogo' size={40} color='purple' />}
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
      {currentProjectSettings?.int_client_six_signal_key && (
        <div className='mt-4 flex flex-col  w-full'>
          <Text
            type='title'
            level={7}
            color='text-primary'
            extraClass='m-0 mb-1'
          >
            API Key
          </Text>
          <Input
            disabled
            placeholder='API Key'
            value={currentProjectSettings?.client6_signal_key}
            style={{ width: '320px', background: '#fff' }}
          />
        </div>
      )}
      <div className='mt-4 flex' data-tour='step-11'>
        {currentProjectSettings?.int_client_six_signal_key ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disconnect
          </Button>
        ) : (
          <Button loading={loading} onClick={() => setShowForm(!showForm)}>
            Connect Now
          </Button>
        )}
        {kbLink && (
          <a
            className='ant-btn-text ml-2 flex items-center px-5'
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
  udpateProjectSettings
})(SixSignalIntegration);
