import React, { useEffect, useState } from 'react';
import {
  Row, Col, Spin, Tabs, Switch, message
} from 'antd';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { connect } from 'react-redux';
import { Text } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import { buildExplainInsights, buildWeeklyInsights } from 'Reducers/insights';
import _ from 'lodash';

const InsightsSettings = ({
  currentProjectSettings,
  currentAgent,
  buildExplainInsights,
  buildWeeklyInsights,
  activeProject,
  fetchProjectSettings

}) => {

  const history = useHistory();

  const [projectSettings, setProjectSettings] = useState({});

  const whiteListedAccounts = [
    'baliga@factors.ai',
    'solutions@factors.ai',
    //   'janani@factors.ai', 
    //   'praveenr@factors.ai',
    //   'sonali@factors.ai',
    //   'praveen@factors.ai',
    //   'ashwin@factors.ai',
  ];


  const toggleEnableWI = (checked) => {

    buildWeeklyInsights(activeProject?.id, { "status": checked }).then((data) => {
      message.success(`Weekly Insights is turned ${checked ? "ON" : "OFF"} for this project`);
      fetchProjectSettings(activeProject?.id);
    }).catch((err) => {
      message.error(`Something went wrong!`);
      console.log("buildWeeklyInsights error", err)
    })

  };

  const toggleEnableExplain = (checked) => {
    buildExplainInsights(activeProject?.id, { "status": checked }).then((data) => {
      message.success(`Explain Insights is turned ${checked ? "ON" : "OFF"} for this project`);
      fetchProjectSettings(activeProject?.id);
    }).catch((err) => {
      message.error(`Something went wrong!`);
      console.log("buildExplainInsights error", err)
    })
  };

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    let activeAccount = currentAgent?.email;
    if (!whiteListedAccounts.includes(activeAccount)) {
      history.push('/');
    }

  }, [currentAgent])

  useEffect(() => {
    setProjectSettings(currentProjectSettings);
  }, [currentProjectSettings, activeProject]);

  return (
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={20}>
          <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 ml-2'}>Build WI/Explain</Text>
        </Col>
        {_.isEmpty(projectSettings) ? <Col span={20}> <Spin /> </Col> : <>
          <Col span={20}>
            <div span={24} className={'flex flex-start items-center mt-2'}>
              <span style={{ width: '50px' }}><Switch checkedChildren="On"
                // disabled={enableEdit} 
                defaultChecked={projectSettings?.is_explain_enabled}
                unCheckedChildren="OFF"
                onChange={toggleEnableExplain}

              /></span>
              <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Weekly Insights</Text>
            </div>
            <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Build Weekly Insights for this project.</Text>
          </Col>
          <Col span={20}>
            <div span={24} className={'flex flex-start items-center mt-2'}>
              <span style={{ width: '50px' }}><Switch checkedChildren="On"
                // disabled={enableEdit} 
                defaultChecked={projectSettings?.is_weekly_insights_enabled}
                unCheckedChildren="OFF"
                onChange={toggleEnableWI}
              /></span>
              <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2'}>Explain Insights</Text>
            </div>
            <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>Build Explain Insights for this project.</Text>
          </Col>
        </>
        }

      </Row>
    </div>
  )

}
const mapStateToProps = (state) => {
  return {
    currentProjectSettings: state.global.currentProjectSettings,
    activeProject: state.global.active_project,
    agents: state.agent.agents,
    currentAgent: state.agent.agent_details
  };
};

export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings, buildExplainInsights, buildWeeklyInsights })(InsightsSettings);
