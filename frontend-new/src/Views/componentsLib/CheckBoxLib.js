import React from 'react';
import {
  Breadcrumb, Row, Col, Divider, Checkbox
} from 'antd';

const plainOptions = ['A', 'B', 'C'];

class CheckBoxLib extends React.Component {
  render() {
    return (
      <>
        <div className="mt-20 mb-8">
          <Divider orientation="left">
            <Breadcrumb>
              <Breadcrumb.Item> Components </Breadcrumb.Item>
              <Breadcrumb.Item> Checkbox </Breadcrumb.Item>
            </Breadcrumb>
          </Divider>
        </div>

        <Row>
          <Col span={4}>
            <Checkbox />
          </Col>
          <Col span={4}>
            <Checkbox>Checkbox</Checkbox>
          </Col>
          <Col span={10}>
            <Checkbox.Group options={plainOptions} defaultValue={['A']} />
          </Col>
        </Row>

      </>

    );
  }
}

export default CheckBoxLib;
