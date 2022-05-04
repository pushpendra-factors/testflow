/* eslint-disable */
import React, { useState, useEffect } from 'react';
import {
  Layout, Breadcrumb, Row, Col, Divider, Button, Radio
} from 'antd';
import { PoweroffOutlined, DownloadOutlined } from '@ant-design/icons';

function ButtonLib() {
  const { Content } = Layout;
  const [size, setSize] = useState(null); 

  const handleSizeChange = (e) => {
    setSize(e.target.value);
  };


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
          <Button size="large" type="primary" loading={true}>Primary Button</Button>
        </Col>
        <Col span={4}>
          <Button size="large" loading={true} >Default Button</Button>
        </Col>
        {/* <Col span={4}>
          <Button size="large" type="dashed">Dashed Button</Button>
        </Col> */}
        <Col span={4}>
          <Button size="large" type="text" loading={true}>Text Button</Button>
        </Col>
        <Col span={4}>
          <Button size="large" type="link" loading={true}>Secondary Button</Button>
        </Col>
      </Row>
      <Row className={'my-6'}>
        <Col span={4}>
          <Button type="primary">Primary Button</Button>
        </Col>
        <Col span={4}>
          <Button>Default Button</Button>
        </Col>
        {/* <Col span={4}>
          <Button type="dashed">Dashed Button</Button>
        </Col> */}
        <Col span={4}>
          <Button type="text">Text Button</Button>
        </Col>
        <Col span={4}>
          <Button type="link">Secondary Button</Button>
        </Col>
      </Row>

      <Row className={'my-6'}>
        <Col span={4}>
          <Button type="primary" icon={<PoweroffOutlined />} >Primary Button</Button>
        </Col>
        <Col span={4}>
          <Button icon={<PoweroffOutlined />}>Default Button</Button>
        </Col>
        {/* <Col span={4}>
          <Button type="dashed" icon={<PoweroffOutlined />}>Dashed Button</Button>
        </Col> */}
        <Col span={4}>
          <Button type="text" icon={<PoweroffOutlined />}>Text Button</Button>
        </Col>
        <Col span={4}>
          <Button type="link" icon={<PoweroffOutlined />}>Secondary Button</Button>
        </Col>
      </Row>

      <Row className={'my-6'}>
        <Col span={4}>
          <Button size="medium" shape="circle"  type="primary" icon={<PoweroffOutlined />} />
        </Col>
        <Col span={4}>
          <Button size="medium" shape="circle"  icon={<PoweroffOutlined />} />
        </Col>
        {/* <Col span={4}>
          <Button type="dashed" icon={<PoweroffOutlined />}/>
        </Col> */}
        <Col span={4}>
          <Button size="medium"  shape="circle" type="text"icon={<PoweroffOutlined />}/>
        </Col>
        <Col span={4}>
          <Button  size="medium" shape="circle" type="link" icon={<PoweroffOutlined />}/>
        </Col>
      </Row>

      <Row className={'my-6'}>
        <Col span={24}>
            <Radio.Group value={size} onChange={handleSizeChange}>
            <Radio.Button value="large">Large</Radio.Button>
            <Radio.Button value="default">Default</Radio.Button>
            <Radio.Button value="small">Small</Radio.Button>
          </Radio.Group>
          <br />
          <br />
          <Button type="primary" size={size}>
            Primary
          </Button>
          <Button size={size}>Default</Button>
          <Button type="dashed" size={size}>
            Dashed
          </Button>
          <br />
          <Button type="link" size={size}>
            Link
          </Button>
          <br />
          <Button type="primary" icon={<DownloadOutlined />} size={size} />
          <Button type="primary" shape="circle" icon={<DownloadOutlined />} size={size} />
          <Button type="primary" shape="round" icon={<DownloadOutlined />} size={size} />
          <Button type="primary" shape="round" icon={<DownloadOutlined />} size={size}>
            Download
          </Button>
          <Button type="primary" icon={<DownloadOutlined />} size={size}>
            Download
          </Button>
        </Col>
      </Row>

    </>

  );
}

export default ButtonLib;
