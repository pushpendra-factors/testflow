 
import React, { useState, useEffect } from 'react';
import Sankey from './SankeyChart';
import Header from '../../AppLayout/Header';
import SearchBar from '../../../components/SearchBar';
import {
  Row, Col, Button, Spin
} from 'antd';
 
 
import { connect, useSelector, useDispatch } from 'react-redux';

import _ from 'lodash';
import { useHistory } from 'react-router-dom';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary'; 
import HeaderContents from './HeaderContents'; 
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
// import matchEventName from './Utils/MatchEventNames';
import QueryBuilder from './QueryBuilder'; 

const PathAnalysisReport = ({ 
  activeProject,
  activeInsights,
  activeQuery

}) => { 
  const [fetchingIngishts, SetfetchingIngishts] = useState(false); 
  const [collapse, setCollapse] = useState(false);
  const history = useHistory();
  const dispatch = useDispatch(); 

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true }); 
    
    return () => {
      dispatch({ type: 'RESET_ACTVIVE_INSIGHTS', payload: null });
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false }); 
    };
  }, [activeProject]); 

  useEffect(()=>{
if(activeInsights){
  setCollapse(true)
}
else{
  setCollapse(false)
}
  },[activeInsights, activeQuery]);
 

  return (
    <>
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Explain Error '} subtitle={'We are facing trouble loading Explain. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>

        {fetchingIngishts ? <Spin size={'large'} className={'fa-page-loader'} /> :
          <>

            <HeaderContents activeQuery={activeQuery} />
            <div className={'fa-container'}>
              <div className={'mt-24'}> 
                <QueryBuilder 
                collapse={collapse}
                setCollapse={setCollapse}
                activeQuery={activeQuery}
                />
                {activeInsights && <div id={'fa-report-container'}>
                  <Sankey activeQuery={activeQuery} sankeyData={activeInsights} />
                </div> 
                  } 
              </div>
            </div>

          </>
        }
      </ErrorBoundary>
    </>
  );
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project, 
    activeQuery: state.pathAnalysis.activeQuery,
    activeInsights: state.pathAnalysis.activeInsights,
  };
};
export default connect(mapStateToProps, null)(PathAnalysisReport);

