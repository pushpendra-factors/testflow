import React, { useState, useEffect, useCallback } from 'react';
import {
  Tabs, Button, Spin
} from 'antd';
import { SVG } from '../../components/factorsComponents';
import { useSelector, useDispatch } from 'react-redux';
import { fetchActiveDashboardUnits } from '../../reducers/dashboard/services';
import { ACTIVE_DASHBOARD_CHANGE } from '../../reducers/types';
import SortableCards from './SortableCards';
import DashboardSubMenu from './DashboardSubMenu';
import ExpandableView from './ExpandableView';
const { TabPane } = Tabs;

function ProjectTabs({ setaddDashboardModal, handleEditClick }) {
  const [widgetModal, setwidgetModal] = useState(false);
  const [widgetModalLoading, setwidgetModalLoading] = useState(false);
  const { active_project } = useSelector(state => state.global);
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(state => state.dashboard);
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

  const handleToggleWidgetModal = (val) => {
    setwidgetModalLoading(true);
    setwidgetModal(val);
    // for canvas to load properly before rendering the charts
    setTimeout(() => {
      window.scrollTo(0, 0);
      setwidgetModalLoading(false);
    }, 1000);
  };

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

  if (dashboards.data.length) {
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
									<SortableCards setwidgetModal={handleToggleWidgetModal} />
								</div>
							</TabPane>
					  );
					})}
				</Tabs>

				<ExpandableView
					loading={widgetModalLoading}
					widgetModal={widgetModal}
					setwidgetModal={setwidgetModal}
				/>

      </>
    );
  }

  return null;
}

export default ProjectTabs;
