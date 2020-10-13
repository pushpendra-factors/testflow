import React, { useState, useEffect } from 'react';
import {
  Row, Col, Tabs, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import WidgetCard from './WidgetCard';
const { TabPane } = Tabs;

function ProjectTabs({ setaddDashboardModal }) {
  const [dataLoading, setDataLoading] = useState(true);
  const [widgetModal, setwidgetModal] = useState(false);

  const operations = <>
  <Button type="text" size={'small'} onClick={() => setaddDashboardModal(true)}><SVG name="plus" color={'grey'}/></Button>
  <Button type="text" size={'small'}><SVG name="edit" color={'grey'}/></Button>
  </>;

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
              tabBarExtraContent={operations}
              >
                <TabPane tab="My Dashboard" key="1">
                   <div className={'fa-container mt-6'}>
                          <Row gutter={[24, 24]} className={'flex justify-start items-stretch'}>
                              <WidgetCard setwidgetModal={setwidgetModal} widthSize={1} id={1}/>
                              <WidgetCard setwidgetModal={setwidgetModal} id={2}/>
                              <WidgetCard setwidgetModal={setwidgetModal} id={3}/>
                              <WidgetCard setwidgetModal={setwidgetModal} id={1}/>
                            </Row>
                   </div>
                </TabPane>
                <TabPane tab="Paid Marketing" key="2">
                    <div className={'fa-container'}>
                        <div className={'py-4 flex justify-center flex-wrap'}>
                            <WidgetCard setwidgetModal={setwidgetModal} id={1}/>
                            <WidgetCard setwidgetModal={setwidgetModal} id={3}/>
                        </div>
                   </div>
                </TabPane>
                <TabPane tab="Campaigns" key="3">
                    <div className={'fa-container'}>
                        <div className={'py-4 flex justify-center flex-wrap'}>
                            <WidgetCard setwidgetModal={setwidgetModal} id={3}/>
                            <WidgetCard setwidgetModal={setwidgetModal} id={2}/>
                        </div>
                   </div>
                </TabPane>
              </Tabs>
            }
          </Col>
        </Row>

    <Modal
        title={null}
        visible={widgetModal}
        footer={null}
        centered={false}
        zIndex={1015}
        mask={false}
        onCancel={() => setwidgetModal(false)}
        // closable={false}
        className={'fa-modal--full-width'}
        >
            <div className={'py-10 flex justify-center'}>
                <div className={'fa-container'}>
                    <Text type={'title'} level={5} weight={'bold'} size={'grey'} extraClass={'m-0'}>Full width Modal</Text>
                    <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Core Query results page comes here..</Text>
                </div>
            </div>
    </Modal>

  </>);
}

export default ProjectTabs;
