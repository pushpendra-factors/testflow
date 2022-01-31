import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select, message
} from 'antd';
import { SVG, Text } from 'factorsComponents'; 
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import { fetchEventNames, getUserProperties, getEventProperties } from 'Reducers/coreQuery/middleware'; 
import { fetchGoalInsights, fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel } from 'Reducers/factors';
import {connect} from 'react-redux';
import { useHistory } from 'react-router-dom';
import _, { isEmpty } from 'lodash';
import FilterBlock from '../../components/QueryComposer/FilterBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
import EventFilterBy from './DrawerUtil/EventFilterBy';
// import MomentTz from 'Components/MomentTz';
import moment from 'moment-timezone';  
import factorsai from 'factorsai';


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

  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';  
  moment.tz.setDefault(timeZone);


  const [TrackedEventNames, SetTrackedEventNames] = useState([]);

  const [EventNames, SetEventNames] = useState([]);
  const [eventCount, SetEventCount] = useState(1);

  const [showDropDown, setShowDropDown] = useState(false);
  const [event1, setEvent1] = useState(null);
  
  const [showDropDown2, setShowDropDown2] = useState(false);
  const [event2, setEvent2] = useState(null);
  
  const [showFtDropDown, setshowFtDropDown] = useState(false);
  const [globalFilter, setglobalFilter] = useState([]);
  const [filterLoader, setfilterLoader] = useState(false);

  const [showDateTime, setShowDateTime] = useState(false);
  const [dateTime, setDateTime] = useState(null);
  const [insightBtnLoading, setInsightBtnLoading] = useState(false);

  const [filtersEvent1, setfiltersEvent1] = useState([]);
  const [showEventFilter1DD, setEventFilter1DD] = useState(false);
  const [filtersEvent2, setfiltersEvent2] = useState([]);
  const [showEventFilter2DD, setEventFilter2DD] = useState(false);

  const [filters, setfilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({
    user: []
});
const [filterDD, setFilterDD] = useState(false);


const operatorMap = {
  "=": "equals",
  "!=": "notEqual",
  contains: "contains",
  "does not contain": "notContains",
  "<": "lesserThan",
  "<=": "lesserThanOrEqual",
  ">": "greaterThan",
  ">=": "greaterThanOrEqual",
};

const getFilters = (filters) => {
  const result = [];
  filters.forEach((filter) => {
    filter.values.forEach((value, index) => {
      result.push({
        en: filter.props[2],
        lop: !index ? "AND" : "OR",
        op: operatorMap[filter.operator],
        pr: filter.props[0],
        ty: filter.props[1],
        va: value,
      });
    });
  });
  return result;
};

  const onChangeGroupSelect1 = (grp, value) => {
    setShowDropDown(false); 
    // console.log("value-->",value);
    // console.log("event value-->",value);
    setEvent1(value[0]); 
  }
  // const onChangeFilter = (grp, value) => {
  //   setshowFtDropDown(false)
  //   console.log('onChangeFilter',value)
  //   setfilterLoader(true);
  //   fetchUserPropertyValues(props.activeProject.id,value[0]).then(res => {
  //     setglobalFilter([ ...globalFilter , 
  //       {
  //         "key": value[0],
  //         "vl": res.data
  //       }
  //   ]); 
  //   console.log('fetchUserPropertyValues',res);
  //   setfilterLoader(false);
  //   });
  //   // setEvent1(value[0]); 
  // }
  const removeFilter = (index) => {
    const fltrs = globalFilter.filter((v, i) => i !== index);
    setglobalFilter(fltrs);
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
  
  // useEffect(()=>{ 
  //   console.log('event-->>',event1);
  //   if(event1){
  //     props.getEventProperties(props.activeProject.id,event1)
  //   }
  // },[event1]); 

  useEffect(()=>{ 
    // if(!props.GlobalEventNames || !factorsModels){
      //   const getData = async () => {
        //     await props.fetchEventNames(props.activeProject.id);
        //     await props.fetchFactorsModels(props.activeProject.id);
        //   };
        //   getData();    
        // }
    // if(props.GlobalEventNames){ 
    //   SetEventNames(props.GlobalEventNames);
    
    // }
    if(props.factors_models){ 
      setDateTime(factorsModels[0]);
    } 
    if(props.activeProject && props.activeProject.id) {
      props.getUserProperties(props.activeProject.id, 'channel')
    }
    if(props.tracked_events){
      const fromatterTrackedEvents = props.tracked_events.map((item)=>{
        return [item.name]
      });
      SetTrackedEventNames(fromatterTrackedEvents);
    }  
  },[props.activeProject, props.tracked_events, props.factors_models, props.goal_insights])

  useEffect(() => {
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.user = props.userProperties;
    let  catAndNumericalProps = [];

    props.userProperties.map((item)=>{ 
      if(item[1]=='categorical' || item[1]=='numerical'){ 
        catAndNumericalProps.push(item); 
      }
    }); 
    assignFilterProps.user = catAndNumericalProps;
    setFilterProperties(assignFilterProps);

  }, [props.userProperties]);


const factorsDataFormatNew = {
  "name": "",
  "rule": {
      "st_en": {
          "na": "",
          "pr": [
              // {
              //     "en": "user",
              //     "lop": "",
              //     "op": "equals",
              //     "pr": "$user_id12",
              //     "ty": "categorical",
              //     "va": "1"
              // }
          ]
      },
      "en_en": {
          "na": "",
          "pr": []
      },
      "gpr" :[],
      "vs": true
  }
};


// const factorsDataFormat = {
//   name: "",
//   rule: {
//       st_en: "",
//       en_en: "",
//       vs: true,
//       rule: {
//           ft: []
//       }
//   }
// };

const getInsights = (projectID, isJourney=false) =>{ 
  setInsightBtnLoading(true); 
  const calcModelId = props.factors_models.filter((item)=>{   
    const generateStringArray = [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]; 
    if (_.isEqual(dateTime,generateStringArray)){  
      return item
    } 
  });
  // console.log("calcModelId",calcModelId[0].mid);
  let gprData =  getFilters(filters);
  let event1pr =  getFilters(filtersEvent1);
  let event2pr =  getFilters(filtersEvent2);
  let factorsData = {
    ...factorsDataFormatNew,
    rule:{
       ...factorsDataFormatNew.rule,
       st_en: {
         na: eventCount===2 ? (event2 ? event1 : null) :null, 
         pr: eventCount===2 ? (event2pr ? event1pr : null) :null,  
        },
        en_en: {
          na: eventCount===2 ? (event2 ? event2 : event1) : event1,
          pr: eventCount===2 ? (event2pr ? event2pr : event1pr) : event1pr, 
        },
        gpr:gprData,
       
    }  
  }  
  
  props.fetchGoalInsights(projectID, isJourney, factorsData, calcModelId[0].mid).then((data)=>{ 
      props.saveGoalInsightRules(factorsData); 
      props.saveGoalInsightModel(dateTime); 
      setInsightBtnLoading(false);
      history.push('/explain/insights');
    }).catch((err)=>{
      console.log("fetchGoalInsights catch",err);
      const ErrMsg = err?.data?.error ? err.data.error : `Oops! Something went wrong!`;
      message.error(ErrMsg);
      setInsightBtnLoading(false);
  }); 

  //Factors RUN_EXPLAIN tracking
  factorsai.track('RUN_EXPLAIN',{'query_type': 'explain'});
  
}  


const delFilter = (index) => {
  const fltrs = filters.filter((v, i) => i !== index);
  setfilters(fltrs);
} 


const addFilter = (val) => {
  // console.log("add filter", val);
  const filterState = [...filters];
  filterState.push(val);
  setfilters(filterState);
}

const closeFilter = () => {
  setFilterDD(false);
}


const renderFilterBlock = () => {
  if(filterProps) {
      const filtrs = [];

      filters.forEach((filt, id) => {
          filtrs.push(
              <div key={id} className={id !== 0? `mt-4 relative` : null}>
                  <FilterBlock activeProject={props.activeProject} 
                      index={id}
                      blockType={'global'} 
                      // filterType={'channel'} 
                      filter={filt}
                      extraClass={'filter-block--row'}
                      delBtnClass={'filter-block--delete'}
                      delIcon={`trash`}
                      deleteFilter={delFilter}
                      // typeProps={{channel: channel}} 
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                  ></FilterBlock>
              </div>
          )
      })

      if(filterDD) {
          filtrs.push(  
              <div key={filtrs.length} className={`mt-4 relative`}>
                  <FilterBlock activeProject={props.activeProject} 
                      blockType={'global'} 
                      // extraClass={styles.filterSelect}
                      delBtnClass={'filter-block--delete'}
                      // typeProps={{channel: channel}} 
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                      insertFilter={addFilter}
                      closeFilter={closeFilter} 
                      operatorProps={{
                        "categorical": [
                          '=',
                          '!='
                        ],
                        "numerical": [
                          '=',
                          '<=',
                          '>='
                        ],
                        "datetime": [
                          '='
                        ]
                      }}
                      
                  ></FilterBlock>
              </div>
          )
      } else {
          filtrs.push(
              <div key={filtrs.length} className={`flex relative justify-start`}> 
                  <Button size={'large'} loading={filterLoader} type={'text'} onClick={() => setFilterDD(true)} className={'mt-2'}><SVG name={'plus'} extraClass={'mr-1'} />{'Add Filter'} </Button>
              </div>
          )
      }
      
      return (<div className={`mt-4 relative`}>{filtrs}</div>);
  }
  
}

  return (
        <Drawer
        title={title(props)}
        placement="left"
        closable={false}
        visible={props.visible}
        onClose={props.onClose}
        getContainer={false}
        width={'650px'}
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
                      
                      <div className={'flex flex-col'}>
                      
                      <div className={'flex items-center justify-between'}>
                        <div className={'flex items-center'}> 
                            {event1 &&  <>
                            <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'} style={{height:'24px', width: '24px'}}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{1}</Text> </div> 
                            <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>Users who perform</Text> 
                            </>}
                            <div className='relative' style={{height: '42px'}}>
                              {!showDropDown && !event1 && <Button onClick={()=>setShowDropDown(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>{eventCount === 2 ? 'Add First event': 'Add an event'}</Button> }
                              { showDropDown && <>
                                <GroupSelect2 
                                  allowEmpty={true}
                                  groupedProperties={TrackedEventNames ? [
                                    {             
                                    label: 'Most Recent',
                                    icon: 'most_recent',
                                    values: TrackedEventNames
                                    }
                                  ]:null}
                                  placeholder="Select Events"
                                  optionClick={(group, val) => onChangeGroupSelect1(group, val)}
                                  onClickOutside={() => setShowDropDown(false)}
                                  /> 
                                </>
                                }

                                {event1 && !showDropDown  && <Button type={'link'} size={'large'} className={'ml-2 fa-button--truncate fa-button--truncate-sm'} ellipsis onClick={()=>{
                                  setShowDropDown(true); 
                                  }} >{event1}</Button> 
                                } 
                            </div>  
                        </div>
                        {event1 &&  <Button size={'large'} type={'text'} onClick={() => setEventFilter1DD(true)} className={'m-0'}><SVG name={'filter'} extraClass={'m-0'} /></Button> }
                      </div>
                        <EventFilterBy event={event1} setfiltersParent={setfiltersEvent1} showEventFilterDD={showEventFilter1DD} setEventFilterDD={setEventFilter1DD} />  
                      </div>
                </div>
              </Col>
          </Row>

          {eventCount === 2 &&
          <Row gutter={[24, 4]}>
              <Col span={24}>
                <div  className={'mt-4'}> 
                      
                <div className={'flex flex-col'}>
                      
                      <div className={'flex items-center justify-between'}>
                        <div className={'flex items-center'}> 

                        {event2 &&  <>
                        <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'} style={{height:'24px', width: '24px'}}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{2}</Text> </div> 
                        <Text type={'title'} level={6} weight={'regular'} color={'grey'} extraClass={'m-0'}>And then</Text>
                        {/* <Text type={'title'} level={6} weight={'bold'} color={'black'} extraClass={'m-0 ml-2'}>performed</Text> */}
                        </>}
                        <div className='relative' style={{height: '42px'}}>
                          {!showDropDown2 && !event2 && <Button onClick={()=>setShowDropDown2(true)} type={'text'} size={'large'}><SVG name={'plus'} size={14} color={'grey'} extraClass={'mr-2'}/>Add next event</Button> }
                          { showDropDown2 && <>

                            <GroupSelect2 
                              allowEmpty={true}
                               groupedProperties={TrackedEventNames ? [
                                {             
                                label: 'Most Recent',
                                icon: 'most_recent',
                                values: TrackedEventNames
                                }
                              ]:null}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeGroupSelect2(group, val)}
                              onClickOutside={() => setShowDropDown2(false)}
                              /> 
                            </>
                            }

                            {event2 && !showDropDown2  && <Button type={'link'} size={'large'} className={'ml-2 fa-button--truncate fa-button--truncate-xs'} ellipsis onClick={()=>{
                              setShowDropDown2(true); 
                              }} >{event2}</Button> 
                            }
                        </div>  
                      </div> 
                        {event2 &&  <Button size={'large'} type={'text'} onClick={() => setEventFilter2DD(true)} className={'m-0'}><SVG name={'filter'} extraClass={'m-0'} /></Button> }
                      </div> 
                          <EventFilterBy event={event2} setfiltersParent={setfiltersEvent2} showEventFilterDD={showEventFilter2DD} setEventFilterDD={setEventFilter2DD} />  
                            {/* {event2 && <EventFilterBy setfiltersParent={setfiltersEvent2} /> } */}
                      </div>
                </div>
              </Col>
          </Row>
          } 


          {/* <div className={' mt-12 border-top--thin'}> 
            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 mt-4'}>Filter by</Text> 
            {renderFilterBlock()}
          </div> */}

          {/* <div className={' mt-5 border-top--thin'}>


        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 mt-4'}>Filter by</Text>

          {globalFilter && globalFilter.map((item,index)=>{
            return <div className={'flex justify-between items-center mt-2'} key={index}>
              <Button type={'link'} size={'large'}>{item.key}</Button>
              <Button size={'small'} className={'fa-button-ghost'} onClick={()=>removeFilter(index)}><SVG name={'trash'} size={16}  extraClass={'m-0'} /></Button> 
            </div>
          })}

          <div className={'relative w-full'}>
          {!showFtDropDown && <Button size={'large'} loading={filterLoader} type={'text'} onClick={()=>setshowFtDropDown(true)} className={'mt-4'}><SVG name={'plus'} extraClass={'mr-1'} />{'Add Filter'} </Button>}
            {showFtDropDown &&  
                         <GroupSelect2 
                              groupedProperties={props.userProperties ? [
                                {             
                                label: 'Most Recent',
                                icon: 'most_recent',
                                values: props.userProperties
                                }
                              ]:null}
                              placeholder="Select Events"
                              optionClick={(group, val) => onChangeFilter(group, val)}
                              onClickOutside={() => setshowFtDropDown(false)}
                              />  
            }  
          </div>
          </div> */}

    <div className={'flex flex-col justify-center items-center'} style={{ height: '150px' }}> 
    </div>
        <div className={'flex justify-between items-center'}>

          <div className={'relative'}>
            {!showDateTime && <Button size={'large'} onClick={()=>setShowDateTime(true)}><SVG name={'calendar'} extraClass={'mr-1'} />{dateTime ? dateTime : 'Select Date Range'} </Button>}
            {showDateTime && 
            <GroupSelect2 
                    groupedProperties={factorsModels ? [
                    {             
                    label: 'Most Recent',
                    icon: 'most_recent',
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
    userProperties: state.coreQuery.userProperties,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    userProperties: state.coreQuery.userProperties,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events
  };
};
export default connect(mapStateToProps, {fetchEventNames, fetchGoalInsights, 
  fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel, getUserProperties, fetchUserPropertyValues, getEventProperties})(CreateGoalDrawer);
