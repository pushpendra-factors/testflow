import React, {useState, useEffect} from 'react';
import { Row, Col, Modal, Button, Menu, Avatar  } from 'antd';  
import {Text, SVG} from 'factorsComponents';   
import { UserOutlined } from '@ant-design/icons';

const { SubMenu } = Menu;


function UserSettingsModal (props){  

  const [visibleBorder, setVisibleBorder] = useState(false);   
    
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
            <Row justify={'space-between'} className={`py-6 m-0 `}>
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
          <Row gutter={[24, 24]} justify={'center'} className={`pt-8 pb-8 m-0 `}>
              <Col span={20}>
                <Text type={'title'} level={2} weight={'bold'} extraClass={`m-0`}>My Account Details</Text>  
                <Text type={'title'} level={7} weight={'regular'} extraClass={`m-0`} color={'grey'}>Jeff Richards (jeff@example.com)</Text>  
              </Col> 
          </Row>
          <Row gutter={[24, 24]} justify={'center'}>
              <Col span={5}> 

              <Menu 
                style={{ width: 256 }}
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
              <Col span={15} className={`mb-10`}>
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
                    <Button>Edit Details</Button>   
                    <Button className={'ml-4'}>Change Password</Button> 
                  </Col>
                </Row>
              </Col> 
          </Row>
        </div>

          
        </Modal>
      </>
      
    );
  
}

export default UserSettingsModal