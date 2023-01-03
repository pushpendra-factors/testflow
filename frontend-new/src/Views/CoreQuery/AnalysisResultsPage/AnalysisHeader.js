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
import { useHistory, useLocation } from 'react-router-dom';
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
import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import AppModal from '../../../components/AppModal';

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
  const [visible, setVisible] = useState(false);
  const [helpMenu, setHelpMenu] = useState(false);
  const savedQueries = useSelector((state) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );

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
    window.addEventListener('popstate', () => {
      window.history.pushState(null, document.title, window.location.href);
    });
  }, []);

  const handleCloseToAnalyse = useCallback(() => {
    if (!savedQueryId && requestQuery !== null) {
      setVisible(true);
    } else {
      history.push({
        pathname: '/analyse'
      });
      onBreadCrumbClick();
    }
  });

  const saveAndClose = () => {
    setVisible(false);
    setShowSaveQueryModal(true);
  };

  const closeWithoutSave = () => {
    history.push({
      pathname: '/analyse'
    });
    onBreadCrumbClick();
  };

  // This checks where to route back if came from Dashboard
  const conditionalRouteBackCheck = () => {
    let navigatedFromDashboardExistingReports =
      location.state?.navigatedFromDashboardExistingReports;
    if (navigatedFromDashboardExistingReports) {
      // Just moving back to / route
      history.push({
        pathname: '/'
      });
    } else {
      // Going Back to specefic Widget Where we came from
      history.push({
        pathname: '/',
        state: { dashboardWidgetId: navigatedFromDashboard.id }
      });
    }
  };
  const handleCloseDashboardQuery = useCallback(() => {
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
  }, [history, navigatedFromDashboard, requestQuery, savedQueryId]);

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
    let navigatedFromDashboardExistingReports =
      location.state?.navigatedFromDashboardExistingReports;
    return (
      <Button
        size='large'
        type='default'
        onClick={
          // This is the condition checking
          navigatedFromDashboardExistingReports || navigatedFromDashboard
            ? handleCloseDashboardQuery
            : handleCloseToAnalyse
        }
      >
        Close
      </Button>
    );
  };

  const renderLogo = () => (
    <Button
      size='large'
      type='text'
      onClick={
        navigatedFromDashboard
          ? handleCloseDashboardQuery
          : handleCloseDashboardQuery
      }
      icon={<SVG size={32} name='Brand' />}
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
            type='link'
            icon={<SVG name='Handshake' size={16} color='blue' />}
            onClick={() => {
              userflow.start(flowID);
            }}
            style={{
              display: 'inline-flex',
              alignItems: 'center'
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
  const setActions = (opt) => {
    if (opt[1] === 'help_doc') {
      window.open('https://help.factors.ai/', '_blank');
    } else if (opt[1] === 'intercom_help') {
      handleIntercomHelp();
    }
  };
  const getHelpMenu = () => {
    return helpMenu === false ? (
      ''
    ) : (
      <FaSelect
        extraClass={styles.additionalops}
        options={[
          ['Help and Support', 'help_doc'],
          ['Talk to us', 'intercom_help']
        ]}
        optionClick={(val) => setActions(val)}
        onClickOutside={() => setHelpMenu(false)}
        posRight={true}
      ></FaSelect>
    );
  };
  return (
    <div
      id='app-header'
      className={cx('bg-white z-50 flex-col  px-8 w-full', {
        fixed: requestQuery
      })}
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
          <div className='pr-2'>{renderSaveQueryComp()}</div>
          {isFromAnalysisPage ? (
            <div className='pr-2 '>
              <div className='relative'>
                <Button
                  size='large'
                  type='text'
                  icon={<QuestionCircleOutlined />}
                  onClick={() => setHelpMenu(!helpMenu)}
                ></Button>
                {getHelpMenu()}
              </div>
            </div>
          ) : (
            ''
          )}
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
        >
          <div className='text-center'>
            <div className='text-center mx-20 my-2'>
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
              Save and Close
            </Button>
            <Button
              type='default'
              style={{ width: '168px', height: '32px' }}
              className='mx-4 my-2'
              onClick={closeWithoutSave}
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
