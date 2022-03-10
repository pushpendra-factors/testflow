import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Button,
  message,
  Select,
  Modal,
  Row,
  Col,
  Input,
  Checkbox,
  Skeleton,
} from 'antd';
import {
  enableBingAdsIntegration,
  createBingAdsIntegration,
  fetchBingAdsIntegration,
  disableBingAdsIntegration,
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';

const BingIntegration = ({
  activeProject,
  agent_details,
  enableBingAdsIntegration,
  setIsStatus,
  createBingAdsIntegration,
  kbLink = false,
  fetchBingAdsIntegration,
  disableBingAdsIntegration,
  bingAds
}) => {
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState(null);

  const onDisconnect = () => {
    setLoading(true);
    disableBingAdsIntegration(activeProject.id)
      .then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Bing Ads integration disconnected!');
        }, 500);
        setIsStatus('');
      })
      .catch((err) => {
        message.error(`${err?.data?.error}`);
        setLoading(false);
        console.log('disconnect failed-->', err);
      });
  };

  const isBingAdsEnabled = () => {
    fetchBingAdsIntegration(activeProject.id);
  };

  useEffect(() => {
    isBingAdsEnabled();
    if (bingAds.status) {
      bingAds.accounts == '' ? setIsStatus('Pending') : setIsStatus('Active');
      setAccounts(bingAds.accounts);
    } else {
      setIsStatus('');
    }
  }, [activeProject, agent_details, bingAds?.status]);

  const enableBingAds = () => {
    setLoading(true);
    createBingAdsIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 200) {
          let hostname = window.location.hostname
          let protocol = window.location.protocol
          let port = window.location.port
          let redirectURL = protocol + "//" + hostname + ":" + port + "?bingadsint=" + activeProject.id
          if (port === undefined || port === '') {
            redirectURL = protocol + "//" + hostname + "?bingadsint=" + activeProject.id
          }
          let url = new URL(r.data.redirect_uri);
          url.searchParams.set('redirect_uri', redirectURL)
          window.location = url.href;
        }
        if (r.status >= 400) {
          message.error('Error fetching Bing Ads accounts');
        }
      })
      .catch((err) => {
        setLoading(false);
        console.log('Bing Ads error-->', err);
        setIsStatus('');
      });
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            subtitle={'Facing issues with Bing Ads integrations'}
          />
        }
        onError={FaErrorLog}
      >
        <div className={'mt-4 flex w-full'}>
          {bingAds.status && (
            <>
              <div
                className={
                  'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'
                }
              >
                <Text
                  type={'title'}
                  level={6}
                  weight={'bold'}
                  extraClass={'m-0'}
                >
                  Connected Account
                </Text>
                <Text
                  type={'title'}
                  level={7}
                  color={'grey'}
                  extraClass={'m-0 mt-2'}
                >
                  Bing Ads sync account details
                </Text>
                {accounts == "" ? 
                <Text
                type={'title'}
                size={10}
                color={'red'}
                extraClass={'m-0 mt-2'}
                >
                  No ads account found or partial integration. Please disconnect and try again.
                </Text>
                :
                <Input
                  size='large'
                  disabled={true}
                  value={
                    accounts
                  }
                  style={{ width: '400px' }}
                />
                }
              </div>
            </>
          )}
        </div>

        <div className={'mt-4 flex'}>
          {!bingAds.status ? 
            <Button
              className={'mr-2'}
              type={'primary'}
              loading={loading}
              onClick={enableBingAds}
            >
              Connect Now
            </Button>
              :
            <Button
              className={'mr-2'}
              loading={loading}
              onClick={() => onDisconnect()}
            >
              Disconnect
            </Button>
          }
          {kbLink && (
            <a className={'ant-btn'} target={'_blank'} href={kbLink}>
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
  agent_details: state.agent.agent_details,
  bingAds: state.global.bingAds,
});

export default connect(mapStateToProps, {
  enableBingAdsIntegration,
  createBingAdsIntegration,
  fetchBingAdsIntegration,
  disableBingAdsIntegration
})(BingIntegration);
