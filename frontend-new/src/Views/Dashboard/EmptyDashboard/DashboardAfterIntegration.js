import React, { useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from '../../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Button } from 'antd';
import { connect } from 'react-redux';
import {
  getHubspotContact,
  setActiveProject,
  fetchDemoProject
} from 'Reducers/global';
import DashboardTemplates from '../../DashboardTemplates';
function DashboardAfterIntegration({
  setaddDashboardModal,
  getHubspotContact,
  currentAgent,
  setActiveProject,
  fetchDemoProject,
  projects
}) {
  const history = useHistory();

  useEffect(() => {
    if (currentAgent?.email != null) {
      let email = currentAgent.email;
      getHubspotContact(email)
        .then((res) => {
          console.log('get hubspot contact success', res.data);
        })
        .catch((err) => {
          console.log(err.data.error);
        });
    }
  }, [currentAgent?.email, getHubspotContact]);

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
        <div className={'rounded-lg border-2 border-gray-200 w-full h-24'}>
          <div className='w-20 float-left mt-2 ml-4 mr-4 mb-1'>
            <img
              alt='nodata'
              src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/NoData.png'
            />
          </div>
          <div className={'mt-4 mb-4'}>
            <Text
              type={'title'}
              level={4}
              color={'grey-2'}
              weight={'bold'}
              extraClass={'m-0 mt-2 mb-1'}
            >
              Have you brought in the data that matters to you?
            </Text>
            <Text
              type={'title'}
              level={7}
              color={'grey'}
              extraClass={'m-0 mb-1'}
            >
              {currentAgent?.first_name}, Factors is best when connected to all
              the data that you want to track
            </Text>
          </div>
          <div className={'float-right -mt-20 pt-2 mr-8'}>
            <Button
              type={'link'}
              style={{ backgroundColor: 'white' }}
              className={'mt-2'}
              onClick={() => history.push('/welcome')}
            >
              Finish Setup
              <SVG
                name={'Arrowright'}
                size={16}
                extraClass={'ml-1'}
                color={'blue'}
              />
            </Button>
          </div>
        </div>
        <DashboardTemplates setaddDashboardModal={setaddDashboardModal} />
      </ErrorBoundary>
    </>
  );
}

const mapStateToProps = (state) => ({
  currentAgent: state.agent.agent_details,
  projects: state.global.projects
});

export default connect(mapStateToProps, {
  getHubspotContact,
  setActiveProject,
  fetchDemoProject
})(DashboardAfterIntegration);
