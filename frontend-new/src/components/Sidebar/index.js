import React from 'react';
import { Layout, Menu, Row, Col } from 'antd';
import styles from './index.module.scss';
import { Link } from 'react-router-dom'; 


function Sidebar() {
  const { Sider } = Layout;

  return (
    <>
    <Sider className="fa-aside" width={`64`} >
      <div className={styles.logo} >
        <img src="./assets/icons/factors.png" alt="Factors.ai" />
      </div> 

      <Row justify="center" align="middle" className="py-4"> 
            <Link to="/"><img className="anticon" src="./assets/icons/home.svg" alt="Home" /></Link>  
      </Row>
      <Row justify="center" align="middle" className="py-4"> 
            <Link to="/components/"><img className="anticon" src="./assets/icons/core-query-white.png" /></Link> 
      </Row>

        
  
    </Sider> 
    </>
  )
}

export default Sidebar;