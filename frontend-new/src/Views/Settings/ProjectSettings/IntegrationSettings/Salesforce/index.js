import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings, enableSalesforceIntegration, fetchSalesforceRedirectURL } from 'Reducers/global';
import {
     Input, Button, message
  } from 'antd'; 
  import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
  import {ErrorBoundary} from 'react-error-boundary';
  import factorsai from 'factorsai';
  
const SalesForceIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive,
    enableSalesforceIntegration,
    fetchSalesforceRedirectURL,
    kbLink = false
}) =>{  
    const [loading, setLoading] = useState(false);
    const [showForm, setShowForm] = useState(false);

    const isSalesforceEnabled = () => {
        return currentProjectSettings && currentProjectSettings.int_salesforce_enabled_agent_uuid && currentProjectSettings.int_salesforce_enabled_agent_uuid != "";
      }

      
useEffect(()=>{
  setIsActive(isSalesforceEnabled());
},[]);
 
 
  const handleRedirectToURL = () =>{
    fetchSalesforceRedirectURL(activeProject.id.toString())
    .then((r)=>{
      if (r.status == 307) {
        window.location = r.data.redirectURL;
      }
    })
  }

  const  onClickEnableSalesforce = () => {

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION',{'name': 'salesforce','activeProjectID': activeProject.id});

    enableSalesforceIntegration(activeProject.id.toString())
      .then((r) => {
        if (r.status == 304) {
          handleRedirectToURL();
        }
      });
  }

  const onDisconnect = () =>{
    setLoading(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_salesforce_enabled_agent_uuid' : ""
    }).then(() => {
        setLoading(false);
        setShowForm(false); 
        setTimeout(() => {
            message.success('Salesforce integration disconnected!'); 
        }, 500);
        setIsActive(false);
    }).catch((err) => {
        message.error(`${err?.data?.error}`);  
        setShowForm(false);
        setLoading(false);
        console.log('change password failed-->', err); 
    });
  }

const isEnabled = isSalesforceEnabled();
return (
    <> 
    <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with Salesforce integrations'} />} onError={FaErrorLog}>  
    <div className={'mt-4 flex'}>
    {isEnabled && <>
      <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Account Connected</Text>
            <Text type={'title'} level={7} color={'grey'}  extraClass={'m-0 mt-2'}>Salesforce sync is enabled</Text>
            <Button loading={loading} className={'mt-4'} onClick={()=>onDisconnect()}>Disconnect</Button>
      </div>
    </>} 
    {!isEnabled && <>
    <Button type={'primary'} loading={loading} onClick={onClickEnableSalesforce}>Enable using Salesforce</Button> 
    {kbLink && <a className={'ant-btn ml-2 '} target={"_blank"} href={kbLink}>View documentation</a>}
    </>
    }
    </div>
    </ErrorBoundary>
    </>
)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
  });
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings, enableSalesforceIntegration, fetchSalesforceRedirectURL })(SalesForceIntegration)