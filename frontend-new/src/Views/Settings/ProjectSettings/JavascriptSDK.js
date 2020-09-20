import React, { useState, useEffect } from 'react';
import {
  Row, Col, Skeleton, Tabs
} from 'antd';
import { Text } from 'factorsComponents';
const { TabPane } = Tabs;

function EditUserDetails() {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 2000);
  });

  const callback = (key) => {
    console.log(key);
  };

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Javascript SDK</Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={24}>
            { dataLoading ? <Skeleton active paragraph={{ rows: 4 }}/>
              : <Tabs defaultActiveKey="1" onChange={callback}>
                <TabPane tab="Setup" key="1">
                  Setup content comes here..
                </TabPane>
                <TabPane tab="Configuration" key="2">
                  Configuration content comes here..
                </TabPane>
              </Tabs>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}

export default EditUserDetails;
