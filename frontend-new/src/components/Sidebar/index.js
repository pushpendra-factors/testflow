import React, {useState, useEffect} from 'react';
import { Layout, Menu, Row, Col } from 'antd';
import styles from './index.module.scss';
import { NavLink } from 'react-router-dom'; 
import { SVG} from 'factorsComponents';
import ModalLib from '../../Views/componentsLib/ModalLib';

function Sidebar() {
  const { Sider } = Layout;

  const [visible, setVisible] = useState(false); 

  const showModal = () => {
      setVisible(true)
  };
 
  const handleCancel = e => { 
      setVisible(false)
  };

      useEffect(() => {
            document.onkeydown = keydown;  
            function keydown (evt) {  
                  if (!evt) evt = event;
                  //Shift+G to trigger grid debugger
                  if (evt.shiftKey && evt.keyCode === 71) { setVisible(true); }  
            } 
      });


  return (
    <>
    <Sider className="fa-aside" width={`64`} >

      <Row justify="center" align="middle" className="py-5"> 
            <NavLink className="active fa-brand-logo" exact to="/"><SVG name={'brand'} size={40} color="white"/></NavLink>  
      </Row>
      <Row justify="center" align="middle" className="pb-2"> 
            <div className={`fa-aside--divider`} />
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" exact to="/"><SVG name={'home'} size={24} color="white"/></NavLink>  
      </Row>
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" disabled exact to="/core-query"><SVG name={'corequery'} size={24} color="white"/></NavLink> 
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" disabled exact to="/key"><SVG name={'key'} size={24} color="white"/></NavLink> 
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" disabled exact to="/bug"><SVG name={'bug'} size={24} color="white"/></NavLink> 
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" disabled exact to="/report"><SVG name={'report'} size={24} color="white"/></NavLink> 
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active" disabled exact to="/notify"><SVG name={'notify'} size={24} color="white"/></NavLink> 
      </Row> 
      <Row justify="center" align="middle" className="py-2"> 
            <NavLink activeClassName="active"   exact to="/components"><SVG name={'hexagon'} size={24} color="white"/></NavLink> 
      </Row> 

        
      <ModalLib visible={visible} handleCancel={handleCancel} />
  
    </Sider> 
    </>
  )
}

export default Sidebar;