import React from 'react';
import { Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery'; 

function AppLayout() {
    const { Content } = Layout;

    return (
        <Layout>
            <Sidebar />
            <Layout className="fa-content-container">
                <Content> 
                    <div className="p-4 bg-white min-h-screen"> 
                        <CoreQuery />
                    </div> 
                </Content>
            </Layout>
        </Layout>
    )
}

export default AppLayout;