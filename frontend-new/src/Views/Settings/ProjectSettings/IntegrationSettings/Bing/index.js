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
  Skeleton
} from 'antd';
import {
  enableBingAdsIntegration,
  createBingAdsIntegration,
  fetchBingAdsIntegration,
  disableBingAdsIntegration
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';

const BingIntegration = ({
  activeProject,
  agent_details,
  integrationCallback,
  enableBingAdsIntegration,
  createBingAdsIntegration,
  fetchBingAdsIntegration,
  disableBingAdsIntegration,
  bingAds
}) => {
  const [loading, setLoading] = useState(false);

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        disableBingAdsIntegration(activeProject.id)
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('Bing Ads integration disconnected!');
            }, 500);
            integrationCallback();
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const enableBingAds = () => {
    setLoading(true);
    createBingAdsIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 200) {
          const { hostname } = window.location;
          const { protocol } = window.location;
          const { port } = window.location;
          let redirectURL = `${protocol}//${hostname}:${port}?bingadsint=${activeProject.id}&email=${agent_details.email}&projectname=${activeProject.name}`;
          if (port === undefined || port === '') {
            redirectURL = `${protocol}//${hostname}?bingadsint=${activeProject.id}&email=${agent_details.email}&projectname=${activeProject.name}`;
          }
          const url = new URL(r.data.redirect_uri);
          url.searchParams.set('redirect_uri', redirectURL);
          window.location = url.href;
          integrationCallback();
        }
        if (r.status >= 400) {
          message.error('Error fetching Bing Ads accounts');
        }
      })
      .catch((err) => {
        setLoading(false);
        console.log('Bing Ads error-->', err);
      });
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Bing Ads integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='flex w-full'>
        {bingAds.status && (
          <div className='flex flex-col w-full mt-4'>
            <Text
              type='title'
              level={7}
              color='character-primary'
              weight='bold'
              extraClass='m-0'
            >
              Selected Bing Account
            </Text>

            {bingAds?.accounts == '' ? (
              <Text type='title' size={10} color='red' extraClass='m-0 mt-2'>
                No ads account found or partial integration. Please disconnect
                and try again.
              </Text>
            ) : (
              <Input
                disabled
                value={bingAds?.accounts}
                style={{ width: 320, marginTop: 8, background: '#fff' }}
              />
            )}
          </div>
        )}
      </div>

      <div className='mt-4 flex'>
        {!bingAds.status ? (
          <Button
            className='mr-2'
            type='primary'
            loading={loading}
            onClick={enableBingAds}
          >
            Connect Now
          </Button>
        ) : (
          <Button
            className='mr-2'
            loading={loading}
            onClick={() => onDisconnect()}
          >
            Disconnect
          </Button>
        )}
      </div>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  agent_details: state.agent.agent_details,
  bingAds: state.global.bingAds
});

export default connect(mapStateToProps, {
  enableBingAdsIntegration,
  createBingAdsIntegration,
  fetchBingAdsIntegration,
  disableBingAdsIntegration
})(BingIntegration);
