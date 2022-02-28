import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
  Button, message
  } from 'antd'; 
  import { FaErrorComp, FaErrorLog } from 'factorsComponents';
  import {ErrorBoundary} from 'react-error-boundary'
  import factorsai from 'factorsai';

const DriftIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive,
    kbLink = false,
    currentAgent
}) =>{  
    const [loading, setLoading] = useState(false); 

    useEffect(() => {
      if (currentProjectSettings?.int_drift) {
        setIsActive(true);
      }
    }, [currentProjectSettings]);

    const sendSlackNotification = () => {
        let webhookURL = 'https://hooks.slack.com/services/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';
        let data = {
            "text": `User ${currentAgent.email} from Project "${activeProject.name}" Activated Integration: Drift`,
            "username" : "Signup User Actions",
            "icon_emoji" : ":golf:"
        }
        let params = {
            method: 'POST',
            body: JSON.stringify(data)
        }

        fetch(webhookURL, params)
        .then((response) => response.json())
        .then((response) => {
            console.log(response);
        })
        .catch((err) => {
            console.log('err',err);
        });
   }

const enableDrift = () => { 
    setLoading(true); 

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION',{'name': 'drift','activeProjectID': activeProject.id});

        udpateProjectSettings(activeProject.id, 
        { 'int_drift' : true 
    }).then(() => {
        setLoading(false); 
        setTimeout(() => {
            message.success('Drift integration enabled!'); 
        }, 500);
        setIsActive(true);
        sendSlackNotification();
    }).catch((err) => { 
        setLoading(false);
        console.log('change password failed-->', err);
        seterrorInfo(err.error);
        setIsActive(false);
    });
  };

  const onDisconnect = () =>{
    setLoading(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_drift' : false 
    }).then(() => {
        setLoading(false); 
        setTimeout(() => {
            message.success('Drift integration disabled!'); 
        }, 500);
        setIsActive(false);
    }).catch((err) => {
        message.error(`${err?.data?.error}`);   
        setLoading(false);
        console.log('change password failed-->', err); 
    });
  }

 

return ( 
    <>
    <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with Facebook integrations'} />} onError={FaErrorLog}>
    <div className={'mt-4 flex'}>
    {
        currentProjectSettings?.int_drift ? <Button loading={loading} onClick={()=>onDisconnect()}>Disable</Button> : 
        <Button type={'primary'} loading={loading} onClick={enableDrift}>Enable Now</Button>
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
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(DriftIntegration)