import React, { useState, useEffect } from 'react';
import {
  Layout, Row, Avatar, Popover, Button, Modal, Col, notification, Tooltip, Tag, Badge
} from 'antd';
import { NavLink, useHistory } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import ModalLib from '../../Views/componentsLib/ModalLib';
import UserSettings from '../../Views/Settings/UserSettings';
import { setActiveProject } from 'Reducers/global';
import { signout } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import { PlusOutlined, PoweroffOutlined, BankOutlined } from '@ant-design/icons';
import CreateNewProject from './CreateNewProject';
import _ from 'lodash';
import NewProject from '../../Views/Settings/SetupAssist/Modals/NewProject';


// const ColorCollection = ['#4C9FC8','#4CBCBD', '#86D3A3', '#F9C06E', '#E89E7B', '#9982B5'];

function Sidebar(props) {
  const { Sider } = Layout;

  const [visible, setVisible] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(true);
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);
  const [searchProjectName, setsearchProjectName] = useState('');
  const [CreateNewProjectModal, setCreateNewProjectModal] = useState(false);
  const history = useHistory();

  const searchProject = (e) => {
    setsearchProjectName(e.target.value);
  };

  const popOvercontent = () => {
    return (
        <div className={'fa-popupcard'}>
          <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Projects</Text>
          {props.projects.length > 6 ? <input onChange={(e) => searchProject(e)} value={searchProjectName} placeholder={'Search Project'} className={'fa-project-list--search'}/> : null}
          <div className={'flex flex-col items-start fa-project-list--wrapper'} >
            {props.projects.filter(project => project.name.toLowerCase().includes(searchProjectName.toLowerCase())).map((project, index) => {
              return <div key={index}
              className={`flex justify-start items-center project-item ${props.active_project.id === project.id ? 'active' : null}`}
              onClick={() => {
                if(props.active_project.id!==project.id){
                  setShowPopOver(false);
                  setchangeProjectModal(true);
                  setselectedProject(project); 
                }
              }}>
                <Avatar size={28} style={{ color: '#fff', backgroundColor: '#52BE95', fontSize: '14px', textTransform: 'uppercase', fontWeight:'400' }}>{`${project?.name?.charAt(0)}`}</Avatar>
                <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 ml-2'}>{project.name}</Text>
              </div>;
            })}

          </div>
          <div className={'fa-popupcard-divider'} />
          <Button size={'large'} type={'text'}
          onClick={() => {
            setShowPopOver(false);
            setCreateNewProjectModal(true);
          }}>
            <span className={'mr-4'}><PlusOutlined /></span> {'Add Projects'}</Button>
          <div className={'fa-popupcard-divider'} />
          <div className={'flex justify-start items-center project-item'}
              onClick={() => {
                setShowPopOver(false);
                showUserSettingsModal();
              }}>
                <Avatar size={28}style={{ color: '#f56a00', backgroundColor: '#fde3cf', fontSize: '12px' }}
                  >{`${props.currentAgent?.first_name?.charAt(0)}${props.currentAgent?.last_name?.charAt(0)}`}</Avatar>
                <Text type={'title'} level={7} extraClass={'m-0 ml-2'}>{'Account Settings'}</Text>
          </div>
          <Button size={'large'} type={'text'}
          onClick={() => {
            setShowPopOver(false);
            props.signout();
          }}>
            <span className={'mr-4'}><PoweroffOutlined /></span> {'Logout'}</Button>

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

  const switchProject = () => {  
    localStorage.setItem('activeProject', selectedProject?.id);
    props.setActiveProject(selectedProject);
    history.push('/');
    notification.success({
      message: 'Project Changed!',
      description: `You are currently viewing data from ${selectedProject.name}`
    });
  }; 

  return (
    <>
      <Sider className="fa-aside" width={'64'} >

        <div className={'flex flex-col h-full justify-between items-center w-full'}>
          <div className={'flex flex-col justify-start items-center w-full '}>
            <Row justify="center" align="middle" className=" w-full py-5 relative">
              <NavLink className="active fa-brand-logo" exact to="/"><SVG name={'brand'} size={40} color="white"/></NavLink>
              {/* <Tag color="gold" className={'fa-tag--beta'}>BETA</Tag> */}
            </Row>
            <Row justify="center" align="middle" className=" w-full pb-2">
              <div className={'fa-aside--divider'} />
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <Tooltip title="Dashboard" placement="right" overlayStyle={{paddingLeft:'12px'}} arrowPointAtCenter={true} mouseEnterDelay={0.3}>
              <NavLink activeClassName="active" exact to="/"><SVG name={'dashboard'} size={24} color="white"/></NavLink>
              </Tooltip>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <Tooltip title="Analyse" placement="right" overlayStyle={{paddingLeft:'12px'}} arrowPointAtCenter={true} mouseEnterDelay={0.3}>
                <NavLink activeClassName="active" exact to="/analyse"><SVG name={'corequery'} size={24} color="white"/></NavLink>
              </Tooltip>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <Tooltip title="Explain" placement="right" overlayStyle={{paddingLeft:'12px'}} arrowPointAtCenter={true} mouseEnterDelay={0.3}>
                <NavLink activeClassName="active" to="/explain"><SVG name={'key'} size={24} color="white"/></NavLink> 
              </Tooltip>
            </Row>
            {/* <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/bug"><SVG name={'bug'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/report"><SVG name={'report'} size={24} color="white"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" exact to="/components"><SVG name={'notify'} size={24} color="white"/></NavLink>
            </Row> */}
            <Row justify="center" align="middle" className=" w-full py-2">
              <Tooltip title="Settings" placement="right" overlayStyle={{paddingLeft:'12px'}} arrowPointAtCenter={true} mouseEnterDelay={0.3}>
                <NavLink activeClassName="active" to="/settings"><SVG name={'hexagon'} size={24} color="white"/></NavLink>
              </Tooltip>
            </Row>
            <Row justify="center" align="middle" style={{marginTop:'50vh'}} className=" w-full py-2">
              <Tooltip title="Setup Assist" placement="right" overlayStyle={{paddingLeft:'12px'}} arrowPointAtCenter={true} mouseEnterDelay={0.3}>
                <NavLink activeClassName="active" to="/project-setup"><SVG name={'Emoji'} size={40} color="white"/></NavLink>
                <Badge dot offset={[25,-35]}></Badge>
              </Tooltip>
            </Row>
          </div>
          <div className={'flex flex-col justify-end items-center w-full pb-8 pt-2'}>
            <Row justify="center" align="middle" className=" w-full py-2">
              <Popover placement="top" overlayClassName={'fa-popupcard--wrapper'} title={false}
              content={popOvercontent}
              visible={ShowPopOver}
              onVisibleChange={(visible) => {
                setShowPopOver(visible);
              }}
              onClick={() => {
                setsearchProjectName('');
                setShowPopOver(true);
              }}
                trigger="click"> 
                  <Avatar shape={'square'}  className={'flex justify-center flex-col items-center fa-aside--avatar'} style={{ color: '#fff', backgroundColor: '#52BE95', fontSize: '16px', textTransform: 'uppercase', fontWeight:'400' }}>{`${props.active_project?.name?.charAt(0)}`}</Avatar>
              </Popover>
            </Row>
          </div>
        </div>

        {/* Popover */}

        {/* Modals triggered from sidebar */}
        <ModalLib visible={visible} handleCancel={handleCancel} />
        <UserSettings visible={ShowUserSettings} handleCancel={closeUserSettingsModal} />
        <NewProject visible={showProjectModal} handleCancel={() => setShowProjectModal(false)} />

        <CreateNewProject
          visible={CreateNewProjectModal}
          setCreateNewProjectModal={setCreateNewProjectModal}
        />

        <Modal
        visible={changeProjectModal}
        zIndex={1020}
        onCancel={() => {
          setchangeProjectModal(false);
          setselectedProject(null);
        }}
        className={'fa-modal--regular fa-modal--slideInDown'}
        okText={'Switch'}
        onOk={() => {
          setShowPopOver(false);
          setchangeProjectModal(false);
          setselectedProject(null);
          switchProject();
        }}
        centered={true}
        transitionName=""
        maskTransitionName=""
        >
          <div className={'p-4'}>
            <Row>
              <Col span={24}>
                <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>Do you want to switch the project?</Text>
                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>You can easily switch between projects. You will be redirected a different dataset.</Text>
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
    projects: state.global.projects,
    active_project: state.global.active_project,
    currentAgent: state.agent.agent_details
  };
};
export default connect(mapStateToProps, { setActiveProject, signout })(Sidebar);
