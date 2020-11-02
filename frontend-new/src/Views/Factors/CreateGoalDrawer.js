import React from 'react';
import {
  Drawer, Button
} from 'antd';
import { SVG, Text } from 'factorsComponents';
import { NavLink } from 'react-router-dom';

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

<div className={' fa--query_block bordered '}>
    <div className={'flex flex-col justify-center items-center'} style={{ height: '300px' }}>
        <p style={{ color: '#bbb' }}>CoreQuery reusable drawer components comes here..</p>
        <p className={'mt-2'} style={{ color: '#bbb' }}>{'Click on \'Find Insights\' to view Insights page.'}</p>
    </div>
        <div className={'flex justify-between items-center'}>
            <Button><SVG name={'calendar'} extraClass={'mr-1'} />Last Week </Button>
            <NavLink to="/factors/insights"><Button type="primary">Find Insights</Button></NavLink>
        </div>
</div>

      </Drawer>
  );
};

export default CreateGoalDrawer;
