import React from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';

function Dashboard() {
  return (<>

            <Header>
                <div className="w-full h-full py-4 flex flex-col justify-center items-center">
                    <SearchBar />
                </div>
            </Header>
            <div className={'mt-16'}>

                    <ProjectTabs />
            </div>

  </>);
}

export default Dashboard;
