import React, { useState, useCallback, useEffect } from 'react';
import { Row, Col, Tabs, Modal, notification, Button } from 'antd';
import { useSelector, useDispatch } from 'react-redux';
import AddDashboardTab from './AddDashboardTab';
import AddWidgetsTab from './AddWidgetsTab';
import { Text, SVG } from '../../../components/factorsComponents';
import {
  createDashboard,
  assignUnitsToDashboard,
  deleteDashboard,
  DeleteUnitFromDashboard,
  updateDashboard,
  fetchActiveDashboardUnits
} from '../../../reducers/dashboard/services';
import {
  DASHBOARD_CREATED,
  DASHBOARD_DELETED,
  WIDGET_DELETED,
  DASHBOARD_UPDATED,
  ADD_DASHBOARD_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  ACTIVE_DASHBOARD_CHANGE
} from '../../../reducers/types';
import styles from './index.module.scss';
import ConfirmationModal from '../../../components/ConfirmationModal';
import factorsai from 'factorsai';
import { Link, useHistory, useLocation } from 'react-router-dom';
import useAutoFocus from 'hooks/useAutoFocus';
import DashboardTemplatesModal from './DashboardTemplatesModal';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { DASHBOARD_KEYS } from 'Constants/localStorage.constants';
import { LoadingOutlined } from '@ant-design/icons';
import { stubFalse } from 'lodash';

