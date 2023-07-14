import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Spin, Tag } from 'antd';
import { connect, useSelector } from 'react-redux';
import _, { isEmpty } from 'lodash';

import { Text, SVG, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { useHistory } from 'react-router-dom';
import {
  fetchSavedPathAnalysis,
  removeSavedQuery,
  fetchPathAnalysisInsights
} from 'Reducers/pathAnalysis';
import PathAnalysisReport from './PathAnalysisReport';
import PathAnalysisLP from './landingPage';

const Factors = ({
  activeProject,
  fetchSavedPathAnalysis,
  currentProjectSettings
}) => {
  const [loadingTable, SetLoadingTable] = useState(true);
  const [fetchingIngishts, SetfetchingIngishts] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const history = useHistory();
  const [loading, setLoading] = useState(true);
  const [durationObj, setDurationObj] = useState();

  useEffect(() => {
    fetchSavedPathAnalysis(activeProject?.id).then(() => {
      setLoading(false);
    });
  }, [activeProject]);

  if (loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }
  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Path Analysis Error '}
            subtitle={
              'We are facing trouble loading Path Analysis. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        {fetchingIngishts ? (
          <Spin size={'large'} className={'fa-page-loader'} />
        ) : (
          <>
            <PathAnalysisLP
              SetfetchingIngishts={SetfetchingIngishts}
              setShowReport={setShowReport}
            />
          </>
        )}
      </ErrorBoundary>
    </>
  );
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
  };
};
export default connect(mapStateToProps, { fetchSavedPathAnalysis })(Factors);
