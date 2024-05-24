import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import {
  fetchProjectSettings,
  udpateProjectSettings,
  addLinkedinAccessToken,
  deleteIntegration
} from 'Reducers/global';
import { Button, message, Select, Modal, Row, Col, Input } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import _ from 'lodash';
import { sendSlackNotification } from '../../../../../utils/slack';
import {
  linkedInScope_rw_ads,
  linkedInScope_rw_conversions
} from './constants';
import { getBackendHost } from '../util';

const LinkedInIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  addLinkedinAccessToken,
  deleteIntegration,
  currentAgent,
  integrationCallback
}) => {
  const [loading, setLoading] = useState(false);
  const [FbResponse, SetFbResponse] = useState(null);
  const [adAccounts, SetAdAccounts] = useState(null);
  const [SelectedAdAccount, SetSelectedAdAccount] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [oauthResponse, setOauthResponse] = useState(false);

  const getAdAccounts = (jsonRes) => {
    const accounts = jsonRes.map((res) => ({ value: res.id, name: res.name }));
    return accounts;
  };

  useEffect(() => {
    const code = localStorage.getItem('Linkedin_code');
    const state = localStorage.getItem('Linkedin_state');
    if (code != '' && state === 'factors') {
      const url = `${getBackendHost()}/integrations/linkedin/auth`;
      fetch(url, {
        method: 'POST',
        body: JSON.stringify({
          code
        })
      })
        .then((response) => {
          if (!response.ok) {
            throw Error;
          }
          return response;
        })
        .then((response) => {
          if (response.status < 400) {
            response.json().then((e) => {
              setOauthResponse(e);
              fetch(`${getHostURL()}/integrations/linkedin/ad_accounts`, {
                method: 'POST',
                body: JSON.stringify({
                  access_token: e?.access_token
                })
              })
                .then((response) => {
                  if (!response.ok) {
                    throw Error;
                  }
                  return response;
                })
                .then((response) => {
                  response.json().then((res) => {
                    const jsonRes = JSON.parse(res);
                    const adAccountsNew = getAdAccounts(jsonRes.elements);
                    SetAdAccounts(adAccountsNew);
                    localStorage.removeItem('Linkedin_code');
                    localStorage.removeItem('Linkedin_state');
                    setShowForm(true);
                  });
                })
                .catch((err) => {
                  message.error('Failed to fetch linkedin/ad_accounts');
                });
            });
          } else {
            console.log('Failed to fetch linkedin/ad_accounts!!');
          }
        })
        .catch((err) => {
          message.error('Failed to fetch linkedin/auth');
        });
    }
  }, []);

  const makeSelectOpt = (value, label) => {
    if (!label) label = value;
    return { value, label: `${label} (${value})` };
  };

  const createSelectOpts = (opts) => {
    const ropts = [];
    for (const k in opts) ropts.push(makeSelectOpt(k, opts[k]));
    return ropts;
  };

  const renderLinkedinLogin = () => {
    if (!currentProjectSettings?.int_linkedin_access_token) {
      const { hostname } = window.location;
      const { protocol } = window.location;
      const { port } = window.location;
      let redirect_uri = `${protocol}//${hostname}:${port}`;
      if (port === undefined || port === '') {
        redirect_uri = `${protocol}//${hostname}`;
      }
      // linkedIn scope check for accounts (defined in ./constants file).
      const scope_rw_ads = linkedInScope_rw_ads.includes(activeProject?.id);
      const scope_rw_conversions = linkedInScope_rw_conversions.includes(
        activeProject?.id
      );
      // linkedIn oauth url gets updated based on the scopes assigned to project.
      const href = `https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=${
        BUILD_CONFIG.linkedin_client_id
      }&redirect_uri=${redirect_uri}&state=factors&scope=r_basicprofile%20r_liteprofile%20r_ads_reporting${
        scope_rw_ads ? '%20rw_ads' : '%20r_ads'
      }${scope_rw_conversions ? '%20rw_conversions' : ''}`;
      return (
        <a href={href} className='ant-btn ant-btn-primary'>
          Connect Now
        </a>
      );
    }
  };

  const convertToString = (e) => {
    const dataString = _.toString(e);
    SetSelectedAdAccount(dataString);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'linkedin',
      activeProjectID: activeProject.id
    });

    if (SelectedAdAccount != '') {
      const data = {
        int_linkedin_ad_account: SelectedAdAccount,
        int_linkedin_refresh_token: oauthResponse.refresh_token,
        int_linkedin_refresh_token_expiry:
          oauthResponse.refresh_token_expires_in,
        project_id: activeProject.id.toString(),
        int_linkedin_access_token: oauthResponse.access_token,
        int_linkedin_access_token_expiry: oauthResponse.expires_in
      };
      addLinkedinAccessToken(data)
        .then(() => {
          fetchProjectSettings(activeProject.id);
          setShowForm(false);
          message.success('LinkedIn integration enabled!');
          sendSlackNotification(
            currentAgent.email,
            activeProject.name,
            'Linkedin'
          );
        })
        .catch((e) => {
          console.log(e);
          message.error(e);
          setShowForm(false);
        });
    }
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
        deleteIntegration(activeProject.id, 'linkedin')
          .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('LinkedIn integration disconnected!');
            }, 500);
            integrationCallback();
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setShowForm(false);
          });
      },
      onCancel: () => {}
    });
  };

  const getAdAccountsOptSrc = () => {
    const opts = {};
    for (const i in adAccounts) {
      const adAccount = adAccounts[i];
      opts[adAccount.value] = adAccount?.name;
    }
    return opts;
  };

  const formComponent = () => {
    if (!currentProjectSettings.int_linkedin_access_token) {
      if (adAccounts != '' && adAccounts?.length != 0) {
        return (
          <Modal
            visible={showForm}
            zIndex={1020}
            afterClose={() => setShowForm(false)}
            className='fa-modal--regular fa-modal--slideInDown'
            centered
            footer={null}
            transitionName=''
            maskTransitionName=''
            closable={false}
          >
            <div className='p-4'>
              <Row>
                <Col span={24}>
                  <Text type='title' level={6} weight='bold' extraClass='m-0'>
                    Choose your LinkedIn Ad account:
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    color='grey'
                    extraClass='m-0 mt-2'
                  >
                    Choose your LinkedIn Ad account to sync reports with Factors
                    for performance reporting
                  </Text>
                </Col>
              </Row>
              <form onSubmit={(e) => handleSubmit(e)} className='w-full'>
                <Row className='mt-6'>
                  <Col span={24}>
                    <div className='w-full'>
                      <div className='w-full pb-2'>
                        <Select
                          mode='multiple'
                          allowClear
                          className='w-full'
                          placeholder='Select Account'
                          onChange={(e) => convertToString(e)}
                          options={createSelectOpts(getAdAccountsOptSrc())}
                        />
                      </div>
                    </div>
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={24}>
                    <div className='flex justify-end'>
                      <Button
                        className='ant-btn-primary'
                        disabled={!SelectedAdAccount}
                        htmlType='submit'
                      >
                        Select
                      </Button>
                    </div>
                  </Col>
                </Row>
              </form>
            </div>
          </Modal>
        );
      }
      //   if (adAccounts != "" && adAccounts.length == 0) {
      //     return <div>You don't have any ad accounts associated to the id you logged in with.</div>
      //   }
    } else if (
      currentProjectSettings?.int_linkedin_ad_account !== '' ||
      currentProjectSettings?.int_linkedin_ad_account !== undefined
    ) {
      return (
        <div className='mt-4 flex flex-col w-full'>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='character-primary'
            extraClass='m-0'
          >
            Selected linkedin Account
          </Text>
          <Input
            disabled
            value={currentProjectSettings?.int_linkedin_ad_account}
            style={{ width: 320, marginTop: 8, background: '#fff' }}
          />
        </div>
      );
    }
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with LinkedIn integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='flex '>{formComponent()}</div>

      {!adAccounts && (
        <div className='mt-4 flex'>
          {currentProjectSettings?.int_linkedin_ad_account ? (
            <Button loading={loading} onClick={() => onDisconnect()}>
              Disconnect
            </Button>
          ) : (
            renderLinkedinLogin()
          )}
        </div>
      )}
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  addLinkedinAccessToken,
  fetchProjectSettings,
  udpateProjectSettings,
  deleteIntegration
})(LinkedInIntegration);
