import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Spin, notification, Button } from 'antd';
import FaDatepicker from 'Components/FaDatepicker';
import ControlledComponent from 'Components/ControlledComponent';
import { SVG as Svg, Text } from 'Components/factorsComponents';
import {
  selectAccountPayload,
  selectEditInsightsMetricStatus,
  selectInsightsConfig
} from 'Reducers/accountProfilesView/selectors';
import {
  fetchInsightsConfig,
  updateInsightsQueryMetric
} from 'Reducers/accountProfilesView/services';
import { setInsightsDuration } from 'Reducers/accountProfilesView/actions';
import { EMPTY_OBJECT } from 'Utils/global';
import { selectSegmentBySegmentId } from 'Reducers/timelines/selectors';
import useFeatureLock from 'hooks/useFeatureLock';
import { SEGMENT_INSIGHTS_LEARN_MORE_LINK } from 'Utils/constants';
import { FEATURES } from 'Constants/plans.constants';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import InsightsWidget from './InsightsWidget';
import SegmentKpisOverview from './SegmentKpisOverview';
import { DEFAULT_DATE_RANGE } from './accountInsightsConstants';
import SegmentCompareDropdown from './SegmentCompareDropdown';
import EditMetricModal from './EditMetricModal';

export default function AccountsInsights() {
  const dispatch = useDispatch();
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  const insightsConfig = useSelector(selectInsightsConfig);
  const accountPayload = useSelector(selectAccountPayload);
  const segment = useSelector((state) =>
    selectSegmentBySegmentId(state, accountPayload?.segment.id)
  );
  const areInsightsAvailable = segment.long_run_comp === true;
  const featureFlag = useFeatureLock(FEATURES.FEATURE_SEGMENT_KPI);

  const featureLocked =
    featureFlag.isLoading === false && featureFlag.isFeatureLocked === true;

  const editMetricStatus = useSelector(selectEditInsightsMetricStatus);
  const activeProject = useSelector((state) => state.global.active_project);
  const [isFetchDone, setIsFetchDone] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editWidget, setEditWidget] = useState(EMPTY_OBJECT);
  const [editWidgetGroupId, setEditWidgetGroupId] = useState(null);

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

  const handleEditModalClose = useCallback(() => {
    setShowEditModal(false);
  }, []);

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

  const handleEditMetric = useCallback((widget, widgetGroupId) => {
    setEditWidgetGroupId(widgetGroupId);
    setEditWidget(widget);
    setShowEditModal(true);
  }, []);

  const handleSave = useCallback(
    (selectedWidget, newName) => {
      if (
        editWidget.q_me !== selectedWidget.value ||
        editWidget.d_name !== newName
      ) {
        dispatch(
          updateInsightsQueryMetric({
            projectId: activeProject.id,
            widgetGroupId: editWidgetGroupId,
            widgetId: editWidget.id,
            metric:
              selectedWidget.value !== editWidget.q_me
                ? selectedWidget.value
                : undefined,
            metricName: newName !== editWidget.d_name ? newName : undefined
          })
        );
      } else {
        handleEditModalClose();
        notification.success({
          message: 'Success',
          description: 'Metric updated successfully',
          duration: 2
        });
      }
    },
    [editWidget, editWidgetGroupId, handleEditModalClose]
  );

  const handleLearnMoreClick = useCallback(() => {
    window.open(SEGMENT_INSIGHTS_LEARN_MORE_LINK, '_blank');
  }, []);

  useEffect(() => {
    if (
      areInsightsAvailable &&
      insightsConfig.completed !== true &&
      insightsConfig.loading !== true &&
      isFetchDone === false &&
      featureLocked === false
    ) {
      setIsFetchDone(true);
      dispatch(fetchInsightsConfig(activeProject.id));
    }
  }, [
    areInsightsAvailable,
    isFetchDone,
    insightsConfig.loading,
    insightsConfig.completed,
    activeProject.id,
    featureLocked
  ]);

  useEffect(() => {
    if (editMetricStatus.completed === true) {
      handleEditModalClose();
    }
  }, [editMetricStatus.completed, handleEditModalClose]);

  const isLoading =
    featureFlag.isLoading === true ||
    insightsConfig.loading === true ||
    (insightsConfig.completed !== true && insightsConfig.error !== true);

  if (areInsightsAvailable === false && featureLocked === false) {
    return (
      <div className='flex justify-center items-center flex-col gap-y-1 flex-1'>
        <img src='../../../../assets/icons/pana.svg' alt='loader-man' />
        <Text type='title' level={8} extraClass='mb-0' color='character-title'>
          Working hard to build your reports...
        </Text>
        <Text type='title' level={8} extraClass='mb-0' color='character-title'>
          Check back in some time
        </Text>
      </div>
    );
  }

  if (featureLocked === true) {
    return (
      <div className='flex justify-center items-center flex-col gap-y-4'>
        <img
          alt='no-data'
          src='../../../../assets/images/locked-segment-insights.svg'
        />
        <div className='flex justify-center items-center flex-col gap-y-2'>
          <Text
            level={3}
            type='title'
            extraClass='mb-0'
            color='character-primary'
          >
            {`Your plan doesn't have this feature`}
          </Text>
          <Text type='title' color='character-secondary' extraClass='mb-0'>
            This feature is not included in your current plan. Please upgrade to
            use this feature
          </Text>
          <div className='flex gap-x-4 justify-center items-center'>
            <Button
              onClick={handleLearnMoreClick}
              icon={<Svg size={16} color='#8c8c8c' name='arrowUpRightSquare' />}
            >
              Learn more
            </Button>
            <Button
              onClick={() =>
                handlePlanUpgradeClick(FEATURES.FEATURE_SEGMENT_KPI)
              }
              type='primary'
            >
              Upgrade now
            </Button>
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-1'>
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
          <div className='flex items-center gap-x-1'>
            <Svg name='pieChart' color='#8C8C8C' />
            <div className='flex items-center gap-x-3'>
              <Text type='title' color='black' extraClass='mb-0' level={7}>
                {accountPayload.segment.name}
              </Text>
              <Svg name='compare' color='#8C8C8C' />
              <SegmentCompareDropdown />
            </div>
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
            onEditMetricClick={handleEditMetric}
            editWidgetGroupId={editWidgetGroupId}
          />
        ))}
        <EditMetricModal
          visible={showEditModal}
          onCancel={handleEditModalClose}
          savedMetricName={editWidget.d_name}
          savedMetric={editWidget.q_me}
          onSave={handleSave}
          isLoading={editMetricStatus.loading}
          metricDataType={editWidget.q_me_ty}
        />
      </div>
    );
  }

  return null;
}
