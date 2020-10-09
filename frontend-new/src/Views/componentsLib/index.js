/* eslint-disable */
import React from 'react';
import {
  Layout
} from 'antd';
import Sidebar from '../../components/Sidebar';
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

  );
}

export default componentsLib;
