import React, { useState, useEffect } from 'react';
import {
  Row, Col, Tabs
} from 'antd';
// import { PlusOutlined, EditOutlined } from '@ant-design/icons';
import WidgetCard from './WidgetCard';
const { TabPane } = Tabs;

// const operations = <>
//   <Button type="text" icon={<PlusOutlined />}/>
//   <Button type="text" icon={<EditOutlined />}/>
// </>;

function ProjectTabs() {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 200);
  });

  return (<>

        <Row className={'mt-2'}>
          <Col span={24}>
            { dataLoading ? <></>
              : <Tabs defaultActiveKey="1"
              className={'fa-tabs--dashboard'}
            //   tabBarExtraContent={operations}
              >
                <TabPane tab="My Dashboard" key="1">
                   <div className={'fa-container'}>
                        <div className={'py-4 flex justify-center flex-wrap'}>
                            <WidgetCard id={1}/>
                            <WidgetCard id={2}/>
                            <WidgetCard id={3}/>
                            <WidgetCard id={1}/>
                        </div>
                   </div>
                </TabPane>
                <TabPane tab="Paid Marketing" key="2">
                    <div className={'fa-container'}>
                        <div className={'py-4 flex justify-center flex-wrap'}>
                            <WidgetCard id={1}/>
                            <WidgetCard id={3}/>
                        </div>
                   </div>
                </TabPane>
                <TabPane tab="Campaigns" key="3">
                    <div className={'fa-container'}>
                        <div className={'py-4 flex justify-center flex-wrap'}>
                            <WidgetCard id={3}/>
                            <WidgetCard id={2}/>
                        </div>
                   </div>
                </TabPane>
              </Tabs>
            }
          </Col>
        </Row>

  </>);
}

export default ProjectTabs;
