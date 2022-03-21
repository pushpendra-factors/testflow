import React, { useEffect, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import { SVG, Text } from '../../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Button, message, notification } from 'antd';
import Header from '../../AppLayout/Header';
import { connect } from 'react-redux';
import { getHubspotContact, setActiveProject, fetchDemoProject } from 'Reducers/global';
import { meetLink } from '../../../utils/hubspot';

function DashboardAfterIntegration({setaddDashboardModal, getHubspotContact, currentAgent, setActiveProject, fetchDemoProject, projects}) {
    const [dataLoading, setdataLoading] = useState(true);
    const [ownerID, setownerID] = useState();
    const history = useHistory();

    const switchProject = () => {
        fetchDemoProject().then((res) => {
            let id = res.data[0];
            let selectedProject = projects.filter(project => project.id === id);
            selectedProject = selectedProject[0];
            localStorage.setItem('activeProject', selectedProject?.id);
            setActiveProject(selectedProject);
            history.push('/');
            notification.success({
              message: 'Project Changed!',
              description: `You are currently viewing data from ${selectedProject.name}`
            });
        });
      };

    useEffect(() => {
        let email = currentAgent.email;
        getHubspotContact(email).then((res) => {
            console.log('get hubspot contact success', res.data)
            setownerID(res.data.hubspot_owner_id)
        }).catch((err) => {
            console.log(err.data.error)
        });
    }, []);

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
                <Header>
                    <div className={'rounded-lg border-2 border-gray-200 w-full h-24 mt-8'}>
                            <div className='w-20 float-left mt-2 ml-4 mr-4 mb-1'>
                                <img src='assets/images/NoData.png'/>
                            </div>
                            <div className={'mt-4 mb-4'}>
                                <Text type={'title'} level={4} color={'grey-2'} weight={'bold'} extraClass={'m-0 mt-2 mb-1'}>
                                    Complete Project Setup
                                </Text>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mb-1'}>
                                    Are you done connecting to all your data sources?
                                </Text>
                            </div>
                            <div className={'float-right -mt-20 pt-2 mr-8'}>
                                <Button type={'link'} style={{backgroundColor:'white'}} className={'mt-2'} onClick={()=> history.push('/welcome')}>Setup Assist<SVG name={'Arrowright'} size={16} extraClass={'ml-1'} color={'blue'} /></Button>
                            </div>
                    </div>
                </Header>

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
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>It should take us a day to bring in and process all your data.</Text>
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Until then, explore the <Link onClick={()=> switchProject()}>Demo Project</Link>
                    </Text>
                    {/* { dataLoading ? 
                    <div className={'rounded-lg border-2 border-gray-400 w-11/12 mt-6'}>
                        <Text type={'title'} level={6} color={'grey'} extraClass={'m-0 mt-2 -mb-1'}>
                           We don’t have any data yet. While we fetching your metrics,
                        </Text>
                        <Button type={'text'} color={'grey-2'} className={'mb-2'} onClick={()=> switchProject()}>Explore our Demo Project<SVG name={'Arrowright'} size={16} extraClass={'ml-1'} color={'grey'} /></Button>
                    </div>
                    : */}
                    <div className={'mt-6'}>
                        <Button type={'primary'} size={'large'} className={'w-full'} onClick={() => setaddDashboardModal(true)}>Create your first dashboard</Button>
                        <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2 mb-2'}>
                            or
                        </Text>
                        <a href={meetLink(ownerID)} target='_blank' ><Button type={'default'} size={'large'} className={'w-full'}>Need Help?</Button></a>
                    </div>
                    {/* } */}
                </div>
                
            </ErrorBoundary>
        </>
    );

}

const mapStateToProps = (state) => ({
    currentAgent: state.agent.agent_details,
    projects: state.global.projects
});

export default connect(mapStateToProps, { getHubspotContact , setActiveProject, fetchDemoProject})(DashboardAfterIntegration);