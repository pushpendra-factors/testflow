import React from 'react';
import { Layout, Menu, Row, Col } from 'antd';
import styles from './index.module.scss';
import { NavLink } from 'react-router-dom'; 
import { SVG} from 'factorsComponents';

function Sidebar() {
  const { Sider } = Layout;

  return (
    <>
    <Sider className="fa-aside" width={`64`} >

      <Row justify="center" align="middle" className="py-8"> 
            <NavLink className="active" exact to="/"><SVG name={'brand'} size={32} color="white"/></NavLink>  
      </Row>

      <Row justify="center" align="middle" className="py-4"> 
            <NavLink activeClassName="active" exact to="/"><SVG name={'home'} size={24} color="white"/></NavLink>  
      </Row>
      <Row justify="center" align="middle" className="py-4"> 
            <NavLink activeClassName="active" to="/components/"><SVG name={'corequery'} size={24} color="white"/></NavLink> 
      </Row> 

        
  
    </Sider> 
    </>
  )
}

export default Sidebar;