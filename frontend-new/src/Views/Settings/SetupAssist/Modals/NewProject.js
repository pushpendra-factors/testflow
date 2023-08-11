import React, { useState } from 'react';
import {
  Row, Input, Button, Modal, Col, Form, message, Progress
} from 'antd';
import { Text, SVG } from 'factorsComponents';

import { connect } from 'react-redux';
import BasicDetails from './BasicDetails';

// const { SubMenu } = Menu;

function NewProject(props) {

  return (
    <>

      <Modal
        title={null}
        visible={props.visible}
        footer={null}
        centered={false}
        // zIndex={1005}
        mask={false}
        closable={false}
        className={'fa-modal--full-width'}
      >

          <div className={'fa-container mb-10'}>
            <Row justify={'space-between'} className={'py-4 px-16 m-0 mt-8'}>
              <Col>
                <SVG name={'brand'} size={50}/>
                <Text type={'title'} level={4} weight={'bold'} color={'grey-2'} extraClass={'m-0 -mt-10 ml-16'}>Create a New Project</Text>
              </Col>
              <Col>
                {/* <Button size={'large'} type="text" onClick={() => props.handleCancel()}><SVG name="times" size={20}></SVG></Button> */}
              </Col>
            </Row>
          </div>

          <BasicDetails handleCancel = {props.handleCancel}/>

      </Modal>  

    </>

  );
}

const mapStateToProps = (state) => {
  return ({
    agent: state.agent.agent_details
  }
  );
};

export default connect(mapStateToProps)(NewProject);
