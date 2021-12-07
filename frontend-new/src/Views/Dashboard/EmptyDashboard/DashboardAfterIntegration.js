import React from 'react';
import { Link } from 'react-router-dom';
// import NoDataChart from '../../../components/NoDataChart';
import { SVG, Text } from '../../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';

function DashboardAfterIntegration() {

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
                <div
                    style={{marginTop:'20em'}}
                    className={
                    'flex justify-center flex-col items-center fa-dashboard--no-data-container'
                    }
                >
                    <img alt='no-data' src='assets/images/Group 880.png' className={'mb-2'} />
                    <Text type={'title'} level={6} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>
                        Create a dashboard to moniter your metrics in one place.
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Learn  <Link to='/project-setup' className={'text-sm font-semibold'}>Dashboard Basics </Link>
                    </Text>

                    <div className={'rounded-lg border-2 border-gray-400 w-11/12 mt-4'}>
                        <Text type={'title'} level={6} color={'grey'} extraClass={'m-0 mt-2 -mb-4'}>
                           We don’t have any data yet. While we fetching your metrics,
                        </Text>
                        <Text type={'title'} level={6} color={'grey-2'} weight={'bold'} extraClass={'m-0 mb-2'}>
                           Explore our Demo Project <span style={{fontSize:'30px'}}>→</span>
                        </Text>
                    </div>
                </div>
                
            </ErrorBoundary>
        </>
    );

}

export default DashboardAfterIntegration;