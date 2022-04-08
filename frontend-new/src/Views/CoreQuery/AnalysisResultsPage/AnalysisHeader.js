import React, {
  useCallback,
  useEffect,
  useState,
  useContext,
  memo,
} from 'react';
import cx from 'classnames';
import moment from 'moment';
import _ from 'lodash';
import { Button, Tabs } from 'antd';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import { EVENT_BREADCRUMB } from 'Utils/constants';
import SaveQuery from '../../../components/SaveQuery';
import { addShadowToHeader } from './analysisResultsPage.helpers';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';

const { TabPane } = Tabs;

function AnalysisHeader({
  queryType,
  onBreadCrumbClick,
  requestQuery,
  queryTitle,
  changeTab,
  activeTab,
  ...rest
}) {
  const history = useHistory();
  const {
    coreQueryState: { navigatedFromDashboard },
  } = useContext(CoreQueryContext);
  const { metadata } = useSelector((state) => state.insights);
  const isInsightsEnabled =
    (metadata?.QueryWiseResult != null &&
      !metadata?.DashboardUnitWiseResult != null) ||
    (!_.isEmpty(metadata?.QueryWiseResult) &&
      !_.isEmpty(metadata?.DashboardUnitWiseResult));

  const showReportTabs = requestQuery && isInsightsEnabled;

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, []);

  const handleCloseToAnalyse = () => {
    history.push({
      pathname: '/analyse',
    });
    onBreadCrumbClick();
  };

  const handleCloseDashboardQuery = useCallback(() => {
    if (!requestQuery) {
      onBreadCrumbClick();
    } else {
      history.push({
        pathname: '/',
        state: { dashboardWidgetId: navigatedFromDashboard.id },
      });
    }
  }, [history, navigatedFromDashboard, requestQuery]);

  const renderReportTitle = () => {
    return (
      <Text
        type={'title'}
        level={5}
        weight={`bold`}
        extraClass={'m-0 mt-1'}
        lineHeight={'small'}
      >
        {queryTitle
          ? `Reports / ${EVENT_BREADCRUMB[queryType]} / ${queryTitle}`
          : `Reports / ${EVENT_BREADCRUMB[queryType]} / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
      </Text>
    );
  };

  const renderReportCloseIcon = () => {
    return (
      <Button
        size={'large'}
        type='text'
        icon={<SVG size={20} name={'close'} />}
        onClick={
          navigatedFromDashboard
            ? handleCloseDashboardQuery
            : handleCloseToAnalyse
        }
      />
    );
  };

  const renderLogo = () => {
    return (
      <Button
        size={'large'}
        type='text'
        icon={<SVG size={32} name='Brand' />}
      />
    );
  };

  const renderSaveQueryComp = () => {
    if (!requestQuery) return null;
    return (
      <SaveQuery
        queryType={queryType}
        requestQuery={requestQuery}
        queryTitle={queryTitle}
        {...rest}
      />
    );
  };

  const renderReportTabs = () => {
    if (!showReportTabs) return null;
    return (
      <div className={'items-center flex justify-center w-full'}>
        <Tabs
          defaultActiveKey={activeTab}
          onChange={changeTab}
          className={'fa-tabs--dashboard'}
        >
          <TabPane tab='Reports' key='1' />
          <TabPane tab='Insights' key='2' />
        </Tabs>
      </div>
    );
  };

  return (
    <div
      id='app-header'
      className={cx(
        'bg-white z-50 flex-col pt-3 px-8 w-11/12 w-full',
        { fixed: requestQuery },
        { 'pb-0': showReportTabs },
        { 'pb-3': !showReportTabs }
      )}
    >
      <div className={'items-center flex justify-between w-full'}>
        <div
          role='button'
          tabIndex={0}
          onClick={onBreadCrumbClick}
          className='flex items-center cursor-pointer'
        >
          {renderLogo()}
          {renderReportTitle()}
        </div>

        <div className='flex items-center gap-x-2'>
          <div className='pr-6 border-r'>{renderSaveQueryComp()}</div>
          {renderReportCloseIcon()}
        </div>
      </div>

      {renderReportTabs()}
    </div>
  );
}

export default memo(AnalysisHeader);
