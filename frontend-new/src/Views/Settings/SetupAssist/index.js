import React, { useState, useEffect } from 'react';
import { Row, Col, Modal, Button } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import Lottie from 'react-lottie';
import {
  fetchProjectSettingsV1,
  getHubspotContact,
  fetchBingAdsIntegration,
  fetchMarketoIntegration
} from 'Reducers/global';
import Website from './Website';
import AdPlatforms from './AdPlatforms';
import CRMS from './CRMS';
import OtherIntegrations from './OtherIntegrations';
import setupAssistData from '../../../assets/lottie/Final Jan 3 Setupassist.json';
import styles from './index.module.scss';
import { meetLink } from '../../../utils/hubspot';
import { ArrowLeftOutlined } from '@ant-design/icons';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';

function SetupAssist({
  currentAgent,
  integration,
  integrationV1,
  activeProject,
  fetchProjectSettingsV1,
  getHubspotContact,
  bingAds,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  marketo
}) {
  const [current, setCurrent] = useState(0);
  const [showModal, setShowModal] = useState(false);
  const [ownerID, setownerID] = useState();
  const [sdkCheck, setsdkCheck] = useState();
  const history = useHistory();
  const location = useLocation();
  const [isBackBtn, setIsBackButton] = useState(false);
  const { isFeatureConnected: isFactorsDeanonymizationConnected } =
    useFeatureLock(FEATURES.INT_FACTORS_DEANONYMISATION);
  useEffect(() => {
    let searchParams = new URLSearchParams(location.search, {
      get: (searchParams, prop) => searchParams.get(prop)
    });
    if (searchParams.get('redirected_from') === 'onboardflow') {
      setIsBackButton(true);
    }
    const { email } = currentAgent;
    getHubspotContact(email)
      .then((res) => {
        setownerID(res.data.hubspot_owner_id);
      })
      .catch((err) => {
        console.log(err.data.error);
      });
  }, []);

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: setupAssistData,
    rendererSettings: {
      preserveAspectRatio: 'xMidYMid slice'
    }
  };

  useEffect(() => {
    // console.log(fetchProjectSettingsV1(activeProject.id));
    // fetchProjectSettingsV1(activeProject.id).then((res) => {
    //   setsdkCheck(res.data.int_completed);
    //   console.log(res);
    // });
    // fetchBingAdsIntegration(activeProject.id);
    // fetchMarketoIntegration(activeProject.id);
  }, [activeProject, sdkCheck]);

  const checkIntegration =
    integration?.int_segment ||
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit ||
    sdkCheck ||
    bingAds?.accounts ||
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null ||
    integration?.int_client_six_signal_key ||
    isFactorsDeanonymizationConnected ||
    integration?.int_rudderstack;

  return (
    <>
      <div className='fa-container'>
        <Row gutter={[24, 24]} justify='center' className='pt-24 pb-12 mt-0 '>
          {isBackBtn ? (
            <Col>
              <Button
                size='large'
                type='text'
                icon={<ArrowLeftOutlined />}
                onClick={() => {
                  history.go(-1);
                }}
              ></Button>
            </Col>
          ) : (
            ''
          )}
          <Col span={checkIntegration ? 17 : 20}>
            <Text type='title' level={2} weight='bold' extraClass='m-0'>
              Let's get started
            </Text>
            <Text
              type='title'
              level={6}
              weight='regular'
              extraClass='m-0'
              color='grey'
            >
              The first step to get up and running with Factors is to get data
              into your project:
            </Text>
            <img
              src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/pop-gift.png'
              alt='gift'
              style={{
                width: '100%',
                maxWidth: '80px',
                marginLeft: '610px',
                marginTop: '-80px'
              }}
            />
          </Col>
          <Col>
            {checkIntegration ? (
              <Button
                type='default'
                size='large'
                style={{
                  borderColor: '#1E89FF',
                  color: '#1E89FF',
                  background: '#fff',
                  marginTop: '1rem'
                }}
                onClick={() => setShowModal(true)}
              >
                <SVG name='dashboard' size={20} color='blue' />
                Go to Dashboards
              </Button>
            ) : null}
          </Col>
        </Row>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={5} style={{ paddingRight: '20px' }}>
            <div>
              <div
                className={`${current === 0 ? styles.divActive : styles.div}`}
                onClick={() => setCurrent(0)}
              >
                <div
                  className={`${
                    current === 0 ? styles.sideActive : styles.side
                  }`}
                />
                <div
                  className={`${
                    current === 0 ? styles.textActive : styles.text
                  }`}
                >
                  <p className={`${styles.text1}`}>Connect with your</p>
                  <p className={`${styles.text2}`}>Ad Platforms</p>
                </div>
                <div className='m-0'>
                  <SVG
                    name='CaretDown'
                    size={20}
                    color='blue'
                    extraClass={`${
                      current === 0 ? styles.caretActive : styles.caret
                    }`}
                  />
                </div>
              </div>

              <div
                className={`${current === 1 ? styles.divActive : styles.div}`}
                onClick={() => setCurrent(1)}
              >
                <div
                  className={`${
                    current === 1 ? styles.sideActive : styles.side
                  }`}
                />
                <div
                  className={`${
                    current === 1 ? styles.textActive : styles.text
                  }`}
                >
                  <p className={`${styles.text1}`}>Connect with your</p>
                  <p className={`${styles.text2}`}>Website Data</p>
                </div>
                <div className='m-0'>
                  <SVG
                    name='CaretDown'
                    size={20}
                    color='blue'
                    extraClass={`${
                      current === 1 ? styles.caretActive : styles.caret
                    }`}
                  />
                </div>
              </div>

              <div
                className={`${current === 2 ? styles.divActive : styles.div}`}
                onClick={() => setCurrent(2)}
              >
                <div
                  className={`${
                    current === 2 ? styles.sideActive : styles.side
                  }`}
                />
                <div
                  className={`${
                    current === 2 ? styles.textActive : styles.text
                  }`}
                >
                  <p className={`${styles.text1}`}>Connect with your</p>
                  <p className={`${styles.text2}`}>CRMs</p>
                </div>
                <div className='m-0'>
                  <SVG
                    name='CaretDown'
                    size={20}
                    color='blue'
                    extraClass={`${
                      current === 2 ? styles.caretActive : styles.caret
                    }`}
                  />
                </div>
              </div>

              <div
                className={`${current === 3 ? styles.divActive : styles.div}`}
                onClick={() => setCurrent(3)}
              >
                <div
                  className={`${
                    current === 3 ? styles.sideActive : styles.side
                  }`}
                />
                <div
                  className={`${
                    current === 3 ? styles.textActive : styles.text
                  }`}
                >
                  <p className={`${styles.text1}`}>More</p>
                  <p className={`${styles.text2}`}>Integrations</p>
                </div>
                <div className='m-0'>
                  <SVG
                    name='CaretDown'
                    size={20}
                    color='blue'
                    extraClass={`${
                      current === 3 ? styles.caretActive : styles.caret
                    }`}
                  />
                </div>
              </div>
            </div>
            <Row className={`${styles.help} mt-4`}>
              <Col>
                <Text
                  type='title'
                  level={5}
                  weight='bold'
                  color='white'
                  extraClass='m-0 ml-6 mt-4'
                >
                  Need help with setup?
                </Text>
                <Text
                  type='title'
                  level={7}
                  color='white'
                  extraClass='m-0 ml-6 mb-6'
                  style={{ opacity: 0.7 }}
                >
                  Setup a call with our rep. We are always happy to assist to
                  you
                </Text>
                <a href={meetLink(ownerID)} target='_blank' rel='noreferrer'>
                  <Button
                    type='text'
                    style={{ color: '#1890FF' }}
                    className='ml-6 mb-4'
                  >
                    Setup a Call
                  </Button>
                </a>
              </Col>
              <Col className={`${styles.callimg}`}>
                <img
                  src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/1to1.svg'
                  alt='cta'
                />
              </Col>
            </Row>
          </Col>
          <Col span={16} style={{ padding: '0px 0px 0px 30px' }}>
            {current === 0 ? (
              <AdPlatforms />
            ) : current === 1 ? (
              <Website setsdkCheck={setsdkCheck} sdkCheck={sdkCheck} />
            ) : current === 2 ? (
              <CRMS />
            ) : (
              <OtherIntegrations />
            )}
          </Col>
        </Row>
      </div>

      <Modal
        title={null}
        visible={showModal}
        footer={null}
        centered
        mask
        maskClosable={false}
        maskStyle={{ backgroundColor: 'rgb(0 0 0 / 70%)' }}
        closable
        onCancel={() => setShowModal(false)}
        className='fa-modal--regular'
      >
        <div className='fa-container'>
          <Row className='mt-8'>
            <Col>
              <Text type='title' level={4} weight='bold' extraClass='m-0'>
                You are leaving project setup assist now
              </Text>
              <Text
                type='title'
                level={7}
                weight='regular'
                extraClass='ml-5'
                color='grey'
              >
                You can always access it from the bottom left corner
              </Text>
            </Col>
          </Row>
          <Row style={{ marginLeft: '80px' }}>
            <Col>
              <Lottie options={defaultOptions} height={200} width={200} />
            </Col>
          </Row>
          <Row style={{ marginLeft: '120px' }} className='pb-4'>
            <Col>
              <Button type='primary' onClick={() => history.push('/')}>
                Got it, continue
              </Button>
            </Col>
          </Row>
        </div>
      </Modal>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentAgent: state.agent.agent_details,
  integration: state.global.currentProjectSettings,
  bingAds: state.global.bingAds,
  marketo: state.global.marketo,
  integrationV1: state.global.projectSettingsV1
});

export default connect(mapStateToProps, {
  fetchProjectSettingsV1,
  getHubspotContact,
  fetchBingAdsIntegration,
  fetchMarketoIntegration
})(SetupAssist);
