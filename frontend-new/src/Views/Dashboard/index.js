import React, { useState } from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
import { Text } from 'factorsComponents';
import AddDashboard from './AddDashboard';
import AddWidgets from './AddWidgets';
import {
  Row, Col, Tabs, Modal
} from 'antd';
const { TabPane } = Tabs;

function Dashboard() {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  return (<>

            <Header>
                <div className="w-full h-full py-4 flex flex-col justify-center items-center">
                    <SearchBar />
                </div>
            </Header>
            <div className={'mt-16'}>
                    <ProjectTabs setaddDashboardModal={setaddDashboardModal} />
            </div>

            <Modal
                title={null}
                visible={addDashboardModal}
                centered={true}
                zIndex={1015}
                width={700}
                onCancel={() => setaddDashboardModal(false)}
                className={'fa-modal--regular p-4'}
                okText={'Next'}
                >
                    <div className={'px-4'}>
                        <Row>
                            <Col span={24}>
                                <Text type={'title'} level={4} weight={'bold'} size={'grey'} extraClass={'m-0'}>New Dashboard</Text>
                            </Col>
                        </Row>
                        <Row>
                            <Col span={24}>
                                    <Tabs defaultActiveKey="1"
                                    className={'fa-tabs'}>
                                        <TabPane tab="Setup" key="1">
                                            <AddDashboard />
                                        </TabPane>
                                        <TabPane tab="Widget" key="2">
                                            <AddWidgets />
                                        </TabPane>
                                    </Tabs>
                            </Col>
                        </Row>
                    </div>
            </Modal>

  </>);
}

export default Dashboard;
