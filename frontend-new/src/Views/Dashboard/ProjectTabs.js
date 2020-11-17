import React, { useState, useEffect, useCallback } from 'react';
import {
    Tabs, Modal, Button, Spin
} from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
// import WidgetCard from './WidgetCard';

import { useSelector, useDispatch } from 'react-redux';
import { fetchActiveDashboardUnits } from '../../reducers/dashboard/services';
import { ACTIVE_DASHBOARD_CHANGE } from '../../reducers/types';
import SortableCards from './SortableCards';
import DashboardSubMenu from './DashboardSubMenu';
const { TabPane } = Tabs;

function ProjectTabs({ setaddDashboardModal, handleEditClick }) {
    const [widgetModal, setwidgetModal] = useState(false);
    const { active_project } = useSelector(state => state.global);
    const { dashboards, activeDashboard, activeDashboardUnits} = useSelector(state => state.dashboard);
    const dispatch = useDispatch();

    const handleTabChange = useCallback((value) => {
        dispatch({
            type: ACTIVE_DASHBOARD_CHANGE,
            payload: dashboards.data.find(d => d.id === parseInt(value))
        });
    }, [dashboards, dispatch]);

    const fetchUnits = useCallback(() => {
        if (active_project.id && activeDashboard.id) {
            fetchActiveDashboardUnits(dispatch, active_project.id, activeDashboard.id);
        }
    }, [active_project.id, activeDashboard.id, dispatch]);

    useEffect(() => {
        fetchUnits();
    }, [fetchUnits]);

    const operations = (
        <>
            <Button type="text" size={'small'} onClick={() => setaddDashboardModal(true)}><SVG name="plus" color={'grey'} /></Button>
            <Button type="text" size={'small'}><SVG name="edit" color={'grey'} /></Button>
        </>
    );

    if (dashboards.loading || activeDashboardUnits.loading) {
        return (
            <div className="flex justify-center items-center w-full h-64">
                <Spin size="large" />
            </div>
        );
    }

    if (dashboards.error || activeDashboardUnits.error) {
        return (
            <div className="flex justify-center items-center w-full h-64">
                Something went wrong!
            </div>
        );
    }

    return (
        <>
            <Tabs
                onChange={handleTabChange}
                activeKey={activeDashboard.id.toString()}
                className={'fa-tabs--dashboard'}
                tabBarExtraContent={operations}
            >
                {dashboards.data.map(d => {
                    return (
                        <TabPane tab={d.name} key={d.id}>
                            <div className={'fa-container mt-6 min-h-screen'}>
                                <DashboardSubMenu dashboard={activeDashboard} handleEditClick={handleEditClick} />
                                <SortableCards />
                            </div>
                        </TabPane>
                    )
                })}
            </Tabs>

            <Modal
                title={null}
                visible={widgetModal}
                footer={null}
                centered={false}
                zIndex={1015}
                mask={false}
                onCancel={() => setwidgetModal(false)}
                className={'fa-modal--full-width'}
            >
                <div className={'py-10 flex justify-center'}>
                    <div className={'fa-container'}>
                        <Text type={'title'} level={5} weight={'bold'} size={'grey'} extraClass={'m-0'}>Full width Modal</Text>
                        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Core Query results page comes here..</Text>
                    </div>
                </div>
            </Modal>

        </>
    );

}

export default ProjectTabs;