function AddDashboard({
  addDashboardModal,
  setaddDashboardModal,
  editDashboard,
  setEditDashboard
}) {
  const [isLoading, setIsLoading] = useState(false);
  const [activeKey, setActiveKey] = useState('1');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dashboardType, setDashboardType] = useState('pv');
  const [apisCalled, setApisCalled] = useState(false);
  const [selectedQueries, setSelectedQueries] = useState([]);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [deleteModal, showDeleteModal] = useState(false);
  const { data: queries } = useSelector((state) => state.queries);
  const { active_project } = useSelector((state) => state.global);
  const { activeDashboardUnits } = useSelector((state) => state.dashboard);
  const dispatch = useDispatch();
  const history = useHistory();
  const { pathname } = useLocation();
  const inputComponentRef = useAutoFocus(addDashboardModal);

  let { isAddNewDashboardModal } = useSelector(
    (state) => state.dashboard_templates_Reducer
  );
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
    setaddDashboardModal(false);
    setActiveKey('1');
    setTitle('');
    setDescription('');
    setDashboardType('pv');
    setApisCalled(false);
    setSelectedQueries([]);
    setEditDashboard(null);

    dispatch({ type: ADD_DASHBOARD_MODAL_CLOSE });
  }, [setaddDashboardModal, setEditDashboard]);

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await deleteDashboard(active_project?.id, editDashboard?.id);
      setDeleteApiCalled(false);
      dispatch({ type: DASHBOARD_DELETED, payload: editDashboard });
      showDeleteModal(false);
      resetState();
    } catch (err) {
      console.log(err);
      setDeleteApiCalled(false);
    }
  }, [editDashboard, dispatch, resetState, active_project?.id]);

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

  const getUnitsAssignRequestBody = useCallback(() => {
    const reqBody = selectedQueries.map((sq) => {
      return {
        description: sq.description,
        query_id: sq.query_id
      };
    });
    return reqBody;
  }, [selectedQueries]);

  const createNewDashboard = useCallback(async () => {
    try {
      setIsLoading(true);
      setApisCalled(true);
      const res = await createDashboard(active_project?.id, {
        name: title,
        description,
        type: dashboardType
      });
      if (selectedQueries.length) {
        const reqBody = getUnitsAssignRequestBody();
        await assignUnitsToDashboard(active_project?.id, res.data.id, reqBody);
      }
      dispatch({ type: DASHBOARD_CREATED, payload: res.data });
      // pathname === '/template' && history.push('/');

      setItemToLocalStorage(DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID, res.data.id);
      // // Doing this to close NewTemplates Modal after creating Dashboard
      dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
      resetState();
      setIsLoading(false);
      dispatch({
        type: ACTIVE_DASHBOARD_CHANGE,
        payload: res.data
      });
      // window.location.reload(); // temporary Fix for empty dashboard
    } catch (err) {
      setIsLoading(false);
      console.log(err.response);
      setApisCalled(false);
    }
  }, [
    active_project?.id,
    dashboardType,
    description,
    dispatch,
    resetState,
    selectedQueries,
    title,
    getUnitsAssignRequestBody
  ]);

  const editExistingDashboard = useCallback(async () => {
    try {
      setIsLoading(true);
      setApisCalled(true);
      const newAddedUnits = selectedQueries.filter(
        (elem) =>
          activeDashboardUnits.data.findIndex((unit) => unit.id === elem.id) ===
          -1
      );
      if (newAddedUnits.length) {
        const reqBody = newAddedUnits.map((unit) => {
          return {
            description: unit.description,
            query_id: unit.query_id
          };
        });
        await assignUnitsToDashboard(
          active_project?.id,
          editDashboard?.id,
          reqBody
        );
      }
      const deletedUnits = activeDashboardUnits.data.filter(
        (elem) =>
          selectedQueries.findIndex((query) => query.id === elem.id) === -1
      );
      if (deletedUnits.length) {
        //just delete the deleted widgets
        const deletePromises = deletedUnits.map((q) => {
          return DeleteUnitFromDashboard(
            active_project?.id,
            editDashboard?.id,
            q.id
          );
        });
        await Promise.all(deletePromises);
        deletedUnits.forEach((unit) => {
          dispatch({ type: WIDGET_DELETED, payload: unit.id });
        });
      }
      await updateDashboard(active_project?.id, editDashboard.id, {
        name: title,
        description,
        type: dashboardType
      });
      dispatch({
        type: DASHBOARD_UPDATED,
        payload: {
          name: title,
          description,
          id: editDashboard.id,
          type: dashboardType
        }
      });

      if (newAddedUnits.length) {
        dispatch(
          fetchActiveDashboardUnits(active_project?.id, editDashboard.id)
        );
      }

      //Factors EDIT_DASHBOARD tracking
      factorsai.track('EDIT_DASHBOARD', {
        dashboard_name: title,
        dashboard_type: dashboardType,
        dashboard_id: editDashboard.id
      });

      dispatch({ type: ADD_DASHBOARD_MODAL_CLOSE });
      setIsLoading(false);
      setApisCalled(false);
      resetState();
    } catch (err) {
      setIsLoading(stubFalse);
      console.log(err);
      setApisCalled(false);
    }
  }, [
    activeDashboardUnits,
    dashboardType,
    selectedQueries,
    active_project?.id,
    description,
    dispatch,
    editDashboard,
    resetState,
    title
  ]);

  const handleOk = useCallback(async () => {
    if (activeKey === '2') {
      if (!editDashboard) {
        createNewDashboard();
      } else {
        editExistingDashboard();
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
  }, [
    activeKey,
    title,
    createNewDashboard,
    editDashboard,
    editExistingDashboard
  ]);

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
          return "I'll add them later";
        }
      }
    }
  }, [activeKey, editDashboard, selectedQueries.length]);

  return (
    <>
      <Modal
        title={null}
        visible={isAddNewDashboardModal}
        centered={true}
        zIndex={1010}
        width={700}
        className={'fa-modal--regular p-4 fa-modal--slideInDown'}
        confirmLoading={apisCalled}
        closable={false}
        okText={getOkText()}
        transitionName=''
        maskTransitionName=''
        footer={null}
      >
        <div>
          <Row>
            <Col span={24}>
              <Text
                type={'title'}
                level={4}
                weight={'bold'}
                size={'grey'}
                extraClass={'m-0'}
              >
                {editDashboard ? 'Edit Dashboard' : 'New Dashboard'}
              </Text>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Tabs
                onChange={handleTabChange}
                activeKey={activeKey}
                className={'fa-tabs'}
              >
                <TabPane className={styles.tabContent} tab='Setup' key='1'>
                  {activeKey === '1' ? (
                    <AddDashboardTab
                      title={title}
                      setTitle={setTitle}
                      description={description}
                      setDescription={setDescription}
                      dashboardType={dashboardType}
                      setDashboardType={setDashboardType}
                      editDashboard={editDashboard}
                      showDeleteModal={showDeleteModal}
                      inputComponentRef={inputComponentRef}
                    />
                  ) : (
                    ''
                  )}
                </TabPane>
                <TabPane className={styles.tabContent} tab='Widget' key='2'>
                  {activeKey === '2' ? (
                    <AddWidgetsTab
                      queries={queries}
                      selectedQueries={selectedQueries}
                      setSelectedQueries={setSelectedQueries}
                      setIsLoading={setIsLoading}
                    />
                  ) : (
                    ''
                  )}
                </TabPane>
              </Tabs>
            </Col>
          </Row>
          <div className='flex mt-6 items-center justify-end'>
            {/* <Link
              to={{
                pathname: '/template',
                state: { fromSelectTemplateBtn: true }
              }}
              className='flex items-center font-semibold gap-2'
              style={{ color: `#1d89ff` }}
            >
              Select from Templates{' '}
              <SVG size={20} name='Arrowright' color={`#1d89ff`} />
            </Link> */}
            <div className='flex gap-3'>
              <Button
                disabled={isLoading}
                type='default'
                size='large'
                onClick={() => {
                  dispatch({ type: ADD_DASHBOARD_MODAL_CLOSE });
                  handleCancel();
                }}
              >
                Cancel
              </Button>
              <Button
                disabled={isLoading}
                type='primary'
                size='large'
                onClick={() => handleOk()}
              >
                {isLoading === true ? <LoadingOutlined /> : ''}{' '}
                {activeKey === '2' ? 'Save' : 'Next'}
              </Button>
            </div>
          </div>
        </div>
      </Modal>
      <DashboardTemplatesModal
        addDashboardModal={addDashboardModal}
        apisCalled={apisCalled}
        setaddDashboardModal={setaddDashboardModal}
        getOkText={getOkText}
      />
      <ConfirmationModal
        visible={deleteModal}
        confirmationText='Are you sure you want to delete this Dashboard?'
        onOk={confirmDelete}
        onCancel={showDeleteModal.bind(this, false)}
        title={`Delete Dashboard - ${editDashboard ? editDashboard.name : ''}`}
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteApiCalled}
      />
    </>
  );
}

export default AddDashboard;
