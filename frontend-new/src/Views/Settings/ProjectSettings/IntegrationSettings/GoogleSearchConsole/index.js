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
  enableSearchConsoleIntegration,
  fetchSearchConsoleCustomerAccounts,
  udpateProjectSettings,
  fetchProjectSettings,
  deleteIntegration
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';
import { getBackendHost } from '../util';
// const GSC_REDIRECT_URI = "/adwords/auth/redirect";
const GSC_REDIRECT_URI_NEW = '/google_organic/v1/auth/redirect';

const GoogleSearchConsole = ({
  activeProject,
  agent_details,
  currentProjectSettings,
  enableSearchConsoleIntegration,
  fetchSearchConsoleCustomerAccounts,
  udpateProjectSettings,
  fetchProjectSettings,
  deleteIntegration,
  integrationCallback
}) => {
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [addNewAccount, setAddNewAccount] = useState(false);
  const [customerAccountsLoaded, setCustomerAccountsLoaded] = useState(false);
  const [customerAccounts, setCustomerAccounts] = useState(false);
  const [selectedGSCAccounts, setSelectedGSCAccounts] = useState([]);
  const [manualAccounts, setManualAccounts] = useState([]);
  const [accountId, setAccountId] = useState(null);
  const [showManageBtn, setShowManageBtn] = useState(true);
  const [showURLModal, setShowURLModal] = useState(false);

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        setCustomerAccounts(false);
        deleteIntegration(activeProject.id, 'google_organic')
          .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setShowModal(false);
            setShowURLModal(false);
            setTimeout(() => {
              message.success('Google integration disconnected!');
            }, 500);
            integrationCallback();
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setShowModal(false);
            setShowURLModal(false);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const isGSCEnabled = () =>
    currentProjectSettings &&
    currentProjectSettings.int_google_organic_enabled_agent_uuid &&
    currentProjectSettings.int_google_organic_enabled_agent_uuid != '';

  const getRedirectURL = () => {
    const params = {
      method: 'GET',
      credentials: 'include'
    };
    const host = getBackendHost();
    const url = `${host + GSC_REDIRECT_URI_NEW}?pid=${activeProject?.id}&aid=${
      agent_details?.uuid
    }`;
    fetch(url, params)
      .then((response) => response.json())
      .then((response) => {
        if (response?.url) {
          window.location = response.url;
        }
      })
      .catch((err) => false);
  };

  const renderSettingInfo = () => {
    const isCustomerAccountChosen =
      currentProjectSettings.int_google_organic_url_prefixes &&
      currentProjectSettings.int_google_organic_url_prefixes != '' &&
      !addNewAccount;

    // get all GSC account when no account is chosen and not account list not loaded.
    // if (isGSCEnabled() && !isCustomerAccountChosen && !customerAccountsLoaded) {
    if (isGSCEnabled() && !customerAccountsLoaded) {
      // setLoadingData(true);
      fetchSearchConsoleCustomerAccounts({ project_id: activeProject.id })
        .then((data) => {
          console.log('fetchSearchConsoleCustomerAccounts', data);
          setCustomerAccountsLoaded(true);
          setCustomerAccounts(data?.urls);
          // setLoadingData(false);
        })
        .catch((error) => {
          console.log('fetchSearchConsoleCustomerAccounts error-->', error);
          message.error('Error while fetching URLs.');
        });
    }
  };

  const enableGSC = () => {
    setLoading(true);
    enableSearchConsoleIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 304) {
          getRedirectURL();
        }
        if (r.status == 200) {
          renderSettingInfo();
          fetchProjectSettings(activeProject.id);
          sendSlackNotification(
            agent_details.email,
            activeProject.name,
            'Google Search Console'
          );
        }
        if (r.status >= 400) {
          setShowManageBtn(true);
          setCustomerAccountsLoaded(false);
          message.error(
            'Oops! We noticed an error whilst trying to fetch your Google Ads account. Please try again.'
          );
        }
      })
      .catch((err) => {
        setLoading(false);
        console.log('Google Ads error-->', err);
      });
  };

  const onAccountSelect = (e) => {
    const selectedGSCAcc = [...selectedGSCAccounts];
    if (e.target.checked) {
      selectedGSCAcc.push(e.target.value);
    } else {
      const index = selectedGSCAcc.indexOf(e.target.value);
      if (index > -1) selectedGSCAcc.splice(index, 1);
    }
    setSelectedGSCAccounts(selectedGSCAcc);
  };

  const addManualAccount = () => {
    const accounts = [...manualAccounts];
    if (accountId != '') {
      accounts.push({
        customer_id: accountId
      });
    }
    setManualAccounts(accounts);
    setShowModal(false);
  };

  const onClickFinishSetup = () => {
    const selectedGSCAcc = selectedGSCAccounts.join(', ');

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'google_organic',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      int_google_organic_url_prefixes: selectedGSCAcc
    }).then(() => {
      setAddNewAccount(false);
      setSelectedGSCAccounts([]);
      message.success('Search Console Accounts updated!');
      setShowManageBtn(true);
      setCustomerAccountsLoaded(false);
    });
  };

  const renderAccountsList = () => {
    const accountRows = [];

    if (!customerAccounts) return;

    for (let i = 0; i < customerAccounts.length; i++) {
      const account = customerAccounts[i];

      accountRows.push(
        <div className='flex mt-2'>
          <Checkbox value={account} onChange={onAccountSelect}>
            {account}
          </Checkbox>
        </div>
      );
    }
    for (let i = 0; i < manualAccounts.length; i++) {
      const account = manualAccounts[i];

      accountRows.push(
        <div className='flex mt-2'>
          <Checkbox value={account} onChange={onAccountSelect}>
            {account}
          </Checkbox>
        </div>
      );
    }

    return (
      <Modal
        visible={showURLModal}
        zIndex={10}
        width={600}
        afterClose={() => setShowURLModal(false)}
        className='fa-modal--regular fa-modal--slideInDown'
        centered
        footer={null}
        closable={false}
        transitionName=''
        maskTransitionName=''
      >
        <div className='flex flex-col w-full p-2'>
          <Text type='title' level={3} weight='bold' extraClass='my-2 pb-2'>
            Google Search Console
          </Text>
          <Text type='title' level={6} weight='bold' extraClass='my-2'>
            Add/Remove URL(s):
          </Text>
          <div className='p-2'>
            <Text
              type='title'
              level={7}
              color='grey'
              weight='bold'
              extraClass='m-0'
            >
              URL(s):
            </Text>
            {accountRows}
          </div>
          <div className='mt-6 flex justify-end'>
            {/* <Button onClick={() => setShowModal(true)}>
              {' '}
              Enter Id Manually{' '}
            </Button> */}
            <Button
              type='primary'
              className='ml-2'
              onClick={onClickFinishSetup}
            >
              {' '}
              Finish Setup{' '}
            </Button>
          </div>
        </div>
      </Modal>
    );
  };

  // const isCustomerAccountSelected = () => {
  //     return currentProjectSettings && currentProjectSettings.int_google_organic_url_prefixes && !addNewAccount;
  // };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Google Search Console integrations' />
      }
      onError={FaErrorLog}
    >
      {currentProjectSettings?.int_google_organic_url_prefixes &&
        currentProjectSettings?.int_google_organic_url_prefixes != '' && (
          <div className='mt-4 flex flex-col  w-full'>
            <Text
              type='title'
              level={6}
              weight='bold'
              color='character-primary'
              extraClass='m-0'
            >
              Connected URL(s)
            </Text>
            <div
              className={`mt-4 p-4 relative ${
                showManageBtn ? 'cursor-pointer' : ''
              } `}
              style={{ background: '#fafafa', borderRadius: 12 }}
              onClick={() => {
                if (showManageBtn) {
                  renderSettingInfo();
                  setShowURLModal(true);
                  setShowManageBtn(false);
                }
              }}
            >
              {currentProjectSettings?.int_google_organic_url_prefixes
                ?.split(',')
                .map((id, i) => (
                  <div
                    className={`flex gap-6 items-center ${
                      i !== 0 ? 'mt-2' : ''
                    } `}
                  >
                    <SVG name='Greentick' size='14' />
                    <Text
                      type='title'
                      level={7}
                      extraClass='m-0'
                      color='chracter-secondary'
                    >
                      {id}
                    </Text>
                  </div>
                ))}
              <div className='absolute' style={{ top: 16, right: 16 }}>
                <SVG name='Edit' size={16} color='#00000073' />
              </div>
            </div>

            {/* <Input
              disabled
              value={currentProjectSettings?.int_google_organic_url_prefixes}
              style={{ width: '320px', marginTop: 8, background: '#fff' }}
            /> */}
          </div>
        )}
      {isGSCEnabled() &&
        showManageBtn &&
        !currentProjectSettings?.int_google_organic_url_prefixes && (
          <div className='mt-4'>
            <Button
              type='primary'
              onClick={() => {
                renderSettingInfo();
                setShowURLModal(true);
                setShowManageBtn(false);
              }}
            >
              Connect URL(s)
            </Button>
          </div>
        )}

      <div className='w-full'>
        {!showManageBtn && !customerAccountsLoaded && <Skeleton />}
      </div>
      <div>{customerAccountsLoaded && renderAccountsList()}</div>

      <div className='mt-4 flex'>
        {!currentProjectSettings?.int_google_organic_enabled_agent_uuid ? (
          <Button
            className='mr-2'
            type='primary'
            loading={loading}
            onClick={enableGSC}
          >
            Enable using Google
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

      <Modal
        visible={showModal}
        zIndex={11}
        afterClose={() => setShowModal(false)}
        className='fa-modal--regular fa-modal--slideInDown'
        centered
        footer={null}
        transitionName=''
        maskTransitionName=''
        closable={false}
      >
        <Row>
          <Col span={24}>
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              Manually add Google Search Console account
            </Text>
          </Col>
        </Row>
        <Row className='mt-4'>
          <Col span={24}>
            <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
              Enter Search Console account ID:
            </Text>
            <Input
              type='text'
              className='mt-2'
              onChange={(e) => setAccountId(e.target.value)}
            />
          </Col>
        </Row>
        <Row className='mt-4'>
          <Col span={24}>
            <div className='flex justify-end'>
              <Button onClick={() => setShowModal(false)}> Cancel </Button>
              <Button
                className='ml-2'
                type='primary'
                onClick={addManualAccount}
              >
                {' '}
                Submit{' '}
              </Button>
            </div>
          </Col>
        </Row>
      </Modal>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  agent_details: state.agent.agent_details,
  currentProjectSettings: state.global.currentProjectSettings
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  enableSearchConsoleIntegration,
  fetchSearchConsoleCustomerAccounts,
  udpateProjectSettings,
  deleteIntegration
})(GoogleSearchConsole);
