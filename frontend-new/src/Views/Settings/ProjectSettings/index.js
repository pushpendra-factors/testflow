import React, { useState } from 'react';
import {
  Row, Col, Menu
} from 'antd';
import { Text } from 'factorsComponents';
import BasicSettings from './BasicSettings';
import SDKSettings from './SDKSettings';
import UserSettings from './UserSettings';
import IntegrationSettings from './IntegrationSettings';
import { fetchProjectAgents } from 'Reducers/agentActions';
import { connect } from 'react-redux';

const MenuTabs = {
  generalSettings: 'General Settings',
  SDK: 'Javascript SDK',
  Users: 'Users',
  Integrations: 'Integrations',
  EventAlias: 'Event Alias'
};

function ProjectSettings({
  activeProjectID,
  fetchProjectAgents
}) {
  const [selectedMenu, setSelectedMenu] = useState(MenuTabs.generalSettings);
  // const [editPasswordModal, setPasswordModal] = useState(false);
  // const [editDetailsModal, setDetailsModal] = useState(false);
  // const [confirmLoading, setConfirmLoading] = useState(false);

  const handleClick = (e) => {
    setSelectedMenu(e.key);
    if (e.key === MenuTabs.Users) {
      fetchProjectAgents(activeProjectID);
    }
  };

  // const handleOk = () => {
  //   setConfirmLoading(true);
  //   setTimeout(() => {
  //     setConfirmLoading(false);
  //     setPasswordModal(false);
  //     setDetailsModal(false);
  //   }, 2000);
  // };

  return (
    <>

      <div className={'fa-container'}>
        <Row gutter={[24, 24]} justify={'center'} className={'pt-16 pb-2 m-0 '}>
          <Col span={20}>
            <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'}>Project Settings</Text>
            <Text type={'title'} level={7} weight={'regular'} extraClass={'m-0'} color={'grey'}>FactorsAI</Text>
          </Col>
        </Row>
        <Row gutter={[24, 24]} justify={'center'}>
          <Col span={5}>

            <Menu
              onClick={handleClick}
              defaultSelectedKeys={MenuTabs.generalSettings}
              mode="inline"
              className={'fa-settings--menu'}>
              <Menu.Item key={MenuTabs.generalSettings}>{MenuTabs.generalSettings}</Menu.Item>
              <Menu.Item key={MenuTabs.SDK}>{MenuTabs.SDK}</Menu.Item>
              <Menu.Item key={MenuTabs.Users}>{MenuTabs.Users}</Menu.Item>
              <Menu.Item key={MenuTabs.Integrations}>{MenuTabs.Integrations}</Menu.Item>
            </Menu>

          </Col>
          <Col span={15}>
            {selectedMenu === MenuTabs.generalSettings && <BasicSettings /> }
            {selectedMenu === MenuTabs.SDK && <SDKSettings /> }
            {selectedMenu === MenuTabs.Users && <UserSettings /> }
            {selectedMenu === MenuTabs.Integrations && <IntegrationSettings /> }
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProjectID: state.global.active_project.id,
  agents: state.agent.agents
});

export default connect(mapStateToProps, { fetchProjectAgents })(ProjectSettings);
