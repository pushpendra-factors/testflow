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
(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="${assetURL}";s.async=true;d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("${projectToken}")})
</script>`}
            </code>
            </pre>
          </Col>
          <Col span={24}>
            <Text type={'title'} level={5} weight={'bold'} color={'grey'} extraClass={'m-0 mt-4'}>Setup 2</Text>
            <Text type={'paragraph'} extraClass={'m-0'}>Send us an event (Enable Auto-track for capturing user visits automatically).</Text>
          </Col>
          <Col span={24}>
            <pre className={'fa-code-block my-4'}>
            <code>
{'factors.track("YOUR_EVENT");'}
            </code>
            </pre>
          </Col>
          <Col span={24}>
            <Text type={'paragraph'} extraClass={'m-0 mt-2'}>For detailed instructions on how to install and initialize the JavaScript SDK please refer to our <a className={'fa-anchor'} href="#!">JavaScript developer documentation.</a></Text>
          </Col>
    </Row>
  );
};

const JSConfig = ({ currentProjectSettings, activeProject, udpateProjectSettings }) => {
  const currentProjectId = activeProject.id;

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

  return (
    <Row>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-2'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" onChange={toggleAutoTrack} defaultChecked={currentProjectSettings.auto_track} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto-track</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.</Text>
    </Col>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" onChange={toggleExcludeBot} defaultChecked={currentProjectSettings.exclude_bot} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Exclude Bot</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.</Text>
    </Col>
    <Col span={24}>
      <div span={24} className={'flex flex-start items-center mt-8'}>
        <span style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" onChange={toggleAutoFormCapture} defaultChecked={currentProjectSettings.auto_form_capture} /></span> <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Auto Form Capture</Text>
      </div>
    </Col>
    <Col span={24} className={'flex flex-start items-center'}>
      <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.</Text>
    </Col>
    </Row>
  );
};

function EditUserDetails({
  activeProject, fetchProjectSettings, currentProjectSettings, udpateProjectSettings
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

  return (
    <>
      <div className={'mb-10 pl-4'}>
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
                  <JSConfig udpateProjectSettings={udpateProjectSettings} currentProjectSettings={currentProjectSettings} activeProject={activeProject} />
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
    activeProject: state.global.active_project
  };
};
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(EditUserDetails);
