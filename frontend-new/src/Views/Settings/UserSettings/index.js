import React, {useState, useEffect} from 'react';
import { Row, Col, Modal, Button, Menu, Avatar, Input  } from 'antd';  
import {Text, SVG} from 'factorsComponents';   
import { UserOutlined } from '@ant-design/icons';

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
              <div className={`mb-10 pl-4`}>
                <Row>
                  <Col>
                    <Text type={'title'} level={3} weight={'bold'} extraClass={`m-0`}>Profile</Text>   
                  </Col>
                </Row>
                <Row className={`mt-2`}>
                  <Col>
                    <Avatar shape="square" size={104} icon={<UserOutlined />} />
                    <Text type={'paragraph'} mini extraClass={`m-0 mt-1`} color={'grey'} >A photo helps personalise your account</Text>    
                  </Col>
                </Row>
                <Row className={`mt-6`}> 
                  <Col>
                    <Text type={'title'} level={7} extraClass={`m-0`}>Name</Text>    
                    <Text type={'title'} level={6} extraClass={`m-0`} weight={'bold'}>Vishnu Baliga</Text>    
                  </Col>
                </Row>
                <Row className={`mt-6`}>
                  <Col>
                    <Text type={'title'} level={7} extraClass={`m-0`}>Email</Text>    
                    <Text type={'title'} level={6} extraClass={`m-0`} weight={'bold'}>baliga@factors.ai</Text>    
                  </Col>
                </Row>
                <Row className={`mt-6`}>
                  <Col>
                    <Text type={'title'} level={7} extraClass={`m-0`}>Mobile</Text>    
                    <Text type={'title'} level={6} extraClass={`m-0`} weight={'bold'}>+91-8123456789</Text>    
                  </Col>
                </Row>
                <Row className={`mt-6`}>
                  <Col>
                    <Text type={'title'} level={7} extraClass={`m-0`}>Password</Text>    
                    <Text type={'title'} level={6} extraClass={`m-0`} weight={'bold'}>&#8226; &#8226; &#8226; &#8226; &#8226; &#8226;</Text>    
                  </Col>
                </Row>
                <Row className={`mt-6`}>
                  <Col className={`flex justify-start items-center`}>
                    <Button onClick={()=>setDetailsModal(true)}>Edit Details</Button>
                    <Button className={'ml-4'} onClick={()=>setPasswordModal(true)} >Change Password</Button> 
                  </Col>
                </Row>
              </div> 
              </Col> 
          </Row>
        </div>

          
        </Modal>


        
        <Modal 
          visible={editPasswordModal}
          zIndex={1020}
          onCancel={()=>setPasswordModal(false)}
          className={`fa-modal--regular`}
          okText={`Update Password`}
          onOk={()=>handleOk()}
          confirmLoading={confirmLoading}
          centered={true}
        >
          <div className={`p-4`}> 
            <Row>
              <Col span={24}>
                <Text type={'title'} level={3} weight={'bold'} extraClass={`m-0`}>Change Password</Text>   
              </Col>
            </Row>
            <Row className={`mt-6`}>
              <Col span={24}>
                <Text type={'title'} level={7} extraClass={`m-0`}>Old Password</Text>    
                <Input  size="large" className={`fa-input w-full`} placeholder="Old Password" />
              </Col>
            </Row>
            <Row className={`mt-6`}>
                <Col span={24}>
                  <Text type={'title'} level={7} extraClass={`m-0`}>New Password</Text>    
                  <Input  size="large" className={`fa-input w-full`} placeholder="New Password" />
                </Col>
            </Row>
            <Row className={`mt-6`}>
                <Col span={24}>
                  <Text type={'title'} level={7} extraClass={`m-0`}>Confirm Password</Text>    
                  <Input  size="large" className={`fa-input w-full`} placeholder="Confirm Password" />
                </Col>
            </Row>
          </div>

        </Modal>

        <Modal 
          visible={editDetailsModal}
          zIndex={1020}
          onCancel={()=>setDetailsModal(false)}
          className={`fa-modal--regular`}
          okText={`Update Password`}
          onOk={()=>handleOk()}
          confirmLoading={confirmLoading}
          centered={true}
        >
          <div className={`p-4`}> 
            <Row>
              <Col span={24}>
                <Text type={'title'} level={3} weight={'bold'} extraClass={`m-0`}>Edit Details</Text>   
              </Col>
            </Row>
            <Row className={`mt-6`}>
              <Col span={24}>
                <Text type={'title'} level={7} extraClass={`m-0`}>Name</Text>    
                <Input  size="large" className={`fa-input w-full`} placeholder="Name" />
              </Col>
            </Row>
            <Row className={`mt-6`}>
                <Col span={24}>
                  <Text type={'title'} level={7} extraClass={`m-0`}>Email</Text>    
                  <Input  size="large" className={`fa-input w-full`} placeholder="Email" />
                </Col>
            </Row>
            <Row className={`mt-6`}>
                <Col span={24}>
                  <Text type={'title'} level={7} extraClass={`m-0`}>Phone</Text>    
                  <Input  size="large" className={`fa-input w-full`} placeholder="Phone" />
                </Col>
            </Row>
          </div>

        </Modal>

      </>
      
    );
  
}

export default UserSettingsModal