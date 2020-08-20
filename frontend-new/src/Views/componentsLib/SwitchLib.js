import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider, Switch  } from 'antd'; 
import { CloseOutlined, CheckOutlined } from '@ant-design/icons';


function SwitchLib() {
    const { Content } = Layout;
  return ( 
 <>
                    <div className="mt-20 mb-8">
                        <Divider orientation="left">
                            <Breadcrumb>  
                                <Breadcrumb.Item> Components </Breadcrumb.Item> 
                                <Breadcrumb.Item> Switch </Breadcrumb.Item> 
                            </Breadcrumb> 
                        </Divider> 
                    </div>
 
                    <Row> 
                        <Col span={4}>
                            <Switch checkedChildren="On" unCheckedChildren="Off" defaultChecked /> 
                        </Col>
                        <Col span={4}> 
                            <Switch size="small" checkedChildren="On" unCheckedChildren="Off"  />
                        </Col>
                        <Col span={4}> 
                        </Col>
                        
                    </Row> 

 

</>

  );
}

export default SwitchLib;
