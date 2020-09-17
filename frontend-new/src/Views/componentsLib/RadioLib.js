/* eslint-disable */
import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider, Radio  } from 'antd';  



class SwitchLib extends React.Component {
    state = { visible: false };

    onChange = e => {
    console.log('radio checked', e.target.value);
    this.setState({
        value: e.target.value,
    });
    };

      render(){
          return ( 
         <>
                            <div className="mt-20 mb-8">
                                <Divider orientation="left">
                                    <Breadcrumb>  
                                        <Breadcrumb.Item> Components </Breadcrumb.Item> 
                                        <Breadcrumb.Item> Radio </Breadcrumb.Item> 
                                    </Breadcrumb> 
                                </Divider> 
                            </div>
         
                            <Row> 
                                <Col span={4}>
                                    <Radio checked></Radio>
                                </Col>
                                <Col span={4}>
                                    <Radio checked>Radio</Radio>
                                </Col>
                                <Col span={10}>
                                <Radio.Group onChange={this.onChange} value={this.state.value}>
                                    <Radio value={1}>A</Radio>
                                    <Radio value={2}>B</Radio>
                                    <Radio value={3}>C</Radio>
                                    <Radio value={4}>D</Radio>
                                </Radio.Group>
                                </Col>
        
                                 
                                
                            </Row> 
        
         
        
        </>
        
          );

      }
}

export default SwitchLib;
