import React from 'react';
import JavascriptSDK from './JavascriptSDK';
import { Row, Col } from 'antd';

function SDKSettings() {
  return (
    <div className={'fa-container'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <JavascriptSDK />
        </Col>
      </Row>
    </div>
  );
}

export default SDKSettings;
