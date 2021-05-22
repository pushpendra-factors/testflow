import React, { useCallback, useEffect, useState, useContext } from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../../../components/factorsComponents';
import { Button } from 'antd';
import moment from 'moment';
import { EVENT_BREADCRUMB } from '../../../utils/constants';
import SaveQuery from '../../../components/SaveQuery';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { useHistory } from 'react-router-dom';

function AnalysisHeader({
  queryType,
  onBreadCrumbClick,
  requestQuery,
  queryTitle,
  setQuerySaved,
  breakdownType,
}) {
  const [showSaveModal, setShowSaveModal] = useState(false);
  const { navigatedFromDashboard } = useContext(CoreQueryContext);
  const history = useHistory();

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

  const handleCloseDashboardQuery = useCallback(() => {
    history.push({
      pathname: '/',
      state: { dashboardWidgetId: navigatedFromDashboard },
    });
  }, [history, navigatedFromDashboard]);

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, [addShadowToHeader]);

  return (
    <div
      id='app-header'
      className={`bg-white z-50	flex fixed items-center justify-between py-3 px-8 ${styles.topHeader}`}
    >
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
            level={7}
            extraClass={'m-0 mt-1'}
            color={'grey'}
            lineHeight={'small'}
          >
            {queryTitle
              ? `Reports / ${EVENT_BREADCRUMB[queryType]} / ${queryTitle}`
              : `Reports / ${
                  EVENT_BREADCRUMB[queryType]
                } / Untitled Analysis${' '}
            ${moment().format('DD/MM/YYYY')}`}
          </Text>
        </div>
      </div>
      <div className='flex items-center'>
        <SaveQuery
          requestQuery={requestQuery}
          visible={showSaveModal}
          setVisible={setShowSaveModal}
          queryType={queryType}
          setQuerySaved={setQuerySaved}
          breakdownType={breakdownType}
        />

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
        ) : null}
      </div>
    </div>
  );
}

export default AnalysisHeader;
