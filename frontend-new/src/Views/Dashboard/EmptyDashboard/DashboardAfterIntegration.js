import React, { useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from '../../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Button } from 'antd';
import { connect } from 'react-redux';
import { getHubspotContact } from 'Reducers/global';
import DashboardTemplates from '../../DashboardTemplates';
import { PathUrls } from 'Routes/pathUrls';
function DashboardAfterIntegration({
  setaddDashboardModal,
  getHubspotContact,
  currentAgent
}) {
  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Dashboard Overview Error'}
            subtitle={
              'We are facing trouble loading dashboards overview. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        <DashboardTemplates setaddDashboardModal={setaddDashboardModal} />
      </ErrorBoundary>
    </>
  );
}

const mapStateToProps = (state) => ({
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  getHubspotContact
})(DashboardAfterIntegration);
