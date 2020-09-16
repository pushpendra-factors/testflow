import React, { useState, useEffect } from 'react';
import {
  Layout, Menu, Row, Col, Avatar
} from 'antd';
import { NavLink, Link } from 'react-router-dom';
import { SVG } from 'factorsComponents';
import ModalLib from '../../Views/componentsLib/ModalLib';
import UserSettings from '../../Views/Settings/UserSettings';
import { UserOutlined } from '@ant-design/icons';

function Sidebar() {
  const { Sider } = Layout;

  const [visible, setVisible] = useState(false);
  const [ShowUserSettings, setShowUserSettings] = useState(false);

  const showUserSettingsModal = () => {
    setShowUserSettings(true);
  };
  const closeUserSettingsModal = () => {
    setShowUserSettings(false);
  };

  const handleCancel = e => {
    setVisible(false);
  };

  useEffect(() => {
    document.onkeydown = keydown;
    function keydown(evt) {
      // Shift+G to trigger grid debugger
      if (evt.shiftKey && evt.keyCode === 71) { setVisible(true); }
    }
  });

  return (
    <>
      <Sider className="fa-aside" width={'64'} >

        <div className={'flex flex-col h-full justify-between items-center w-full'}>
          <div className={'flex flex-col justify-start items-center w-full '}>
            <Row justify="center" align="middle" className=" w-full py-5">
              <NavLink className="active fa-brand-logo" exact to="/"><SVG name={'brand'} size={40} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full pb-2">
              <div className={'fa-aside--divider'} />
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" exact to="/"><SVG name={'home'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/core-query"><SVG name={'corequery'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/key"><SVG name={'key'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/bug"><SVG name={'bug'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/report"><SVG name={'report'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" exact to="/components"><SVG name={'notify'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" to="/settings"><SVG name={'hexagon'} size={24} color="white"/></NavLink>
            </Row>

          </div>
          <div className={'flex flex-col justify-end items-center w-full pb-8 pt-2'}>
            <Row justify="center" align="middle" className=" w-full py-2">
              <Link to={'#'} onClick={() => showUserSettingsModal()} >
                <Avatar
                  src={'https://zos.alipayobjects.com/rmsportal/ODTLcjxAfvqbxHnVXCYX.png'}
                  className={'flex justify-center items-center fa-aside--avatar'}
                />
              </Link>
            </Row>
          </div>
        </div>

        {/* Modals triggered from sidebar */}
        <ModalLib visible={visible} handleCancel={handleCancel} />
        <UserSettings visible={ShowUserSettings} handleCancel={closeUserSettingsModal} />
      </Sider>
    </>
  );
}

export default Sidebar;
