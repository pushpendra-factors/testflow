import React, {
  useCallback,
  useContext,
  useEffect,
  useState,
  useRef,
} from 'react';
import AnalysisHeader from './AnalysisHeader';
import ReportContent from './ReportContent';
import { FaErrorComp, FaErrorLog, SVG, Text } from '../../../components/factorsComponents';
import QueryComposer from '../../../components/QueryComposer';
import AttrQueryComposer from '../../../components/AttrQueryComposer';
import { ErrorBoundary } from 'react-error-boundary';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import WeeklyInsights from '../WeeklyInsights';
import { fetchWeeklyIngishts } from '../../../reducers/insights';
import { connect, useDispatch } from 'react-redux';
import styles from './index.module.scss';

import ComposerBlock from 'Components/QueryCommons/ComposerBlock';

import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION
} from '../../../utils/constants';

import { Collapse, Button, Modal } from 'antd';
import { isArray } from 'highcharts';

const { Panel } = Collapse;

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  breakdownType,
  activeProject,
  fetchWeeklyIngishts,
  ...rest
}) {
  const renderedCompRef = useRef(null);
  const { setNavigatedFromDashboard } = useContext(CoreQueryContext);
  const [activeTab, setActiveTab] = useState(1);

  const [queryOpen, setQueryOpen] = useState(true);
  // const [insights, setInsights] = useState(null);
  const dispatch = useDispatch();

  const handleBreadCrumbClick = useCallback(() => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
  }, [setNavigatedFromDashboard, setShowResult]);

  const getCurrentSorter = useCallback(() => {
    if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
      return renderedCompRef.current.currentSorter;
    }
    return {};
  }, []);

  function changeTab(key) {
    // console.log('current tab is=-->>',key);
    setActiveTab(key);
  }

  useEffect(() => {
    return () => {
      dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: false });
      dispatch({ type: 'RESET_WEEKLY_INSIGHTS', payload: false });
    };
  }, [dispatch, activeProject]);

  useEffect(() => {
    if(requestQuery) {
      setQueryOpen(false);
    } else {
      setQueryOpen(true);
    }
    

  }, [requestQuery])

  const renderQueryHeader = () => {
    if(queryOpen) return null;

    let eventList = [];
    if(requestQuery.cl === 'attribution') {
      eventList = [requestQuery.query.ce];
    } else {
      eventList = isArray(requestQuery) ? requestQuery[0].ewp : requestQuery.ewp;
    }

    return (
      <ComposerBlock blockTitle={'EVENTS'} isOpen={true} showIcon={false}
        extraClass={``}>
        <div className={`flex`}>{<Button

          type='link'
        >
          {requestQuery && eventList[0] && eventList[0].na}
        </Button>}
          {` `}
          {<span className={`${styles.query_header__content__trail}`}>...</span>}
        </div>
      </ComposerBlock>
    )
    // return (
    //   <div className={`fa--query_block`}>
    //     {requestQuery && eventList[0] 
    //     && <Text level={6} type={'title'} extraClass={'m-0'} weight={'bold'}>{eventList[0].na}</Text>}
    //   </div>
    // )
    // return (<div className={`${styles.query_header} flex flex-col fa-act-header`}>
    //     <span className={`${styles.query_header__link}`}>{queryType} Analysis</span>
    //     <div className={`${styles.query_header__content}`}>
    //       {requestQuery && eventList?.map((ev, index) => {
    //         return (
    //           <>
    //             <Text level={6} type={'title'} extraClass={'m-0'} weight={'bold'}>{ev?.na}</Text>
    //             <span className={`${styles.query_header__content__trail}`}>...</span>
    //           </>
    //         )
    //       })}
          
    //     </div>
    // </div>)
  }

  const renderComposer = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
      return (
        <QueryComposer
          queries={rest.composerFunctions.queries}
          runQuery={rest.composerFunctions.runQuery}
          eventChange={rest.composerFunctions.queryChange}
          queryType={queryType}
          queryOptions={rest.queryOptions}
          setQueryOptions={rest.composerFunctions.setExtraOptions}
          runFunnelQuery={rest.composerFunctions.runFunnelQuery}
          activeKey={rest.composerFunctions.activeKey}
          collapse={rest.composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return <AttrQueryComposer runAttributionQuery={rest.composerFunctions.runAttributionQuery}
        collapse={rest.composerFunctions.showResult}
        setCollapse={() => setQueryOpen(false)}
      />;
    }
  }

  const renderQueryComposerNew = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_ATTRIBUTION) {
      return (
        <div className={`query_card_cont ${queryOpen? `query_card_open` : `query_card_close`}`} onClick={(e) => !queryOpen && setQueryOpen(true)}>
          <div className={`query_composer`}>{renderComposer()}</div>
          <Button size={'large'} className={`query_card_expand`}><SVG name={'expand'} size={20} ></SVG>Expand</Button>
        </div>
      )
    }

    /*
      <Collapse activeKey={queryOpen? ["1"] : ["0"]} onChange={() => setQueryOpen(!queryOpen)} className={`fa-query-edit ${queryOpen? 'query-open': ''}`} expandIcon={() => 
            <SVG size={30}  name={`sliders`} extraClass={`query_header_icon`}></SVG>}>
            <Panel id={queryOpen? 'query_header' : ''} header={renderQueryHeader()} key="1">
              <div>
                
              </div>
            </Panel>
          </Collapse>
    */
    return null;
  }

  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={handleBreadCrumbClick}
        queryType={queryType}
        queryTitle={querySaved}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
        changeTab={changeTab}
        activeTab={activeTab}
        getCurrentSorter={() => getCurrentSorter()}
      />
      <div className='mt-24 px-8'>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Analyse Results Error'}
              subtitle={
                'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
          {Number(activeTab) === 1 && (
            <>
              {renderQueryComposerNew()}
              {requestQuery && <ReportContent
                breakdownType={breakdownType}
                queryTitle={querySaved}
                queryType={queryType}
                renderedCompRef={renderedCompRef}
                {...rest}
              />}
            </>
          )}

          {Number(activeTab) === 2 && (
            <WeeklyInsights
              requestQuery={requestQuery}
              queryType={queryType}
              queryTitle={querySaved}
            />
          )}
        </ErrorBoundary>
      </div>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  insights: state.insights,
});

export default connect(mapStateToProps, { fetchWeeklyIngishts })(ReportsLayout);
