import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider,Skeleton, Button  } from 'antd';
import Sidebar from '../../components/Sidebar'; 
import {Text, SVG} from 'factorsComponents';
import { Link } from 'react-router-dom'; 
import TextLib from './TextLib';
import ButtonLib from './ButtonLib';
import ColorLib from './ColorLib';
import SwitchLib from './SwitchLib';
import RadioLib from './RadioLib';
import CheckBoxLib from './CheckBoxLib';
import ModalLib from './ModalLib';


function componentsLib() {
    const { Content } = Layout;
  return ( 
        <Layout>
        <Sidebar />
                <Layout className="fa-content-container">
                <Content> 
                    <div className="px-16 pt-8 pb-20 bg-white min-h-screen"> 


                    <ColorLib />
                    <TextLib />
                    <ButtonLib />
                    <SwitchLib />
                    <RadioLib />
                    <CheckBoxLib />
                    {/* <ModalLib /> */}
                


                    </div> 
                </Content>
            </Layout>
        </Layout>

  );
}

export default componentsLib;
