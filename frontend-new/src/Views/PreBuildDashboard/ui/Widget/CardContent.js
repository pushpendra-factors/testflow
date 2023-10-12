import React, { useEffect, useMemo } from 'react';
import { Spin } from 'antd';
import { get } from 'lodash';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  DASHBOARD_WIDGET_SECTION,
  reverse_user_types,
  presentationObj,
  QUERY_TYPE_KPI,
  apiChartAnnotations,
  CHART_TYPE_TABLE
} from 'Utils/constants';
import NoDataChart from 'Components/NoDataChart';
import { useSelector } from 'react-redux';
import {
  FaErrorComp,
  SVG,
  Text,
  FaErrorLog
} from 'Components/factorsComponents';
import KPIAnalysis from '../../../Dashboard/KPIAnalysis';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  DASHBOARD_PRESENTATION_KEYS
} from 'Components/SaveQuery/saveQuery.constants';
import { ErrorBoundary } from 'react-error-boundary';

function CardContent({ unit, resultState, durationObj, breakdown }) {
  let content = null;

  const queryType = 'kpi';


  if (resultState.loading) {
    content = (
      <div className='flex justify-center items-center w-full h-full'>
        <Spin size='small' />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  if (resultState.apiCallStatus && !resultState.apiCallStatus.required) {
    content = (
      <div className='flex justify-center flex-col items-center w-full h-full px-2 text-center'>
        <SVG name='nodata' />
        <Text type='title' color='grey' extraClass='mb-0'>
          {resultState.apiCallStatus.message}
        </Text>
      </div>
    );
  }

  if (resultState.data) {
    const reportSelectedChart = get(
      unit,
      'chart_setting.ty',
      apiChartAnnotations[CHART_TYPE_TABLE]
    );

    const selectedDashboardPresentation = get(
      unit,
      'chart_setting.pr',
      DEFAULT_DASHBOARD_PRESENTATION
    );

    const dashboardPresentation =
    selectedDashboardPresentation === DASHBOARD_PRESENTATION_KEYS.CHART
    ? reportSelectedChart
    : apiChartAnnotations[CHART_TYPE_TABLE];

    const kpiData = unit.me.map(obj => {
      const { inter_e_type, na, d_na, ...rest } = obj; // Use destructuring to exclude "inter_e_type" and "na"
      return { ...rest, label: d_na };; // Return the object without "inter_e_type" and "na"
    })


    if (queryType === QUERY_TYPE_KPI) {
      content = (
        <KPIAnalysis
          kpis={kpiData}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          section={DASHBOARD_WIDGET_SECTION}
          breakdown={breakdown.length ? [
            {
              property: breakdown?.[0]?.na,
              prop_type: 'categorical',
              display_name: breakdown?.[0]?.d_na,
            }
          ] : []}
          unit={unit}
          arrayMapper={[]}
          durationObj={durationObj}
        />
      );
    }
  }

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
      {content}
    </ErrorBoundary>
  );
}

export default CardContent;
