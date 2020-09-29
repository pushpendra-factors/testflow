import React, { useState, useEffect } from 'react';
import {
  Layout, Row, Avatar, Popover, Button, Modal, Col
} from 'antd';
import { NavLink } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import ModalLib from '../../Views/componentsLib/ModalLib';
import UserSettings from '../../Views/Settings/UserSettings';
import { setActiveProject } from '../../reducers/global';
import { connect } from 'react-redux';

function Sidebar(props) {
  const { Sider } = Layout;

  const [visible, setVisible] = useState(false);
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);

  const popOvercontent = () => {
    return (
        <div className={'fa-popupcard'}>
          <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Projects</Text>
          <div className={'flex flex-col items-start'}>
            {props.projects.map((project, index) => {
              return <Button type={'text'} key={index} onClick={() => { 
                setShowPopOver(false); 
                setchangeProjectModal(true);
                setselectedProject(project);
              }}>{project.name}</Button>;
            })}
          </div>
          <div className={'fa-popupcard-divider'} />
          <Button type={'text'}>{'Add Projects'}</Button>
          <div className={'fa-popupcard-divider'} />
          <Button type={'text'} onClick={() => { setShowPopOver(false); showUserSettingsModal(); }}>{'Account Settings'}</Button>
        </div>
    );
  };

  const showUserSettingsModal = () => {
    setShowUserSettings(true);
  };
  const closeUserSettingsModal = () => {
    setShowUserSettings(false);
  };

  const handleCancel = () => {
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
              <Popover placement="leftTop" title={false} content={popOvercontent} visible={ShowPopOver} onClick={() => setShowPopOver(true)} trigger="click">
                {/* <Link to={'#'} onClick={() => showUserSettingsModal()} > */}
                  <Avatar
                    src={'https://zos.alipayobjects.com/rmsportal/ODTLcjxAfvqbxHnVXCYX.png'}
                    className={'flex justify-center items-center fa-aside--avatar'}
                  />
                {/* </Link> */}
              </Popover>
            </Row>
          </div>
        </div>

        {/* Popover */}

        {/* Modals triggered from sidebar */}
        <ModalLib visible={visible} handleCancel={handleCancel} />
        <UserSettings visible={ShowUserSettings} handleCancel={closeUserSettingsModal} />

        <Modal
        visible={changeProjectModal}
        zIndex={1020}
        onCancel={()=>{
          setchangeProjectModal(false);
          setselectedProject(null);
        }}
        className={'fa-modal--regular'}
        okText={'Switch'}
        onOk={()=> {
          props.setActiveProject(selectedProject);
          setShowPopOver(false);
          setchangeProjectModal(false);
          setselectedProject(null);
        }}
        // confirmLoading={props.confirmLoading}
        centered={true}
        >
          <div className={'p-4'}>
            <Row>
              <Col span={24}>
                <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>Do you want to switch the project?</Text>
                <Text type={'title'} level={7}  color={'grey'} extraClass={'m-0 mt-2'}>You can easily switch between projects. You will be redirected a different dataset.</Text>
              </Col>
            </Row> 
          </div>

        </Modal>

      </Sider>
    </>
  );
}
const mapStateToProps = (state) => { 
  return {
    projects: state.global.projects
  };
};
export default connect(mapStateToProps, { setActiveProject })(Sidebar);
