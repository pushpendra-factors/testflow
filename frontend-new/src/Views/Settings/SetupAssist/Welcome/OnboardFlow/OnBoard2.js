import {
  Button,
  Divider,
  Form,
  Input,
  message,
  Modal,
  Radio,
  Row,
  Space
} from 'antd';
import factorsai from 'factorsai';
import { SVG, Text } from 'Components/factorsComponents';
import React, { useCallback, useEffect, useState } from 'react';
import SixSignal from 'Views/Settings/ProjectSettings/IntegrationSettings/SixSignal';
import styles from './index.module.scss';
import { udpateProjectSettings } from 'Reducers/global';
import { useSelector, connect, useDispatch } from 'react-redux';
import {
  ENABLE_STEP_AND_MOVE_TO_NEXT,
  TOGGLE_DISABLED_STATE_NEXT_BUTTON,
  TOGGLE_FACTORS_6SIGNAL_REQUEST
} from 'Reducers/types';
import { sendSlackNotification } from 'Utils/slack';
import { useHistory } from 'react-router-dom';

const HorizontalCard = ({
  isDropdown,
  setIsModalRequestAccess,
  setIsStep2Done,
  udpateProjectSettings
}) => {
  const dispatch = useDispatch();
  const int_client_six_signal_key = useSelector(
    (state) => state?.global?.currentProjectSettings?.int_client_six_signal_key
  );
  const activeProject = useSelector((state) => state?.global?.active_project);
  const currentProjectSettings = useSelector(
    (state) => state?.global?.currentProjectSettings
  );
  const factors6SignalKeyRequested = useSelector(
    (state) => state?.onBoardFlow?.factors6SignalKeyRequested
  );
  const [selectedOption, setSelectedOption] = useState(undefined);
  const [isOpen, setIsOpen] = useState(false);
  const RadioHandle = (e) => {
    let value = e.target.value;
    console.log(value);
    setSelectedOption(value);
  };
  const openDropDown = useCallback(() => setIsOpen(true), []);
  useEffect(() => {
    if (int_client_six_signal_key) setSelectedOption('1');
  }, [int_client_six_signal_key]);
  const handleVerify6Signal = (values) => {
    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: '6Signal',
      activeProjectID: activeProject.id
    });
    udpateProjectSettings(activeProject.id, {
      client6_signal_key: values.api_key,
      int_client_six_signal_key: true
    })
      .then(() => {
        setTimeout(() => {
          message.success('6Signal integration successful');
        }, 500);

        dispatch({
          type: TOGGLE_DISABLED_STATE_NEXT_BUTTON,
          payload: { step: '2', state: true }
        });
      })
      .catch((err) => {
        console.error(err);
      });
  };
  const handleFactors6SignalKeyRequest = useCallback(() => {
    setIsModalRequestAccess(true);
  }, []);
  return (
    <Row className={styles['horizontalCard']}>
      <div className={styles['horizontalCardContent']}>
        <div className={styles['horizontalCardLeft']}>
          <div style={{ display: 'grid', placeContent: 'center' }}>
            {' '}
            <SVG name={'SixSignalLogo'} size={40} color='purple' />{' '}
          </div>
          <div>
            <Text
              type={'title'}
              level={6}
              weight={'bold'}
              style={{ margin: 0 }}
            >
              Integrate with 6signal by 6sense
            </Text>
            <div>
              Gain insight into who is visiting your website and where they are
              in the buying journey
            </div>
          </div>
        </div>
        <div className={styles['horizontalCardRight']}>
          <Button onClick={isOpen === false ? openDropDown : null}>
            Connect
          </Button>
        </div>
      </div>
      {isOpen === true ? (
        <>
          <Divider style={{ margin: '20px 0 5px 0' }} />
          <div className={styles['cardExtension']}>
            <div className={styles['eachExtension']}>
              <Radio
                name='test'
                defaultChecked={selectedOption == '1'}
                checked={selectedOption == '1'}
                value='1'
                onClick={RadioHandle}
              />
              <div>
                <div
                  onClick={() => setSelectedOption('1')}
                  style={{ cursor: 'pointer' }}
                >
                  <Text
                    type={'title'}
                    level={7}
                    weight={'bold'}
                    style={{ margin: 0 }}
                  >
                    Use your own 6Signal API Key
                  </Text>
                  <div>
                    If you have a Clearbit API key, add it below to use it
                    directly in Factors. You usage will not be capped by your
                    Factors plan.
                  </div>
                </div>
                <div style={{ padding: '10px 0' }}>
                  <div style={{ padding: '2px 0' }}>API Key</div>
                  <Space direction='horizontal'>
                    <Form
                      onFinish={handleVerify6Signal}
                      style={{ display: 'flex' }}
                    >
                      <Form.Item
                        name='api_key'
                        rules={[
                          { required: true, message: 'API Key is required' }
                        ]}
                      >
                        <Input
                          defaultValue={
                            currentProjectSettings?.int_client_six_signal_key
                              ? currentProjectSettings.client6_signal_key
                              : ''
                          }
                          style={{ margin: 0, width: '320px' }}
                          placeholder='Enter/Paste API key here'
                          disabled={
                            selectedOption != '1'
                              ? true
                              : currentProjectSettings?.int_client_six_signal_key
                          }
                        />
                      </Form.Item>
                      {currentProjectSettings?.int_client_six_signal_key ? (
                        <span
                          role='img'
                          aria-label='integration_done'
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            padding: '0 10px'
                          }}
                        >
                          <SVG name='greentick' />
                        </span>
                      ) : (
                        <Button
                          style={{ margin: '0 10px' }}
                          onClick={() => {}}
                          disabled={selectedOption != '1'}
                          htmlType='submit'
                        >
                          Verify key
                        </Button>
                      )}
                    </Form>
                  </Space>
                </div>
              </div>
            </div>
            <div className={styles['eachExtension']}>
              <Radio
                name='test'
                value='2'
                defaultChecked={selectedOption == '2'}
                checked={selectedOption == '2'}
                onClick={RadioHandle}
              />
              <div>
                <div
                  onClick={() => setSelectedOption('2')}
                  style={{ cursor: 'pointer' }}
                >
                  <Text
                    type={'title'}
                    level={7}
                    weight={'bold'}
                    style={{ margin: 0 }}
                  >
                    Use the Factors 6signal Api key
                  </Text>
                  <div>
                    In case you don’t have your own, get started immediately
                    using the Factors’ API key. Usage will be capped according
                    to your plan and conversation with the success team.
                  </div>
                </div>
                <div style={{ padding: '10px 0' }}>
                  <Space direction='horizontal'>
                    <Button
                      size='large'
                      className={styles['btn']}
                      onClick={
                        factors6SignalKeyRequested === false
                          ? handleFactors6SignalKeyRequest
                          : ''
                      }
                      disabled={selectedOption != '2'}
                    >
                      {factors6SignalKeyRequested === true ? (
                        <span>
                          {' '}
                          <SVG name='Greentick' /> Access Requested
                        </span>
                      ) : (
                        'Request Access'
                      )}
                    </Button>
                  </Space>
                </div>
              </div>
            </div>

            {/* <Radio value={3}>C</Radio>
          <Radio value={4}>D</Radio> */}
          </div>
        </>
      ) : (
        ''
      )}
    </Row>
  );
};
const OnBoard2 = ({ isStep2Done, setIsStep2Done, udpateProjectSettings }) => {
  const [isModalRequestAccess, setIsModalRequestAccess] = useState(false);
  const activeProject = useSelector((state) => state?.global?.active_project);
  const currentAgent = useSelector((state) => state?.agent?.agent_details);
  const history = useHistory();
  const dispatch = useDispatch();
  const int_client_six_signal_key = useSelector(
    (state) => state?.global?.currentProjectSettings?.int_client_six_signal_key
  );
  const factors6SignalKeyRequested = useSelector(
    (state) => state?.onBoardFlow?.factors6SignalKeyRequested
  );

  useEffect(() => {
    dispatch({
      type: TOGGLE_DISABLED_STATE_NEXT_BUTTON,
      payload: { step: '2', state: int_client_six_signal_key }
    });
  }, [int_client_six_signal_key]);
  const handleFactors6SignalSetup = () => {
    sendSlackNotification(
      currentAgent.email,
      activeProject.name,
      'factors6Signal_Test'
    );
    if (factors6SignalKeyRequested === false)
      dispatch({
        type: TOGGLE_FACTORS_6SIGNAL_REQUEST
      });
    dispatch({
      type: ENABLE_STEP_AND_MOVE_TO_NEXT,
      payload: { step: 2, state: true, moveTo: 2 }
    });
    history.push('/welcome/visitoridentification/3');

    message.success('Requested for Factors 6 Signal Key');
  };
  return (
    <div className={styles['onBoardContainer']}>
      {/* <SixSignal setIsActive={() => {}} kbLink={true} /> */}
      <div>
        <Text type={'title'} level={6} weight={'bold'}>
          Integrations to push
        </Text>
        <HorizontalCard
          isDropdown={true}
          setIsModalRequestAccess={setIsModalRequestAccess}
          isStep2Done={isStep2Done}
          setIsStep2Done={setIsStep2Done}
          udpateProjectSettings={udpateProjectSettings}
        />
        {console.log('SD')}
        <Modal
          visible={isModalRequestAccess}
          onOk={() => setIsModalRequestAccess(false)}
          onCancel={() => setIsModalRequestAccess(false)}
          bodyStyle={{ borderRadius: '20px' }}
          footer={null}
        >
          <div>
            <Text
              type={'title'}
              level={6}
              weight={'bold'}
              style={{ margin: '0 0 10px 0' }}
            >
              Request have been sent
            </Text>
            <p>
              We have received your request. We will get back to you within half
              a day.
            </p>
            <div
              style={{ width: '100%', display: 'flex', justifyContent: 'end' }}
            >
              <Button
                type='primary'
                style={{
                  padding: '0 22px',
                  height: '40px',
                  borderRadius: '6px'
                }}
                onClick={handleFactors6SignalSetup}
              >
                Continue with setup
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentAgent: state.agent.agent_details,
  projects: state.global.projects
});
export default connect(mapStateToProps, { udpateProjectSettings })(OnBoard2);
