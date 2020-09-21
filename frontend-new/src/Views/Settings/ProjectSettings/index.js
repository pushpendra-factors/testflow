import React, { useState } from 'react';
import {
  Row, Col, Menu
} from 'antd';
import { Text } from 'factorsComponents';
import ViewProjectSettings from './ViewProjectSettings';
import JavascriptSDK from './JavascriptSDK';

const MenuTabs = {
  generalSettings: 'General Settings',
  SDK: 'Javascript SDK',
  Users: 'Users',
  Integrations: 'Integrations',
  EventAlias: 'Event Alias'
};

function UserSettingsModal() {
  const [selectedMenu, setSelectedMenu] = useState(MenuTabs.generalSettings);
  // const [editPasswordModal, setPasswordModal] = useState(false);
  // const [editDetailsModal, setDetailsModal] = useState(false);
  // const [confirmLoading, setConfirmLoading] = useState(false);

  const handleClick = (e) => {
    setSelectedMenu(e.key);
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
              className={'fa-settings--menu'}
            >
              <Menu.Item key={MenuTabs.generalSettings}>{MenuTabs.generalSettings}</Menu.Item>
              <Menu.Item key={MenuTabs.SDK}>{MenuTabs.SDK}</Menu.Item>
              <Menu.Item key={MenuTabs.Users}>{MenuTabs.Users}</Menu.Item>
              <Menu.Item key={MenuTabs.Integrations}>{MenuTabs.Integrations}</Menu.Item>
              <Menu.Item key={MenuTabs.EventAlias}>{MenuTabs.EventAlias}</Menu.Item>
            </Menu>

          </Col>
          <Col span={15}>

          {selectedMenu === MenuTabs.generalSettings &&
            <ViewProjectSettings
              // editDetails={() => setDetailsModal(true)}
              // editPassword={() => setPasswordModal(true)}
            />
          }
          {selectedMenu === MenuTabs.SDK &&
            <JavascriptSDK />
          }

          </Col>
        </Row>
      </div>

    </>

  );
}

export default UserSettingsModal;
