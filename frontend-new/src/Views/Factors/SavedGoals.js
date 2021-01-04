import React, {useEffect, useState} from 'react';
import { Text, SVG } from 'factorsComponents'; 
import { Table, Avatar } from 'antd';
import { useHistory } from 'react-router-dom'; 
import {  fetchGoalInsights, saveGoalInsightRules } from 'Reducers/factors';
import { connect } from 'react-redux';

const columns = [
    {
      title: 'Saved Goals',
      dataIndex: 'title',
      key: 'title',
      render: (text) => <Text type={'title'} level={6} extraClass={'cursor-pointer m-0'} >{text}</Text>
    },
    {
      title: 'Created By',
      dataIndex: 'author',
      key: 'author',
      render: (text) => <div className="flex items-center">
        {text === "System Generated" ? <Text type={'title'} level={7} color={'grey'} extraClass={'cursor-pointer m-0'} >{`System Generated`}</Text> : <>
          <Avatar src="assets/avatar/avatar.png" className={'mr-2'} /><Text type={'title'} level={6} extraClass={'cursor-pointer m-0 ml-2'} >{text}</Text>
          </>
        }
          </div>
    }
  ];

const SavedGoals = ({goals, fetchGoalInsights,  factors_models, agents, saveGoalInsightRules, SetfetchingIngishts}) => {

    const [loadingTable, SetLoadingTable] = useState(true);
    const [dataSource, setdataSource] = useState(null);
    const history = useHistory();

    useEffect(() => {
        SetLoadingTable(true);
        setdataSource(null);
        if (goals && agents) {
          const formattedArray = [];
          goals.map((goal, index) => {
            let createdUser = '';
            agents.map((agent) => {
              if (agent.uuid === goal.created_by) {
                createdUser = `${agent.first_name} ${agent.last_name}`;
              }
            });
            formattedArray.push({
              key: index,
              title: goal.name,
              author: createdUser ? createdUser : 'System Generated',
              rule: goal.rule,
              project_id: goal.project_id
            });
            setdataSource(formattedArray);
          });
          SetLoadingTable(false);
        }
      }, [goals]);

      const getInsights = (project_id, rule, name) => {

        SetfetchingIngishts(true);
        const isJourney = !_.isEmpty(rule?.rule?.st_en); 
        const ruleData = {
          name: name,
          rule: rule
        }
        const getData = async () => {
          await fetchGoalInsights(project_id, isJourney, ruleData, factors_models[0].mid); 
        };
        getData().then(()=>{  
          saveGoalInsightRules(ruleData);  
          history.push('/explain/insights');   
          SetfetchingIngishts(false)
        });
      };


  return (
            <Table loading={loadingTable} className="ant-table--custom mt-8" columns={columns} dataSource={dataSource} pagination={false} 
            onRow={(record, rowIndex) => {
            return {
                onClick: event => {
                getInsights(record.project_id,record.rule, record.title) 
                }, // click row 
            };
        }}
        />
  );
};


const mapStateToProps = (state) => {
    return { 
      goals: state.factors.goals,
      agents: state.agent.agents,
      factors_models: state.factors.factors_models,
    };
  };


export default connect(mapStateToProps, {fetchGoalInsights,  saveGoalInsightRules})(SavedGoals);
