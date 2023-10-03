import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Switch,
  Menu,
  Dropdown,
  Button,
  Tabs,
  Table,
  Tag,
  Space,
  message,
  notification
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { MoreOutlined } from '@ant-design/icons';
import ContentGroupForm from './ContentGroupForm';
import { fetchContentGroup, deleteContentGroup } from 'Reducers/global';
import ConfirmationModal from '../../../../components/ConfirmationModal';

function ContentGroups({
  fetchContentGroup,
  deleteContentGroup,
  activeProject,
  contentGroup
}) {
  const [showSmartForm, setShowSmartForm] = useState(false);
  const [tableLoading, setTableLoading] = useState(false);
  const [tableData, setTableData] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);

  useEffect(() => {
    if (activeProject?.id) {
      setTableLoading(true);
      fetchContentGroup(activeProject.id).then(() => {
        setTableLoading(false);
      });
    }
  }, [activeProject]);

  useEffect(() => {
    const dataColumn = [];
    contentGroup.forEach((prop) => {
      //harcoded type
      dataColumn.push({
        content_group_name: prop.content_group_name,
        content_group_description: prop.content_group_description,
        rule: prop.rule.length,
        actions: prop
      });
    });
    setTableData(dataColumn);
  }, [contentGroup]);

  const menu = (obj) => {
    return (
      <Menu>
        <Menu.Item key='0' onClick={() => showDeleteWidgetModal(obj.id)}>
          <a>Remove</a>
        </Menu.Item>
        <Menu.Item key='1' onClick={() => editProp(obj)}>
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
      render: (text) => <span className={'font-medium'}>{text}</span>
    },
    {
      title: 'Description',
      dataIndex: 'content_group_description',
      key: 'content_group_description',
      render: (text) => <span className={'text-gray-700'}>{text}</span>
    },
    {
      title: 'Values',
      dataIndex: 'rule',
      key: 'rule',
      render: (text) => <span className={'ml-3 text-gray-700'}>{text}</span>
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      render: (obj) => (
        <Dropdown overlay={() => menu(obj)} trigger={['click']}>
          <Button
            type='text'
            icon={
              <MoreOutlined
                rotate={90}
                style={{ color: 'gray', fontSize: '18px' }}
              />
            }
          />
        </Dropdown>
      )
    }
  ];

  const editProp = (obj) => {
    setSelectedGroup(obj);
    setShowSmartForm(true);
  };

  const confirmRemove = (id) => {
    return deleteContentGroup(activeProject.id, id).then(
      (res) => {
        fetchContentGroup(activeProject.id);
        notification.success({
          message: 'Success',
          description: 'Deleted content group successfully ',
          duration: 5
        });
      },
      (err) => {
        notification.error({
          message: 'Error',
          description: err.data,
          duration: 5
        });
      }
    );
  };

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [deleteWidgetModal]);

  return (
    <div className={'fa-container'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <div className={'mb-10 pl-4'}>
            {!showSmartForm && (
              <>
                <Row>
                  <Col span={12}>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                      id={'fa-at-text--page-title'}
                    >
                      Content Groups
                    </Text>
                  </Col>
                  <Col span={12}>
                    <div className={'flex justify-end'}>
                      <Button
                        size={'large'}
                        onClick={() => {
                          setShowSmartForm(true);
                        }}
                      >
                        <SVG name={'plus'} extraClass={'mr-2'} size={16} />
                        Add New
                      </Button>
                    </div>
                  </Col>
                </Row>
                <Row className={'mt-4'}>
                  <Col span={24}>
                    <div className={'mt-6'}>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0'}
                      >
                        Create logical collections of related URLs, such as blog
                        articles or product pages, to analyze their impact on
                        leads, revenue, and pipeline stages.
                      </Text>
                      <Text
                        type={'title'}
                        level={7}
                        color={'grey-2'}
                        extraClass={'m-0 mt-4'}
                      >
                        Compare the performance of different content groups,
                        identify optimization opportunities, and enhance your
                        content marketing efforts to drive better results.{' '}
                        <a href='https://help.factors.ai/en/articles/7284125-content-groups' target='_blank'>
                          Learn more
                        </a>
                      </Text>

                      <Table
                        className='fa-table--basic mt-8'
                        columns={columns}
                        dataSource={tableData}
                        pagination={false}
                        loading={tableLoading}
                      />
                    </div>
                  </Col>
                </Row>
              </>
            )}
            {showSmartForm && (
              <>
                <ContentGroupForm
                  selectedGroup={selectedGroup}
                  setShowSmartProperty={(showVal) => {
                    setShowSmartForm(showVal);
                    setSelectedGroup(null);
                    fetchContentGroup(activeProject.id);
                  }}
                />
              </>
            )}
            <ConfirmationModal
              visible={deleteWidgetModal ? true : false}
              confirmationText='Do you really want to remove this content group?'
              onOk={confirmDelete}
              onCancel={showDeleteWidgetModal.bind(this, false)}
              title='Remove Content Group'
              okText='Confirm'
              cancelText='Cancel'
              confirmLoading={deleteApiCalled}
            />
          </div>
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  contentGroup: state.global.contentGroup,
  agents: state.agent.agents,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchContentGroup,
  deleteContentGroup
})(ContentGroups);
