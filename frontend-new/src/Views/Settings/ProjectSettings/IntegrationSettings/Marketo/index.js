import React, { useState } from 'react';
import { connect } from 'react-redux';
import { Button, message, Input, Modal } from 'antd';
import {
  enableMarketoIntegration,
  createMarketoIntegration,
  fetchMarketoIntegration,
  disableMarketoIntegration
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';

const MarketoIntegration = ({
  activeProject,
  agent_details,
  enableMarketoIntegration,
  createMarketoIntegration,
  fetchMarketoIntegration,
  disableMarketoIntegration,
  marketo
}) => {
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState(null);

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        disableMarketoIntegration(activeProject.id)
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('Marketo integration disconnected!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
            console.log('disconnect failed-->', err);
          });
      },
      onCancel: () => {}
    });
  };

  const enableMarketo = () => {
    setLoading(true);
    createMarketoIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 200) {
          const { hostname } = window.location;
          const { protocol } = window.location;
          const { port } = window.location;
          let redirectURL = `${protocol}//${hostname}:${port}?marketoInt=${activeProject.id}&email=${agent_details.email}&projectname=${activeProject.name}`;
          if (port === undefined || port === '') {
            redirectURL = `${protocol}//${hostname}?markketoInt=${activeProject.id}&email=${agent_details.email}&projectname=${activeProject.name}`;
          }
          const url = new URL(r.data.redirect_uri);
          url.searchParams.set('redirect_uri', redirectURL);
          window.location = url.href;
        }
        if (r.status >= 400) {
          message.error('Error fetching Marketo accounts');
        }
      })
      .catch((err) => {
        setLoading(false);
        console.log('Marketo error-->', err);
      });
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Marketo integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='mt-4 flex'>
        {!marketo.status ? (
          <Button type='primary' loading={loading} onClick={enableMarketo}>
            Connect Marketo
          </Button>
        ) : (
          <Button loading={loading} onClick={() => onDisconnect()}>
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
  marketo: state.global.marketo
});

export default connect(mapStateToProps, {
  enableMarketoIntegration,
  createMarketoIntegration,
  fetchMarketoIntegration,
  disableMarketoIntegration
})(MarketoIntegration);
