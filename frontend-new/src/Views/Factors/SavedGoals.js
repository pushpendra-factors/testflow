import React, { useEffect, useState } from 'react';
import { Text, SVG } from 'factorsComponents';
import { Button, Table, Avatar, Menu, Dropdown, Modal, message } from 'antd';
import { useHistory } from 'react-router-dom';
import { fetchGoalInsights, saveGoalInsightRules, removeSavedGoal, fetchFactorsGoals } from 'Reducers/factors';
import { connect } from 'react-redux';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import moment from 'moment';


const { confirm } = Modal;




const SavedGoals = ({ goals, fetchGoalInsights, factors_models, agents, saveGoalInsightRules, SetfetchingIngishts, removeSavedGoal, fetchFactorsGoals }) => {

  const [loadingTable, SetLoadingTable] = useState(true);
  const [dataSource, setdataSource] = useState(null);
  const history = useHistory();

  const menu = (values) => {
    return (
      <Menu>
        <Menu.Item key="0" onClick={() => confirmRemove(values)}>
          <a>Delete Goal</a>
        </Menu.Item>
      </Menu>
    );
  };

  const columns = [
    {
      title: 'Saved Goals',
      dataIndex: 'actions',
      key: 'actions',
      render: (goal) => <Text type={'title'} level={7} weight={'bold'} extraClass={'cursor-pointer m-0'} onClick={() => getInsights(goal.project_id, goal.rule, goal.name)} >{goal.name}</Text>
    },
    {
      title: 'Created By',
      dataIndex: 'author',
      key: 'author',
      render: (text) => <div className="flex items-center">
        {text === "System Generated" ? <Text type={'title'} level={7} color={'grey'} extraClass={'cursor-pointer m-0'} >{`System Generated`}</Text> : <>
          <Avatar src="assets/avatar/avatar.png" className={'mr-2'} size={24} /><Text type={'title'} level={6} extraClass={'cursor-pointer m-0 ml-2'} >{text}</Text>
        </>
        }
      </div>
    },
    {
      title: 'Date',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => <Text type={'title'} level={7}  extraClass={'m-0'}>{moment(date).format('MMM DD, YYYY')}</Text>
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


  const confirmRemove = (goalValues) => {
    const goalId = {
      id: goalValues.id
    }
    confirm({
      title: 'Are you sure you want to remove this Goal?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        removeSavedGoal(goalValues.project_id, goalId).then(() => {
          message.success('Goal Removed!');
          fetchFactorsGoals(goalValues.project_id);
        }).catch((err) => {
          message.error(err);
        });
      }
    });

  };



  useEffect(() => {
    SetLoadingTable(true);
    setdataSource(null);
    if (goals && agents) {
      const formattedArray = [];
      goals.map((goal, index) => {
        let createdUser = '';
        if (goal.is_active) {
          agents.map((agent) => {
            if (agent.uuid === goal.created_by) {
              createdUser = `${agent.first_name} ${agent.last_name}`;
            }
          });
          formattedArray.push({
            key: index,
            author: createdUser ? createdUser : 'System Generated',
            rule: goal.rule,
            project_id: goal.project_id,
            created_at: goal?.created_at,
            actions: goal
          });
        }
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
    getData().then(() => {
      saveGoalInsightRules(ruleData);
      history.push('/explain/insights');
      SetfetchingIngishts(false)
    });
  };


  return (
    <Table loading={loadingTable} className="fa-table--basic mt-8" columns={columns} dataSource={dataSource} pagination={false}

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


export default connect(mapStateToProps, { fetchGoalInsights, saveGoalInsightRules, removeSavedGoal, fetchFactorsGoals })(SavedGoals);
