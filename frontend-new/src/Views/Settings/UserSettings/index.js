import React from 'react';
import { Row, Col, Modal, Button  } from 'antd';  
import {Text, SVG} from 'factorsComponents'; 

class ModalLib extends React.Component { 
  render() {
    return (
      <>
        
        <Modal
          title={null}
          visible={this.props.visible}
          onCancel={this.props.handleCancel} 
          footer={null} 
          centered={false}
          zIndex={1015}
          mask={false}
          closable={false}
          className={`fa-modal--full-width`}
        > 

        <div className={`fa-container`}>
          <Row gutter={[24, 24]}>
              <Col span={12}>
                <Text type={'title'} level={2} weight={'bold'}>My Account Details</Text> 
            </Col> 
          </Row>
        </div>

          
        </Modal>
      </>
      
    );
  }
}

export default ModalLib