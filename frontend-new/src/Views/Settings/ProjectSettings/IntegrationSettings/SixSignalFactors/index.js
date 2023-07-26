import React from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import ConnectedScreen from './ConnectedScreen';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';

function SixSignalFactorsIntegration({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent
}) {
  const { isFeatureConnected: isFactorsDeanonymisationConnected } =
    useFeatureLock(FEATURES.INT_FACTORS_DEANONYMISATION);
  useEffect(() => {
    if (isFactorsDeanonymisationConnected) {
      setIsActive(true);
    }
  }, [isFactorsDeanonymisationConnected, setIsActive]);

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with 6Signal Factors integrations' />
      }
      onError={FaErrorLog}
    >
      {isFactorsDeanonymisationConnected && <ConnectedScreen />}

      <div className='mt-4 flex' data-tour='step-11'>
        {kbLink && (
          <a
            className='ant-btn ml-2 '
            target='_blank'
            href={kbLink}
            rel='noreferrer'
          >
            View documentation
          </a>
        )}
      </div>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings
})(SixSignalFactorsIntegration);
