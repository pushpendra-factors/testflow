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
  let [helpMenu, setHelpMenu] = useState(false);

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

  const handleCloseButton = () => {
    if (navigatedFromDashboard?.id) {
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
    <Button size='large' type='default' onClick={handleCloseButton}>
      Close
    </Button>
  );

  const renderLogo = () => (
    <Button
      size='large'
      type='text'
      onClick={handleCloseButton}
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
    </div>
  );
}

export default memo(AttributionHeader);
