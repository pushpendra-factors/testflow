import React, { useState, useEffect } from 'react';
import {
  Row, Col, Skeleton, Tabs, Switch, message, Table, Button
} from 'antd';
import { Text } from 'factorsComponents';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { fetchClickableElements, toggleClickableElement } from '../../../../reducers/settings/middleware';
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
window.factors=window.factors||function(){this.q=[];var i=new CustomEvent("FACTORS_QUEUED_EVENT"),n=function(t,e){this.q.push({k:t,a:e}),window.dispatchEvent(i)};return this.track=function(t,e,i){n("track",arguments)},this.init=function(t,e,i){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=i,window.dispatchEvent(new CustomEvent("FACTORS_INIT_EVENT"))},this.reset=function(){n("reset",arguments)},this.page=function(t,e){n("page",arguments)},this.updateEventProperties=function(t,e){n("updateEventProperties",arguments)},this.identify=function(t,e){n("identify",arguments)},this.addUserProperties=function(t){n("addUserProperties",arguments)},this.getUserId=function(){n("getUserId",arguments)},this.call=function(){var t={k:"",a:[]};if(arguments&&1<=arguments.length){for(var e=1;e<arguments.length;e++)t.a.push(arguments[e]);t.k=arguments[0]}this.q.push(t),window.dispatchEvent(i)},this.init("${projectToken}"),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,d=document.getElementsByTagName("script")[0],document.head.insertBefore(t,d)}(); 
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

