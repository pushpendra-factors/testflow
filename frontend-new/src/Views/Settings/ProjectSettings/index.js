import React, { useEffect, useState } from 'react';
import { Row, Col, Menu } from 'antd';
import BasicSettings from './BasicSettings';
import SDKSettings from './SDKSettings';
import UserSettings from './UserSettings';
import IntegrationSettings from './IntegrationSettings';
import Events from './Events';
import Properties from './PropertySettings';
import { fetchSmartEvents } from 'Reducers/events';
import { connect } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import ContentGroups from './ContentGroups';
import Touchpoints from './Touchpoints';
import CustomKPI from './CustomKPI';
import Attributions from './Attributions';
import ExplainDP from './ExplainDataPoints';

const MenuTabs = {
  generalSettings: 'General Settings',
  SDK: 'Javascript SDK',
  Users: 'Users',
  Integrations: 'Integrations',
  EventAlias: 'Event Alias',
  Events: 'Events',
  Properties: 'Properties',
  MarketingInteractions: 'Marketing Touchpoints',
  ContentGroups: 'Content Groups',
  Touchpoints: 'Touchpoints',
  Attributions: 'Attributions',
  CustomKPI: 'Custom KPIs',
  ExplainDP: 'Top Events and Properties',
};

function ProjectSettings({ activeProject, fetchSmartEvents }) {
  const [selectedMenu, setSelectedMenu] = useState(MenuTabs.generalSettings);
  const history = useHistory();
  let location = useLocation();

  const handleClick = (e) => {
    setSelectedMenu(e.key);
    if (e.key == 'Integrations') {
      history.push(`/settings/#${e.key.toLowerCase()}`);
    } else {
      history.push(`/settings`);
    }

    if (e.key === MenuTabs.Events) {
      fetchSmartEvents(activeProject.id);
    }
  };

  useEffect(() => {
    if (location.hash == '#integrations') {
      setSelectedMenu(MenuTabs.Integrations);
    }
  }, []);

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Settings Error'}
            subtitle={
              'We are facing trouble loading project settings. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        <div className={'ml-12 mt-24'}>
          <Row gutter={[24, 24]}>
            <Col span={20} className={`px-24`}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
                Project Settings
              </Text>
              <Text
                type={'title'}
                level={6}
                weight={'regular'}
                extraClass={'m-0'}
                color={'grey'}
              >
                {activeProject.name}
              </Text>
            </Col>
          </Row>
          <Row gutter={[24, 24]}>
            <Col span={6}>
              <Menu
                onClick={handleClick}
                defaultSelectedKeys={selectedMenu}
                mode='inline'
                className={'fa-settings--menu'}
              >
                <Menu.Item key={MenuTabs.generalSettings}>
                  {MenuTabs.generalSettings}
                </Menu.Item>
                <Menu.Item key={MenuTabs.SDK}>{MenuTabs.SDK}</Menu.Item>
                <Menu.Item key={MenuTabs.Users}>{MenuTabs.Users}</Menu.Item>
                <Menu.Item data-tour='step-10' key={MenuTabs.Integrations}>
                  {MenuTabs.Integrations}
                </Menu.Item>
                <Menu.Item key={MenuTabs.Touchpoints}>
                  {MenuTabs.Touchpoints}
                </Menu.Item>
                <Menu.Item key={MenuTabs.Attributions}>
                  {MenuTabs.Attributions}
                </Menu.Item>
                <Menu.Item key={MenuTabs.Events}>{MenuTabs.Events}</Menu.Item>
                <Menu.Item key={MenuTabs.Properties}>
                  {MenuTabs.Properties}
                </Menu.Item>
                <Menu.Item key={MenuTabs.ContentGroups}>
                  {MenuTabs.ContentGroups}
                </Menu.Item>
                <Menu.Item key={MenuTabs.CustomKPI}>
                  {MenuTabs.CustomKPI}
                </Menu.Item>
                <Menu.Item key={MenuTabs.ExplainDP}>
                  {MenuTabs.ExplainDP}
                </Menu.Item>
              </Menu>
            </Col>
            <Col span={15}>
              {selectedMenu === MenuTabs.generalSettings && <BasicSettings />}
              {selectedMenu === MenuTabs.SDK && <SDKSettings />}
              {selectedMenu === MenuTabs.Users && <UserSettings />}
              {selectedMenu === MenuTabs.Integrations && (
                <IntegrationSettings />
              )}
              {selectedMenu == MenuTabs.Touchpoints && <Touchpoints />}
              {selectedMenu == MenuTabs.Attributions && <Attributions />}
              {selectedMenu === MenuTabs.Events && <Events />}
              {selectedMenu === MenuTabs.Properties && <Properties />}
              {selectedMenu === MenuTabs.ContentGroups && <ContentGroups />}
              {selectedMenu === MenuTabs.CustomKPI && <CustomKPI />}
              {selectedMenu === MenuTabs.ExplainDP && <ExplainDP />}
            </Col>
          </Row>
        </div>
      </ErrorBoundary>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

export default connect(mapStateToProps, { fetchSmartEvents })(ProjectSettings);
