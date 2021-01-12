import React, { useState, useEffect } from 'react';
import {
  Row, Col, Switch, Avatar, Button, Tabs, Table, Tag, Space
} from 'antd';
import { Text, SVG } from 'factorsComponents'; 
import { connect } from 'react-redux';
import SmartEventsForm from './SmartEvents/SmartEventsForm';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware';

const { TabPane } = Tabs;



function Events({smart_events, fetchEventNames, activeProject}) { 

    const [smartEvents, setsmartEvents] = useState(null); 
    const [showSmartEventForm, setShowSmartEventForm] = useState(false); 

const columns = [

    {
      title: 'Diplay name',
      dataIndex: 'name',
      key: 'name', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    }
  ];
   

  useEffect(()=>{
    fetchEventNames(activeProject.id);
    if(smart_events){
        let smartEventsArray = [];
        smart_events?.map((item,index)=>{
            smartEventsArray.push({
                key: index,
                name: item.name, 
                source: item?.expr?.source, 
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
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Events</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} onClick={() => setShowSmartEventForm(true)}><SVG name={'plus'} extraClass={'mr-2'} size={16} />New Event</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-4'}>
            <Col span={24}>  
            <div className={'mt-6'}>
                <Tabs defaultActiveKey="1" >
                            <TabPane tab="Smart Events" key="1">
                                    <Table className="ant-table--custom mt-4" 
                                    columns={columns} 
                                    dataSource={smartEvents} 
                                    pagination={false}
                                    />
                            </TabPane>
                </Tabs> 
            </div>  
        </Col> 
        </Row>
        </>
        }
        {showSmartEventForm && <>  
                <SmartEventsForm setShowSmartEventForm={setShowSmartEventForm} /> 
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

  export default connect(mapStateToProps, {fetchEventNames})(Events); 