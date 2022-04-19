import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Button,
  message,
  Input,
} from 'antd';
import {
  enableMarketoIntegration,
  createMarketoIntegration,
  fetchMarketoIntegration,
  disableMarketoIntegration,
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';

const MarketoIntegration = ({
  activeProject,
  agent_details,
  enableMarketoIntegration,
  setIsStatus,
  createMarketoIntegration,
  kbLink = false,
  fetchMarketoIntegration,
  disableMarketoIntegration,
  marketo
}) => {
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState(null);

  const onDisconnect = () => {
    setLoading(true);
    disableMarketoIntegration(activeProject.id)
      .then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Marketo integration disconnected!');
        }, 500);
        setIsStatus('');
      })
      .catch((err) => {
        message.error(`${err?.data?.error}`);
        setLoading(false);
        console.log('disconnect failed-->', err);
      });
  };

  const isMarketoEnabled = () => {
    fetchMarketoIntegration(activeProject.id);
  };

  useEffect(() => {
    isMarketoEnabled();
    if (marketo.status) {
      marketo.accounts == '' ? setIsStatus('Pending') : setIsStatus('Active');
      setAccounts(marketo.accounts);
    } else {
      setIsStatus('');
    }
  }, [activeProject, agent_details, marketo?.status]);

  const enableMarketo = () => {
    setLoading(true);
    createMarketoIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 200) {
          let hostname = window.location.hostname
          let protocol = window.location.protocol
          let port = window.location.port
          let redirectURL = protocol + "//" + hostname + ":" + port + "?marketoInt=" + activeProject.id + "&email=" + agent_details.email + "&projectname=" + activeProject.name
          if (port === undefined || port === '') {
            redirectURL = protocol + "//" + hostname + "?markketoInt=" + activeProject.id + "&email=" + agent_details.email + "&projectname=" + activeProject.name
          }
          let url = new URL(r.data.redirect_uri);
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
        setIsStatus('');
      });
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            subtitle={'Facing issues with Marketo integrations'}
          />
        }
        onError={FaErrorLog}
      >
        <div className={'mt-4 flex w-full'}>
          {marketo.status && (
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
                  Marketo sync account details
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
          {!marketo.status ? 
            <Button
              className={'mr-2'}
              type={'primary'}
              loading={loading}
              onClick={enableMarketo}
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
  marketo: state.global.marketo,
});

export default connect(mapStateToProps, {
  enableMarketoIntegration,
  createMarketoIntegration,
  fetchMarketoIntegration,
  disableMarketoIntegration
})(MarketoIntegration);
