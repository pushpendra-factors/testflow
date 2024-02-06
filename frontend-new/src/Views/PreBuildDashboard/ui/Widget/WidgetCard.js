import React, {
  useRef,
  useEffect,
  useCallback,
  useState,
  useMemo
} from 'react';
import _ from 'lodash';
import { connect, useSelector } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { Text, SVG } from 'Components/factorsComponents';
import { QUERY_TYPE_KPI } from 'Utils/constants';
import { getQueryData } from 'Views/PreBuildDashboard/state/services';
import {
  getPredefinedQuery,
  transformWidgetResponse
} from 'Views/PreBuildDashboard/utils';
import { selectActivePreDashboard } from 'Reducers/dashboard/selectors';
import { Select, Tooltip } from 'antd';
import CampaignMetricsDropdown from './CampaignMetricsDropdown';
import styles from './index.module.scss';
import { DashboardContext } from '../../../../contexts/DashboardContext';
import { getValidGranularityForSavedQueryWithSavedGranularity } from '../../../Dashboard/utils';
import { initialState } from '../../../CoreQuery/utils';
import CardContent from './CardContent';

const { Option } = Select;

function WidgetCard({
  unit,
  // showDeleteWidgetModal,
  durationObj,
  setOldestRefreshTime,
  onDataLoadSuccess,
  dashboardRefreshState
  // handleWidgetRefresh
}) {
  const hasComponentUnmounted = useRef(false);
  const cardRef = useRef(null);
  const history = useHistory();
  const location = useLocation();
  const [resultState, setResultState] = useState(initialState);
  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  const activeDashboard = useSelector((state) =>
    selectActivePreDashboard(state)
  );
  const [appliedBreakdown, setAppliedBreakdown] = useState(unit?.g_by);
  const dashboardFilters = useSelector(
    (state) => state.preBuildDashboardConfig.filters
  );

  const durationWithSavedFrequency = useMemo(() => {
    const savedFrequency = null;
    const queryType = 'kpi';
    if (queryType === QUERY_TYPE_KPI) {
      const frequency = getValidGranularityForSavedQueryWithSavedGranularity({
        durationObj,
        savedFrequency
      });
      return {
        ...durationObj,
        frequency
      };
    }
    return durationObj;
  }, [durationObj]);

  useEffect(() => {
    if (
      location.state &&
      location.state.dashboardWidgetId &&
      location.state.dashboardWidgetId === unit.inter_id
    ) {
      window.scrollTo({
        top: cardRef.current.getBoundingClientRect().top - 180,
        behavior: 'smooth'
      });
      location.state = undefined;
      window.history.replaceState(null, '');
    }
  }, [location, unit]);

  const getData = useCallback(
    async (refresh = false) => {
      try {
        hasComponentUnmounted.current = false;
        setResultState({
          ...initialState,
          loading: true
        });

        const queryType = QUERY_TYPE_KPI;
        const apiCallStatus = {
          required: true,
          message: null
        };

        let lastRefreshedAt = null;
        if (apiCallStatus.required) {
          const payload = getPredefinedQuery(
            unit,
            durationWithSavedFrequency,
            dashboardFilters,
            appliedBreakdown?.[0]
          );

          const res = await getQueryData(
            activeProject.id,
            payload,
            activeDashboard?.inter_id
          );

          if (unit?.inter_id === 1) {
            res.data = transformWidgetResponse(res.data.result || res.data);
          }

          if (!hasComponentUnmounted.current) {
            onDataLoadSuccess({ unitId: unit.inter_id });
          }
          if (queryType === QUERY_TYPE_KPI && !hasComponentUnmounted.current) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );

            setResultState({
              ...initialState,
              data: res.data.result || res.data
            });
          }

          if (lastRefreshedAt != null && !hasComponentUnmounted.current) {
            setOldestRefreshTime((currValue) => {
              if (currValue == null || lastRefreshedAt < currValue) {
                return lastRefreshedAt;
              }
              return currValue;
            });
          }
        } else {
          setResultState({
            ...initialState,
            apiCallStatus
          });
        }
      } catch (err) {
        console.log(err);
        console.log(err.response);
        if (!hasComponentUnmounted.current) {
          onDataLoadSuccess({ unitId: unit.inter_id });
        }
        setResultState({
          ...initialState,
          error: true
        });
      }
    },
    [
      durationWithSavedFrequency,
      activeProject.id,
      activeDashboard?.inter_id,
      onDataLoadSuccess,
      setOldestRefreshTime,
      appliedBreakdown,
      dashboardFilters
    ]
  );

  useEffect(() => {
    getData();
    return () => {
      hasComponentUnmounted.current = true;
    };
  }, [getData, durationWithSavedFrequency, appliedBreakdown, dashboardFilters]);

  // const handleDelete = useCallback(() => {
  //   showDeleteWidgetModal(unit);
  // }, [unit, showDeleteWidgetModal]);

  // const onWidgetRefresh = useCallback(() => {
  //   handleWidgetRefresh(unit.inter_id);
  // }, [unit.inter_id, handleWidgetRefresh]);

  // const getMenu = () => (
  //   <Menu>
  //     <Menu.Item key='0'>
  //       <a onClick={handleDelete} href='#!'>
  //         Delete Widget
  //       </a>
  //     </Menu.Item>
  //     <Menu.Item key='1'>
  //       <a onClick={onWidgetRefresh} href='#!'>
  //         Refresh
  //       </a>
  //     </Menu.Item>
  //   </Menu>
  // );

  const handleEditQuery = useCallback(() => {
    history.push({
      pathname: '/report/quick-board',
      state: {
        query: unit,
        filter: dashboardFilters,
        web_analytics: true,
        navigatedFromDashboard: unit
      }
    });
  }, [history, unit]);

  const contextValue = useMemo(
    () => ({
      handleEditQuery
    }),
    [handleEditQuery]
  );

  function handleBreakdownChange(value) {
    const result = unit?.g_by?.filter((item) => value === item.na);
    setAppliedBreakdown(result);
  }

  // metric change

  const [currMetricsValue, setCurrMetricsValue] = useState(0);

  const kpiData = unit?.me?.map((obj) => {
    const { inter_e_type, ty, na, d_na, ...rest } = obj;
    return { ...rest, metric: na, label: d_na, metricType: ty };
  });

  return (
    <div
      className={`${unit?.d_na.split(' ').join('-')} ${
        unit?.className
      } py-3 flex widget-card-top-div`}
    >
      <div
        id={`card-${unit?.inter_id}`}
        ref={cardRef}
        className={`fa-dashboard--widget-card h-full w-full flex ${styles.widgetCardCustomCSS}`}
      >
        <div className='flex justify-between items-start w-full'>
          <div className='w-full flex flex-1 flex-col h-full justify-between'>
            <div
              className={`${styles.widgetCard} flex items-center justify-between px-4`}
            >
              <div
                className='widget-card--title-container py-3 flex truncate cursor-pointer items-center w-full mr-2'
                onClick={handleEditQuery}
              >
                <div className='flex  items-center'>
                  <Tooltip title={unit?.d_na} mouseEnterDelay={0.2}>
                    <Text
                      ellipsis
                      type='title'
                      level={6}
                      weight='bold'
                      extraClass='widget-card--title m-0 mr-1 flex'
                    >
                      {unit?.d_na}
                    </Text>
                  </Tooltip>
                </div>
                <SVG
                  extraClass='widget-card--expand-icon ml-1'
                  size={20}
                  color='grey'
                  name='arrowright'
                />
              </div>
              {unit?.g_by?.length ? (
                <div className='mr-4'>
                  <Select
                    value={appliedBreakdown?.[0]?.d_na}
                    onChange={handleBreakdownChange}
                    style={{ minWidth: 120 }}
                    className='fa-select'
                    suffixIcon={
                      <SVG name='caretDown' size={16} extraClass='-mt-1' />
                    }
                  >
                    {unit?.g_by?.map((val) => (
                      <Option value={val.na}>{val.d_na}</Option>
                    ))}
                  </Select>
                </div>
              ) : null}
              <div className='flex items-center'>
                {resultState.apiCallStatus &&
                resultState.apiCallStatus.required &&
                resultState.apiCallStatus.message ? (
                  <Tooltip
                    mouseEnterDelay={0.2}
                    title={resultState.apiCallStatus.message}
                  >
                    <div className='cursor-pointer'>
                      <SVG color='#dea069' name='warning' />
                    </div>
                  </Tooltip>
                ) : null}
                {/* <Dropdown
                  placement='bottomRight'
                  overlay={getMenu()}
                  trigger={['hover']}
                >
                  <Button
                    type='text'
                    icon={<SVG size={20} name='threedot' color='grey' />}
                  />
                </Dropdown> */}
              </div>
            </div>
            {!unit?.g_by?.length ? (
              <div>
                <CampaignMetricsDropdown
                  metrics={kpiData}
                  currValue={currMetricsValue}
                  setCurrMetricsValue={setCurrMetricsValue}
                  metricsValue={resultState?.data?.[1]?.rows?.[0]}
                />
              </div>
            ) : null}
            <DashboardContext.Provider value={contextValue}>
              <CardContent
                durationObj={durationWithSavedFrequency}
                unit={unit}
                resultState={resultState}
                breakdown={appliedBreakdown}
                currMetricsValue={currMetricsValue}
              />
            </DashboardContext.Provider>
          </div>
        </div>
      </div>
      <div
        id={`resize-${unit.inter_id}`}
        className='fa-widget-card--resize-container'
      >
        <span style={{ padding: '5px 8px' }} />
      </div>
    </div>
  );
}

export default connect(null, {
  // fetchWeeklyInsights: fetchWeeklyInsightsAction
})(React.memo(WidgetCard));
