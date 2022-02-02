import React, { useEffect, useState } from 'react';
import {  Button , Badge, message, Spin} from 'antd';
import { SVG, Text} from 'factorsComponents';
import { Link } from 'react-router-dom'; 
import {connect} from 'react-redux'; 
import _, { isEmpty } from 'lodash';  
import GroupSelect2 from '../../../components/QueryComposer/GroupSelect2';  
import { fetchGoalInsights} from 'Reducers/factors'; 
import moment from 'moment-timezone';

function Header({factors_insight_rules, factors_models, fetchGoalInsights, activeProject, goal_insights, factors_insight_model, savedName}) {

  const [showDateTime, setShowDateTime] = useState(false);
  const [dateTime, setDateTime] = useState(null);
  const [fetchingIngishts, SetfetchingIngishts] = useState(false);

  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';  
moment.tz.setDefault(timeZone);


  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  } 
  const factorsModels = !_.isEmpty(factors_models) && _.isArray(factors_models) ? factors_models.map((item)=>{return [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]}) : [];

  const onChangeDateTime = (grp, value) => { 
    setDateTime(value); 
    setShowDateTime(false);  
    SetfetchingIngishts(true);
    const calcModelId = factors_models.filter((item)=>{   
      const generateStringArray = [`[${item.mt}] ${readableTimstamp(item.st)} - ${readableTimstamp(item.et)}`]; 
      if (_.isEqual(value,generateStringArray)){  
        return item
      } 
    }); 

    fetchGoalInsights(activeProject.id, goal_insights.type, factors_insight_rules, calcModelId[0].mid).then((data)=>{
      SetfetchingIngishts(false);
      }).catch((err)=>{
        SetfetchingIngishts(false);
        console.log("fetchGoalInsights catch",err);
        const ErrMsg = err?.data?.error ? err.data.error : `Oops! Something went wrong!`;
        message.error(ErrMsg); 
    }); 

  }

  if(factors_insight_rules){
      return ( 
        <div className={'fa-container pb-5'}>
             <div className="flex flex-col justify-between border-bottom--thin-2 pb-4" style={{borderBottomWidth:'3px'}}> 
                    <Text type={'title'} level={2} color={'grey'} weight={'bold'} color={'grey-3'} extraClass={'m-0 '}>{_.isEmpty(factors_insight_rules.name) ? (savedName ? savedName : 'Untitled Name') : factors_insight_rules.name }</Text>
                    <div className="flex items-center">
                        <Badge count={'Goal'} className={'fa-custom-badge'} />
                        {factors_insight_rules.rule?.st_en?.na ? <>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>{factors_insight_rules?.rule?.st_en?.na}</Text> 
                            <Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'}>and</Text>
                        </> : null}
                        <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>{factors_insight_rules?.rule?.en_en?.na}</Text>
                        {!_.isEmpty(factors_insight_rules?.rule?.gpr) ? <>
                            <Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'}>where</Text>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>Untitled</Text>
                        </> : null
                        }
                    </div>
            </div>
            <div className="flex flex-col py-4" >
              <div className={'absolute'} style={{bottom:0,left:0, paddingLeft:'50px'}}>
                <div className={'relative'}>
                {!showDateTime && <Button  loading={fetchingIngishts} onClick={()=>setShowDateTime(true)}><SVG name={'calendar'} extraClass={'mr-1'} />{dateTime ? dateTime : factors_insight_model} </Button>}
                {showDateTime && 
                <GroupSelect2 
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
              </div>
            </div>
        </div>
      );
  }
  else return null
}

const mapStateToProps = (state) => {
  return { 
    factors_insight_rules: state.factors.factors_insight_rules,
    factors_insight_model: state.factors.factors_insight_model,
    factors_models: state.factors.factors_models,
    activeProject: state.global.active_project,
    goal_insights: state.factors.goal_insights,
  };
};
export default connect(mapStateToProps, {fetchGoalInsights})(Header);
