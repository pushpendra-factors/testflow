import React from 'react';
import { Layout, Menu, Row, Col } from 'antd';
import styles from './index.module.scss';
import { Link } from 'react-router-dom'; 
import { SVG} from 'factorsComponents';

function Sidebar() {
  const { Sider } = Layout;

  return (
    <>
    <Sider className="fa-aside" width={`64`} >

      <Row justify="center" align="middle" className="py-8"> 
            <Link to="/"><SVG name={'brand'} size={32} color="white"/></Link>  
      </Row>

      <Row justify="center" align="middle" className="py-4"> 
            <Link to="/"><SVG name={'home'} size={24} color="white"/></Link>  
      </Row>
      <Row justify="center" align="middle" className="py-4"> 
            <Link to="/components/"><SVG name={'corequery'} size={24} color="white"/></Link> 
      </Row> 

        
  
    </Sider> 
    </>
  )
}

export default Sidebar;