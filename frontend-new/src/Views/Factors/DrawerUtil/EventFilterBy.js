import React, { useEffect, useState } from 'react';
import {
  Drawer, Button, Row, Col, Select, message
} from 'antd';
import { SVG, Text } from 'factorsComponents'; 
// import GroupSelect from '../../../components/QueryComposer/GroupSelect';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware';
import { fetchGoalInsights, fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel } from 'Reducers/factors';
import {connect} from 'react-redux';
import { useHistory } from 'react-router-dom'; 
import moment from 'moment'; 
import FilterBlock from '../../../components/QueryComposer/FilterBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';


const EventFilterBy = (props) => { 
  const [TrackedEventNames, SetTrackedEventNames] = useState([]); 
  const [filterLoader, setfilterLoader] = useState(false); 
  const [filters, setfilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({
    user: [],
    event: []
});
// const [filterDD, setFilterDD] = useState(false);

 

  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  } 
  const factorsModels = !_.isEmpty(props.factors_models) && _.isArray(props.factors_models) ? props.factors_models.map((item)=>{return [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]}) : [];
  

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

  useEffect(() => {
    const assignFilterProps = Object.assign({}, filterProps);
    // console.log("assignFilterProps---->",assignFilterProps);
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
                      blockType={'global'} 
                      // filterType={'channel'} 
                      filter={filt}
                      extraClass={'filter-block--row'}
                      delBtnClass={'filter-block--delete--mini'}
                      delIcon={`times`}
                      deleteFilter={delFilter}
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
                      blockType={'global'} 
                      // extraClass={styles.filterSelect}
                      extraClass={'filter-block--row'}
                      delBtnClass={'filter-block--delete--mini'}
                      // typeProps={{channel: channel}} 
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                      insertFilter={addFilter}
                      closeFilter={closeFilter} 
                      operatorProps={{
                        "categorical": [
                          '=',
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
    userProperties: state.coreQuery.userProperties,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    userProperties: state.coreQuery.userProperties,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events
  };
};
export default connect(mapStateToProps, {fetchEventNames, fetchGoalInsights, 
  fetchFactorsModels, saveGoalInsightRules, saveGoalInsightModel, getUserProperties, fetchUserPropertyValues})(EventFilterBy);
