import React, { useState } from 'react';
import {
  Row, Col, Modal, Button, Menu
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import EditPassword from './EditPassword';
import EditUserDetails from './EditUserDetails';
import ViewUserDetails from './ViewUserDetails';
import ProjectDetails from './ProjectDetails';
import { connect } from 'react-redux';

// const { SubMenu } = Menu;

const MenuTabs = {
  projects: 'projects',
  accounts: 'account'
};

function UserSettings(props) {
  const [selectedMenu, setSelectedMenu] = useState(MenuTabs.accounts);
  const [editPasswordModal, setPasswordModal] = useState(false);
  const [editDetailsModal, setDetailsModal] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const { agent } = props;

  const handleClick = (e) => {
    setSelectedMenu(e.key);
    // console.log('click ', e.key);
  };

  const handleOk = () => {
    setConfirmLoading(true);
    setTimeout(() => {
      setConfirmLoading(false);
      setPasswordModal(false);
      setDetailsModal(false);
    }, 2000);
  };

  return (
    <>

      <Modal
        title={null}
        visible={props.visible}
        footer={null}
        centered={false}
        // zIndex={1005}
        mask={false}
        closable={false}
        className={'fa-modal--full-width'}
      >

        <div className={'fa-modal--header'}>
          <div className={'fa-container'}>
            <Row justify={'space-between'} className={'py-4 m-0 '}>
              <Col>
                <SVG name={'brand'} size={40}/>
              </Col>
              <Col>
                <Button size={'large'} type="text" onClick={() => props.handleCancel()}><SVG name="times"></SVG></Button>
              </Col>
            </Row>
          </div>
        </div>

        <div className={'fa-container'}>
          <Row gutter={[24, 24]} justify={'center'} className={'pt-4 pb-2 m-0 '}>
            <Col span={20}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>My Account Details</Text>
              <Text type={'title'} level={6} weight={'regular'} extraClass={'m-0'} color={'grey'}>{agent ? `${agent?.first_name} ${agent?.last_name} (${agent.email})` : ''}</Text>
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify={'center'}>
            <Col span={5}>

              <Menu
                onClick={handleClick}
                defaultSelectedKeys={MenuTabs.accounts}
                mode="inline"
                className={'fa-settings--menu'}
              >
                <Menu.Item key={MenuTabs.accounts}>My Profile</Menu.Item>
                <Menu.Item key={MenuTabs.projects}>Projects</Menu.Item>
              </Menu>

            </Col>
            <Col span={15}>

              {(selectedMenu === MenuTabs.accounts) &&
                  <ViewUserDetails
                    editDetails={() => setDetailsModal(true)}
                    editPassword={() => setPasswordModal(true)}
                  />
              }
              {(selectedMenu === MenuTabs.projects) &&
                  <ProjectDetails />
              }

            </Col>
          </Row>
        </div>

      </Modal>

      <EditPassword
        visible={editPasswordModal}
        onCancel={() => setPasswordModal(false)}
        onOk={() => handleOk()}
        confirmLoading={confirmLoading}
      />

      <EditUserDetails
        visible={editDetailsModal}
        zIndex={1020}
        onCancel={() => setDetailsModal(false)}
        onOk={() => handleOk()}
        confirmLoading={confirmLoading}
      />

    </>

  );
}

const mapStateToProps = (state) => {
  return ({
    agent: state.agent.agent_details
  }
  );
};

export default connect(mapStateToProps)(UserSettings);
