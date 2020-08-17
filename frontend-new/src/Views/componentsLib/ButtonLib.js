import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider, Button  } from 'antd'; 
import { PoweroffOutlined } from '@ant-design/icons'; 

function ButtonLib() {
    const { Content } = Layout;
  return ( 
 <>
                    <div className="mt-20 mb-8">
                        <Divider orientation="left">
                            <Breadcrumb>  
                                <Breadcrumb.Item> Components </Breadcrumb.Item> 
                                <Breadcrumb.Item> Button </Breadcrumb.Item> 
                            </Breadcrumb> 
                        </Divider> 
                    </div>

                    <Row> 
                        <Col span={4}>
                            <Button type="primary">Primary Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button>Default Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="dashed">Dashed Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="text">Text Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="link">Link Button</Button> 
                        </Col> 
                    </Row>

                    <Row className={`my-6`}> 
                        <Col span={4}>
                            <Button type="primary"      icon={<PoweroffOutlined />} >Primary Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button icon={<PoweroffOutlined />}>Default Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="dashed" icon={<PoweroffOutlined />}>Dashed Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="text" icon={<PoweroffOutlined />}>Text Button</Button> 
                        </Col>
                        <Col span={4}>
                            <Button type="link" icon={<PoweroffOutlined />}>Link Button</Button> 
                        </Col> 
                    </Row>

                    <Row className={`my-6`}> 
                        <Col span={4}>
                            <Button type="primary"      icon={<PoweroffOutlined />} />
                        </Col>
                        <Col span={4}>
                            <Button icon={<PoweroffOutlined />} />
                        </Col>
                        <Col span={4}>
                            <Button type="dashed" icon={<PoweroffOutlined />}/>
                        </Col>
                        <Col span={4}>
                            <Button type="text" icon={<PoweroffOutlined />}/>
                        </Col>
                        <Col span={4}>
                            <Button type="link" icon={<PoweroffOutlined />}/>
                        </Col> 
                    </Row>

 

</>

  );
}

export default ButtonLib;
