import React, { useEffect, useState, useCallback } from 'react';
import { getWebAnalyticsRequestBody } from '../utils';
import { initialState } from '../../CoreQuery/utils';
import { useSelector, useDispatch } from 'react-redux';
import { getWebAnalyticsData } from '../../../reducers/coreQuery/services';
import { Spin } from 'antd';
import TableUnits from './TableUnits';
import CardUnit from './CardUnit';
import NoDataChart from '../../../components/NoDataChart';
import { DASHBOARD_LAST_REFRESHED } from '../../../reducers/types';
import ErrorBoundary from '../../../ErrorBoundary';
import { FaErrorComp, FaErrorLog } from 'Components/factorsComponents';

function WebsiteAnalytics({
  webAnalyticsUnits,
  setwidgetModal,
  durationObj,
  dashboardRefreshState,
  resetDashboardRefreshState
}) {
  const { active_project } = useSelector((state) => state.global);
  const [resultState, setResultState] = useState(initialState);
  const dispatch = useDispatch();
  const [lastRefesh, setLastRefesh] = useState(null);
  const fetchData = useCallback(
    async (refresh = false) => {
      try {
        const reqBody = getWebAnalyticsRequestBody(
          webAnalyticsUnits,
          durationObj
        );
        setResultState({ ...initialState, loading: true });
        const dashboardId = webAnalyticsUnits[0].dashboard_id;
        const response = await getWebAnalyticsData(
          active_project.id,
          reqBody,
          dashboardId,
          refresh,
          false
        );

        setResultState({
          ...initialState,
          data: response.data.result,
          refreshed_at: response.data.refreshed_at
        });
        setLastRefesh(response?.data?.refreshed_at);
        resetDashboardRefreshState();
      } catch (err) {
        console.log(err);
        resetDashboardRefreshState();
        setResultState({ ...initialState, error: true });
      }
    },
    [
      active_project.id,
      durationObj,
      webAnalyticsUnits,
      resetDashboardRefreshState
    ]
  );

  useEffect(() => {
    if (dashboardRefreshState.inProgress) {
      fetchData(true);
    } else {
      fetchData(false);
    }
  }, [dashboardRefreshState.inProgress, durationObj, fetchData]);

  useEffect(() => {
    dispatch({
      type: DASHBOARD_LAST_REFRESHED,
      payload: lastRefesh
    });
  }, [lastRefesh, dispatch]);

  if (resultState.loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (resultState.error) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  if (resultState.data) {
    const tableUnits = webAnalyticsUnits.filter(
      (unit) => unit.presentation === 'pt'
    );
    const cardUnits = webAnalyticsUnits.filter(
      (unit) => unit.presentation === 'pc'
    );

    return (
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='small'
            title='Widget Error'
            subtitle='We are facing trouble loading this widget. Drop us a message on the in-app chat.'
            className='h-full'
          />
        }
        onError={FaErrorLog}
      >
        {cardUnits.length ? (
          <CardUnit
            resultState={resultState}
            setwidgetModal={setwidgetModal}
            cardUnits={cardUnits}
            data={resultState.data}
          />
        ) : null}
        {tableUnits.length ? (
          <TableUnits
            resultState={resultState}
            setwidgetModal={setwidgetModal}
            tableUnits={tableUnits}
            data={resultState.data}
          />
        ) : null}
      </ErrorBoundary>
    );
  }

  return null;
}

export default WebsiteAnalytics;
