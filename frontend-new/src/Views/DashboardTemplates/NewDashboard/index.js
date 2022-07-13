import React, { useState, useCallback } from 'react';
import { Row, Col, Tabs, Modal, notification } from 'antd';
import { useSelector, useDispatch } from 'react-redux';
import AddDashboardTab from './AddDashboardTab';
import AddWidgetsTab from './AddWidgetsTab';
import { useHistory } from 'react-router-dom';

import { Text } from '../../../components/factorsComponents';
import {
  createDashboard,
  assignUnitsToDashboard,
} from '../../../reducers/dashboard/services';
import {
  DASHBOARD_CREATED,
} from '../../../reducers/types';
import styles from './index.module.scss';

function NewDashboard({
  AddDashboardDetailsVisible,
  setAddDashboardDetailsVisible,
}) {
  const history = useHistory();
  const [activeKey, setActiveKey] = useState('1');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dashboardType, setDashboardType] = useState('pr');
  const [apisCalled, setApisCalled] = useState(false);
  const [selectedQueries, setSelectedQueries] = useState([]);
  const { data: queries } = useSelector((state) => state.queries);
  const { active_project } = useSelector((state) => state.global);
  const dispatch = useDispatch();

  const { TabPane } = Tabs;


  const resetState = useCallback(() => {
    setActiveKey('1');
    setTitle('');
    setDescription('');
    setDashboardType('pr');
    setApisCalled(false);
    setSelectedQueries([]);
    setAddDashboardDetailsVisible(false);
  }, [setAddDashboardDetailsVisible]);

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
          duration: 5,
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
        query_id: sq.query_id,
      };
    });
    return reqBody;
  }, [selectedQueries]);

  const createNewDashboard = useCallback(async () => {
    try {
      setApisCalled(true);
      const res = await createDashboard(active_project.id, {
        name: title,
        description,
        type: dashboardType,
      });
      if (selectedQueries.length) {
        const reqBody = getUnitsAssignRequestBody();
        await assignUnitsToDashboard(active_project.id, res.data.id, reqBody);
      }
      dispatch({ type: DASHBOARD_CREATED, payload: res.data });
      resetState();
      window.location.reload(); // temporary Fix for empty dashboard
    } catch (err) {
      console.log(err.response);
      setApisCalled(false);
    }
  }, [
    active_project.id,
    dashboardType,
    description,
    dispatch,
    resetState,
    selectedQueries,
    title,
    getUnitsAssignRequestBody,
  ]);

  const handleOk = useCallback(async () => {
    if (activeKey === '2') {
      createNewDashboard();
      history.push('/');
    } else {
      if (!title.trim().length) {
        notification.error({
          message: 'Incorrect Input!',
          description: 'Please Enter dashboard title',
          duration: 5,
        });
        return false;
      }
      setActiveKey('2');
    }
  }, [
    activeKey,
    title,
    createNewDashboard,
  ]);

  const getOkText = useCallback(() => {
    if (activeKey === '1') {
      return 'Next';
    } else {
        if (selectedQueries.length) {
          return 'Create Dashboard';
        } else {
          return "I'll add them later";
        }
    }
  }, [activeKey, selectedQueries.length]);

  return (
    <>
      <Modal
        title={null}
        visible={AddDashboardDetailsVisible}
        centered={true}
        zIndex={1005}
        width={700}
        onCancel={handleCancel}
        onOk={handleOk}
        className={'fa-modal--regular p-4 fa-modal--slideInDown'}
        confirmLoading={apisCalled}
        closable={false}
        okText={getOkText()}
        transitionName=''
        maskTransitionName=''
        okButtonProps={{ size: 'large' }}
        cancelButtonProps={{ size: 'large' }}
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
                New Dashboard
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
                  <AddDashboardTab
                    title={title}
                    setTitle={setTitle}
                    description={description}
                    setDescription={setDescription}
                    dashboardType={dashboardType}
                    setDashboardType={setDashboardType}
                  />
                </TabPane>
                <TabPane className={styles.tabContent} tab='Widget' key='2'>
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
    </>
  );
}

export default NewDashboard;
