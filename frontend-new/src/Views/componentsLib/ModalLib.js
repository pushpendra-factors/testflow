/* eslint-disable */
import React from 'react';
import { Breadcrumb, Row, Col, Divider, Modal, Button  } from 'antd';  

class ModalLib extends React.Component {
  state = { visible: false };

  showModal = () => {
    this.setState({
      visible: true,
    });
  };

  handleOk = e => {
    console.log(e);
    this.setState({
      visible: false,
    });
  };

  handleCancel = e => { 
    this.setState({
      visible: false,
    });
  };

  render() {
    return (
      <>

{/*       
<div className="mt-20 mb-8">
                                <Divider orientation="left">
                                    <Breadcrumb>  
                                        <Breadcrumb.Item> Components </Breadcrumb.Item> 
                                        <Breadcrumb.Item> Grid </Breadcrumb.Item> 
                                    </Breadcrumb> 
                                </Divider> 
                            </div>
         
                            <Row> 
                                <Col span={4}>
                                <Button type="primary" onClick={this.showModal}> Open Modal </Button>
                                </Col>  
                            </Row>  */}


        
        <Modal
          title="Grid Debugger"
          visible={this.props.visible}
          onCancel={this.props.handleCancel}
        //   onOk={this.handleOk}
        //   onCancel={this.handleCancel}
          footer={null} 
          centered={false}
          zIndex={1010}
          mask={false}
          className={`fa-demo-grid--modal`}
        >
 
        <div className={`fa-container`}>
          <Row gutter={[24, 24]}>
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
              <Col className={`fa-demo-grid--col`} span={2}><div className={`fa-demo-grid--col-content`} /></Col>
              <Col className={`fa-demo-grid--col even`} span={2}><div className={`fa-demo-grid--col-content`} /></Col> 
          </Row>
        </div>

          
        </Modal>
      </>
      
    );
  }
}

export default ModalLib