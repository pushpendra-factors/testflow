import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings, addLinkedinAccessToken, deleteIntegration } from 'Reducers/global';
import {
    Button, message, Select, Modal, Row, Col, Input
} from 'antd';  
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import {ErrorBoundary} from 'react-error-boundary';

const LinkedInIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings,
    setIsActive,
    addLinkedinAccessToken,
    kbLink = false,
    deleteIntegration
}) => {
    const [loading, setLoading] = useState(false);
    const [FbResponse, SetFbResponse] = useState(null);
    const [adAccounts, SetAdAccounts] = useState(null);
    const [SelectedAdAccount, SetSelectedAdAccount] = useState(null);
    const [showForm, setShowForm] = useState(false);
    const [oauthResponse, setOauthResponse] = useState(false);


    const getAdAccounts = (jsonRes) => {
        let accounts = jsonRes.map(res => {
            return { "value": res.id, "name": res.name }
        })
        return accounts
    }

    const getHostURL = () => {
        // return isDevelopment() ? BUILD_CONFIG.adwords_service_host : BUILD_CONFIG.backend_host;
        return BUILD_CONFIG.backend_host;
    }

    useEffect(() => {
      if (currentProjectSettings?.int_linkedin_ad_account) {
        setIsActive(true);
      }
    }, [currentProjectSettings]);

    useEffect(() => {
        let code = localStorage.getItem("Linkedin_code");
        let state = localStorage.getItem("Linkedin_state");
        if (code != '' && state === 'factors') {
            let url = getHostURL() + '/integrations/linkedin/auth'
            fetch(url, {
                method: 'POST',
                body: JSON.stringify({
                    'code': code
                })
            }).then(response => { 
                if (!response.ok) {
                    throw Error;
                }
                return response;
            }).then(response => { 
                if (response.status < 400) {
                    response.json().then(e => { 
                        setOauthResponse(e)
                        fetch(getHostURL() + '/integrations/linkedin/ad_accounts', {
                            method: 'POST',
                            body: JSON.stringify({
                                'access_token': e?.access_token
                            })
                        }
                        ).then(response => { 
                            if (!response.ok) {
                                throw Error;
                            }
                            return response;
                        }).then(response => {  
                            response.json().then(res => {
                                let jsonRes = JSON.parse(res)
                                let adAccountsNew = getAdAccounts(jsonRes.elements)
                                SetAdAccounts(adAccountsNew); 
                                localStorage.removeItem('Linkedin_code');
                                localStorage.removeItem('Linkedin_state');
                                setShowForm(true);
                            });
                        }).catch((err) => {
                            message.error('Failed to fetch linkedin/ad_accounts');
                        });
                    })
                }
                else{
                    console.log("Failed to fetch linkedin/ad_accounts!!")
                }
            }).catch((err) => {
                message.error('Failed to fetch linkedin/auth');
            });
        }
    }, []);

    const makeSelectOpt = (value, label) => {
        if (!label) label = value;
        return { value: value, label: label }
    }

    const createSelectOpts = (opts) => {
        let ropts = [];
        for (let k in opts) ropts.push(makeSelectOpt(k, opts[k]));
        return ropts;
    }


    const renderLinkedinLogin = () => {
        if (!(currentProjectSettings?.int_linkedin_access_token)) {
            let hostname = window.location.hostname
            let protocol = window.location.protocol
            let port = window.location.port
            let redirect_uri = protocol + "//" + hostname + ":" + port
            if (port === undefined || port === '') {
                redirect_uri = protocol + "//" + hostname
            }
            let href = `https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=${BUILD_CONFIG.linkedin_client_id}&redirect_uri=${redirect_uri}&state=factors&scope=r_basicprofile%20r_liteprofile%20r_ads_reporting%20r_ads`
            return (
                <a href={href} className='ant-btn ant-btn-primary'> Enable using LinkedIn </a>
            )
        } 
    }

    const handleSubmit = e => {
        e.preventDefault(); 
        if (SelectedAdAccount != "") {
            const data = {
                "int_linkedin_ad_account": SelectedAdAccount,
                "int_linkedin_refresh_token": oauthResponse["refresh_token"],
                "int_linkedin_refresh_token_expiry": oauthResponse["refresh_token_expires_in"],
                "project_id": activeProject.id.toString(),
                "int_linkedin_access_token": oauthResponse['access_token'],
                "int_linkedin_access_token_expiry": oauthResponse["expires_in"]
            }
            addLinkedinAccessToken(data).then(() => {
                fetchProjectSettings(activeProject.id);
                setShowForm(false);
                setIsActive(true);
                message.success('LinkedIn integration enabled!');
            }).catch((e) => {
                console.log(e);
                message.error(e);
                setShowForm(false);
                setIsActive(false);
            });

        }
    }

    const onDisconnect = () =>{
        setLoading(true);
        deleteIntegration(activeProject.id, 'linkedin')
        .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setShowForm(false); 
            setTimeout(() => {
                message.success('LinkedIn integration disconnected!'); 
            }, 500);
            setIsActive(false);
        }).catch((err) => {
            message.error(`${err?.data?.error}`);  
            setShowForm(false);
            setLoading(false);
            console.log('change password failed-->', err); 
        });
      }

    const getAdAccountsOptSrc = () => {
        let opts = {}
        for (let i in adAccounts) {
            let adAccount = adAccounts[i];
            opts[adAccount.value] = adAccount.label;
        }
        return opts;
    }


    const formComponent = () => {
        if (!(currentProjectSettings.int_linkedin_access_token)) {
            if (adAccounts != "" && adAccounts?.length != 0) {
                return (
                    <>
                        <Modal
                            visible={showForm}
                            zIndex={1020}
                            afterClose={() => setShowForm(false)}
                            className={'fa-modal--regular fa-modal--slideInDown'}
                            centered={true}
                            footer={null}
                            transitionName=""
                            maskTransitionName=""
                            closable={false}
                        >
                            <div className={'p-4'}>
                                <Row>
                                    <Col span={24}>
                                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Choose your LinkedIn Ad account:</Text>
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Choose your LinkedIn Ad account to sync reports with Factors for performance reporting</Text>
                                    </Col>
                                </Row>
                                            <form onSubmit={e => handleSubmit(e)} className="w-full">
                                <Row className={'mt-6'}>
                                    <Col span={24}>
                                        <div className="w-full">
                                                <div className="w-full pb-2">
                                                    <Select
                                                        className="w-full"
                                                        value={SelectedAdAccount}
                                                        onChange={e => SetSelectedAdAccount(e)}
                                                        options={createSelectOpts(getAdAccountsOptSrc())}
                                                    />
                                                </div>
                                        </div>
                                    </Col>
                                </Row>
                                <Row className={'mt-2'}>
                                    <Col span={24}>
                                        <div className={'flex justify-end'}>
                                            <Button className='ant-btn-primary' htmlType="submit">Select</Button>
                                        </div>
                                    </Col>
                                </Row>
                                            </form>
                            </div>
                        </Modal>

                        {/* <div className="p-2">
            <form onSubmit={e => this.handleSubmit(e)}>
              <div className="w-50 pb-2">
                <h5>Choose your ad account:</h5>
                <Select
                value={SelectedAdAccount}
                onChange={this.handleChange}
                options={createSelectOpts(this.getAdAccountsOptSrc())}
                />
              </div>
              <input className="btn btn-primary shadow-none" type="submit" value="Submit"/>
            </form>
          </div> */}

                    </>
                )
            }
            //   if (adAccounts != "" && adAccounts.length == 0) {
            //     return <div>You don't have any ad accounts associated to the id you logged in with.</div>
            //   }
        } else {
            if (currentProjectSettings?.int_linkedin_ad_account !== "" || currentProjectSettings?.int_linkedin_ad_account !== undefined) {
                return (
                    <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Connected Account</Text>
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Selected LinkedIn Ad Account</Text>
                        <Input size="large" disabled={true} value={currentProjectSettings?.int_linkedin_ad_account} style={{ width: '400px' }} />
                        <Button loading={loading} className={'mt-4'} onClick={()=>onDisconnect()}>Disconnect</Button>
                    </div>
                )
            }
        }
        return
    }

    return (
        <>
            <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with LinkedIn integrations'} />} onError={FaErrorLog}> 
            <div className={'mt-4 flex w-6/12'}>
                {formComponent()}
            </div>

            { !adAccounts && <div className={'mt-4 flex'}>
                {renderLinkedinLogin()}
                {kbLink && <a className={'ant-btn ml-2 '} target={"_blank"} href={kbLink}>View documentation</a>}
            </div>}
            </ErrorBoundary>
        </>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
});

export default connect(mapStateToProps, { addLinkedinAccessToken, fetchProjectSettings, udpateProjectSettings, deleteIntegration })(LinkedInIntegration)
