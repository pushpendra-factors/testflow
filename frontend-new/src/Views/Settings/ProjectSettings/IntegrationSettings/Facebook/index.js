import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import {
  fetchProjectSettings,
  udpateProjectSettings,
  addFacebookAccessToken,
  deleteIntegration
} from 'Reducers/global';
import { Button, message, Select, Modal, Row, Col, Input } from 'antd';
import FacebookLogin from 'react-facebook-login';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import _ from 'lodash';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const FacebookIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  addFacebookAccessToken,
  kbLink = false,
  deleteIntegration,
  currentAgent
}) => {
  const [loading, setLoading] = useState(false);
  const [FbResponse, SetFbResponse] = useState(null);
  const [FbAdAccounts, SetFbAdAccounts] = useState(null);
  const [SelectAdAccount, SetSelectAdAccount] = useState(null);
  const [showForm, setShowForm] = useState(false);

  useEffect(() => {
    if (currentProjectSettings?.int_facebook_ad_account) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const makeSelectOpt = (value, label) => {
    if (!label) label = value;
    return { value: value, label: `${label} (${value})` };
  };

  const createSelectOpts = (opts) => {
    let ropts = [];
    for (let k in opts) ropts.push(makeSelectOpt(k, opts[k]));
    return ropts;
  };

  const responseFacebook = (response) => {
    SetFbResponse(response);
    if (response.id != undefined) {
      fetch(
        `https://graph.facebook.com/v17.0/${response.id}/adaccounts?access_token=${response.accessToken}&fields=id,name`
      )
        .then((res) =>
          res.json().then((r) => {
            if (r.data?.length != 0) {
              let adAccounts = r.data.map((account) => {
                return { value: account.id, label: account.name };
              });
              SetFbAdAccounts(adAccounts);
              setShowForm(true);
            } else {
              message.error(
                "You don't have any ad accounts associated to the id you logged in with."
              );
            }
          })
        )
        .catch((err) => console.log('responseFacebook error->>>', err));
    }
  };

  const renderFacebookLogin = () => {
    if (!currentProjectSettings.int_facebook_access_token) {
      return (
        <FacebookLogin
          appId={BUILD_CONFIG.facebook_app_id}
          fields='name,email,picture'
          scope='ads_read,email'
          callback={responseFacebook}
          cssClass='ant-btn ant-btn-primary'
        />
      );
    }
    // else {
    //   return (
    //     <div>Logged In</div>
    //   )
    // }
  };

  const convertToString = (e) => {
    let dataString = _.toString(e);
    SetSelectAdAccount(dataString);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'facebook',
      activeProjectID: activeProject.id
    });

    if (SelectAdAccount != '') {
      const data = {
        int_facebook_user_id: FbResponse.id,
        int_facebook_email: FbResponse.email,
        int_facebook_ad_account: SelectAdAccount,
        project_id: activeProject.id.toString(),
        int_facebook_access_token: FbResponse.accessToken
      };
      addFacebookAccessToken(data)
        .then(() => {
          fetchProjectSettings(activeProject.id);
          setShowForm(false);
          setIsActive(true);
          message.success('Facebok integration enabled!');
          sendSlackNotification(
            currentAgent.email,
            activeProject.name,
            'Facebook'
          );
        })
        .catch((e) => {
          console.log(e);
          message.error(e);
          setShowForm(false);
          setIsActive(false);
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
        deleteIntegration(activeProject.id, 'facebook')
          .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('Facebook integration disconnected!');
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

  const getAdAccountsOptSrc = () => {
    let opts = {};
    for (let i in FbAdAccounts) {
      let adAccount = FbAdAccounts[i];
      opts[adAccount.value] = adAccount.label;
    }
    return opts;
  };

  const formComponent = () => {
    if (!currentProjectSettings?.int_facebook_access_token) {
      if (FbAdAccounts) {
        return (
          <Modal
            visible={showForm}
            zIndex={1020}
            afterClose={() => setShowForm(false)}
            className={'fa-modal--regular fa-modal--slideInDown'}
            centered={true}
            footer={null}
            transitionName=''
            maskTransitionName=''
            closable={false}
          >
            <div className={'p-4'}>
              <Row>
                <Col span={24}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={'m-0'}
                  >
                    Choose your Facebook Ad account:
                  </Text>
                  <Text
                    type={'title'}
                    level={7}
                    color={'grey'}
                    extraClass={'m-0 mt-2'}
                  >
                    Choose your Facebook Ad account to pull in reports from
                    Facebook, Instagram and Facebook Audience Network
                  </Text>
                </Col>
              </Row>
              <form onSubmit={(e) => handleSubmit(e)} className='w-full'>
                <Row className={'mt-6'}>
                  <Col span={24}>
                    <div className='w-full'>
                      <div className='w-full pb-2'>
                        <Select
                          mode='multiple'
                          allowClear
                          className='w-full'
                          placeholder={'Select Account'}
                          onChange={(e) => convertToString(e)}
                          options={createSelectOpts(getAdAccountsOptSrc())}
                        />
                      </div>
                    </div>
                  </Col>
                </Row>
                <Row className={'mt-2'}>
                  <Col span={24}>
                    <div className={'flex justify-end'}>
                      <Button className='ant-btn-primary' htmlType='submit'>
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
      //   else {
      //     return <div>You don't have any ad accounts associated to the id you logged in with.</div>
      //   }
    } else {
      if (
        currentProjectSettings?.int_facebook_ad_account !== '' ||
        currentProjectSettings?.int_facebook_ad_account !== undefined
      ) {
        return (
          <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2'}>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
              Connected Account
            </Text>
            <Text
              type={'title'}
              level={7}
              color={'grey'}
              extraClass={'m-0 mt-2'}
            >
              Selected Facebook Ad Account
            </Text>
            <Input
              size='large'
              disabled={true}
              value={currentProjectSettings?.int_facebook_ad_account}
              style={{ width: '400px' }}
            />
            <Button
              loading={loading}
              className={'mt-4'}
              onClick={() => onDisconnect()}
            >
              Disconnect
            </Button>
          </div>
        );
      }
    }
  };

  // const apiData = [
  //   {value: "act_506913550667906", label: "act_506913550667906"},
  //   {value: "act_300992258107471", label: "act_300992258107471"}
  // ]

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp subtitle={'Facing issues with Facebook integrations'} />
        }
        onError={FaErrorLog}
      >
        <div className={'mt-4 flex w-6/12'}>{formComponent()}</div>

        {!currentProjectSettings?.int_facebook_access_token && (
          <div className={'mt-4 flex'}>
            {renderFacebookLogin()}
            {kbLink && (
              <a className={'ant-btn ml-2 '} target={'_blank'} href={kbLink}>
                View documentation
              </a>
            )}
          </div>
        )}
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
  addFacebookAccessToken,
  fetchProjectSettings,
  udpateProjectSettings,
  deleteIntegration
})(FacebookIntegration);
