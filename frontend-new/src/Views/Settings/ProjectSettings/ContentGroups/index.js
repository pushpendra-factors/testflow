import React, { useState, useEffect } from 'react';
import {
  Row, Col, Switch, Menu, Dropdown, Button, Tabs, Table, Tag, Space, message
} from 'antd';
import { Text, SVG } from 'factorsComponents'; 
import { connect } from 'react-redux';
import { MoreOutlined } from '@ant-design/icons';
import ContentGroupForm from './ContentGroupForm';
import { fetchContentGroup } from 'Reducers/global';


function ContentGroups({fetchContentGroup, activeProject, contentGroup, agents, currentAgent}) { 

    const [showSmartForm, setShowSmartForm] = useState(false);
    const [tableLoading, setTableLoading] = useState(false);
    const [tableData, setTableData] = useState([]); 
    const [selectedGroup, setSelectedGroup] = useState(null);


    useEffect(() => {
      if (activeProject?.id) {
          setTableLoading(true);
          fetchContentGroup(activeProject.id).then(() => {
              setTableLoading(false);
          })
      }

  }, [activeProject]);

  useEffect(() => {
    const dataColumn = [];
    contentGroup.forEach((prop) => {
        //harcoded type
        dataColumn.push({ content_group_name: prop.content_group_name, content_group_description: prop.content_group_description, rule: prop.rule.length, actions: prop })
    })
    setTableData(dataColumn);
}, [contentGroup])



    const menu = (obj) => {
      return (
      <Menu> 
        <Menu.Item key="0" onClick={() => editProp(obj)}>
            <a>Edit</a>
          </Menu.Item>
      </Menu>
      );
    };

const columns = [

    {
      title: 'Title',
      dataIndex: 'content_group_name',
      key: 'content_group_name', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Description',
      dataIndex: 'content_group_description',
      key: 'content_group_description', 
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
        title: 'Values',
        dataIndex: 'rule',
        key: 'rule', 
        render: (text) => <span className={'capitalize'}>{text}</span>
      },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      render: (obj) => (
        <Dropdown overlay={() => menu(obj)} trigger={['hover']}>
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  const editProp = (obj) => {
    setSelectedGroup(obj);
    setShowSmartForm(true);
}

 

  return (
    <>
        <div className={'mb-10 pl-4'}>
        {!showSmartForm && <> 
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Content Groups</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} onClick={() =>   {setShowSmartForm(true)}}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>
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
                dataSource={tableData} 
                pagination={false}
                />
            </div>  
        </Col> 
        </Row> 
        </>
        }
        {showSmartForm && <>  
                <ContentGroupForm selectedGroup={selectedGroup} setShowSmartForm={setShowSmartForm} /> 
        </>
        }
      </div>
    </>

  );
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    contentGroup: state.global.contentGroup,
    agents: state.agent.agents, 
    currentAgent: state.agent.agent_details
  });

  export default connect(mapStateToProps, {fetchContentGroup})(ContentGroups); 