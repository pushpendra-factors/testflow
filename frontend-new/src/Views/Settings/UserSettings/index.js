import React, {useState, useEffect} from 'react';
import { Row, Col, Modal, Button, Menu, Avatar, Input  } from 'antd';  
import {Text, SVG} from 'factorsComponents';   
import { UserOutlined } from '@ant-design/icons';
import EditPassword from './EditPassword';
import EditUserDetails from './EditUserDetails';
import ViewUserDetails from './ViewUserDetails';

const { SubMenu } = Menu;


function UserSettingsModal (props){  
   
  const [editPasswordModal, setPasswordModal] = useState(false);   
  const [editDetailsModal, setDetailsModal] = useState(false);   
  const [confirmLoading, setConfirmLoading] = useState(false);   
    

  const handleOk = () => {
    setConfirmLoading(true)
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
          zIndex={1015}
          mask={false}
          closable={false}
          className={`fa-modal--full-width`}
        > 

        <div className={`fa-modal--header`}>
          <div className={`fa-container`}>
            <Row justify={'space-between'} className={`py-4 m-0 `}>
                <Col>
                  <SVG name={'brand'} size={40}/>
                </Col> 
                <Col>
                <Button type="text" onClick={()=>props.handleCancel()}><SVG name="times"></SVG></Button>
                </Col> 
            </Row>
          </div>
        </div>

        <div className={`fa-container`}> 
          <Row gutter={[24, 24]} justify={'center'} className={`pt-4 pb-2 m-0 `}>
              <Col span={20}>
                <Text type={'title'} level={2} weight={'bold'} extraClass={`m-0`}>My Account Details</Text>  
                <Text type={'title'} level={7} weight={'regular'} extraClass={`m-0`} color={'grey'}>Jeff Richards (jeff@example.com)</Text>  
              </Col> 
          </Row>
          <Row gutter={[24, 24]} justify={'center'}>
              <Col span={5}> 

              <Menu  
                defaultSelectedKeys={['1']} 
                mode="inline"
                className={`fa-settings--menu`}
              >  
              <Menu.Item key="1">My Profile</Menu.Item> 
              <Menu.Item key="2">Projects</Menu.Item> 
              <Menu.Item key="3">Notifications</Menu.Item> 
              <Menu.Item key="4">Saved for Later</Menu.Item> 
              <Menu.Item key="5">Data and Privacy</Menu.Item> 
              </Menu>

              </Col> 
              <Col span={15}>
                <ViewUserDetails 
                  editDetails={()=>setDetailsModal(true)}
                  editPassword={()=>setPasswordModal(true)}
                />
              </Col> 
          </Row>
        </div>

          
        </Modal>


        
        <EditPassword 
          visible={editPasswordModal}  
          onCancel={()=>setPasswordModal(false)} 
          onOk={()=>handleOk()}
          confirmLoading={confirmLoading} 
        /> 

        <EditUserDetails 
          visible={editDetailsModal}
          zIndex={1020}
          onCancel={()=>setDetailsModal(false)} 
          onOk={()=>handleOk()}
          confirmLoading={confirmLoading} 
        /> 

      </>
      
    );
  
}

export default UserSettingsModal