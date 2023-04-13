import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select, message
} from 'antd';
import { SVG, Text } from 'factorsComponents';  
import { fetchEventNames, getUserProperties, getEventProperties } from 'Reducers/coreQuery/middleware';
import { fetchGoalInsights, fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel } from 'Reducers/factors';
import {connect} from 'react-redux';
import { useHistory } from 'react-router-dom'; 
import FilterBlock from '../../../components/QueryComposer/FilterBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services'; 
// import MomentTz from 'Components/MomentTz'; 
import moment from 'moment-timezone';


const EventFilterBy = (props) => { 
  const [TrackedEventNames, SetTrackedEventNames] = useState([]); 
  const [filterLoader, setfilterLoader] = useState(false); 
  const [filters, setfilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
}); 
// const [filterDD, setFilterDD] = useState(false);

const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';  
moment.tz.setDefault(timeZone);

  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  }  
  
  useEffect(()=>{   
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

  useEffect(()=>{
    if (!props.eventProperties[props.event]) {
      props.getEventProperties(props.activeProject.id, props.event);
    }
    setfilters([]);
  },[props.event])


  useEffect(() => {
    const assignFilterProps = Object.assign({}, filterProps); 
    assignFilterProps.user = props.userProperties;
    let  catAndNumericalProps = []; 

    //removing numerical type for both events and user properties
    if (props.event && props.eventProperties[props.event]) {
      let numericalEventProps =  props.eventProperties[props.event]?.filter((item)=>{  
        if(item[2]=='categorical'){ 
          return item
        }
      });
      assignFilterProps.event =  numericalEventProps;
    }
    
    props.userProperties.map((item)=>{
        if(item[2]=='categorical'){ 
          catAndNumericalProps.push(item); 
        }
      }); 
      
      assignFilterProps.user = catAndNumericalProps;
      setFilterProperties(assignFilterProps);

  }, [props.userProperties, props.eventProperties]);

 


const delFilter = (index) => {
  const fltrs = filters.filter((v, i) => i !== index);
  setfilters(fltrs);
  props.setfiltersParent(fltrs);
} 


const addFilter = (val) => { 
  const filterState = [...filters];
  filterState.push(val);
  setfilters(filterState);
  props.setfiltersParent(filterState);
}

const closeFilter = () => {
  props.setEventFilterDD(false);
}
 

const renderFilterBlock = () => {
  if(filterProps) {
      const filtrs = [];

      filters.forEach((filt, id) => {
          filtrs.push(
              <div key={id} className={`mt-0 relative flex flex-grow w-full`}>
                  <FilterBlock activeProject={props.activeProject} 
                      index={id}
                      blockType={'event'} 
                      // filterType={'channel'} 
                      filter={filt}
                      extraClass={'filter-block--row'}
                      delBtnClass={'filter-block--delete--mini'}
                      delIcon={`times`}
                      deleteFilter={delFilter}
                      event={{label: props.event}}
                      // typeProps={{channel: channel}} 
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                  ></FilterBlock>
              </div>
          )
      })

      if(props.showEventFilterDD) {
          filtrs.push(  
              <div key={filtrs.length} className={`mt-0 relative flex flex-grow w-full`}>
                  <FilterBlock activeProject={props.activeProject} 
                      blockType={'event'} 
                      // extraClass={styles.filterSelect}
                      extraClass={'filter-block--row'}
                      delBtnClass={'filter-block--delete--mini'}
                      // typeProps={{channel: channel}} 
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                      insertFilter={addFilter}
                      closeFilter={closeFilter} 
                      event={{label: props.event}}
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
      }
      
      return (<div className={`relative flex flex-col w-full`}> 
          {filtrs} 
        </div>);
  }
   
}

  return ( 


          <div className={'relative flex flex-grow w-full'}>  
            {renderFilterBlock()}
          </div>

       
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project, 
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events,
    userProperties: state.coreQuery.userProperties, 
    eventProperties: state.coreQuery.eventProperties,
  };
};
export default connect(mapStateToProps, {fetchEventNames, fetchGoalInsights, 
  fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel, getUserProperties, fetchUserPropertyValues, getEventProperties})(EventFilterBy);
