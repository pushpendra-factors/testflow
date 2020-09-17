/* eslint-disable */
import React from 'react';
import {
  Layout, Breadcrumb, Row, Col, Divider, Skeleton, Button
} from 'antd';
import Sidebar from '../../components/Sidebar';
import { Text, SVG } from 'factorsComponents';
import { Link } from 'react-router-dom';
import TextLib from './TextLib';
import ButtonLib from './ButtonLib';
import ColorLib from './ColorLib';
import SwitchLib from './SwitchLib';
import RadioLib from './RadioLib';
import CheckBoxLib from './CheckBoxLib';
import ModalLib from './ModalLib';
import IconsLib from './IconsLib';

function componentsLib() {
  const { Content } = Layout;
  return (
    <Layout>
      <Sidebar />
      <Layout className="fa-content-container">
        <Content>
          <div className="pt-4 pb-24 bg-white min-h-screen">
            <div className={'fa-container'}>
              <ColorLib />
              <TextLib />
              <ButtonLib />
              <SwitchLib />
              <RadioLib />
              <CheckBoxLib />
              <IconsLib />
              {/* <ModalLib /> */}
            </div>
          </div>
        </Content>
      </Layout>
    </Layout>

  );
}

export default componentsLib;
