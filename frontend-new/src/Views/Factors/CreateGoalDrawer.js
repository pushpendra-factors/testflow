import React from 'react';
import {
  Drawer, Button
} from 'antd';
import { SVG, Text } from 'factorsComponents';

const title = (props) => {
  return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex'}>
          <SVG name={'templates_cq'} size={24} />
          <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>New Goal</Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Button size={'large'} type="text" onClick={() => props.onClose()}><SVG name="times"></SVG></Button>
        </div>
      </div>
  );
};

const CreateGoalDrawer = (props) => {
  return (
        <Drawer
        title={title(props)}
        placement="left"
        closable={false}
        visible={props.visible}
        onClose={props.onClose}
        getContainer={false}
        width={'600px'}
        className={'fa-drawer'}
      >

      </Drawer>
  );
};

export default CreateGoalDrawer;
