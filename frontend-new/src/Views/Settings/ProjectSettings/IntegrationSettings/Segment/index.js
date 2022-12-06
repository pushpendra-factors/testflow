import React, { useState, useEffect, useRef } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Input, Button, message, Modal, Form, Row, Col, Avatar } from 'antd';
import { Text, SVG, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

function SegmentIntegration({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent
}) {
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [copySuccess, setCopySuccess] = useState('');
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const textAreaRef = useRef(null);

  currentProjectSettings =
    currentProjectSettings?.project_settings || currentProjectSettings;

  useEffect(() => {
    if (currentProjectSettings?.int_segment) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const copyToClipboard = async () => {
    textAreaRef.current.select();
    try {
      await navigator.clipboard.writeText(activeProject?.private_token);
      setCopySuccess('Copied!');
    } catch (err) {
      setCopySuccess('Failed to copy!');
    }
  };

  const enableSegment = () => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'segment',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      int_segment: true
    })
      .then(() => {
        copyToClipboard();
        fetchProjectSettings(activeProject.id);
        setLoading(false);
        setShowForm(false);
        setTimeout(() => {
          message.success('Segment integration enabled!');
        }, 500);
        setIsActive(true);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          'Segment'
        );
      })
      .catch((err) => {
        setShowForm(false);
        setLoading(false);
        message.error(`${err?.data?.error}`);
        console.log('change password failed-->', err);
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
          int_segment: false
        })
          .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('Segment integration disabled!');
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
        <FaErrorComp subtitle='Facing issues with Segment integrations' />
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
          <Form form={form} className='w-full' onChange={onChange}>
            <Row>
              <Col span={24}>
                <Avatar
                  size={40}
                  shape='square'
                  icon={<SVG name='Segment_ads' size={40} color='purple' />}
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
                  Integrate with Segment
                </Text>
                <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
                  First, take your API key and configure Factors as a
                  destination in your Segment Workspace. Once done, enable all
                  the data sources inside Segment that you would like to send to
                  Factors. We start bringing in data only once you've completed
                  these steps.
                </Text>
              </Col>
            </Row>
            <Row className='mt-6'>
              <Col span={24}>
                <Text type='title' level={7} color='grey-2' extraClass='m-0'>
                  API Key
                </Text>
                <Input
                  size='large'
                  className='fa-input w-full'
                  placeholder='API Key'
                  ref={textAreaRef}
                  value={activeProject?.private_token}
                />
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
                  <Button
                    disabled={loading}
                    size='large'
                    onClick={onReset}
                    className='mr-2'
                  >
                    {' '}
                    Cancel{' '}
                  </Button>
                  <Button
                    loading={loading}
                    type='primary'
                    size='large'
                    onClick={enableSegment}
                  >
                    {' '}
                    {copySuccess || 'Copy Code'}
                  </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </div>
      </Modal>
      {currentProjectSettings?.int_segment && (
        <div className='mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Integration Details
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0 mt-1 mb-3'>
            First, take your API key and configure Factors as a destination in
            your Segment Workspace. Once done, enable all the data sources
            inside Segment that you would like to send to Factors. We start
            bringing in data only once you've completed these steps.
          </Text>
          <div>
            <Input
              size='large'
              ref={textAreaRef}
              placeholder='API Key'
              value={activeProject?.private_token}
              style={{ width: '300px' }}
            />
            <Button
              type='link'
              icon={<SVG name='TextCopy' size={16} color='blue' />}
              onClick={copyToClipboard}
              size='large'
            >
              {copySuccess || 'Copy'}
            </Button>
          </div>
        </div>
      )}
      <div className='mt-4 flex'>
        {currentProjectSettings?.int_segment ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disable
          </Button>
        ) : (
          <Button
            type='primary'
            loading={loading}
            onClick={() => setShowForm(!showForm)}
          >
            Get API Key
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
})(SegmentIntegration);
