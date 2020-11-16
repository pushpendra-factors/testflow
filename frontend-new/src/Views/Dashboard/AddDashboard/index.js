import React, { useState, useCallback, useEffect } from 'react';
import {
  Row, Col, Tabs, Modal, notification
} from 'antd';
import { useSelector, useDispatch } from 'react-redux';
import AddDashboardTab from './AddDashboardTab';
import AddWidgetsTab from './AddWidgetsTab';
import { Text } from '../../../components/factorsComponents';
import { createDashboard, assignUnitsToDashboard } from '../../../reducers/dashboard/services';
import { DASHBOARD_CREATED, DASHBOARD_DELETED } from '../../../reducers/types';
import styles from './index.module.scss';
import ConfirmationModal from '../../../components/ConfirmationModal';

function AddDashboard({
  addDashboardModal, setaddDashboardModal, editDashboard, setEditDashboard
}) {
  const [activeKey, setActiveKey] = useState('1');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dashboardType, setDashboardType] = useState('pr');
  const [apisCalled, setApisCalled] = useState(false);
  const [selectedQueries, setSelectedQueries] = useState([]);
  // const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [deleteModal, showDeleteModal] = useState(false);
  const { data: queries } = useSelector(state => state.queries);
  const { active_project } = useSelector(state => state.global);
  const { activeDashboardUnits } = useSelector(state => state.dashboard);
  const dispatch = useDispatch();

  const { TabPane } = Tabs;

  useEffect(() => {
    if (editDashboard) {
      setTitle(editDashboard.name);
      setDescription(editDashboard.description);
      setDashboardType(editDashboard.type);
      setSelectedQueries([...activeDashboardUnits.data]);
    }
  }, [editDashboard, activeDashboardUnits.data]);

  const resetState = useCallback(() => {
    setActiveKey('1');
    setTitle('');
    setDescription('');
    setDashboardType('pr');
    setApisCalled(false);
    setSelectedQueries([]);
    setEditDashboard(null);
    setaddDashboardModal(false);
  }, [setaddDashboardModal, setEditDashboard]);

  const confirmDelete = useCallback(async () => {
    try {
      // setDeleteApiCalled(true);
      // await DeleteDashboard(active_project.id, editDashboard);
      // setDeleteApiCalled(false);
      dispatch({ type: DASHBOARD_DELETED, payload: editDashboard });
      showDeleteModal(false);
      resetState();
    } catch (err) {
      console.log(err);
      // setDeleteApiCalled(false);
    }
  }, [editDashboard, dispatch, resetState]);

  const handleCancel = useCallback(() => {
    if (!apisCalled) {
      resetState();
    }
  }, [resetState, apisCalled]);

  const handleTabChange = useCallback(() => {
    if (activeKey === '2') {
      setActiveKey('1');
    } else {
      if (!title.trim().length) {
        notification.error({
          message: 'Incorrect Input!',
          description: 'Please Enter dashboard title',
          duration: 5
        });
        return false;
      }
      setActiveKey('2');
    }
  }, [activeKey, title]);

  const handleOk = useCallback(async () => {
    if (activeKey === '2') {
      try {
        setApisCalled(true);
        const res = await createDashboard(active_project.id, { name: title, description, type: dashboardType });
        if (selectedQueries.length) {
          const reqBody = selectedQueries.map(sq => {
            const settings = {};
            if (sq.query.query_group) {
              settings.chart = 'pl';
            } else {
              settings.chart = 'pb';
            }
            return {
              settings: JSON.stringify(settings),
              title: sq.title,
              description: sq.description,
              query_id: sq.query_id
            };
          });
          await assignUnitsToDashboard(active_project.id, res.data.id, reqBody);
        }
        dispatch({ type: DASHBOARD_CREATED, payload: res.data });
        resetState();
      } catch (err) {
        console.log(err.response);
        setApisCalled(false);
      }
    } else {
      if (!title.trim().length) {
        notification.error({
          message: 'Incorrect Input!',
          description: 'Please Enter dashboard title',
          duration: 5
        });
        return false;
      }
      setActiveKey('2');
    }
  }, [activeKey, title, dashboardType, description, resetState, active_project.id, dispatch, selectedQueries]);

  const getOkText = useCallback(() => {
    if (activeKey === '1') {
      return 'Next';
    } else {
      if (editDashboard) {
        return 'Update Dashboard';
      } else {
        if (selectedQueries.length) {
          return 'Create Dashboard';
        } else {
          return 'I\'ll add them later';
        }
      }
    }
  }, [activeKey, editDashboard, selectedQueries.length]);

  return (
    <>
      <Modal
        title={null}
        visible={addDashboardModal}
        centered={true}
        zIndex={1005}
        width={700}
        onCancel={handleCancel}
        onOk={handleOk}
        className={'fa-modal--regular p-4'}
        confirmLoading={apisCalled}
        closable={false}
        okText={getOkText()}
      >
        <div className={'px-4'}>
          <Row>
            <Col span={24}>
              <Text type={'title'} level={4} weight={'bold'} size={'grey'} extraClass={'m-0'}>New Dashboard</Text>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Tabs onChange={handleTabChange} activeKey={activeKey} className={'fa-tabs'}>
                <TabPane className={styles.tabContent} tab="Setup" key="1">
                  <AddDashboardTab
                    title={title}
                    setTitle={setTitle}
                    description={description}
                    setDescription={setDescription}
                    dashboardType={dashboardType}
                    setDashboardType={setDashboardType}
                    editDashboard={editDashboard}
                    showDeleteModal={showDeleteModal}
                  />
                </TabPane>
                <TabPane className={styles.tabContent} tab="Widget" key="2">
                  <AddWidgetsTab
                    queries={queries}
                    selectedQueries={selectedQueries}
                    setSelectedQueries={setSelectedQueries}
                  />
                </TabPane>
              </Tabs>
            </Col>
          </Row>
        </div>
      </Modal>
      <ConfirmationModal
        visible={deleteModal}
        confirmationText="Are you sure you want to delete this Dashboard?"
        onOk={confirmDelete}
        onCancel={showDeleteModal.bind(this, false)}
        title="Delete Dashboard"
        okText="Confirm"
        cancelText="Cancel"
      />
    </>
  );
}

export default AddDashboard;
