import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select
} from 'antd';
import { SVG, Text } from 'factorsComponents'; 
import GroupSelect from '../../components/QueryComposer/GroupSelect';
import { fetchEventNames } from 'Reducers/coreQuery/middleware';
import { fetchGoalInsights, fetchFactorsModels } from 'Reducers/factors';
import {connect} from 'react-redux';
import { useHistory } from 'react-router-dom';
import _ from 'lodash';
import moment from 'moment';
 

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
  const history = useHistory();
  const { Option } = Select;

  const [EventNames, SetEventNames] = useState([]);
  const [eventCount, SetEventCount] = useState(1);

  const [showDropDown, setShowDropDown] = useState(false);
  const [event1, setEvent1] = useState(null);
  
  const [showDropDown2, setShowDropDown2] = useState(false);
  const [event2, setEvent2] = useState(null);

  const [showDateTime, setShowDateTime] = useState(false);
  const [dateTime, setDateTime] = useState(null);
  const [insightBtnLoading, setInsightBtnLoading] = useState(false);

  const onChangeGroupSelect1 = (grp, value) => {
    setShowDropDown(false);
    setEvent1(value[0]); 
  }
  const onChangeGroupSelect2 = (grp, value) => {
    setShowDropDown2(false);
    setEvent2(value[0]); 
  }
  const onChangeDateTime = (grp, value) => {
    setShowDateTime(false); 
    setDateTime(value); 
  }

  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  } 
  const factorsModels = !_.isEmpty(props.factors_models) && _.isArray(props.factors_models) ? props.factors_models.map((item)=>{return [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]}) : [];
  
  useEffect(()=>{ 
    // if(!props.GlobalEventNames || !factorsModels){
      //   const getData = async () => {
        //     await props.fetchEventNames(props.activeProject.id);
        //     await props.fetchFactorsModels(props.activeProject.id);
        //   };
        //   getData();    
        // }
    if(props.GlobalEventNames){ 
      SetEventNames(props.GlobalEventNames); 
    }  
  },[props.activeProject, props.GlobalEventNames, props.factors_models])

const factorsDataFormat = {
  name: "123",
  rule: {
      st_en: "",
      en_en: "",
      vs: true,
      rule: {
          ft: []
      }
  }
};

const getInsights = (projectID, isJourney=false) =>{  
  setInsightBtnLoading(true); 
  const calcModelId = props.factors_models.filter((item)=>{   
    const generateStringArray = [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]; 
    if (_.isEqual(dateTime,generateStringArray)){  
      return item
    } 
  });
  // console.log("calcModelId",calcModelId[0].mid);
  
  let factorsData = {
    ...factorsDataFormat,
    rule:{
       ...factorsDataFormat.rule,
       st_en: event2 ? event1 : null,
       en_en: event2 ? event2 : event1

    }  
  } 
  const getData = async () => {
    await props.fetchGoalInsights(projectID, isJourney, factorsData, calcModelId[0].mid); 
  };
  getData().then(()=>{
    setInsightBtnLoading(false);
    history.push('/explain/insights'); 
  });
}

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
                        </>}
                        <div className='relative' style={{height: '42px'}}>
                          {!showDropDown && !event1 && <Button onClick={()=>setShowDropDown(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>{eventCount === 2 ? 'Add First event': 'Add an event'}</Button> }
                          { showDropDown && <>
                            <GroupSelect 
                              groupedProperties={EventNames ? [
                                {             
                                label: 'MOST RECENT',
                                icon: 'fav',
                                values: EventNames
                                }
                              ]:null}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect1(group, val)}
                              onClickOutside={() => setShowDropDown(false)}
                              /> 
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
                        <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'} style={{height:'24px', width: '24px'}}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{2}</Text> </div> 
                        <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>And then</Text>
                        {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                        </>}
                        <div className='relative' style={{height: '42px'}}>
                          {!showDropDown2 && !event2 && <Button onClick={()=>setShowDropDown2(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>Add next event</Button> }
                          { showDropDown2 && <>

                            <GroupSelect 
                               groupedProperties={EventNames ? [
                                {             
                                label: 'MOST RECENT',
                                icon: 'fav',
                                values: EventNames
                                }
                              ]:null}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect2(group, val)}
                              onClickOutside={() => setShowDropDown2(false)}
                              /> 
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

    <div className={'flex flex-col justify-center items-center'} style={{ height: '50px' }}> 
    </div>
        <div className={'flex justify-between items-center'}>

          <div className={'relative'}>
            {!showDateTime && <Button size={'large'} onClick={()=>setShowDateTime(true)}><SVG name={'calendar'} extraClass={'mr-1'} />{dateTime ? dateTime : 'Select Date Range'} </Button>}
            {showDateTime && 
            <GroupSelect 
                    groupedProperties={factorsModels ? [
                    {             
                    label: 'MOST RECENT',
                    icon: 'fav',
                    values: factorsModels
                    }
                  ]:null}
                  placeholder="Select Date Range "
                  optionClick={(group, val) => onChangeDateTime(group, val)}
                  onClickOutside={() => setShowDateTime(false)}
                />  
            }
          </div> 
            <Button type="primary" size={'large'} loading={insightBtnLoading} disabled={!(event1 && dateTime)} onClick={()=>getInsights(props.activeProject.id, eventCount===2?true:false )}>Find Insights</Button>
        </div>
</div>

      </Drawer>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project, 
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    factors_models: state.factors.factors_models
  };
};
export default connect(mapStateToProps, {fetchEventNames, fetchGoalInsights, fetchFactorsModels})(CreateGoalDrawer);
