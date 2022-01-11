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
import { saveQuery, updateQuery } from '../../reducers/coreQuery/services';
import { useSelector, useDispatch, connect } from 'react-redux';
import { QUERY_CREATED, QUERY_UPDATED } from '../../reducers/types';
import { saveQueryToDashboard } from '../../reducers/dashboard/services';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  apiChartAnnotations,
  CHART_TYPE_TABLE,
  DASHBOARD_TYPES,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI,
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
  savedQueryId,
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

  const startOfWeek = MomentTz().startOf('week').utc().unix();
  const todayNow = MomentTz().utc().unix();

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
      const querySettings = {
        ...getCurrentSorter(),
        chart: dashboardPresentation,
      };
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
        querySettings.attributionMetrics = JSON.stringify(attributionMetrics);
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
      } else if (queryType === QUERY_TYPE_KPI) {
        query = {
          ...requestQuery,
          qG: requestQuery.qG.map((q) => {
            return {
              ...q,
              fr: startOfWeek,
              to: todayNow,
              gbt: q.gbt ? 'date' : '',
            };
          }),
        };
      } else if (queryType === QUERY_TYPE_PROFILE) {
        query = {
          ...requestQuery,
        };
      }
      let res;
      if (!savedQueryId) {
        const type = addToDashboard ? 1 : 2;
        res = await saveQuery(
          active_project.id,
          title,
          query,
          type,
          querySettings
        );
        dispatch({ type: QUERY_CREATED, payload: res.data });
        setQuerySaved({ name: title, id: res.data.id });
      } else {
        const reqBody = {
          title,
          settings: querySettings,
        };
        if (addToDashboard) {
          reqBody.type = 1;
        }
        res = await updateQuery(active_project.id, savedQueryId, reqBody);
        dispatch({
          type: QUERY_UPDATED,
          queryId: savedQueryId,
          payload: {
            title,
            settings: querySettings,
          },
        });
        setQuerySaved({ name: title, id: savedQueryId });
      }

      if (addToDashboard) {
        const reqBody = {
          query_id: savedQueryId ? savedQueryId : res.data.id,
        };
        await saveQueryToDashboard(
          active_project.id,
          selectedDashboards.join(', '),
          reqBody
        );
      }
      notification.success({
        message: 'Report Saved Successfully',
        duration: 5,
      });

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
    getCurrentSorter,
    savedQueryId,
  ]);

  let dashboardHelpText = 'Create a dashboard widget for regular monitoring';
  const chartOptions = (
    <div className='mt-4'>
      <Radio.Group
        value={dashboardPresentation}
        onChange={handlePresentationChange}
        className={styles.radioGroup}
      >
        {getSaveChartOptions(queryType, requestQuery)}
      </Radio.Group>
    </div>
  );

  let dashboardList;

  if (addToDashboard) {
    dashboardHelpText = 'This widget will appear on the following dashboards:';
    dashboardList = (
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
        {savedQueryId ? 'Edit' : 'Save'}
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
            {savedQueryId ? 'Edit Report' : 'Create New Report'}
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
          {chartOptions}
          {/* <div className={`pt-2 ${styles.linkText}`}>Help others to find this query easily?</div> */}
          <React.Fragment>
            <div className={'pt-6 flex items-center'}>
              <Switch
                onChange={toggleAddToDashboard}
                checked={addToDashboard}
                className={styles.switchBtn}
                checkedChildren='On'
                unCheckedChildren='Off'
                disabled={queryType === QUERY_TYPE_PROFILE}
              />
              {queryType != QUERY_TYPE_PROFILE ? (
                <Text extraClass='m-0' type='title' level={6} weight='bold'>
                  Add to Dashboard
                </Text>
              ) : (
                <Text
                  extraClass='m-0 italic'
                  type='title'
                  level={9}
                  color='grey'
                >
                  Add to Dashboard Unavailable for Profiles
                </Text>
              )}
            </div>
            {queryType != QUERY_TYPE_PROFILE ? (
              <Text
                extraClass={`pt-1 ${styles.noteText}`}
                mini
                type={'paragraph'}
              >
                {dashboardHelpText}
              </Text>
            ) : null}
            {dashboardList}
          </React.Fragment>
        </div>
      </Modal>
    </>
  );
}

export default connect(null, { fetchWeeklyIngishtsMetaData })(SaveQuery);
