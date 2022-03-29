import React, { useState, useEffect } from 'react';
import {
  Row, Col, Skeleton, Tabs, Switch, message
} from 'antd';
import { Text } from 'factorsComponents';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { connect } from 'react-redux';
const { TabPane } = Tabs;

const ViewSetup = ({ activeProject }) => {
  const projectToken = activeProject.token;
  // eslint-disable-next-line
  const assetURL = BUILD_CONFIG.sdk_asset_url;


  return (
    <Row>
          <Col span={24}>
            <Text type={'title'} level={5} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2'}>Setup 1</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>Add the below javascript code on every page between the {'<head>'} and {'</head>'} tags.</Text>
          </Col>
          <Col span={24}>
            <pre className={'fa-code-block my-4'}>
            <code>
{`<script>
(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="${assetURL}";s.async=true;d=document.getElementsByTagName("script")[0];document.head.insertBefore(s,d)})(function(){factors.init("${projectToken}")})
</script>`}
            </code>
            </pre>
          </Col>
          <Col span={24}> 
            <Text type={'paragraph'} extraClass={'m-0 mt-2 mb-2'}>For detailed help or instructions to setup via GTM (Google Tag Manager), please refer to our <a className={'fa-anchor'} href="https://help.factors.ai/en/articles/5754974-placing-factors-s-javascript-sdk-on-your-website" target='_blank'>JavaScript developer documentation.</a></Text> 
          </Col>
          <Col span={24}>
            <Text type={'title'} level={5} weight={'bold'} color={'grey'} extraClass={'m-0 mt-4'}>Setup 2 (Optional)</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>Send us an event (Enable Auto-track for capturing user visits automatically).</Text>
          </Col>
          <Col span={24}>
            <pre className={'fa-code-block my-4'}>
            <code>
{'factors.track("YOUR_EVENT");'}
            </code>
            </pre>
          </Col>
    </Row>
  );
};

const JSConfig = ({ currentProjectSettings, activeProject, udpateProjectSettings, agents, currentAgent }) => {
  const [enableEdit, setEnableEdit] = useState(false);

  const currentProjectId = activeProject.id;

  useEffect(() => {
    setEnableEdit(false);
    agents && currentAgent && agents.map((agent) => {
      console.log(agent,currentAgent);
      if (agent.uuid === currentAgent.uuid) {
        if (agent.role === 1) {
          setEnableEdit(true);
        }
      }
    }); 
  }, [activeProject, agents, currentAgent]);


  const toggleAutoTrack = (checked) => { 
    udpateProjectSettings(currentProjectId, { auto_track: checked }).catch((err) => {
      console.log('Oops! something went wrong-->', err);
      message.error('Oops! something went wrong.');
    }); 
  };

  const toggleExcludeBot = (checked) => { 
      udpateProjectSettings(currentProjectId, { exclude_bot: checked }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      }); 
  };

  const toggleAutoFormCapture = (checked) => { 
      udpateProjectSettings(currentProjectId, { auto_form_capture: checked }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      });  
  };
 
  const toggleAutoTrackSPAPageView = (checked) => { 
    udpateProjectSettings(currentProjectId, { auto_track_spa_page_view: checked }).catch((err) => {
      console.log('Oops! something went wrong-->', err);
      message.error('Oops! something went wrong.');
    }); 
  };

  return (
    <Row>
      {enableEdit &&  <Col span={24}>
        <Text type={'title'} level={7}  color={'grey'} extraClass={'m-0 my-2'}>*Only Admin(s) can change configurations.</Text>
    </Col>
      }
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-2'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On"  disabled={enableEdit} unCheckedChildren="OFF" onChange={toggleAutoTrack} defaultChecked={currentProjectSettings.auto_track} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto-track</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Track standard events such as page_view, page_load time, page_spent_time and button clicks for each user</Text>
    </Col>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" disabled={enableEdit} unCheckedChildren="OFF" onChange={toggleAutoTrackSPAPageView} defaultChecked={currentProjectSettings.auto_track_spa_page_view} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto-track Single Page Application</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Track standard events such as page_view, page_load time, page_spent_time and button clicks for each user, on Single Page Applications like React, Angular, Vue, etc,.</Text>
    </Col>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" disabled={enableEdit} unCheckedChildren="OFF" onChange={toggleExcludeBot} defaultChecked={currentProjectSettings.exclude_bot} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Exclude Bot</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Automatically exclude bot traffic from website traffic using Factor’s proprietary algorithm</Text>
    </Col>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" disabled={enableEdit} unCheckedChildren="OFF" onChange={toggleAutoFormCapture} defaultChecked={currentProjectSettings.auto_form_capture} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto Form Capture</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Automatically track personal identification information such as email and phone number from Form Submissions</Text>
    </Col>
    </Row>
  );
};

function EditUserDetails({
  activeProject, fetchProjectSettings, currentProjectSettings, udpateProjectSettings, agents, currentAgent
}) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false);
    });
  }, [activeProject]);

  const callback = (key) => {
    console.log(key);
  };

  currentProjectSettings = currentProjectSettings?.project_settings || currentProjectSettings;

  return (
    <>
      <div className={'mb-4 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Javascript SDK</Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={24}>
            { dataLoading ? <Skeleton active paragraph={{ rows: 4 }}/>
              : <Tabs defaultActiveKey="1" onChange={callback}>
                <TabPane tab="Setup" key="1">
                  <ViewSetup currentProjectSettings={currentProjectSettings} activeProject={activeProject} />
                </TabPane>
                <TabPane tab="Configuration" key="2">
                  <JSConfig 
                  udpateProjectSettings={udpateProjectSettings} 
                  currentProjectSettings={currentProjectSettings} 
                  activeProject={activeProject}
                  agents={agents}
                  currentAgent={currentAgent}
                   />
                </TabPane>
              </Tabs>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}
const mapStateToProps = (state) => {
  return {
    currentProjectSettings: state.global.currentProjectSettings,
    activeProject: state.global.active_project,
    agents: state.agent.agents, 
    currentAgent: state.agent.agent_details
  };
};
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(EditUserDetails);
