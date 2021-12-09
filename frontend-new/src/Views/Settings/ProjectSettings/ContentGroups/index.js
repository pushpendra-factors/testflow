import React, { useState, useEffect } from 'react';
import {
  Row, Col, Switch, Menu, Dropdown, Button, Tabs, Table, Tag, Space, message
} from 'antd';
import { Text, SVG } from 'factorsComponents'; 
import { connect } from 'react-redux';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware';
import { MoreOutlined } from '@ant-design/icons';
import {removeSmartEvents, fetchSmartEvents} from 'Reducers/events';
import ContentGroupForm from './ContentGroupForm';

const { TabPane } = Tabs;



function ContentGroups({smart_events, fetchEventNames, activeProject, removeSmartEvents, fetchSmartEvents}) { 

    const [smartEvents, setsmartEvents] = useState(null); 
    const [showSmartEventForm, setShowSmartEventForm] = useState(false); 
    const [seletedEvent, setSeletedEvent] = useState(null); 


    const menu = (values) => {
      return (
      <Menu> 
        <Menu.Item key="0" onClick={() => confirmRemove(values)}>
          <a>Remove Event</a> 
        </Menu.Item> 
      </Menu>
      );
    };

const columns = [

    {
      title: 'Title',
      dataIndex: 'name',
      key: 'name', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Description',
      dataIndex: 'source',
      key: 'source', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
        title: 'Values',
        dataIndex: 'source',
        key: 'source', 
        render: (text) => <span className={'capitalize'}>{text}</span>
      },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (values) => (
        <Dropdown overlay={() => menu(values)} trigger={['hover']}>
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  const editEvent = (values) =>{
    setSeletedEvent(values); 
    setShowSmartEventForm(true);
  }

  const confirmRemove = (values) =>{ 
    removeSmartEvents(values?.project_id, values?.id).then(()=>{
      message.success("Custom Event removed!")
      fetchSmartEvents(values?.project_id);
    }).catch((err)=>{
      message.error(err?.data?.error);
      console.log('error in removing Smartevent:', err)
    });
    return false
  }

  useEffect(()=>{
    fetchEventNames(activeProject.id);
    if(smart_events){
        let smartEventsArray = [];
        smart_events?.map((item,index)=>{
            smartEventsArray.push({
                key: index,
                name: item.name, 
                source: item?.expr?.source, 
                actions: item, 
              });
        });
        setsmartEvents(smartEventsArray);

    }
  },[smart_events])
  

  return (
    <>
        <div className={'mb-10 pl-4'}>
        {!showSmartEventForm && <> 
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Content Groups</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} onClick={() =>   {setSeletedEvent(null);setShowSmartEventForm(true)}}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-4'}>
            <Col span={24}>  
            <div className={'mt-6'}>
                <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0'}>A content group refers to a collection of logically related URLs that makes up your overall website’s content. For example a collection of blog articles written with a specific intend on your blog. By defining a content group to identify all such pages on the site, you can analyse common traits across many such pages at one go. You can define upto 3 content groups. Learn <a href='#'>more</a></Text>
                <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0 mt-2'}>Currently, content groups can be used to drill down the factors default event <code>Website Session</code></Text>
                
                <Table className="fa-table--basic mt-4" 
                columns={columns} 
                dataSource={smartEvents} 
                pagination={false}
                />
            </div>  
        </Col> 
        </Row> 
        </>
        }
        {showSmartEventForm && <>  
                <ContentGroupForm seletedEvent={seletedEvent} setShowSmartEventForm={setShowSmartEventForm} /> 
        </>
        }
      </div>
    </>

  );
}

const mapStateToProps = (state) => ({
    smart_events: state.events.smart_events,
    activeProject: state.global.active_project,
  });

  export default connect(mapStateToProps, {fetchEventNames, removeSmartEvents, fetchSmartEvents})(ContentGroups); 