import React from 'react';
import { Text } from 'factorsComponents';
import {
  Row, Col
} from 'antd';

function AddWidgets() {
  return (<>
         <Row className={'py-20'} justify={'center'} gutter={[24, 24]}>
            <Col span={18}>
                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Widgets make up your dashboard and are created from saved queries You must create a query first and save it before you can add it here. <a>Learn more.</a></Text>
            </Col>
        </Row>

  </>);
}

export default AddWidgets;
