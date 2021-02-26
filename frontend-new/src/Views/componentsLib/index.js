/* eslint-disable */
import React from 'react';
import {
  Layout
} from 'antd';
import TextLib from './TextLib';
import ButtonLib from './ButtonLib';
import ColorLib from './ColorLib';
import SwitchLib from './SwitchLib';
import RadioLib from './RadioLib';
import CheckBoxLib from './CheckBoxLib';
import ModalLib from './ModalLib';
import IconsLib from './IconsLib';
import NumberLib from './NumberLib';

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
              <NumberLib />

              {/* <ModalLib /> */}
            </div> 

  );
}

export default componentsLib;
