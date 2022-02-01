import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
     Input, Button, message
  } from 'antd'; 
  import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
  import {ErrorBoundary} from 'react-error-boundary';
  import factorsai from 'factorsai';

const SegmentIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive,
    kbLink = false
}) =>{  
    const [loading, setLoading] = useState(false);
    const [showForm, setShowForm] = useState(false);

    currentProjectSettings = currentProjectSettings?.project_settings || currentProjectSettings;
    
    useEffect(() => {
      if (currentProjectSettings?.int_segment) {
        setIsActive(true);
      }
    }, [currentProjectSettings]);

const enableSegment = () => { 
    setLoading(true);

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION',{'name': 'segment','activeProjectID': activeProject.id});

    setShowForm(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_segment' : true 
    }).then(() => {
        setLoading(false);
        setShowForm(false); 
        setTimeout(() => {
            message.success('Segment integration enabled!'); 
        }, 500);
        setIsActive(true);
    }).catch((err) => {
        setShowForm(false);
        setLoading(false);
        console.log('change password failed-->', err);
        seterrorInfo(err.error);
        setIsActive(false);
    });
  };

  const onDisconnect = () =>{
    setLoading(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_segment' : false 
    }).then(() => {
        setLoading(false);
        setShowForm(false); 
        setTimeout(() => {
            message.success('Segment integration disabled!'); 
        }, 500);
        setIsActive(false);
    }).catch((err) => {
        message.error(`${err?.data?.error}`);  
        setShowForm(false);
        setLoading(false);
        console.log('change password failed-->', err); 
    });
  }

 

return (
    <> 
    <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with Segment integrations'} />} onError={FaErrorLog}> 
        {
            currentProjectSettings?.int_segment &&  <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Integration Details</Text>
            <Text type={'title'} level={7} color={'grey'}  extraClass={'m-0 mt-2'}>API Key</Text>
            <Input size="large" disabled={true} placeholder="API Key" value={activeProject?.private_token} style={{width:'400px'}}/>
        </div>
        }
        <div className={'mt-4 flex'}>
        {currentProjectSettings?.int_segment ? <Button loading={loading} onClick={()=>onDisconnect()}>Disable</Button> : <Button type={'primary'} loading={loading} onClick={enableSegment}>Enable Now</Button>
        }
            {kbLink && <a className={'ant-btn ml-2 '} target={"_blank"} href={kbLink}>View documentation</a>}
        </div>
    </ErrorBoundary>
    </>
)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
  });
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(SegmentIntegration)