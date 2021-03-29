import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
  Button, message
  } from 'antd'; 

const DriftIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive
}) =>{  
    const [loading, setLoading] = useState(false); 

useEffect(()=>{
    fetchProjectSettings(activeProject.id).then(()=>{ 
      if(currentProjectSettings?.int_drift){
        setIsActive(true);
      }
    })
},[]);

const enableDrift = () => { 
    setLoading(true); 
        udpateProjectSettings(activeProject.id, 
        { 'int_drift' : true 
    }).then(() => {
        setLoading(false); 
        setTimeout(() => {
            message.success('Drift integration enabled!'); 
        }, 500);
        setIsActive(true);
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
    <div className={'mt-4 flex'}>
    {
        currentProjectSettings?.int_drift ? <Button loading={loading} onClick={()=>onDisconnect()}>Disable</Button> : 
        <Button type={'primary'} loading={loading} onClick={enableDrift}>Enable Now</Button>
    }
        <Button className={'ml-2 '}>View documentation</Button> 
    </div>
    </>
)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
  });
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(DriftIntegration)