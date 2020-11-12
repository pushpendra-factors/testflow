import React, { useState, useEffect, useCallback } from 'react';
import {
  Tabs, Modal, Button, Spin, Select
} from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
import WidgetCard from './WidgetCard';
import { ReactSortable } from 'react-sortablejs';
import {
  LockOutlined, ReloadOutlined, UserAddOutlined, MoreOutlined, EditOutlined, UnlockOutlined
} from '@ant-design/icons';
import { useSelector, useDispatch } from 'react-redux';
import { fetchActiveDashboardUnits } from '../../reducers/dashboard/services';
import { ACTIVE_DASHBOARD_CHANGE } from '../../reducers/types';
const { TabPane } = Tabs;
const { Option } = Select;

const widgetCardCollection = [
  {
    id: 1, type: 1, size: 1, title: 'item 1'
  },
  {
    id: 2, type: 2, size: 1, title: 'item 2'
  },
  {
    id: 3, type: 3, size: 1, title: 'item 3'
  },
  {
    id: 4, type: 1, size: 1, title: 'item 4'
  },
  {
    id: 5, type: 2, size: 2, title: 'item 5'
  },
  {
    id: 6, type: 1, size: 1, title: 'item 6'
  },
  {
    id: 7, type: 3, size: 1, title: 'item 7'
  },
  {
    id: 8, type: 2, size: 1, title: 'item 8'
  },
  {
    id: 9, type: 3, size: 1, title: 'item 9'
  },
  {
    id: 10, type: 1, size: 2, title: 'item 10'
  },
  {
    id: 11, type: 2, size: 3, title: 'item 11'
  }
];

const DashboardSubMenu = ({ dashboard }) => {
  let btn = null;

  if (dashboard.type === 'pr') {
    btn = (
      <Button
        style={{ display: 'flex' }}
        size={'large'}
        type={'text'}
        className={'m-0 fa-button-ghost items-center p-0 py-2'}
      >
        <UnlockOutlined /> Public.
      </Button>
    )
  } else {
    btn = (
      <Button
        style={{ display: 'flex' }}
        size={'large'}
        type={'text'}
        className={'m-0 fa-button-ghost items-center p-0 py-2'}
      >
        <LockOutlined /> Private.
      </Button>
    )
  }

  return (
    <div className={'flex justify-between items-center px-4 mb-4'}>
      <div className={'flex justify-between items-center'}>
        <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>Date from</Text>
        <Select className={'fa-select mx-2 mr-4 ml-4'} defaultValue="Last 30 days">
          <Option value="jack">1 Month</Option>
          <Option value="lucy2">2 Months</Option>
          <Option value="lucy3">6 Months</Option>
          <Option value="lucy4">1 Year</Option>
          <Option value="lucy5">1+ Year</Option>
        </Select>
        {btn}
        <Button size={'large'} type={'text'} className={'m-0 fa-button-ghost flex items-center p-0 py-2'}><EditOutlined /> Edit</Button>
      </div>
      <div className={'flex justify-between items-center'}>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center flex m-0 fa-button-ghost p-0 py-2'}><ReloadOutlined /> Refresh Data.</Button>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><UserAddOutlined /></Button>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><MoreOutlined /></Button>

      </div>
    </div>
  );
};

function ProjectTabs({ setaddDashboardModal }) {
  const [widgetModal, setwidgetModal] = useState(false);
  const [widgets, setWidgets] = useState(widgetCardCollection);
  const { active_project } = useSelector(state => state.global);
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(state => state.dashboard);
  const { data: savedQueries } = useSelector(state => state.queries);
  const dispatch = useDispatch();

  const fetchUnits = useCallback(() => {
    if (active_project.id && activeDashboard.id) {
      fetchActiveDashboardUnits(dispatch, active_project.id, activeDashboard.id);
    }
  }, [active_project.id, activeDashboard.id, dispatch]);

  useEffect(() => {
    fetchUnits();
  }, [fetchUnits]);

  const handleTabChange = useCallback((value) => {
    dispatch({
      type: ACTIVE_DASHBOARD_CHANGE,
      payload: dashboards.data.find(d => d.id === parseInt(value))
    });
  }, [dashboards, dispatch]);

  const onDrop = (newState) => {
    setWidgets(newState);
  };

  const resizeWidth = (index, operator) => {
    const newArray = [...widgets];
    let newSize = 1;
    const currentWidth = newArray[index].size;
    if (operator === '+') {
      if (currentWidth !== 3) {
        newSize = currentWidth + 1;
      } else {
        newSize = 3;
      }
    } else {
      if (currentWidth !== 0) {
        (currentWidth - 1 === 0) ? newSize = 1 : newSize = currentWidth - 1;
      } else {
        newSize = 1;
      }
    }
    newArray[index] = { ...newArray[index], size: newSize };
    setWidgets(newArray);
  };

  const operations = (
    <>
      <Button type="text" size={'small'} onClick={() => setaddDashboardModal(true)}><SVG name="plus" color={'grey'} /></Button>
      <Button type="text" size={'small'}><SVG name="edit" color={'grey'} /></Button>
    </>
  );

  const loading = dashboards.loading || activeDashboardUnits.loading;
  const error = dashboards.error || activeDashboardUnits.error;

  if (loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  const units = activeDashboardUnits.data
    .filter(unit => {
      const idx = savedQueries.findIndex(sq => sq.id === unit.query_id);
      return idx > -1;
    })
    .map(unit => {
      const savedQuery = savedQueries.find(sq => sq.id === unit.query_id);
      return { ...unit, query: savedQuery };
    });

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
              {d.id === activeDashboard.id ? (
                <div className={'fa-container mt-6 min-h-screen'}>
                  <DashboardSubMenu dashboard={activeDashboard} />
                  <ReactSortable list={widgets} setList={onDrop}>
                    <div className="flex flex-wrap">
                      {units.map(unit => {
                        return (
                          <WidgetCard
                            key={unit.id}
                            widthSize={3}
                            resizeWidth={resizeWidth}
                            unit={unit}
                            dashboard={d}
                          />
                        );
                      })}
                    </div>
                  </ReactSortable>
                </div>
              ) : null}
            </TabPane>
          );
        })}
      </Tabs>

      <Modal
        title={null}
        visible={widgetModal}
        footer={null}
        centered={false}
        zIndex={1005}
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
    </>
  );
}

export default React.memo(ProjectTabs);
