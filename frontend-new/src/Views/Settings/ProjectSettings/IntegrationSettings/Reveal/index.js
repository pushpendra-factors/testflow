import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect, useDispatch } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
  Row,
  Col,
  Modal,
  Input,
  Form,
  Button,
  notification,
  message,
  Avatar
} from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';

const RevealIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent
}) => {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const dispatch = useDispatch();

  useEffect(() => {
    if (currentProjectSettings?.int_clear_bit) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const onFinish = (values) => {
    setLoading(true);

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'reveal',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      clearbit_key: values.api_key,
      int_clear_bit: true
    })
      .then(() => {
        setLoading(false);
        setShowForm(false);
        setTimeout(() => {
          message.success('Clearbit integration successful');
        }, 500);
        setIsActive(true);
        sendSlackNotification(currentAgent.email, activeProject.name, 'Reveal');
        dispatch(fetchFeatureConfig(activeProject.id));
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
          clearbit_key: '',
          int_clear_bit: false
        })
          .then(() => {
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('Clearbit integration disconnected!');
            }, 500);
            setIsActive(false);
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
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp subtitle={'Facing issues with Clearbit integrations'} />
        }
        onError={FaErrorLog}
      >
        <Modal
          visible={showForm}
          zIndex={1020}
          onCancel={onReset}
          afterClose={() => setShowForm(false)}
          className={'fa-modal--regular fa-modal--slideInDown'}
          centered={true}
          footer={null}
          closable={false}
          transitionName=''
          maskTransitionName=''
        >
          <div className={'p-4'}>
            <Form
              form={form}
              onFinish={onFinish}
              className={'w-full'}
              onChange={onChange}
            >
              <Row>
                <Col span={24}>
                  <Avatar
                    size={40}
                    shape={'square'}
                    icon={
                      <SVG name={'ClearbitLogo'} size={40} color={'purple'} />
                    }
                    style={{ backgroundColor: '#F5F6F8' }}
                  />
                </Col>
              </Row>
              <Row>
                <Col span={24}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={'m-0 mt-2'}
                  >
                    Integrate with Clearbit Reveal
                  </Text>
                  <Text
                    type={'title'}
                    level={7}
                    color={'grey'}
                    extraClass={'m-0 mt-2'}
                  >
                    Add your Backend API key (i.e, Clearbit Secret Key) to
                    connect with your Clearbit account.
                  </Text>
                </Col>
              </Row>
              <Row className={'mt-6'}>
                <Col span={24}>
                  <Form.Item
                    name='api_key'
                    rules={[
                      {
                        required: true,
                        message: 'Please input your Clearbit API Key'
                      }
                    ]}
                  >
                    <Input
                      size='large'
                      className={'fa-input w-full'}
                      placeholder='Clearbit API Key'
                    />
                  </Form.Item>
                </Col>
                {errorInfo && (
                  <Col span={24}>
                    <div
                      className={
                        'flex flex-col justify-center items-center mt-1'
                      }
                    >
                      <Text
                        type={'title'}
                        color={'red'}
                        size={'7'}
                        className={'m-0'}
                      >
                        {errorInfo}
                      </Text>
                    </div>
                  </Col>
                )}
              </Row>
              <Row className={'mt-6'}>
                <Col span={24}>
                  <div className={'flex justify-end'}>
                    {/* <Button disabled={loading} size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>  */}
                    <Button
                      loading={loading}
                      type='primary'
                      size={'large'}
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
        {currentProjectSettings?.int_clear_bit && (
          <div
            className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}
          >
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
              Connected Account
            </Text>
            <Text
              type={'title'}
              level={7}
              color={'grey'}
              extraClass={'m-0 mt-2'}
            >
              API Key
            </Text>
            <Input
              size='large'
              disabled={true}
              placeholder='API Key'
              value={currentProjectSettings.clearbit_key}
              style={{ width: '400px' }}
            />
          </div>
        )}
        <div className={'mt-4 flex'} data-tour='step-11'>
          {currentProjectSettings?.int_clear_bit ? (
            <Button loading={loading} onClick={() => onDisconnect()}>
              Disconnect
            </Button>
          ) : (
            <Button
              type={'primary'}
              loading={loading}
              onClick={() => setShowForm(!showForm)}
            >
              Connect Now
            </Button>
          )}
          {kbLink && (
            <a className={'ant-btn ml-2 '} target={'_blank'} href={kbLink}>
              View documentation
            </a>
          )}
        </div>
      </ErrorBoundary>
    </>
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
})(RevealIntegration);
