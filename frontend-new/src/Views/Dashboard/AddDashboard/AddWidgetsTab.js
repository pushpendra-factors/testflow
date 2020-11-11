import React from 'react';
import { Text } from '../../../components/factorsComponents';
import {
  Row, Col
} from 'antd';

function AddWidgetsTab({ queries }) {

  if (!queries.length) {
    return (
      <Row className={'py-20'} justify={'center'} gutter={[24, 24]}>
        <Col span={18}>
          <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Widgets make up your dashboard and are created from saved queries You must create a query first and save it before you can add it here. <a>Learn more.</a></Text>
        </Col>
      </Row>
    );
  }

  return (
    <div className="widget-selection">
      <div className="mt-3">

      </div>
    </div>
  )

}

export default AddWidgetsTab;
