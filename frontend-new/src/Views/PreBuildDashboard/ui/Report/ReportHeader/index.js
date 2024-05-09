import React, { useCallback, useEffect, useContext, memo } from 'react';
import cx from 'classnames';
import { Button } from 'antd';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import { addShadowToHeader } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import { PathUrls } from 'Routes/pathUrls';

function ReportHeader({
  isFromAnalysisPage,
  queryType,
  onBreadCrumbClick,
  requestQuery,
  queryTitle,
  changeTab,
  activeTab,
  savedQueryId,
  ...rest
}) {
  const history = useHistory();
  const {
    coreQueryState: { navigatedFromDashboard }
  } = useContext(CoreQueryContext);

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, []);

  useEffect(() => {
    window.history.pushState(null, document.title, window.location.href);
    window.addEventListener('popstate', () => {
      window.history.pushState(null, document.title, window.location.href);
    });
  }, []);

  // This checks where to route back if came from Dashboard
  const conditionalRouteBackCheck = useCallback(() => {
    if (navigatedFromDashboard?.inter_id) {
      history.push({
        pathname: PathUrls.PreBuildDashboard,
        state: { dashboardWidgetId: navigatedFromDashboard.inter_id }
      });
    } else {
      history.push(PathUrls.PreBuildDashboard);
    }
  }, [history, navigatedFromDashboard.inter_id]);

  const handleCloseFromLogo = useCallback(() => {
    conditionalRouteBackCheck();
  }, [conditionalRouteBackCheck]);

  const handleCloseDashboardQuery = useCallback(() => {
    conditionalRouteBackCheck();
  }, [conditionalRouteBackCheck]);

  const renderReportTitle = () => (
    <Text
      type='title'
      level={5}
      weight='bold'
      extraClass='m-0 mt-1'
      lineHeight='small'
    >
      {queryTitle ? `${queryTitle}` : `Quick board report`}
    </Text>
  );

  const renderReportCloseIcon = () => (
    <Button size='large' type='default' onClick={handleCloseDashboardQuery}>
      Close
    </Button>
  );
  const renderLogo = () => (
    <Button
      size='large'
      type='text'
      onClick={
        navigatedFromDashboard ? handleCloseFromLogo : handleCloseFromLogo
      }
      icon={<SVG size={32} name='Brand' />}
    />
  );

  return (
    <div
      id='app-header'
      className={cx('bg-white z-50 flex-col  px-8 w-full', {
        fixed: requestQuery
      })}
      style={{ borderBottom: requestQuery ? '1px solid lightgray' : 'none' }}
    >
      <div className='items-center flex justify-between w-full pt-3 pb-3'>
        <div
          role='button'
          tabIndex={0}
          className='flex items-center cursor-pointer'
        >
          {renderLogo()}
          {renderReportTitle()}
        </div>

        <div className='flex items-center'>{renderReportCloseIcon()}</div>
      </div>
    </div>
  );
}

export default memo(ReportHeader);
