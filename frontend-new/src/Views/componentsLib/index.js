import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider,Skeleton, Button  } from 'antd';
import Sidebar from '../../components/Sidebar'; 
import {Text, SVG} from 'factorsComponents';
import { Link } from 'react-router-dom'; 
import TextLib from './TextLib';
import ButtonLib from './ButtonLib';
import ColorLib from './ColorLib';

function componentsLib() {
    const { Content } = Layout;
  return ( 
        <Layout>
        <Sidebar />
                <Layout className="fa-content-container">
                <Content> 
                    <div className="px-16 pt-8 pb-20 bg-white min-h-screen"> 


                   <TextLib />
                    <ButtonLib />
                    <ColorLib />
                


                    </div> 
                </Content>
            </Layout>
        </Layout>

  );
}

export default componentsLib;
