import React, { useState } from 'react';
import { Layout, Row, Tooltip } from 'antd';
import { NavLink, useHistory } from 'react-router-dom';
import { SVG } from 'factorsComponents';
import ModalLib from '../../Views/componentsLib/ModalLib';
import { setActiveProject } from 'Reducers/global';
import { updateAgentInfo, fetchAgentInfo, fetchProjectAgents } from 'Reducers/agentActions';
import { signout } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import _ from 'lodash';
import { useTour } from '@reactour/tour';

// const ColorCollection = ['#4C9FC8','#4CBCBD', '#86D3A3', '#F9C06E', '#E89E7B', '#9982B5'];

function Sidebar(props) {
  const { Sider } = Layout;

  const [visible, setVisible] = useState(false);
  const { isOpen, setIsOpen } = useTour();

  const handleCancel = () => {
    setVisible(false);
  };

  return (
    <>
      <Sider className='fa-aside' width={'64'}>
        <div className={'flex flex-col h-full justify-between items-center w-full'}>
          <div className={'flex flex-col justify-start items-center w-full '}>
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Dashboard'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink
                  data-tour='step-1'
                  isOpen={isOpen}
                  onClick={() => setIsOpen(!isOpen)}
                  activeClassName='active'
                  exact
                  to='/'
                >
                  <SVG name={'dashboard'} size={24} color='#0E2647' />
                </NavLink>
              </Tooltip>
            </Row>
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Analyse'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink data-tour='step-5' activeClassName='active' exact to='/analyse'>
                  <SVG name={'corequery'} size={24} color='#0E2647' />
                </NavLink>
              </Tooltip>
            </Row>
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Explain'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink data-tour='step-6' activeClassName='active' to='/explain'>
                  <SVG name={'key'} size={24} color='#0E2647' />
                </NavLink>
              </Tooltip>
            </Row>
            {/* <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/bug"><SVG name={'bug'} size={24} color="#0E2647"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" disabled exact to="/report"><SVG name={'report'} size={24} color="#0E2647"/></NavLink>
            </Row>
            <Row justify="center" align="middle" className=" w-full py-2">
              <NavLink activeClassName="active" exact to="/components"><SVG name={'notify'} size={24} color="#0E2647"/></NavLink>
            </Row> */}
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Settings'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink data-tour='step-7' activeClassName='active' to='/settings'>
                  <SVG name={'hexagon'} size={24} color='#0E2647' />
                </NavLink>
              </Tooltip>
            </Row>
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Settings'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink activeClassName='active' to='/configure'>
                  <SVG name={'configure'} size={24} color='#0E2647' />
                </NavLink>
              </Tooltip>
            </Row>
          </div>
          <div className={'flex flex-col justify-end items-center w-full pb-8 pt-2'}>
            <Row justify='center' align='middle' className=' w-full py-2'>
              <Tooltip
                title='Setup Assist'
                placement='right'
                overlayStyle={{ paddingLeft: '12px' }}
                arrowPointAtCenter={true}
                mouseEnterDelay={0.3}
              >
                <NavLink activeClassName='active' to='/welcome'>
                  <SVG name={'Emoji'} size={40} color='#0E2647' />
                </NavLink>
                {/* <Badge dot offset={[25,-35]}></Badge> */}
              </Tooltip>
            </Row>
          </div>
        </div>
        <ModalLib visible={visible} handleCancel={handleCancel} />
      </Sider>
    </>
  );
}
const mapStateToProps = (state) => {
  return {
    projects: state.global.projects,
    active_project: state.global.active_project,
    currentAgent: state.agent.agent_details,
    agents: state.agent.agents,
  };
};
export default connect(mapStateToProps, {
  fetchProjectAgents,
  setActiveProject,
  signout,
  updateAgentInfo,
  fetchAgentInfo,
})(Sidebar);
