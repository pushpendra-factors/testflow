import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select
} from 'antd';
import { SVG, Text } from 'factorsComponents';
import { NavLink } from 'react-router-dom';
import GroupSelect from '../../components/QueryComposer/GroupSelect';
import { fetchEventNames } from 'Reducers/coreQuery/middleware';
import {connect} from 'react-redux';

const { Option } = Select;

 

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

  const [EventNames, SetEventNames] = useState();


  const onChangeGroupSelect1 = (grp, value) => {
    setShowDropDown(false);
    setEvent1(value[0]);
    // console.log(`selectedevent1-- ${grp} ${value[0]}`);
  }
  const onChangeGroupSelect2 = (grp, value) => {
    setShowDropDown2(false);
    setEvent2(value[0]);
    // console.log(`selected-event2-- ${grp} ${value[0]}`);
  }

  // const onChange = (value) => {
  //   setShowDropDown(false);
  //   setEvent1(value);
  //   console.log(`selected ${value}`);
  // }
  // const onChangeDropDown2 = (value) => {
  //   setShowDropDown2(false);
  //   setEvent2(value);
  //   console.log(`onChangeDropDown2 ${value}`);
  // }
  
  // const onBlur = ()  =>{
  //   console.log('blur');
  // }
  
  // const onFocus = ()  =>{
  //   console.log('focus');
  // }
  
  // const onSearch = (val) => {
  //   console.log('search:', val);
  // }
  // const onChangeGroupSelect = (val,grp) => {
  //   console.log('onChangeGroupSelect:', val, grp);
  // }

  useEffect(()=>{

    if(!props.GlobalEventNames){
      const getData = async () => {
        await props.fetchEventNames(props.activeProject.id);
      };
      getData();  
    } 
    if(props.GlobalEventNames){
      // const EventNames1 = props.GlobalEventNames.map((item)=>{
      //   return [item]
      // });
      SetEventNames(props.GlobalEventNames);
      
    }  
  },[props.GlobalEventNames])

const [eventCount, SetEventCount] = useState(1);

const [showDropDown, setShowDropDown] = useState(false);
const [event1, setEvent1] = useState(null);

const [showDropDown2, setShowDropDown2] = useState(false);
const [event2, setEvent2] = useState(null);

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

          <Row gutter={[24, 4]}>
                  <Col span={12}>
                      <div className={`fa-dasboard-privacy--card border-radius--medium p-4 ${eventCount===1 ? 'selected': null}`} onClick={()=>SetEventCount(1)}>
                          <div className={'flex flex-col justify-between items-start'}>  
                                  <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Analyze a single event</Text>
                                  <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Eg: users who joined the webinar</Text> 
                          </div>
                      </div>
                  </Col>
                  <Col span={12}>
                      <div className={`fa-dasboard-privacy--card border-radius--medium p-4 ${eventCount===2 ? 'selected': null}`} onClick={()=>SetEventCount(2)}>
                          <div className={'flex flex-col justify-between items-start'}>  
                                  <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Analyze a user journey</Text>
                                  <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Eg: Visited pricing and then signed up</Text> 
                          </div>
                      </div>
                  </Col>
          </Row> 
          
          <Row gutter={[24, 4]}>
              <Col span={24}>
                <div  className={'mt-4'}> 
                      
                      <div className={'flex items-center'}>
                        {event1 &&  <>
                        <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'} style={{height:'24px', width: '24px'}}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{1}</Text> </div> 
                        <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>Users who</Text>
                        {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                        </>}
                        <div className='relative' style={{height: '42px'}}>
                          {!showDropDown && !event1 && <Button onClick={()=>setShowDropDown(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>{eventCount === 2 ? 'Add First event': 'Add an event'}</Button> }
                          { showDropDown && <>
                            <GroupSelect 
                              groupedProperties={[
                                {             
                                label: 'Most Recent',
                                icon: 'fav',
                                values: EventNames
                                }
                              ]}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect1(group, val)}
                              // onClickOutside={() => closeDropDown()}
                              />
                          {/* <Select
                              showSearch
                              style={{ width: 280, position: 'absolute', top:0, left:'8px' }}
                              placeholder="Search Events"
                              optionFilterProp="children"
                              onChange={onChange}
                              onFocus={onFocus}
                              size={'large'}
                              onBlur={onBlur}
                              onSearch={onSearch} 
                              filterOption={(input, option) =>
                                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                              }
                            > 
                            {EventNames.map((item,index)=>{
                              return <Option key={index} value={item}>{item}</Option> 
                            })}; 
                            </Select> */}
                            </>
                            }

                            {event1 && !showDropDown  && <Button type={'link'} size={'large'} style={{maxWidth: '220px',textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }} className={'ml-2'} ellipsis onClick={()=>{
                              setShowDropDown(true); 
                              }} >{event1}</Button> 
                            } 
                        </div> 
                      </div>
                </div>
              </Col>
          </Row>

          {eventCount === 2 &&
          <Row gutter={[24, 4]}>
              <Col span={24}>
                <div  className={'mt-4'}> 
                      
                      <div className={'flex items-center'}>
                        {event2 &&  <>
                        <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'} style={{height:'24px', width: '24px'}}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{1}</Text> </div> 
                        <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>And then</Text>
                        {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                        </>}
                        <div className='relative' style={{height: '42px'}}>
                          {!showDropDown2 && !event2 && <Button onClick={()=>setShowDropDown2(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>Add next event</Button> }
                          { showDropDown2 && <>

                            <GroupSelect 
                              groupedProperties={[
                                {             
                                label: 'Most Recent',
                                icon: 'fav',
                                values: EventNames
                                }
                              ]}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect2(group, val)}
                              // onClickOutside={() => closeDropDown()}
                              />
                          
                          {/* <Select
                              showSearch
                              style={{ width: 280, position: 'absolute', top:0, left:'8px' }}
                              placeholder="Search Events"
                              optionFilterProp="children"
                              onChange={onChangeDropDown2}
                              onFocus={onFocus}
                              size={'large'}
                              onBlur={onBlur}
                              onSearch={onSearch} 
                              filterOption={(input, option) =>
                                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                              }
                            > 
                            {EventNames.map((item,index)=>{
                              return <Option key={index} value={item}>{item}</Option> 
                            })}; 
                            </Select> */}
                            </>
                            }

                            {event2 && !showDropDown2  && <Button type={'link'} size={'large'} style={{maxWidth: '220px',textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }} className={'ml-2'} ellipsis onClick={()=>{
                              setShowDropDown2(true); 
                              }} >{event2}</Button> 
                            } 
                        </div> 
                      </div>
                </div>
              </Col>
          </Row>
          } 

    <div className={'flex flex-col justify-center items-center'} style={{ height: '300px' }}>
        {/* <p style={{ color: '#bbb' }}>CoreQuery reusable drawer components comes here..</p>
        <p className={'mt-2'} style={{ color: '#bbb' }}>{'Click on \'Find Insights\' to view Insights page.'}</p> */}
    </div>
        <div className={'flex justify-between items-center'}>
            <Button size={'large'}><SVG name={'calendar'} extraClass={'mr-1'} />Last Week </Button>
            <NavLink to="/factors/insights"><Button type="primary" size={'large'}>Find Insights</Button></NavLink>
        </div>
</div>

      </Drawer>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project, 
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values, 
  };
};
export default connect(mapStateToProps, {fetchEventNames})(CreateGoalDrawer);
