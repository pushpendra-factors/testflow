import React, { useState, useCallback, useContext } from 'react';
import MomentTz from 'Components/MomentTz';
import {
  Button,
  Modal,
  Input,
  Switch,
  Select,
  Radio,
  notification,
} from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import { saveQuery } from '../../reducers/coreQuery/services';
import { useSelector, useDispatch, connect } from 'react-redux';
import { QUERY_CREATED } from '../../reducers/types';
import { saveQueryToDashboard } from '../../reducers/dashboard/services';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  apiChartAnnotations,
  CHART_TYPE_TABLE,
  DASHBOARD_TYPES,
} from '../../utils/constants';
import { getSaveChartOptions } from '../../Views/CoreQuery/utils';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import { fetchWeeklyIngishtsMetaData } from '../../reducers/insights';

function SaveQuery({
  requestQuery,
  visible,
  setVisible,
  queryType,
  setQuerySaved,
  fetchWeeklyIngishtsMetaData,
  getCurrentSorter,
}) {
  const [title, setTitle] = useState('');
  const [addToDashboard, setAddToDashboard] = useState(false);
  const [selectedDashboards, setSelectedDashboards] = useState([]);
  const [dashboardPresentation, setDashboardPresentation] = useState(
    apiChartAnnotations[CHART_TYPE_TABLE]
  );
  const [apisCalled, setApisCalled] = useState(false);
  const { attributionMetrics } = useContext(CoreQueryContext);
  const { active_project } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const dispatch = useDispatch();

  const startOfWeek =  MomentTz().startOf('week').utc().unix();
  const todayNow =  MomentTz().utc().unix();

  const handleTitleChange = useCallback((e) => {
    setTitle(e.target.value);
  }, []);

  const resetModalState = useCallback(() => {
    setTitle('');
    setSelectedDashboards([]);
    setAddToDashboard(false);
    setDashboardPresentation(apiChartAnnotations[CHART_TYPE_TABLE]);
    setVisible(false);
  }, [setVisible]);

  const handleSaveCancel = useCallback(() => {
    if (!apisCalled) {
      resetModalState();
    }
  }, [resetModalState, apisCalled]);

  const handleSelectChange = useCallback(
    (value) => {
      const resp = value.map((v) => {
        return dashboards.data.find((d) => d.name === v).id;
      });
      setSelectedDashboards(resp);
    },
    [dashboards.data]
  );

  const handlePresentationChange = useCallback((e) => {
    setDashboardPresentation(e.target.value);
  }, []);

  const toggleAddToDashboard = useCallback(
    (val) => {
      setAddToDashboard(val);
    },
    [setAddToDashboard]
  );

  const getSelectedDashboards = useCallback(() => {
    return selectedDashboards.map((s) => {
      return dashboards.data.find((d) => d.id === s).name;
    });
  }, [dashboards.data, selectedDashboards]);

  const handleSave = useCallback(async () => {
    if (!title.trim().length) {
      notification.error({
        message: 'Incorrect Input!',
        description: 'Please Enter query title',
        duration: 5,
      });
      return false;
    }
    if (addToDashboard && !selectedDashboards.length) {
      notification.error({
        message: 'Incorrect Input!',
        description: 'Please select atleast one dashboard',
        duration: 5,
      });
      return false;
    }

    try {
      setApisCalled(true);
      let query;
      const querySettings = getCurrentSorter();
      if (queryType === QUERY_TYPE_FUNNEL) {
        query = {
          ...requestQuery,
          fr: startOfWeek,
          to: todayNow,
        };
      } else if (queryType === QUERY_TYPE_ATTRIBUTION) {
        query = {
          ...requestQuery,
          query: {
            ...requestQuery.query,
            from: startOfWeek,
            to: todayNow,
          },
        };
      } else if (queryType === QUERY_TYPE_EVENT) {
        query = {
          query_group: requestQuery.map((q) => {
            return {
              ...q,
              fr: startOfWeek,
              to: todayNow,
              gbt: q.gbt ? 'date' : '',
            };
          }),
        };
      } else if (queryType === QUERY_TYPE_CAMPAIGN) {
        query = {
          ...requestQuery,
          query_group: requestQuery.query_group.map((q) => {
            return {
              ...q,
              fr: startOfWeek,
              to: todayNow,
              gbt: q.gbt ? 'date' : '',
            };
          }),
        };
      }
      const type = addToDashboard ? 1 : 2;
      const res = await saveQuery(
        active_project.id,
        title,
        query,
        type,
        querySettings
      );
      if (addToDashboard) {
        const settings = {
          chart: dashboardPresentation,
          attributionMetrics: JSON.stringify(attributionMetrics),
        };
        const reqBody = {
          settings,
          description: '',
          title,
          query_id: res.data.id,
        };
        await saveQueryToDashboard(
          active_project.id,
          selectedDashboards.join(','),
          reqBody
        );
      }
      notification.success({
        message: 'Report Saved Successfully',
        duration: 5,
      });
      dispatch({ type: QUERY_CREATED, payload: res.data });
      setQuerySaved(title);
      setApisCalled(false);
      fetchWeeklyIngishtsMetaData(active_project.id);
      resetModalState();
    } catch (err) {
      setApisCalled(false);
      console.log(err);
      console.log(err.response);
      notification.error({
        message: 'Error!',
        description: 'Something went wrong.',
        duration: 5,
      });
    }
  }, [
    title,
    active_project.id,
    requestQuery,
    dispatch,
    resetModalState,
    addToDashboard,
    dashboardPresentation,
    selectedDashboards,
    queryType,
    setQuerySaved,
    attributionMetrics,
    fetchWeeklyIngishtsMetaData,
    getCurrentSorter
  ]);

  let dashboardHelpText = 'Create a dashboard widget for regular monitoring';
  let dashboardOptions = null;

  if (addToDashboard) {
    dashboardHelpText = 'This widget will appear on the following dashboards:';

    dashboardOptions = (
      <>
        <div className='mt-5'>
          <Select
            mode='multiple'
            style={{ width: '100%' }}
            placeholder={'Please Select'}
            onChange={handleSelectChange}
            className={styles.multiSelectStyles}
            value={getSelectedDashboards()}
          >
            {dashboards.data
              .filter((d) => d.class === DASHBOARD_TYPES.USER_CREATED)
              .map((d) => {
                return (
                  <Select.Option value={d.name} key={d.id}>
                    {d.name}
                  </Select.Option>
                );
              })}
          </Select>
        </div>
        <div className='mt-2'>
          <Radio.Group
            value={dashboardPresentation}
            onChange={handlePresentationChange}
            className={styles.radioGroup}
          >
            {getSaveChartOptions(queryType, requestQuery)}
          </Radio.Group>
        </div>
      </>
    );
  }

  return (
    <>
      <Button
        onClick={setVisible.bind(this, true)}
        type='primary'
        size={'large'}
        icon={<SVG name={'save'} size={20} color={'white'} />}
      >
        Save
      </Button>

      <Modal
        centered={true}
        visible={visible}
        width={900}
        title={null}
        onOk={handleSave}
        onCancel={handleSaveCancel}
        className={'fa-modal--regular p-4 fa-modal--slideInDown'}
        okText={'Save'}
        closable={false}
        confirmLoading={apisCalled}
        transitionName=''
        maskTransitionName=''
      >
        <div className='p-4'>
          <Text extraClass='m-0' type={'title'} level={3} weight={'bold'}>
            Create New Report
          </Text>
          <div className='pt-6'>
            <Text
              type={'title'}
              level={7}
              extraClass={`m-0 ${styles.inputLabel}`}
            >
              Title
            </Text>
            <Input
              onChange={handleTitleChange}
              value={title}
              className={'fa-input'}
              size={'large'}
            />
          </div>
          {/* <div className={`pt-2 ${styles.linkText}`}>Help others to find this query easily?</div> */}
          <div className={'pt-6 flex items-center'}>
            <Switch
              onChange={toggleAddToDashboard}
              checked={addToDashboard}
              className={styles.switchBtn}
              checkedChildren='On'
              unCheckedChildren='Off'
            />
            <Text extraClass='m-0' type='title' level={6} weight='bold'>
              Add to Dashboard
            </Text>
          </div>
          <Text extraClass={`pt-1 ${styles.noteText}`} mini type={'paragraph'}>
            {dashboardHelpText}
          </Text>
          {dashboardOptions}
        </div>
      </Modal>
    </>
  );
}

export default connect(null, { fetchWeeklyIngishtsMetaData })(SaveQuery);
