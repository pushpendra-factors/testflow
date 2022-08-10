import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings, enableHubspotIntegration } from 'Reducers/global';
import {
  Row, Col, Modal, Input, Form, Button, notification, message
} from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary'
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';


const HubspotIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent,
  enableHubspotIntegration
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);

  const isHubspotEnabled = () => {
    return currentProjectSettings && currentProjectSettings?.int_hubspot && currentProjectSettings?.int_hubspot_refresh_token != "";
  }


  useEffect(() => {
    setIsActive(isHubspotEnabled());
  }, []);

  const onClickEnableHubspot = () => {
    setLoading(true);

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', { 'name': 'hubspot', 'activeProjectID': activeProject.id });

    enableHubspotIntegration(activeProject.id).then((r) => {
      setLoading(false);
      sendSlackNotification(currentAgent.email, activeProject.name, 'Hubspot');
      if (r.status == 307) {
        window.location = r.data.redirect_url;
      }
    }).catch((err) => {
      setLoading(false);
      message.error(`${err?.data?.error}`);
      setIsActive(false);
    });
  };

  const onDisconnect = () => {
    setLoading(true);
    udpateProjectSettings(activeProject.id,
      {
        'int_hubspot_api_key': '',
        'int_hubspot': false
      }).then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Hubspot integration disconnected!');
        }, 500);
        setIsActive(false);
      }).catch((err) => {
        message.error(`${err?.data?.error}`);
        setLoading(false);
      });
  }


  const isEnabled = isHubspotEnabled();

  return (
    <>
      <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with Hubspot integrations'} />} onError={FaErrorLog}>

        {
          isEnabled && <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Connected Account</Text>
            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>API Key</Text>
            <Input size="large" disabled={true} placeholder="API Key" value={currentProjectSettings?.int_hubspot_api_key} style={{ width: '400px' }} />
          </div>
        }
        <div className={'mt-4 flex'} data-tour='step-11'>
          {isEnabled ? <Button loading={loading} onClick={() => onDisconnect()}>Disconnect</Button> : <Button type={'primary'} loading={loading} onClick={onClickEnableHubspot}>Enable using Hubspot</Button>
          }
          {kbLink && <a className={'ant-btn ml-2 '} target={"_blank"} href={kbLink}>View documentation</a>}
        </div>
      </ErrorBoundary>
    </>
  )
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details,
});

export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings, enableHubspotIntegration })(HubspotIntegration)