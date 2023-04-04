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
import {
  CaretDownOutlined,
  CaretUpOutlined,
  LoadingOutlined
} from '@ant-design/icons';

const HorizontalCard = ({
  isDropdown,
  setIsModalRequestAccess,
  setIsStep2Done,
  udpateProjectSettings,
  title,
  description,
  icon,
  type,
  onSuccess,
  onDeactivate,
  api_key = '',
  isActivated = false
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const onFinishFailed = () => {};
  const Type1Form = () => {
    return (
      <div>
        <Form
          onFinish={onSuccess}
          onFinishFailed={onFinishFailed}
          style={{
            display: 'flex',
            flexWrap: 'nowrap',
            margin: '10px 0'
          }}
        >
          <Button
            htmlType='submit'
            style={{ margin: '0 0px', padding: '0 10px' }}
            icon={isActivated === true ? <SVG name='Greentick' /> : ''}
          >
            {isActivated === true ? 'Request Sent' : 'Activate'}
          </Button>
        </Form>
      </div>
    );
  };
  const onFinish = async (values) => {
    try {
      if (isActivated) {
        onDeactivate();
        return;
      }
      if (values.api_key === api_key) {
        message.success('API Key Already Set!');
        return;
      }
      setIsLoading(true);
      await onSuccess(values);
      setIsLoading(false);
    } catch (e) {
      setIsLoading(false);
      message.error('Some error Occured!');
    }
  };
  const Type2Form = () => {
    return (
      <Row style={{ margin: '10px 0' }}>
        <Row style={{ width: '100%' }}>
          <Text type={'title'} level={7} weight={'bold'} style={{ margin: 0 }}>
            Enter your API Key
          </Text>
        </Row>
        <Row style={{ width: '100%' }}>
          <Form
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
            style={{
              display: 'flex',
              flexWrap: 'nowrap',
              margin: '10px 0'
            }}
            initialValues={{ api_key: api_key }}
          >
            <Form.Item
              name='api_key'
              rules={[{ required: true, message: 'Please enter API Key!' }]}
              style={{ margin: '0 10px' }}
            >
              <Input
                disabled={isActivated}
                placeholder='eg: xxxxxxxxxxxxxxxx'
                style={{ minWidth: '320px' }}
              />
            </Form.Item>
            <Button
              htmlType='submit'
              style={{ margin: '0 10px', padding: '0 10px' }}
              icon={
                isActivated === true ? (
                  <SVG name='Greentick' />
                ) : isLoading === true ? (
                  <LoadingOutlined />
                ) : (
                  ''
                )
              }
            >
              {isActivated === true ? 'Activated' : 'Activate'}
            </Button>
          </Form>
        </Row>
      </Row>
    );
  };
  return (
    <Row className={styles['horizontalCard']}>
      <div className={styles['horizontalCardContent']}>
        <div className={styles['horizontalCardLeft']}>
          <div
            style={{ display: 'grid', placeContent: 'baseline' }}
            className={styles['brand']}
          >
            {' '}
            <div className={styles['brandItem']}>{icon}</div>
            {/* <SVG name={'SixSignalLogo'} size={40} color='purple' />{' '} */}
          </div>
          <div>
            <Text
              type={'title'}
              level={6}
              weight={'bold'}
              style={{ margin: 0 }}
            >
              {title}
            </Text>
            <div>{description}</div>
            {type === 2 ? <Type2Form /> : type === 1 ? <Type1Form /> : ''}
          </div>
        </div>
        <div className={styles['horizontalCardRight']}>
          {/* <Button onClick={isOpen === false ? openDropDown : null}>
            Connect
          </Button> */}
        </div>
      </div>
    </Row>
  );
};
const OnBoard2 = ({ isStep2Done, setIsStep2Done, udpateProjectSettings }) => {
  const [isModalRequestAccess, setIsModalRequestAccess] = useState(false);
  const activeProject = useSelector((state) => state?.global?.active_project);
  const [isThirdPartyOpen, setIsThirdPartyOpen] = useState(false);
  const currentAgent = useSelector((state) => state?.agent?.agent_details);
  const history = useHistory();
  const dispatch = useDispatch();
  const { int_client_six_signal_key, int_clear_bit, clearbit_key } =
    useSelector((state) => state?.global?.currentProjectSettings);
  const { client6_signal_key } = useSelector(
    (state) => state?.global?.currentProjectSettings
  );
  const factors6SignalKeyRequested = useSelector(
    (state) => state?.onBoardFlow?.factors6SignalKeyRequested
  );
  const is_deanonymization_requested = useSelector(
    (state) =>
      state?.global?.currentProjectSettings?.is_deanonymization_requested
  );

  const [isLoadingFactors6Signal, setIsLoadignFactors6Signal] = useState(false);
  // useEffect(() => {
  //   dispatch({
  //     type: TOGGLE_DISABLED_STATE_NEXT_BUTTON,
  //     payload: { step: '2', state: int_client_six_signal_key }
  //   });
  // }, [int_client_six_signal_key]);
  const handleFactors6SignalSetup = () => {
    setIsLoadignFactors6Signal(true);
    udpateProjectSettings(activeProject.id, {
      is_deanonymization_requested: true
    })
      .then(() => {
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

        setIsLoadignFactors6Signal(false);
        history.push('/welcome/visitoridentification/3');

        message.success('Requested for Factors 6 Signal Key');
      })
      .catch((error) => {
        message.error(error);
        console.error(error);

        setIsLoadignFactors6Signal(false);
      });
  };
  const handleClient6SignalKeyActivate = (values) => {
    return new Promise((resolve, reject) => {
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
          resolve(true);
        })
        .catch((err) => {
          console.error(err);
          reject(err);
        });
    });
  };
  const onClient6SignalDeactivate = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        udpateProjectSettings(activeProject.id, {
          client6_signal_key: '',
          int_client_six_signal_key: false
        })
          .then(() => {
            setTimeout(() => {
              message.success('6Signal integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
          });
      },
      onCancel: () => {}
    });
  };
  const onClearBitActivate = (values) => {
    // setLoading(true);
    return new Promise((resolve, reject) => {
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
          // setLoading(false);
          // setShowForm(false);
          setTimeout(() => {
            message.success('Clearbit integration successful');
          }, 500);
          // setIsActive(true);
          resolve(true);
          sendSlackNotification(
            currentAgent.email,
            activeProject.name,
            'Reveal'
          );
        })
        .catch((err) => {
          // setShowForm(false);
          // setLoading(false);
          // seterrorInfo(err?.error);
          // setIsActive(false);
          reject();
        });
    });
  };

  const onClearBitDeactivate = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        // setLoading(true);
        udpateProjectSettings(activeProject.id, {
          clearbit_key: '',
          int_clear_bit: false
        })
          .then(() => {
            // setLoading(false);
            // setShowForm(false);
            setTimeout(() => {
              message.success('Clearbit integration disconnected!');
            }, 500);
            // setIsActive(false);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            // setShowForm(false);
            // setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const handleClearBitKeyActivate = (values) => {};
  return (
    <div className={styles['onBoardContainer']}>
      {/* <SixSignal setIsActive={() => {}} kbLink={true} /> */}
      <div>
        <HorizontalCard
          isDropdown={true}
          setIsModalRequestAccess={setIsModalRequestAccess}
          isStep2Done={isStep2Done}
          setIsStep2Done={setIsStep2Done}
          udpateProjectSettings={udpateProjectSettings}
          title={'Activate Factors Deanonymisation'}
          description={
            'Use Factors API key to get started with Website Visitor Identification immediately. Your usage will be charged based on our plan'
          }
          icon={<SVG size={32} name='Brand' />}
          type={1}
          onSuccess={() => {
            if (is_deanonymization_requested === true) {
              message.success('Already Requested!');
            } else {
              setIsModalRequestAccess(true);
            }
          }}
          api_key={''}
          isActivated={is_deanonymization_requested}
        />
        <Divider />

        {isThirdPartyOpen === true ? (
          <div className={styles['toggleMenu']}>
            <Text type={'title'} level={6} weight={'bold'}>
              Third party integrations
            </Text>
            <HorizontalCard
              isDropdown={true}
              setIsModalRequestAccess={setIsModalRequestAccess}
              isStep2Done={isStep2Done}
              setIsStep2Done={setIsStep2Done}
              udpateProjectSettings={udpateProjectSettings}
              title={'6Sense by 6Signal'}
              description={
                'If you have a 6Signal API key, add it below to use it directly in Factors. You usage will be charged as per 6Signals plans.'
              }
              icon={<SVG size={32} name='SixSignalLogo' />}
              type={2}
              onSuccess={handleClient6SignalKeyActivate}
              onDeactivate={onClient6SignalDeactivate}
              api_key={
                int_client_six_signal_key === true ? client6_signal_key : ''
              }
              isActivated={int_client_six_signal_key}
            />
            <HorizontalCard
              isDropdown={true}
              setIsModalRequestAccess={setIsModalRequestAccess}
              isStep2Done={isStep2Done}
              setIsStep2Done={setIsStep2Done}
              udpateProjectSettings={udpateProjectSettings}
              title={'Clearbit reveal'}
              description={
                'If you have a 6Signal API key, add it below to use it directly in Factors. You usage will be charged as per 6Signals plans.'
              }
              icon={<SVG size={32} name='ClearbitLogo' />}
              type={2}
              onSuccess={onClearBitActivate}
              onDeactivate={onClearBitDeactivate}
              api_key={int_clear_bit === true ? clearbit_key : ''}
              isActivated={int_clear_bit}
            />
          </div>
        ) : (
          ''
        )}
        <div
          style={{
            color: '#1890FF',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            width: 'fit-content'
          }}
          onClick={() => setIsThirdPartyOpen((prev) => !prev)}
        >
          {isThirdPartyOpen === false ? (
            <>
              I have a third party API key <CaretDownOutlined />
            </>
          ) : (
            <>
              Show less
              <CaretUpOutlined />
            </>
          )}
        </div>
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
              We have received your request, Factors deanonymisation will be
              activated for you shortly.
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
                {isLoadingFactors6Signal === true ? <LoadingOutlined /> : ''}{' '}
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
