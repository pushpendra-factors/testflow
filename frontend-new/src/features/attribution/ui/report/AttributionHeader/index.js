import React, { useEffect, memo, useState, useContext } from 'react';
import _ from 'lodash';
import { Button, Tabs } from 'antd';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from 'factorsComponents';
import { QuestionCircleOutlined } from '@ant-design/icons';
import SaveAttributionQuery from 'Attribution/ui/report/SaveAttributionQuery';
import { addShadowToHeader } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';

import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import AppModal from 'Components/AppModal';
import useQuery from 'hooks/useQuery';
const { TabPane } = Tabs;

function AttributionHeader({
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
  const [ShowAddToDashModal, setShowAddToDashModal] = useState(false);
  // for showing modal on closing unsaved report
  const [visible, setVisible] = useState(false);
  const [showSaveOrUpdateModal, setShowSaveOrUpdateModal] = useState(false);
  let [helpMenu, setHelpMenu] = useState(false);
  const routerQuery = useQuery();
  const paramQueryId = routerQuery.get('queryId');

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
    window.addEventListener('popstate', function (event) {
      window.history.pushState(null, document.title, window.location.href);
    });
  }, []);

  const handleCloseButton = (close = false) => {
    if (!savedQueryId && requestQuery !== null && !close) {
      setVisible(true);
    } else if (navigatedFromDashboard?.id) {
      history.push({
        pathname: ATTRIBUTION_ROUTES.reports,
        state: { dashboardWidgetId: navigatedFromDashboard.id }
      });
    } else {
      history.push(ATTRIBUTION_ROUTES.reports);
    }
  };

  const renderReportTitle = () => (
    <Text
      type='title'
      level={5}
      weight='bold'
      extraClass='m-0 mt-1'
      lineHeight='small'
    >
      {queryTitle || 'New Attribution'}
    </Text>
  );

  const renderReportCloseIcon = () => (
    <Button
      size='large'
      type='default'
      onClick={() => handleCloseButton(false)}
    >
      Close
    </Button>
  );

  const renderLogo = () => (
    <Button
      size='large'
      type='text'
      onClick={() => handleCloseButton(true)}
      icon={<SVG size={32} name='Brand' />}
    />
  );

  const renderSaveQueryComp = () => {
    if (!requestQuery) return null;

    return (
      <SaveAttributionQuery
        showSaveQueryModal={showSaveQueryModal}
        setShowSaveQueryModal={setShowSaveQueryModal}
        ShowAddToDashModal={ShowAddToDashModal}
        setShowAddToDashModal={setShowAddToDashModal}
        queryType={queryType}
        requestQuery={requestQuery}
        queryTitle={queryTitle}
        savedQueryId={savedQueryId}
        showSaveOrUpdateModal={showSaveOrUpdateModal}
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

  const saveAndClose = () => {
    setVisible(false);
    if (paramQueryId) {
      setShowSaveOrUpdateModal({ update: true });
      setTimeout(() => {
        handleCloseButton(true);
      }, 1500);
    } else {
      setShowSaveOrUpdateModal({ save: true });
    }
  };

  return (
    <div id='app-header' className='bg-white z-50 flex-col  px-8 w-full fixed'>
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
          <div className='pr-2'>{renderSaveQueryComp()}</div>
          {renderReportCloseIcon()}
        </div>
      </div>

      {renderReportTabs()}
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
            {paramQueryId ? 'Save and Close' : 'Save as New'}
          </Button>
          <Button
            type='default'
            style={{ width: '168px', height: '32px' }}
            className='mx-4 my-2'
            onClick={() => handleCloseButton(true)}
          >
            Close without saving
          </Button>
        </div>
      </AppModal>
    </div>
  );
}

export default memo(AttributionHeader);
