import React, { useCallback, useEffect, useState, useContext } from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../../../components/factorsComponents';
import { Button } from 'antd';
import moment from 'moment';
import { EVENT_BREADCRUMB } from '../../../utils/constants';
import SaveQuery from '../../../components/SaveQuery';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { useHistory } from 'react-router-dom';
import { Tabs } from 'antd';
import { useSelector } from 'react-redux';
import _ from 'lodash';

const { TabPane } = Tabs;


function AnalysisHeader({
  queryType,
  onBreadCrumbClick,
  requestQuery,
  queryTitle,
  setQuerySaved,
  breakdownType,
  changeTab,
  activeTab,
  getCurrentSorter
}) {
  const [showSaveModal, setShowSaveModal] = useState(false);
  const {
    coreQueryState: { navigatedFromDashboard },
  } = useContext(CoreQueryContext);
  const history = useHistory();

  const { metadata } = useSelector((state) => state.insights);

  const addShadowToHeader = useCallback(() => {
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
          document.documentElement ||
          document.body.parentNode ||
          document.body
        ).scrollTop;
    if (scrollTop > 0) {
      document.getElementById('app-header').style.filter =
        'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
    } else {
      document.getElementById('app-header').style.filter = 'none';
    }
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

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, [addShadowToHeader]);

  // const isInsightsEnabled = metadata?.QueryWiseResult!=null || !metadata?.DashboardUnitWiseResult!=null ||  !_.isEmpty(metadata.QueryWiseResult) || !_.isEmpty(metadata?.DashboardUnitWiseResult)
  const isInsightsEnabled = (metadata?.QueryWiseResult != null && !metadata?.DashboardUnitWiseResult != null) || (!_.isEmpty(metadata?.QueryWiseResult) && !_.isEmpty(metadata?.DashboardUnitWiseResult))
  // console.log('isInsightsEnabled',isInsightsEnabled);
  return (
    <div
      id='app-header'
      className={`bg-white z-50 ${requestQuery && 'fixed'} flex-col pt-3 px-8 w-11/12 ${isInsightsEnabled ? 'pb-0 border-bottom--thin-2' : "pb-3"} ${styles.topHeader}`}
    >
      <div className={'items-center flex justify-between w-full'}>

        <div
          onClick={onBreadCrumbClick}
          className='flex items-center cursor-pointer'
        >
          <Button
            size={'large'}
            type='text'
            icon={<SVG size={32} name='Brand' />}
            className={'mr-2'}
          />
          <div>
            <Text
              type={'title'}
              level={5}
              weight={`bold`}
              extraClass={'m-0 mt-1'}
              
              lineHeight={'small'}
            >
              {queryTitle
                ? `Reports /  ${queryTitle}`
                : `Reports / ${EVENT_BREADCRUMB[queryType]
                } / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
            </Text>
          </div>
        </div>
        <div className='flex items-center'>
          {requestQuery && <SaveQuery
            requestQuery={requestQuery}
            visible={showSaveModal}
            setVisible={setShowSaveModal}
            queryType={queryType}
            setQuerySaved={setQuerySaved}
            breakdownType={breakdownType}
            getCurrentSorter={getCurrentSorter}
          />}

          {/* <Button
          size={'large'}
          type='text'
          icon={<SVG size={20} name={'threedot'} />}
          className={'ml-2'}
        ></Button> */}

          {navigatedFromDashboard ? (
            <Button
              size={'large'}
              type='text'
              icon={<SVG size={20} name={'close'} />}
              className={'ml-2'}
              onClick={handleCloseDashboardQuery}
            ></Button>
          ) : <Button
          size={'large'}
          type='text'
          icon={<SVG size={20} name={'close'} />}
          className={'ml-2'}
          onClick={handleCloseToAnalyse}
        ></Button>}
        </div>

      </div>




      {isInsightsEnabled &&
        <div className={'items-center flex justify-center w-full'}>
          <Tabs defaultActiveKey={activeTab} onChange={changeTab} className={'fa-tabs--dashboard'}>
            <TabPane tab="Reports" key="1" />
            <TabPane tab="Insights" key="2" />
          </Tabs>
        </div>
      }


    </div>
  );
}

export default AnalysisHeader;