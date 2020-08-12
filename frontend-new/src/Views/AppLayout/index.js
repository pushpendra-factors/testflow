import React from 'react';
import { Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import {Text} from '../../components/factorsComponents';

function AppLayout() {
    const { Content } = Layout;

    return (
        <Layout>
            <Sidebar />
            <Layout className="site-layout">
                <Content>


                    <div className="p-4 bg-white min-h-screen">
                        <Text level={1} >Heading Style</Text>
                        <Text level={2} >Heading Style</Text>
                        <Text level={3} >Heading Style</Text>
                        <Text level={4} >Heading Style</Text>
                        <Text level={5} >Heading Style</Text>
                        <Text level={6} >Heading Style</Text>
                    
                        <Text level={5} >Heading Style</Text>
                        <Text level={6} >Heading Style</Text>
                        {/* <CoreQuery /> */}
                    </div>

                </Content>
            </Layout>
        </Layout>
    )
}

export default AppLayout;