import React from 'react';
import { Row, Col, Modal } from 'antd';
import { Text } from '../../../../components/factorsComponents';
import { useSelector, useDispatch } from 'react-redux';

import { createDashboardFromTemplate } from '../../../../reducers/dashboard_templates/services';
import { fetchDashboards } from '../../../../reducers/dashboard/services';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { fetchKPIConfig, fetchPageUrls } from '../../../../reducers/kpi';
import {
  fetchAttrContentGroups,
  fetchGroups,
  fetchQueries,
  fetchSmartPropertyRules
} from '../../../../reducers/coreQuery/services';
import {
  getUserPropertiesV2,
  fetchEventNames,
  getGroupProperties
} from '../../../../reducers/coreQuery/middleware';
import { useHistory } from 'react-router-dom';

function CopyDashboardModal({
  showCopyDashBoardModal,
  setShowCopyDashBoardModal,
  setShowTemplates
}) {
  const history = useHistory();
  const { active_project } = useSelector((state) => state.global);
  const { activeTemplate } = useSelector((state) => state.dashboardTemplates);
  const { dashboards } = useSelector((state) => state.dashboard);
  const dispatch = useDispatch();
  // console.log('r',state);

  const fetchDashboardItems = () => {
    dispatch(fetchDashboards(active_project.id));
    dispatch(fetchQueries(active_project.id));
    dispatch(fetchGroups(active_project.id));
    dispatch(fetchKPIConfig(active_project.id));
    dispatch(fetchPageUrls(active_project.id));
    // dispatch(deleteQueryTest())
    fetchEventNames(active_project.id);
    getUserPropertiesV2(active_project.id);
    getGroupProperties(active_project.id);
    dispatch(fetchSmartPropertyRules(active_project.id));
    fetchWeeklyIngishtsMetaData(active_project.id);
    dispatch(fetchAttrContentGroups(active_project.id));
  };
  const handleOk = async () => {
    try {
      const res = await createDashboardFromTemplate(
        active_project.id,
        activeTemplate.id
      );
      console.log('Dashboards Created');
      setShowTemplates(false);
      fetchDashboardItems();
      if (res) {
        history.push('/');
      }
    } catch (err) {
      console.log(err.response);
    }
    setShowCopyDashBoardModal(false);
  };
  const handleCancel = () => {
    setShowCopyDashBoardModal(false);
  };
  return (
    <Modal
      centered={true}
      width={'30%'}
      onCancel={handleCancel}
      onOk={handleOk}
      className={'fa-modal--regular p-4 fa-modal--slideInDown'}
      closable={true}
      okText={'Create Copy'}
      cancelText={'Cancel'}
      okButtonProps={{ size: 'large' }}
      cancelButtonProps={{ size: 'large' }}
      transitionName=''
      maskTransitionName=''
      visible={showCopyDashBoardModal}
    >
      <Row className={'pt-4'}>
        <Col>
          <Text type='title' level={4} weight={'bold'}>
            Do you want to create a copy?
          </Text>
        </Col>
        <Col>
          <Text type='paragraph' level={7} color={'grey'} weight={'bold'}>
            Creating a copy will replicate this dashboard into your Project
          </Text>
        </Col>
      </Row>
    </Modal>
  );
}

export default CopyDashboardModal;
