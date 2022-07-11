import React from 'react';
import { Text, SVG } from '../../../components/factorsComponents';
import {
  Row, Col, Input, Button
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
// const { Option } = Select;

function AddDashboardTab({
  title, setTitle, description, setDescription, dashboardType, setDashboardType
}) {
  return (
    <>
            <Row className={'pt-4'} gutter={[24, 24]}>
                <Col span={12}>
                    <Text type={'title'} level={7} extraClass={'m-0'}>Title</Text>
                    <Input onChange={(e) => setTitle(e.target.value)} value={title} className={'fa-input'} size={'large'} placeholder="Dashboard Title" />
                </Col>
                <Col span={12}>
                    <Text type={'title'} level={7} extraClass={'m-0'}>Description (Optional)</Text>
                    <Input onChange={(e) => setDescription(e.target.value)} value={description} className={'fa-input'} size={'large'} placeholder="Description (Optional)" />
                </Col>
            </Row>
            <Row className={'pt-2'} gutter={[24, 4]}>
                <Col span={24}>
                    <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Who can access this dashboard?</Text>
                </Col>
                <Col span={24}>
                    <Row gutter={[24, 4]}>
                        <Col span={12}>
                            <div onClick={setDashboardType.bind(this, 'pr')} className={`${dashboardType === 'pr' ? 'fa-dasboard-privacy--card selected' : 'fa-dasboard-privacy--card'} border-radius--medium p-4`}>
                                <div className={'flex justify-between items-start'}>
                                    <div>
                                        <SVG name={'lock'} color={'grey'} extraClass={'mr-2 mt-1'} />
                                    </div>
                                    <div>
                                        <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Private</Text>
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Only you have access to the contents of Private Dashboards.</Text>
                                    </div>
                                </div>
                            </div>
                        </Col>
                        <Col span={12}>
                            <div onClick={setDashboardType.bind(this, 'pv')} className={`${dashboardType === 'pv' ? 'fa-dasboard-privacy--card selected' : 'fa-dasboard-privacy--card'} border-radius--medium p-4`}>
                                <div className={'flex justify-between items-start'}>
                                    <div>
                                        <SVG name={'globe'} color={'grey'} extraClass={'mr-2 mt-1'} />
                                    </div>
                                    <div>
                                        <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Public</Text>
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Everyone in your organization has access to this dashboard.</Text>
                                    </div>
                                </div>
                            </div>
                        </Col>
                    </Row>
                </Col>
            </Row>
    </>
  );
}

export default AddDashboardTab;
