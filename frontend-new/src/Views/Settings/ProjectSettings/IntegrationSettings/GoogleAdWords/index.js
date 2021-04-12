import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
    Button, message, Select, Modal, Row, Col, Input, Checkbox, Skeleton
} from 'antd';
const ADWORDS_REDIRECT_URI = "/adwords/v1/auth/redirect"; 
import { enableAdwordsIntegration, fetchAdwordsCustomerAccounts, udpateProjectSettings, fetchProjectSettings } from 'Reducers/global';
const isDevelopment = () => {
    return ENV === "development"
}
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import {ErrorBoundary} from 'react-error-boundary'
const getAdwordsHostURL = () => {
    // return isDevelopment() ? BUILD_CONFIG.adwords_service_host : BUILD_CONFIG.backend_host;
    return BUILD_CONFIG.backend_host;
}


const GoogleIntegration = ({
    activeProject,
    agent_details,
    currentProjectSettings,
    enableAdwordsIntegration,
    setIsActive,
    fetchAdwordsCustomerAccounts,
    udpateProjectSettings,
    fetchProjectSettings
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

    const isIntAdwordsEnabled = () => {
        return currentProjectSettings && currentProjectSettings.int_adwords_enabled_agent_uuid && currentProjectSettings.int_adwords_enabled_agent_uuid != "";
    }

    const getRedirectURL = () => {
        let host = getAdwordsHostURL();
        return host + ADWORDS_REDIRECT_URI + "?pid=" + activeProject?.id + "&aid=" + agent_details?.uuid;
    }
    useEffect(() => {
        if (isIntAdwordsEnabled()) {
            setIsActive(true);
        }
    }, [activeProject, agent_details]);

    const enableAdwords = () => {
        setLoading(true);
        enableAdwordsIntegration(activeProject.id).then((r) => {
            setLoading(false);
            console.log("rrrrr", r)
            if (r.status == 304) {
                window.location = getRedirectURL();
                return
            }
            if (r.status == 200) {
                renderSettingInfo();
                message.success('Google Ads integration enabled!');
                fetchProjectSettings(activeProject.id).then(() => {
                    if (currentProjectSettings?.int_facebook_ad_account) {
                        setIsActive(true);
                    }
                });
            }
            if(r.status >= 400){
                setShowManageBtn(true);
                setCustomerAccountsLoaded(false);
                message.error('Error fetching Google Ad accounts'); 
            }
            setIsActive(true);
        }).catch((err) => {
            setLoading(false);
            console.log('change password failed-->', err);
            setIsActive(false);
        });
    };

    const onAccountSelect = (e) => {
        let selectedAdwordsAcc = [...selectedAdwordsAccounts]
        if (e.target.checked) {
            selectedAdwordsAcc.push(e.target.value)
        } else {
            let index = selectedAdwordsAcc.indexOf(e.target.value)
            if (index > -1) selectedAdwordsAcc.splice(index, 1)
        }
        setSelectedAdwordsAccounts(selectedAdwordsAcc);
    }

    const addManualAccount = () => {
        let accounts = [...manualAccounts]
        if (accountId != "") {
            accounts.push(
                {
                    customer_id: accountId
                }
            )
        }
        setManualAccounts(accounts);
        setShowModal(false);
    }

    const onClickFinishSetup = () => {
        let selectedAdwordsAcc = selectedAdwordsAccounts.join(",")
        udpateProjectSettings(activeProject.id,
            { 'int_adwords_customer_account_id': selectedAdwordsAcc }).then(() => {
                setAddNewAccount(false);
                setSelectedAdwordsAccounts([]);
                message.success('Adwords Accounts updated!');
                setShowManageBtn(true);
                setCustomerAccountsLoaded(false);
            });
    }


    const renderAccountsList = () => { 
        let accountRows = [];

        if (!customerAccounts) return;

        for (let i = 0; i < customerAccounts.length; i++) {
            let account = customerAccounts[i];

            accountRows.push(
                <tr>
                    <td style={{ border: 'none', padding: '5px' }}>
                        <Checkbox value={account.customer_id} onChange={onAccountSelect} />
                    </td>
                    <td style={{ border: 'none', padding: '5px' }}>{account.customer_id}</td>
                    <td style={{ border: 'none', padding: '5px' }}>{account.name}</td>
                </tr>
            )
        }
        for (let i = 0; i < manualAccounts.length; i++) {
            let account = manualAccounts[i];

            accountRows.push(
                <tr>
                    <td style={{ border: 'none', padding: '5px' }}>
                        <Checkbox value={account.customer_id} onChange={onAccountSelect} />
                    </td>
                    <td style={{ border: 'none', padding: '5px' }}>{account.customer_id}</td>
                    <td style={{ border: 'none', padding: '5px' }}>{account.name}</td>
                </tr>
            )
        }


        return (
            <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
                <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Add/Remove Accounts</Text>
                <table>
                    <thead>
                        <tr>
                            <td style={{ border: 'none', padding: '5px' }}></td>
                            <td style={{ border: 'none', padding: '5px' }}><Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Customer Id</Text></td>
                            <td style={{ border: 'none', padding: '5px' }}><Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Customer Name</Text></td>
                        </tr>
                    </thead>
                    <tbody>{accountRows}</tbody>
                </table>
                <div className={'mt-4'} >
                    <Button onClick={() => setShowModal(true)}> Enter Id Manually </Button>
                    <Button type={'primary'} className={'ml-2'} onClick={onClickFinishSetup}> Finish Setup </Button>
                </div>
            </div>
        );
    }



    // const isCustomerAccountSelected = () => {
    //     return currentProjectSettings && currentProjectSettings.int_adwords_customer_account_id && !addNewAccount;
    // };

    const renderSettingInfo = () => {
 
        let isCustomerAccountChosen = currentProjectSettings.int_adwords_customer_account_id &&
        currentProjectSettings.int_adwords_customer_account_id != "" && !addNewAccount;
        
        // get all adwords account when no account is chosen and not account list not loaded. 
        // if (isIntAdwordsEnabled() && !isCustomerAccountChosen && !customerAccountsLoaded) {
        if (isIntAdwordsEnabled() && !customerAccountsLoaded) {
            // setLoadingData(true); 
            fetchAdwordsCustomerAccounts({ "project_id": activeProject.id }).then((data) => {
                setCustomerAccountsLoaded(true);
                setCustomerAccounts(data?.customer_accounts);
                // setLoadingData(false);
            });
        }
 
    } 
    return (
        <>
                <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with GoogleAdWords integrations'} />} onError={FaErrorLog}>

                

            <div className={'mt-4 flex w-full'}>
                {currentProjectSettings?.int_adwords_customer_account_id && <>
                    <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Connected Accounts</Text>
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Adwords sync account details</Text>
                        <Input size="large" disabled={true} value={currentProjectSettings?.int_adwords_customer_account_id} style={{ width: '400px' }} />
                    </div>
                </>}
            </div>
            <div className={'w-full'}>
                {
                    isIntAdwordsEnabled() && showManageBtn && <div className={'mt-4'}>
                        <Button onClick={() => { renderSettingInfo(); setShowManageBtn(false); }}>Manage Accounts</Button>
                    </div>
                }
            </div>
            <div className={'w-full'}>
                {
                    !showManageBtn && !customerAccountsLoaded && <Skeleton />
                }
            </div>
            <div>
                {
                    customerAccountsLoaded && renderAccountsList()
                }
            </div>

            <div className={'mt-4 flex'}>
                {
                    !currentProjectSettings?.int_adwords_enabled_agent_uuid && <>
                        <Button type={'primary'} loading={loading} onClick={enableAdwords}>Enable using Google</Button>
                        <Button className={'ml-2 '}>View documentation</Button>
                    </>
                }
            </div>

            <Modal
                visible={showModal}
                zIndex={1020}
                afterClose={() => setShowModal(false)}
                className={'fa-modal--regular fa-modal--slideInDown'}
                centered={true}
                footer={null}
                transitionName=""
                maskTransitionName=""
                closable={false}
            >
                <Row>
                    <Col span={24}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Manually add Google Adwords account</Text>
                    </Col>
                </Row>
                <Row className={'mt-4'}>
                    <Col span={24}>
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Enter adwords account ID:</Text>
                        <Input type="text" className={'mt-2'} onChange={(e) => setAccountId(e.target.value)} />
                    </Col>
                </Row>
                <Row className={'mt-4'}>
                    <Col span={24}>
                        <div className={'flex justify-end'}>
                            <Button onClick={() => setShowModal(false)}> Cancel </Button>
                            <Button className={'ml-2'} type={'primary'} onClick={addManualAccount}> Submit </Button>
                        </div>
                    </Col>
                </Row>
            </Modal>

            </ErrorBoundary>
        </>
    )
}




const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    agent_details: state.agent.agent_details,
    currentProjectSettings: state.global.currentProjectSettings
});

export default connect(mapStateToProps, { fetchProjectSettings, enableAdwordsIntegration, fetchAdwordsCustomerAccounts, udpateProjectSettings })(GoogleIntegration);