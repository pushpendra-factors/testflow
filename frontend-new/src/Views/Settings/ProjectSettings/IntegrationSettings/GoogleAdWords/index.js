import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
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
  Spin
} from 'antd';
import NumberFormat from 'react-number-format';

import {
  enableAdwordsIntegration,
  fetchAdwordsCustomerAccounts,
  udpateProjectSettings,
  fetchProjectSettings,
  deleteIntegration
} from 'Reducers/global';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import logger from 'Utils/logger';
import {
  ADWORDS_REDIRECT_URI_NEW,
  ADWORDS_INTERNAL_REDIRECT_URI,
  INTEGRATION_HOME_PAGE,
  getBackendHost
} from '../util';
import { sendSlackNotification } from '../../../../../utils/slack';

const GoogleIntegration = ({
  activeProject,
  agent_details,
  currentProjectSettings,
  enableAdwordsIntegration,
  fetchAdwordsCustomerAccounts,
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
  const [selectedAdwordsAccounts, setSelectedAdwordsAccounts] = useState([]);
  const [manualAccounts, setManualAccounts] = useState([]);
  const [accountId, setAccountId] = useState(null);
  const [showManageBtn, setShowManageBtn] = useState(true);
  const [showURLModal, setShowURLModal] = useState(false);
  const [managerIDArr, SetManagerIDArr] = useState({});
  const history = useHistory();

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
        deleteIntegration(activeProject.id, 'adwords')
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

  const isIntAdwordsEnabled = () =>
    currentProjectSettings &&
    currentProjectSettings.int_adwords_enabled_agent_uuid &&
    currentProjectSettings.int_adwords_enabled_agent_uuid != '';

  const getRedirectURL = () => {
    const params = {
      method: 'GET',
      credentials: 'include'
    };
    const host = getBackendHost();
    const url = `${host + ADWORDS_REDIRECT_URI_NEW}?pid=${
      activeProject?.id
    }&aid=${agent_details?.uuid}`;
    fetch(url, params)
      .then((response) => response.json())
      .then((response) => {
        if (response?.url) {
          window.location = response.url;
        }
      })
      .catch((err) => {
        logger.error('Error!', err);
      });
  };

  const renderSettingInfo = () => {
    const isCustomerAccountChosen =
      currentProjectSettings.int_adwords_customer_account_id &&
      currentProjectSettings.int_adwords_customer_account_id != '' &&
      !addNewAccount;

    // get all adwords account when no account is chosen and not account list not loaded.
    // if (isIntAdwordsEnabled() && !isCustomerAccountChosen && !customerAccountsLoaded) {
    if (isIntAdwordsEnabled() && !customerAccountsLoaded) {
      setLoadingData(true);
      fetchAdwordsCustomerAccounts({ project_id: activeProject.id })
        .then((data) => {
          setCustomerAccountsLoaded(true);
          setCustomerAccounts(data?.customer_accounts);
          setLoadingData(false);
        })
        .catch((error) => {
          message.error('Error while fetching Google Ads Customer Accounts.');
          setLoadingData(false);
        });
    }
  };

  useEffect(() => {
    if (isIntAdwordsEnabled()) {
      if (window.location.href.indexOf(ADWORDS_INTERNAL_REDIRECT_URI) > -1) {
        renderSettingInfo();
        setShowURLModal(true);
      }
    }
  }, [activeProject, agent_details, currentProjectSettings]);

  const enableAdwords = () => {
    setLoading(true);
    enableAdwordsIntegration(activeProject.id)
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
            'Google Adwords'
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

  const onManagerIDSelect = (Id, e) => {
    const validatedManagerID = e.target.value.replace(/-/g, '');
    SetManagerIDArr({
      ...managerIDArr,
      [Id]: validatedManagerID
    });
  };
  const onAccountSelect = (e) => {
    const selectedAdwordsAcc = [...selectedAdwordsAccounts];
    if (e.target.checked) {
      selectedAdwordsAcc.push(e.target.value);
    } else {
      const index = selectedAdwordsAcc.indexOf(e.target.value);
      if (index > -1) selectedAdwordsAcc.splice(index, 1);
    }
    setSelectedAdwordsAccounts(selectedAdwordsAcc);
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
    const selectedAdwordsAcc = selectedAdwordsAccounts.join(', ');

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'adwords',
      activeProjectID: activeProject.id
    });

    const mappedAccounts = selectedAdwordsAccounts.reduce(
      (a, v) => ({ ...a, [v]: managerIDArr[v] ? managerIDArr[v] : '' }),
      {}
    );

    const accountsData = {
      int_adwords_customer_account_id: selectedAdwordsAcc,
      int_adwords_client_manager_map: mappedAccounts
    };

    udpateProjectSettings(activeProject.id, accountsData).then(() => {
      setAddNewAccount(false);
      setSelectedAdwordsAccounts([]);
      setShowURLModal(false);
      setCustomerAccounts([]);
      setCustomerAccountsLoaded(false);
      message.success('Adwords Accounts updated!');
      setShowManageBtn(true);
    });
  };

  const renderAccountsList = () => {
    const accountRows = [];

    if (!customerAccounts) return;

    for (let i = 0; i < customerAccounts.length; i++) {
      const account = customerAccounts[i];

      accountRows.push(
        <tr style={{ 'border-bottom': '1px solid #eee' }}>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            <Checkbox value={account.customer_id} onChange={onAccountSelect} />
          </td>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            {account.customer_id}
          </td>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            {account.descriptiveName ? account.descriptiveName : '-'}
          </td>
          <td
            style={{ border: 'none', paddingTop: '5px', paddingBottom: '5px' }}
          >
            {account.manager_id ? account.manager_id : '-'}
          </td>
        </tr>
      );
    }
    for (let i = 0; i < manualAccounts.length; i++) {
      const account = manualAccounts[i];
      accountRows.push(
        <tr>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            <Checkbox value={account.customer_id} onChange={onAccountSelect} />
          </td>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            {account.customer_id}
          </td>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            {account.name ? account.name : '-'}
          </td>
          <td style={{ border: 'none', paddingTop: '5px' }}>
            <Input
              size='small'
              style={{ width: '180px' }}
              onChange={(e) => onManagerIDSelect(account.customer_id, e)}
            />
          </td>
        </tr>
      );
    }

    return accountRows;
  };

  // const isCustomerAccountSelected = () => {
  //     return currentProjectSettings && currentProjectSettings.int_adwords_customer_account_id && !addNewAccount;
  // };

  useEffect(() => {
    const mapManagerAccount = {};
    if (customerAccounts) {
      customerAccounts?.map(
        (account) =>
          (mapManagerAccount[account.customer_id] = account.manager_id)
      );

      SetManagerIDArr({
        ...managerIDArr,
        ...mapManagerAccount
      });
    }
  }, [customerAccounts]);

  const closeCustomerManagerIDModal = () => {
    setShowURLModal(false);
    setCustomerAccounts([]);
    setCustomerAccountsLoaded(false);
    // history.push(INTEGRATION_HOME_PAGE);
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with GoogleAdWords integrations' />
      }
      onError={FaErrorLog}
    >
      <Modal
        visible={showURLModal}
        zIndex={10}
        width={600}
        afterClose={() => closeCustomerManagerIDModal()}
        className='fa-modal--regular fa-modal--slideInDown'
        centered
        footer={null}
        onCancel={() => closeCustomerManagerIDModal()}
        transitionName=''
        maskTransitionName=''
      >
        <div className='flex flex-col w-full p-2'>
          <Text type='title' level={3} weight='bold' extraClass='my-2 pb-2'>
            Google Ads
          </Text>
          <Text type='title' level={6} weight='bold' extraClass='my-2'>
            Add/Remove Accounts
          </Text>
          <table>
            <thead>
              <tr>
                <td style={{ border: 'none', padding: '5px' }} />
                <td style={{ border: 'none', padding: '5px' }}>
                  <Text type='title' level={7} color='grey' extraClass='m-0'>
                    Customer Id
                  </Text>
                </td>
                <td style={{ border: 'none', padding: '5px' }}>
                  <Text type='title' level={7} color='grey' extraClass='m-0'>
                    Customer Name
                  </Text>
                </td>
                <td style={{ border: 'none', padding: '5px' }}>
                  <Text type='title' level={7} color='grey' extraClass='m-0'>
                    Manager Id (if applicable)
                  </Text>
                </td>
              </tr>
            </thead>
            {customerAccountsLoaded ? (
              <tbody>{renderAccountsList()}</tbody>
            ) : (
              <div className='p-4'>
                <Spin />
              </div>
            )}
          </table>
          <div className='mt-6 flex justify-end'>
            <Button
              onClick={() => setShowModal(true)}
              disabled={!customerAccountsLoaded}
            >
              {' '}
              Enter Id Manually{' '}
            </Button>
            <Button
              type='primary'
              disabled={!customerAccountsLoaded}
              className='ml-2'
              onClick={onClickFinishSetup}
            >
              {' '}
              Finish Setup{' '}
            </Button>
          </div>
        </div>
      </Modal>

      {currentProjectSettings?.int_adwords_customer_account_id && (
        <div className='mt-4 flex flex-col  w-full '>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='character-primary'
            extraClass='m-0'
          >
            Connected Accounts
          </Text>
          <Text
            type='title'
            level={7}
            color='character-secondary'
            extraClass='m-0 mt-1'
          >
            Synced Adwords account details
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
              }
            }}
          >
            {currentProjectSettings?.int_adwords_customer_account_id
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
                    <NumberFormat
                      displayType='text'
                      thousandSeparator='-'
                      value={id}
                      format='### ### ####'
                    />
                  </Text>
                  {currentProjectSettings?.int_adwords_client_manager_map &&
                    currentProjectSettings?.int_adwords_client_manager_map?.[
                      id
                    ] && (
                      <Text
                        type='title'
                        level={7}
                        extraClass='m-0 '
                        color='chracter-secondary'
                      >
                        {
                          currentProjectSettings
                            ?.int_adwords_client_manager_map[id]
                        }
                      </Text>
                    )}
                </div>
              ))}
            <div className='absolute' style={{ top: 16, right: 16 }}>
              <SVG name='Edit' size={16} color='#00000073' />
            </div>
          </div>

          {/* <Input
                  size='large'
                  disabled={true}
                  value={
                    currentProjectSettings?.int_adwords_customer_account_id
                  }
                  style={{ width: '400px' }}
                /> */}
        </div>
      )}
      {isIntAdwordsEnabled() &&
        !currentProjectSettings?.int_adwords_customer_account_id && (
          <div className='mt-4'>
            <Button
              type='primary'
              loading={loading}
              onClick={() => {
                renderSettingInfo();
                setShowURLModal(true);
                // setShowManageBtn(false);
              }}
            >
              Connect Account(s)
            </Button>
          </div>
        )}

      <div className='w-full'>
        {!showManageBtn && !customerAccountsLoaded && <Skeleton />}
      </div>
      {/* <div>{customerAccountsLoaded && renderAccountsList()}</div> */}

      <div className='mt-4 flex'>
        {!currentProjectSettings?.int_adwords_enabled_agent_uuid ? (
          <Button
            className='mr-2'
            type='primary'
            loading={loading}
            onClick={enableAdwords}
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
        size='large'
      >
        <Row>
          <Col span={24}>
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              Manually add Google Adwords account
            </Text>
          </Col>
        </Row>
        <Row className='mt-4'>
          <Col span={24}>
            <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
              Enter adwords account ID:
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
  enableAdwordsIntegration,
  fetchAdwordsCustomerAccounts,
  udpateProjectSettings,
  deleteIntegration
})(GoogleIntegration);
