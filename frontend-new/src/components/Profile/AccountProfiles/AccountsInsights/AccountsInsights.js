import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Spin } from 'antd';
import FaDatepicker from 'Components/FaDatepicker';
import ControlledComponent from 'Components/ControlledComponent';
import { SVG as Svg, Text } from 'Components/factorsComponents';
import {
  selectAccountPayload,
  selectInsightsConfig
} from 'Reducers/accountProfilesView/selectors';
import { fetchInsightsConfig } from 'Reducers/accountProfilesView/services';
import { setInsightsDuration } from 'Reducers/accountProfilesView/actions';
import InsightsWidget from './InsightsWidget';
import SegmentKpisOverview from './SegmentKpisOverview';
import { DEFAULT_DATE_RANGE } from './accountInsightsConstants';
import SegmentCompareDropdown from './SegmentCompareDropdown';

export default function AccountsInsights() {
  const dispatch = useDispatch();
  const insightsConfig = useSelector((state) => selectInsightsConfig(state));
  const activeProject = useSelector((state) => state.global.active_project);
  const [isFetchDone, setIsFetchDone] = useState(false);
  const accountPayload = useSelector((state) => selectAccountPayload(state));

  useEffect(() => {
    if (
      insightsConfig.completed !== true &&
      insightsConfig.loading !== true &&
      isFetchDone === false
    ) {
      setIsFetchDone(true);
      dispatch(fetchInsightsConfig(activeProject.id));
    }
  }, [insightsConfig.loading, activeProject.id, insightsConfig.completed]);

  const isLoading =
    insightsConfig.loading === true ||
    (insightsConfig.completed !== true && insightsConfig.error !== true);

  const nonAccountsWidgets = useMemo(
    () =>
      insightsConfig.config?.filter(
        (widget) => widget.wid_g_d_name !== 'Account Analysis'
      ),
    [insightsConfig.config]
  );

  const accountsAnalysisWidget = useMemo(
    () =>
      insightsConfig.config.find(
        (widget) => widget.wid_g_d_name === 'Account Analysis'
      ),
    [insightsConfig.config]
  );

  const handleDurationChange = useCallback(
    (duration) => {
      dispatch(
        setInsightsDuration({
          segmentId: accountPayload.segment.id,
          range: duration
        })
      );
    },
    [accountPayload?.segment?.id]
  );

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <Spin size='large' />
      </div>
    );
  }

  if (insightsConfig.completed === true) {
    return (
      <div className='flex flex-col gap-y-4'>
        <div className='flex justify-between items-center'>
          <FaDatepicker
            presetRange
            placement='bottomLeft'
            onSelect={handleDurationChange}
            buttonSize='default'
            className='datepicker-minWidth'
            range={
              insightsConfig.dateRange[accountPayload?.segment?.id] ??
              DEFAULT_DATE_RANGE
            }
            todayPicker={false}
            yesterdayPicker={false}
            last14daysPicker={false}
            last7daysPicker={false}
            last28daysPicker={false}
          />
          <div className='flex items-center gap-x-2'>
            <Svg name='pieChart' color='#8C8C8C' />
            <Text type='title' color='black' extraClass='mb-0' level={7}>
              {accountPayload.segment.name}
            </Text>
            <Svg name='compare' color='#8C8C8C' />
            <SegmentCompareDropdown />
          </div>
        </div>
        <ControlledComponent controller={accountsAnalysisWidget != null}>
          <SegmentKpisOverview
            dateRange={
              insightsConfig.dateRange[accountPayload?.segment?.id] ??
              DEFAULT_DATE_RANGE
            }
            widget={accountsAnalysisWidget}
          />
        </ControlledComponent>
        {nonAccountsWidgets.map((widget) => (
          <InsightsWidget
            dateRange={
              insightsConfig.dateRange[accountPayload?.segment?.id] ??
              DEFAULT_DATE_RANGE
            }
            key={widget.wid_g_id}
            widget={widget}
          />
        ))}
      </div>
    );
  }

  return null;
}
