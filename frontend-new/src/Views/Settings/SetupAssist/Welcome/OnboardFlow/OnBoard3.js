import { LoadingOutlined } from '@ant-design/icons';
import { Button, Row, message, Alert, Divider, Modal } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import React, { useCallback, useEffect, useState } from 'react';
import { connect, useSelector } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import {
  disableSlackIntegration,
  enableHubspotIntegration,
  enableSalesforceIntegration,
  enableSlackIntegration,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  fetchSalesforceRedirectURL,
  udpateProjectSettings
} from 'Reducers/global';
import styles from './index.module.scss';
import factorsai from 'factorsai';
import { sendSlackNotification } from 'Utils/slack';

const HorizontalCard = ({
  title,
  description,
  icon,
  is_connected,
  onClickConnect
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const onClick = async () => {
    setIsLoading(true);
    if (onClickConnect) await onClickConnect();
    setIsLoading(false);
  };
  return (
    <Row className={styles['horizontalCard']}>
      <div className={styles['horizontalCardContent']}>
        <div className={styles['horizontalCardLeft']}>
          <div style={{ display: 'grid', placeContent: 'center' }}>{icon}</div>
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
          </div>
        </div>
        <div className={styles['horizontalCardRight']}>
          <Button
            onClick={onClick}
            // icon={isLoading === true ? <LoadingOutlined /> : null}
          >
            {is_connected ? (
              <>
                <SVG name='Greentick' /> Already Connected
              </>
            ) : (
              <>{isLoading ? <LoadingOutlined /> : ''} Connect</>
            )}
          </Button>
        </div>
      </div>
    </Row>
  );
};
const OnBoard3 = ({
  enableSlackIntegration,
  enableHubspotIntegration,
  enableSalesforceIntegration,
  fetchSalesforceRedirectURL,
  udpateProjectSettings,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  disableSlackIntegration
}) => {
  const activeProject = useSelector((state) => state?.global?.active_project);
  const currentAgent = useSelector((state) => state.agent.agent_details);
  const int_slack = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_slack
  );
  const { int_hubspot, int_salesforce_enabled_agent_uuid } = useSelector(
    (state) => state?.global?.currentProjectSettings
  );
  const history = useHistory();
  const {
    int_client_six_signal_key,
    int_factors_six_signal_key,
    int_clear_bit,
    is_deanonymization_requested
  } = useSelector((state) => state?.global?.currentProjectSettings);
  const int_completed = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );

  const is_onboarding_completed = useSelector(
    (state) => state?.global?.currentProjectSettings?.is_onboarding_completed
  );

  const [isLoadingDone, setIsLoadingDone] = useState(false);
  const checkIsValid = (step) => {
    if (step == 1) {
      return int_completed;
    } else if (step == 2) {
      return (
        int_client_six_signal_key ||
        is_deanonymization_requested ||
        int_clear_bit ||
        int_factors_six_signal_key
      );
    }
    return false;
  };
  useEffect(() => {
    if (checkIsValid(1) && checkIsValid(2)) {
      if (is_deanonymization_requested === false)
        udpateProjectSettings(activeProject.id, {
          is_onboarding_completed: true
        });
    }
  }, []);
  const onConnectSlack = () => {
    return new Promise((resolve, reject) => {
      enableSlackIntegration(activeProject.id, window.location.href)
        .then((r) => {
          if (r.status === 200) {
            window.open(r.data.redirectURL, '_self');
          }
          if (r.status >= 400) {
            message.error('Error fetching slack redirect url');
            reject();
          }
        })
        .catch((err) => {
          console.log('Slack error-->', err);
          reject();
        });
    });
  };
  const onDisconnectSlack = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        disableSlackIntegration(activeProject.id)
          .then(() => {
            setTimeout(() => {
              message.success('Slack integration disconnected!');
            }, 500);
            fetchProjectSettingsV1(activeProject.id);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
          });
      },
      onCancel: () => {}
    });
  };

  const onClickEnableHubspot = () => {
    return new Promise((resolve, reject) => {
      //Factors INTEGRATION tracking
      factorsai.track('INTEGRATION', {
        name: 'hubspot',
        activeProjectID: activeProject.id
      });
      enableHubspotIntegration(activeProject.id)
        .then((r) => {
          sendSlackNotification(
            currentAgent.email,
            activeProject.name,
            'Hubspot'
          );
          console.log(r);
          if (r.status == 307) {
            window.location = r.data.redirect_url;
          }
        })
        .catch((err) => {
          message.error(`${err?.data?.error}`);
          reject();
        });
    });
  };

  const onDisconnectHubspot = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        udpateProjectSettings(activeProject.id, {
          int_hubspot_api_key: '',
          int_hubspot: false
        })
          .then(() => {
            setTimeout(() => {
              message.success('Hubspot integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
          });
      },
      onCancel: () => {}
    });
  };
  const handleRedirectToURL = () => {
    fetchSalesforceRedirectURL(activeProject.id).then((r) => {
      if (r.status == 307) {
        window.location = r.data.redirectURL;
      }
    });
  };
  const onClickEnableSalesforce = () => {
    return new Promise((resolve, reject) => {
      //Factors INTEGRATION tracking
      factorsai.track('INTEGRATION', {
        name: 'salesforce',
        activeProjectID: activeProject.id
      });

      enableSalesforceIntegration(activeProject.id)
        .then((r) => {
          sendSlackNotification(
            currentAgent.email,
            activeProject.name,
            'Salesforce'
          );
          if (r.status == 304) {
            handleRedirectToURL();
          }
        })
        .catch((error) => {
          message.error(error);
          reject();
        });
    });
  };

  const onDisconnectSalesForce = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        udpateProjectSettings(activeProject.id, {
          int_salesforce_enabled_agent_uuid: ''
        })
          .then(() => {
            setTimeout(() => {
              message.success('Salesforce integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
          });
      },
      onCancel: () => {}
    });
  };

  const completeUserOnboard = () => {
    setIsLoadingDone(true);
    if (is_onboarding_completed === true) {
      setTimeout(() => {
        setIsLoadingDone(false);
        history.push('/');
      }, 500);
      return;
    }

    udpateProjectSettings(activeProject.id, {
      is_onboarding_completed: true
    })
      .then(() => {
        history.push('/');
        setIsLoadingDone(false);
      })
      .catch((e) => {
        message.error(e);
        setIsLoadingDone(false);
      });
  };

  return (
    <div className={styles['onBoardContainer']}>
      <Alert
        className={styles['notification']}
        style={{ borderRadius: '5px' }}
        description={
          <div
            style={{
              display: 'flex',
              width: '100%',
              justifyContent: 'space-between'
            }}
          >
            <div>
              <Text
                type={'title'}
                level={6}
                weight={'bold'}
                style={{ margin: 0 }}
              >
                Necessary Integrations completed
              </Text>
              Awesome! All necessary integrations are now complete. You can
              integrate additional applications below or get started with your
              first Dashboard.
            </div>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <Button
                style={{ border: '1px solid #E5E5E5' }}
                onClick={completeUserOnboard}
              >
                {isLoadingDone === true ? <LoadingOutlined /> : ''}Go to
                Dashboard
              </Button>
            </div>
          </div>
        }
        icon={'🎉'}
        showIcon
      />
      {/* <SixSignal setIsActive={() => {}} kbLink={true} /> */}
      <div style={{ padding: '30px 0 20px 0' }}>
        <Text type={'title'} level={6} weight={'bold'}>
          Additional Integrations{' '}
          <span style={{ color: 'rgba(0, 0, 0, 0.45)' }}>(Optional)</span>
        </Text>{' '}
      </div>
      <HorizontalCard
        title={'Slack'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={<SVG name={'Slack'} size={40} extraClass={'inline mr-2 -mt-2'} />}
        is_connected={int_slack}
        onClickConnect={int_slack ? onDisconnectSlack : onConnectSlack}
      />
      <Divider style={{ margin: '5px 0' }} />
      <HorizontalCard
        title={'Hubspot'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={
          <SVG
            name={'Hubspot_ads'}
            size={40}
            extraClass={'inline mr-2 -mt-2'}
          />
        }
        is_connected={int_hubspot}
        onClickConnect={
          int_hubspot ? onDisconnectHubspot : onClickEnableHubspot
        }
      />
      <Divider style={{ margin: '5px 0' }} />
      <HorizontalCard
        title={'Salesforce'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={
          <SVG
            name={'Salesforce_ads'}
            size={40}
            extraClass={'inline mr-2 -mt-2'}
          />
        }
        is_connected={int_salesforce_enabled_agent_uuid}
        onClickConnect={
          int_salesforce_enabled_agent_uuid
            ? onDisconnectSalesForce
            : onClickEnableSalesforce
        }
      />
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, {
  enableSlackIntegration,
  enableHubspotIntegration,
  enableSalesforceIntegration,
  fetchSalesforceRedirectURL,
  udpateProjectSettings,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  disableSlackIntegration
})(OnBoard3);
