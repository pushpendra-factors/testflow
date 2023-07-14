import React, {
  useCallback,
  useEffect,
  useContext,
  memo,
  useState
} from 'react';
import cx from 'classnames';
import moment from 'moment';
import { isEmpty } from 'lodash';
import { Button, Dropdown, Menu, Modal, Tabs } from 'antd';
import { useSelector } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import {
  EVENT_BREADCRUMB
  // QUERY_TYPE_ATTRIBUTION,
  // QUERY_TYPE_FUNNEL,
  // QUERY_TYPE_KPI
} from 'Utils/constants';
// import userflow from 'userflow.js';
// import { USERFLOW_CONFIG_ID } from 'Utils/userflowConfig';
import { QuestionCircleOutlined } from '@ant-design/icons';
import SaveQuery from '../../../components/SaveQuery';
import { addShadowToHeader } from './analysisResultsPage.helpers';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
// import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import AppModal from '../../../components/AppModal';
import { PathUrls } from 'Routes/pathUrls';

const { TabPane } = Tabs;

function AnalysisHeader({
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
  useEffect(() => {
    if (window.Intercom) {
      window.Intercom('update', { hide_default_launcher: true });
    }
    return () => {
      if (window.Intercom) {
        window.Intercom('update', { hide_default_launcher: false });
      }
    };
  }, []);

  const history = useHistory();
  const {
    coreQueryState: { navigatedFromDashboard, navigatedFromAnalyse }
  } = useContext(CoreQueryContext);
  const { metadata } = useSelector((state) => state.insights);
  const { active_insight: activeInsight } = useSelector(
    (state) => state.insights
  );
  const isInsightsEnabled =
    (metadata?.QueryWiseResult != null &&
      !metadata?.DashboardUnitWiseResult != null) ||
    (!isEmpty(metadata?.QueryWiseResult) &&
      !isEmpty(metadata?.DashboardUnitWiseResult));

  const showReportTabs = requestQuery && isInsightsEnabled;

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

  const saveAndClose = () => {
    setVisible(false);
    if (navigatedFromDashboard?.id || navigatedFromAnalyse?.key) {
      setShowUpdateQuery(true);
      setTimeout(() => {
        location?.state?.navigatedFromDashboardExistingReports ||
        navigatedFromDashboard
          ? conditionalRouteBackCheck()
          : closeWithoutSave();
      }, 1500);
    } else {
      setShowSaveQueryModal(true);
    }
  };

  const closeWithoutSave = () => {
    history.push({
      pathname: '/analyse'
    });
    onBreadCrumbClick();
  };

  // This checks where to route back if came from Dashboard
  const conditionalRouteBackCheck = useCallback(() => {
    let navigatedFromDashboardExistingReports =
      location.state?.navigatedFromDashboardExistingReports;
    if (navigatedFromDashboardExistingReports) {
      // Just moving back to / route
      history.push(PathUrls.Dashboard);
    } else {
      // Going Back to specefic Widget Where we came from
      history.push({
        pathname: PathUrls.Dashboard,
        state: { dashboardWidgetId: navigatedFromDashboard.id }
      });
    }
  }, [
    history,
    location.state?.navigatedFromDashboardExistingReports,
    navigatedFromDashboard.id
  ]);

  const handleCloseFromLogo = useCallback(() => {
    if (!savedQueryId && requestQuery !== null) {
      Modal.confirm({
        title:
          'This report is not yet saved. Would you like to save this before leaving?',
        okText: 'Save report',
        cancelText: 'Don’t save',
        closable: true,
        centered: true,
        onOk: () => {
          setShowSaveQueryModal(true);
        },
        onCancel: () => {
          conditionalRouteBackCheck();
        }
      });
    } else {
      conditionalRouteBackCheck();
    }
  }, [conditionalRouteBackCheck, requestQuery, savedQueryId]);

  const handleCloseDashboardQuery = useCallback(() => {
    if (!savedQueryId && requestQuery !== null) {
      setVisible(true);
    } else {
      conditionalRouteBackCheck();
    }
  }, [conditionalRouteBackCheck, requestQuery, savedQueryId]);

  const renderReportTitle = () => (
    <Text
      type='title'
      level={5}
      weight='bold'
      extraClass='m-0 mt-1'
      lineHeight='small'
    >
      {queryTitle
        ? `Reports / ${EVENT_BREADCRUMB[queryType]} / ${queryTitle}`
        : `Reports / ${EVENT_BREADCRUMB[queryType]} / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
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
    if (!requestQuery) {
      // if (
      //   queryType === QUERY_TYPE_ATTRIBUTION ||
      //   queryType === QUERY_TYPE_FUNNEL ||
      //   queryType === QUERY_TYPE_KPI
      // ) {
      //   let flowID = '';
      //   if (queryType === QUERY_TYPE_ATTRIBUTION) {
      //     flowID = USERFLOW_CONFIG_ID?.AttributionQueryBuilder;
      //   }
      //   if (queryType === QUERY_TYPE_FUNNEL) {
      //     flowID = USERFLOW_CONFIG_ID?.FunnelSQueryBuilder;
      //   }
      //   if (queryType === QUERY_TYPE_KPI) {
      //     flowID = USERFLOW_CONFIG_ID?.KPIQueryBuilder;
      //   }

      //   return (
      //     <Button
      //       size='large'
      //       type='link'
      //       icon={<SVG name='Handshake' size={16} color='blue' />}
      //       onClick={() => {
      //         userflow.start(flowID);
      //       }}
      //       style={{
      //         display: 'inline-flex',
      //         alignItems: 'center'
      //       }}
      //     >
      //       Walk me through
      //     </Button>
      //   );
      // }
      return null;
    }

    return (
      <SaveQuery
        showSaveQueryModal={showSaveQueryModal}
        setShowSaveQueryModal={setShowSaveQueryModal}
        showUpdateQuery={showUpdateQuery}
        queryType={queryType}
        requestQuery={requestQuery}
        queryTitle={queryTitle}
        savedQueryId={savedQueryId}
        {...rest}
      />
    );
  };

  const renderReportTabs = () => {
    if (!showReportTabs) return null;
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

  let handleIntercomHelp = () => {
    const w = window;
    const ic = w.Intercom;
    if (typeof ic === 'function') {
      setHideIntercomState(!hideIntercomState);
      ic('update', { hide_default_launcher: !hideIntercomState });
      ic(!hideIntercomState === true ? 'hide' : 'show');
    }
  };

  const handleActionMenuClick = (e) => {
    if (e?.key === '6') {
      handleIntercomHelp();
    } else if (e?.key === '7') {
      window.open('https://help.factors.ai/', '_blank');
    }
  };

  const actionMenu = (
    <Menu
      onClick={handleActionMenuClick}
      className={`${styles.antdActionMenu}`}
    >
      <Menu.Item key='1' disabled={!savedQueryId}>
        <SVG
          name={'envelope'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Email this report
      </Menu.Item>
      <Menu.Item key='2' disabled={!savedQueryId}>
        <SVG
          name={'SlackStroke'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Share to slack
      </Menu.Item>
      <Menu.Item key='3' disabled={!savedQueryId}>
        <SVG
          name={'addtodash'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Add to Dashboard
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key='4' disabled={!savedQueryId}>
        <SVG
          name={'edit'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Edit Details
      </Menu.Item>
      <Menu.Item key='5' disabled={!savedQueryId}>
        <SVG
          name={'TrashLight'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Delete
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key='6'>
        <SVG
          name={'headset'}
          size={18}
          color={'grey'}
          extraClass={'inline mr-2'}
        />
        Talk to us
      </Menu.Item>
      <Menu.Item key='7'>
        <QuestionCircleOutlined
          style={{ fontSize: '15px', marginRight: '12px' }}
        />
        Help and Support
      </Menu.Item>
    </Menu>
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
          // onClick={onBreadCrumbClick}
          className='flex items-center cursor-pointer'
        >
          {renderLogo()}
          {renderReportTitle()}
        </div>

        <div className='flex items-center'>
          {isFromAnalysisPage ? (
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
          )}
          <div className='pr-2'>{renderSaveQueryComp()}</div>
          {renderReportCloseIcon()}
        </div>
      </div>

      {renderReportTabs()}
      <div>
        <AppModal
          visible={visible}
          onCancel={() => setVisible(false)}
          footer={null}
          width={300}
          height={200}
          style={{ position: 'absolute', top: 60, right: 30 }}
          mask={false}
        >
          <div className='text-center'>
            <div className='text-center mx-24 my-2'>
              <SVG name={'Files'} />
            </div>
            <Text
              type='title'
              level={6}
              color='grey-2'
              className='mx-6 my-2 w-11/12'
            >
              This report contains unsaved progress.{' '}
            </Text>
            <Button
              type='primary'
              style={{ width: '168px', height: '32px' }}
              className='mx-4 my-2'
              onClick={saveAndClose}
            >
              {navigatedFromDashboard?.id || navigatedFromAnalyse?.key
                ? 'Save and Close'
                : 'Save as New'}
            </Button>
            <Button
              type='default'
              style={{ width: '168px', height: '32px' }}
              className='mx-4 my-2'
              onClick={() =>
                location?.state?.navigatedFromDashboardExistingReports ||
                navigatedFromDashboard
                  ? conditionalRouteBackCheck()
                  : closeWithoutSave()
              }
            >
              Close without saving
            </Button>
          </div>
        </AppModal>
      </div>
    </div>
  );
}

export default memo(AnalysisHeader);
