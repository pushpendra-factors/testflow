import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  fetchProjectSettings,
  udpateProjectSettings,
  enableSalesforceIntegration,
  fetchSalesforceRedirectURL
} from 'Reducers/global';
import { Input, Button, message, Modal } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const SalesForceIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  enableSalesforceIntegration,
  fetchSalesforceRedirectURL,
  currentAgent
}) => {
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);

  const isSalesforceEnabled = () =>
    currentProjectSettings &&
    currentProjectSettings.int_salesforce_enabled_agent_uuid &&
    currentProjectSettings.int_salesforce_enabled_agent_uuid != '';

  const handleRedirectToURL = () => {
    fetchSalesforceRedirectURL(activeProject.id.toString()).then((r) => {
      if (r.status == 307) {
        window.location = r.data.redirectURL;
      }
    });
  };

  const onClickEnableSalesforce = () => {
    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'salesforce',
      activeProjectID: activeProject.id
    });

    enableSalesforceIntegration(activeProject.id.toString()).then((r) => {
      sendSlackNotification(
        currentAgent.email,
        activeProject.name,
        'Salesforce'
      );
      if (r.status == 304) {
        handleRedirectToURL();
      }
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
          int_salesforce_enabled_agent_uuid: ''
        })
          .then(() => {
            setLoading(false);
            setShowForm(false);
            setTimeout(() => {
              message.success('Salesforce integration disconnected!');
            }, 500);
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

  const isEnabled = isSalesforceEnabled();
  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Salesforce integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='mt-4 flex'>
        {isEnabled && (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disconnect
          </Button>
        )}
        {!isEnabled && (
          <Button
            type='primary'
            loading={loading}
            onClick={onClickEnableSalesforce}
          >
            Connect Salesforce
          </Button>
        )}
      </div>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings,
  enableSalesforceIntegration,
  fetchSalesforceRedirectURL
})(SalesForceIntegration);
