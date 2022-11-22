import React, {
  useCallback,
  useEffect,
  useContext,
  memo,
  useState
} from 'react';
import cx from 'classnames';
import moment from 'moment';
import _, { get } from 'lodash';
import { Button, Modal, Tabs } from 'antd';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import {
  EVENT_BREADCRUMB,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI
} from 'Utils/constants';
import userflow from 'userflow.js';
import { USERFLOW_CONFIG_ID } from 'Utils/userflowConfig';
import { QuestionCircleOutlined } from '@ant-design/icons';
import SaveQuery from '../../../components/SaveQuery';
import { addShadowToHeader } from './analysisResultsPage.helpers';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { EMPTY_ARRAY } from 'Utils/global';


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
  const savedQueries = useSelector((state) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );


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
    coreQueryState: { navigatedFromDashboard }
  } = useContext(CoreQueryContext);
  const { metadata } = useSelector((state) => state.insights);
  const { active_insight: activeInsight } = useSelector(
    (state) => state.insights
  );
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

  useEffect(() => {
    window.history.pushState(null, document.title, window.location.href);
    window.addEventListener('popstate', ()=> {
        window.history.pushState(null, document.title,  window.location.href);
    });
  }, [])

  const handleCloseToAnalyse = useCallback(() => {
    if(!savedQueryId && requestQuery !== null) {
      Modal.confirm({
        title: 'This report is not yet saved. Would you like to save this before leaving?',
        okText: 'Save report',
        cancelText: 'Don’t save',
        closable: true,
        centered: true,
        onOk: () => {
          setShowSaveQueryModal(true);
        },
        onCancel: () => {
          history.push({
            pathname: '/analyse'
          });
          onBreadCrumbClick();
        }
      });
    } else {
      history.push({
        pathname: '/analyse'
      });
      onBreadCrumbClick();
    }
  }, [history, requestQuery, savedQueryId]);

  const handleCloseDashboardQuery = useCallback(() => {
    if(!savedQueryId && requestQuery !== null) {
      Modal.confirm({
        title: 'This report is not yet saved. Would you like to save this before leaving?',
        okText: 'Save report',
        cancelText: 'Don’t save',
        closable: true,
        centered: true,
        onOk: () => {
          setShowSaveQueryModal(true);
        },
        onCancel: () => {
          history.push({
            pathname: '/',
            state: { dashboardWidgetId: navigatedFromDashboard.id }
          });
        }
      }); 
    } else {
      history.push({
        pathname: '/',
        state: { dashboardWidgetId: navigatedFromDashboard.id }
      });
    }
  }, [history, navigatedFromDashboard, requestQuery, savedQueryId]);

  const renderReportTitle = () => (
    <Text
      type="title"
      level={5}
      weight="bold"
      extraClass="m-0 mt-1"
      lineHeight="small"
    >
      {queryTitle
        ? `Reports / ${EVENT_BREADCRUMB[queryType]} / ${queryTitle}`
        : `Reports / ${EVENT_BREADCRUMB[queryType]} / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
    </Text>
  );

  const renderReportCloseIcon = () => (
    <Button
      size="large"
      type="text"
      icon={<SVG size={20} name="close" />}
      onClick={
        navigatedFromDashboard
          ? handleCloseDashboardQuery
          : handleCloseToAnalyse
      }
    />
  );

  const renderLogo = () => (
    <Button
      size="large"
      type="text"
      onClick={
        navigatedFromDashboard
        ? handleCloseDashboardQuery
        : handleCloseDashboardQuery
      }
      icon={<SVG size={32} name="Brand" />}
    />
  );

  const renderSaveQueryComp = () => {
    if (!requestQuery) {
      if (
        queryType === QUERY_TYPE_ATTRIBUTION ||
        queryType === QUERY_TYPE_FUNNEL ||
        queryType === QUERY_TYPE_KPI
      ) {
        let flowID = '';
        if (queryType === QUERY_TYPE_ATTRIBUTION) {
          flowID = USERFLOW_CONFIG_ID?.AttributionQueryBuilder;
        }
        if (queryType === QUERY_TYPE_FUNNEL) {
          flowID = USERFLOW_CONFIG_ID?.FunnelSQueryBuilder;
        }
        if (queryType === QUERY_TYPE_KPI) {
          flowID = USERFLOW_CONFIG_ID?.KPIQueryBuilder;
        }

        return (
          <Button
            type="link"
            icon={<SVG name="Handshake" size={16} color="blue" />}
            onClick={() => {
              userflow.start(flowID);
            }}
          >
            Walk me through
          </Button>
        );
      }
      return null;
    }
    return (
      <SaveQuery
        showSaveQueryModal={showSaveQueryModal}
        setShowSaveQueryModal={setShowSaveQueryModal}
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
      <div className="items-center flex justify-center w-full -mt-2">
        <Tabs
          defaultActiveKey={activeTab}
          onChange={changeTab}
          className="fa-tabs--dashboard"
        >
          <TabPane tab="Reports" key="1" />
          <TabPane tab="Insights" key="2" />
        </Tabs>
      </div>
    );
  };

  return (
    <div
      id="app-header"
      className={cx('bg-white z-50 flex-col  px-8 w-full', {
        fixed: requestQuery
      })}
    >
      <div className="items-center flex justify-between w-full pt-3 pb-3">
        <div
          role="button"
          tabIndex={0}
          // onClick={onBreadCrumbClick}
          className="flex items-center cursor-pointer"
        >
          {renderLogo()}
          {renderReportTitle()}
        </div>

        <div className="flex items-center">
          <div className="pr-2">{renderSaveQueryComp()}</div>
          {isFromAnalysisPage ? 
                <div className="pr-2 ">
                <div className='relative'>
                <Button
                  size="large"
                  type="text"
                  shape='circle'
                  // icon={<SVG name={`Handshake`} size={16} color={'blue'} />}
                  onClick={() => {
                    const w = window;
                    const ic = w.Intercom;
                    if (typeof ic === 'function') {
                      setHideIntercomState(!hideIntercomState);
                      ic('update', { hide_default_launcher: !hideIntercomState });
                      ic(!hideIntercomState === true ? 'hide' : 'show');
                    }
                  }}
                >
                  <QuestionCircleOutlined />
                </Button>
                </div>
              </div>
              :
              ''
            }
          {renderReportCloseIcon()}
        </div>
      </div>

      {renderReportTabs()}
    </div>
  );
}

export default memo(AnalysisHeader);
