import React, {
  useCallback,
  useEffect,
  useContext,
  memo,
  useState
} from 'react';
import cx from 'classnames';
import { Button, Tabs, Tooltip } from 'antd';
import { useSelector } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import { addShadowToHeader } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { CoreQueryContext } from 'Context/CoreQueryContext';
// import FaSelect from 'Components/FaSelect';
import { PathUrls } from 'Routes/pathUrls';

const { TabPane } = Tabs;

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
  const [hideIntercomState, setHideIntercomState] = useState(true);
  const [showSaveQueryModal, setShowSaveQueryModal] = useState(false);
  const [showUpdateQuery, setShowUpdateQuery] = useState(false);
  const [visible, setVisible] = useState(false);
  // const savedQueries = useSelector((state) =>
  //   get(state, 'queries.data', EMPTY_ARRAY)
  // );

  let location = useLocation();

  const history = useHistory();
  const {
    coreQueryState: { navigatedFromDashboard, navigatedFromAnalyse }
  } = useContext(CoreQueryContext);
  const { metadata } = useSelector((state) => state.insights);
  const { active_insight: activeInsight } = useSelector(
    (state) => state.insights
  );
  // const isInsightsEnabled =
  //   (metadata?.QueryWiseResult != null &&
  //     !metadata?.DashboardUnitWiseResult != null) ||
  //   (!isEmpty(metadata?.QueryWiseResult) &&
  //     !isEmpty(metadata?.DashboardUnitWiseResult));

  // const showReportTabs = requestQuery && isInsightsEnabled;

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
      {queryTitle
        ? `${queryTitle}`
        : `Quick board report`}
    </Text>
  );

  const renderReportCloseIcon = () => {
    // Here instead of ContextAPIs we can get this state from location state. which makes it simpler to access variables across routes
    // let navigatedFromDashboardExistingReports =
    //   location.state?.navigatedFromDashboardExistingReports;
    return (
      <Button size='large' type='default' onClick={handleCloseDashboardQuery}>
        Close
      </Button>
    );
  };

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

  const renderSaveQueryComp = () => {
    const showShareButton = true;

    return (
      <Tooltip
        placement='bottom'
        title={`${
          showShareButton
            ? 'Share'
            : 'Only weekly visitor reports can be shared for easy access'
        }`}
      >
        <Button
          // onClick={handleShareClick}
          size='large'
          type='primary'
          icon={
            <SVG
              name={'link'}
              color={`${showShareButton ? '#fff' : '#b8b8b8'}`}
            />
          }
          disabled={!showShareButton}
        >
          Share
        </Button>
      </Tooltip>
    );
  };

  const renderReportTabs = () => {
    // if (!showReportTabs) return null;
    if (!activeInsight?.Enabled) return null;
    return (
      <div className='items-center flex justify-center w-full -mt-2'>
        <Tabs
          defaultActiveKey={activeTab}
          onChange={changeTab}
          className='fa-tabs--dashboard'
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
      className={cx('bg-white z-50 flex-col  px-8 w-full', {
        fixed: requestQuery
      })}
      style={{ borderBottom: requestQuery ? '1px solid lightgray' : 'none' }}
    >
      <div className='items-center flex justify-between w-full pt-3 pb-3'>
        <div
          role='button'
          tabIndex={0}
          // onClick={onBreadCrumbClick}
          className='flex items-center cursor-pointer'
        >
          {renderLogo()}
          {renderReportTitle()}
        </div>

        <div className='flex items-center'>
          {/* {isFromAnalysisPage ? (
            <div className='pr-2 '>
              <div className='relative gap-x-2 mr-2'>
                <Dropdown overlay={actionMenu} placement='bottomRight'>
                  <Button
                    type='text'
                    icon={<SVG name={'threedot'} size={25} />}
                  />
                </Dropdown>
              </div>
            </div>
          ) : (
            ''
          )} */}
          {/* <div className='pr-2'>{renderSaveQueryComp()}</div> */}
          {renderReportCloseIcon()}
        </div>
      </div>

      {/* {renderReportTabs()} */}
    </div>
  );
}

export default memo(ReportHeader);