const GTMSetup = ({ activeProject }) => {
  const projectToken = activeProject.token;
  // eslint-disable-next-line
  const assetURL = BUILD_CONFIG.sdk_asset_url;


  return (
    <Row>
          <Col span={24}>
            <Text type={'title'} level={5} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2 mb-1'}>Setup 1</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>1. Sign in to <span className={'underline'}><a href='https://tagmanager.google.com/' target='_blank'>Google Tag Manager</a></span>, select “Workspace”, and “Add a new tag”</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>2. Name it “Factors tag”. Select <span className={'italic'}>Edit</span> on Tag Configuration</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>3. Under custom, select <span className={'italic'}>custom HTML</span></Text>
            <Text type={'paragraph'} extraClass={'m-0'}>4. Copy the below tracking script and <span className={'italic'}>paste</span> it on the HTML field, Select <span className={'font-extrabold'}>Save</span></Text>
          </Col>
          <Col span={24}>
            <pre className={'fa-code-block my-4'}>
            <code>
{`<script>
window.factors=window.factors||function(){this.q=[];var i=new CustomEvent("FACTORS_QUEUED_EVENT"),n=function(t,e){this.q.push({k:t,a:e}),window.dispatchEvent(i)};return this.track=function(t,e,i){n("track",arguments)},this.init=function(t,e,i){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=i,window.dispatchEvent(new CustomEvent("FACTORS_INIT_EVENT"))},this.reset=function(){n("reset",arguments)},this.page=function(t,e){n("page",arguments)},this.updateEventProperties=function(t,e){n("updateEventProperties",arguments)},this.identify=function(t,e){n("identify",arguments)},this.addUserProperties=function(t){n("addUserProperties",arguments)},this.getUserId=function(){n("getUserId",arguments)},this.call=function(){var t={k:"",a:[]};if(arguments&&1<=arguments.length){for(var e=1;e<arguments.length;e++)t.a.push(arguments[e]);t.k=arguments[0]}this.q.push(t),window.dispatchEvent(i)},this.init("${projectToken}"),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,d=document.getElementsByTagName("script")[0],document.head.insertBefore(t,d)}(); 
</script>`}
            </code>
            </pre>
          </Col>
          <Col span={24}>
            <Text type={'paragraph'} extraClass={'m-0'}>5. In the <span className={'italic'}>Triggers</span> popup, select <span className={'italic'}>Add Trigger</span> and select <span className={'italic'}>All Pages</span></Text>
            <Text type={'paragraph'} extraClass={'m-0'}>6. The trigger has been added. Click on <span className={'font-extrabold'}>Publish</span> at the top of your GTM window!</Text>
          </Col>
          <Col span={24}> 
            <Text type={'paragraph'} extraClass={'m-0 mt-4 mb-2'}>For detailed help or instructions to setup via GTM (Google Tag Manager), please refer to our <a className={'fa-anchor'} href="https://help.factors.ai/en/articles/5754974-placing-factors-s-javascript-sdk-on-your-website" target='_blank'>JavaScript developer documentation.</a></Text> 
          </Col>
          <Col span={24}>
            <Text type={'title'} level={5} weight={'bold'} color={'grey'} extraClass={'m-0 mt-4'}>Setup 2 (Optional)</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>Send us custom events that you define using GTM’s triggers (Enable Auto-track for capturing user visits automatically).</Text>
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

  const toggleClickCapture = (checked) => { 
    udpateProjectSettings(currentProjectId, { auto_click_capture: checked }).catch((err) => {
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
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" onChange={toggleClickCapture} defaultChecked={currentProjectSettings.auto_click_capture} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto Click Capture</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Starts discovering available buttons and anchors on the website. After discovery, it will be listed under Click Tracking Configurations and can be enabled for tracking as events.</Text>
    </Col>
    </Row>
  );
};

const ClickTrackConfiguration = ({ 
  activeProject, 
  currentProjectSettings, 
  fetchClickableElements, 
  clickableElements, 
  toggleClickableElement 
}) => {

  var columns = [
    {
      title: 'Display Name',
      dataIndex: 'displayName',
      key: 'displayName',
      width: 300,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: 'Tracking',
      dataIndex: 'tracking',
      key: 'tracking',
      render: (p) => (
        <Switch checkedChildren="On" unCheckedChildren="OFF" defaultChecked={p.enabled} onChange={p.toggler}/>
      ),
      align: 'right',
    }
  ];

  var dataSource = [];
  for (let i=0; i<clickableElements.length; i++) {
    dataSource.push({
      index: clickableElements[i].id,
      displayName: clickableElements[i].display_name,
      type: (clickableElements[i].element_type),
      tracking: { 
        enabled: clickableElements[i].enabled, 
        toggler: () => toggleClickableElement(activeProject.id, clickableElements[i].id, clickableElements[i].enabled) 
      }
    });
  }

  return (
      <Row className={'mt-8'}>
        <Col span={24}>
          <Table
            className={'fa-table--basic'}
            columns={columns}
            dataSource={dataSource}
            pagination={false}
          />
        </Col>
      </Row>
  );
}

function JavascriptSDK({
  activeProject, 
  fetchProjectSettings, 
  currentProjectSettings, 
  udpateProjectSettings, 
  agents, 
  currentAgent, 
  fetchClickableElements,
  toggleClickableElement,
  clickableElements
}) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false);
    })

    fetchClickableElements(activeProject.id).then(() => {
      setDataLoading(false);
    });

  }, [activeProject]);
  

  const callback = (key) => {
    console.log(key);
  };

  currentProjectSettings = currentProjectSettings?.project_settings || currentProjectSettings;

  const renderTabs = () => {
    let tabs = [
      <TabPane tab="GTM Setup" key="1">
      <GTMSetup currentProjectSettings={currentProjectSettings} activeProject={activeProject} />
    </TabPane>,
    <TabPane tab="Manual Setup" key="2">
      <ViewSetup currentProjectSettings={currentProjectSettings} activeProject={activeProject} />
    </TabPane>,
    <TabPane tab="General Configuration" key="3">
      <JSConfig 
      udpateProjectSettings={udpateProjectSettings} 
      currentProjectSettings={currentProjectSettings} 
      activeProject={activeProject}
      agents={agents}
      currentAgent={currentAgent}
       />
    </TabPane>
    ]

   if (currentProjectSettings.auto_click_capture)
    tabs.push(
    <TabPane tab="Click Tracking Configuration" key="4">
      <ClickTrackConfiguration 
        activeProject={activeProject} 
        currentProjectSettings={currentProjectSettings} 
        clickableElements={clickableElements}
        toggleClickableElement={toggleClickableElement}
      />
    </TabPane>
    );

   return tabs;
  }

  return (
    <>
      <div className={'mb-4 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Javascript SDK</Text>
          </Col>
        </Row>
        <Row>
          <Col span={24}>
            <Text type={'title'} level={6} color={'grey-2'} extraClass={'m-0 my-1'}>Your website data will be visible on the platform from the time the your javascript SDK is placed on your site. Hence, no historical data prior to the setup would be available on the platform.</Text>
            <Text type={'title'} level={6} color={'grey-2'} extraClass={'m-0'}>The website data you see in Factors is real-time.</Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={24}>
            { dataLoading ? <Skeleton active paragraph={{ rows: 4 }}/> : <Tabs defaultActiveKey="1" onChange={callback}>{renderTabs()} </Tabs> }
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
    currentAgent: state.agent.agent_details,
    clickableElements: state.settings.clickableElements,
  };
};
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings, 
  fetchClickableElements, toggleClickableElement })(JavascriptSDK);
