import React, { useState, useEffect, useCallback } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  Row,
  Col,
  Menu,
  Dropdown,
  Button,
  Table,
  notification,
  Tabs,
  Badge,
  Switch,
  Modal,
  Space,
  Tag
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import ModalFlow from 'Components/ModalFlow';
import {
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN
} from 'Reducers/types';
import {
  fetchSavedWorkflows,
  fetchWorkflowTemplates,
  removeSavedWorkflow
} from 'Reducers/workflows';
import MomentTz from 'Components/MomentTz';
import TableSearchAndRefresh from 'Components/TableSearchAndRefresh';
import { Stub, StubOld } from './Stub';
import { getAlertTemplatesTransformation } from './utils';
import WorkflowBuilder from './WorkflowBuilder';
import WorkflowEmptyImg from '../../../../assets/images/workflow-empty-screen.png';
import WorkflowHubspotThumbnail from '../../../../assets/images/workflow-hubspot-thumbnail.png';
import logger from 'Utils/logger';

const Workflows = ({
  fetchSavedWorkflows,
  fetchWorkflowTemplates,
  activeProject,
  removeSavedWorkflow,
  savedWorkflows
}) => {
  const dispatch = useDispatch();
  const { confirm } = Modal;
  const [alertTemplates, setAlertTemplates] = useState(false);
  const [builderMode, setBuilderMode] = useState(false);
  const [alertId, setAlertId] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [selectedTemp, setSelectedTemp] = useState(false);

  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);

  const [showSearch, setShowSearch] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [searchTableData, setSearchTableData] = useState([]);

  const dashboard_templates_modal_state = useSelector(
    (state) => state.dashboardTemplatesController
  );
  useEffect(() => {
    setTableLoading(true);
    fetchSavedWorkflows(activeProject?.id)
      .then((res) => {
        setTableLoading(false);
      })
      .catch((err) => {
        logger.log('saved workflows fetch error=>', err);
        setTableLoading(false);
      });

    fetchWorkflowTemplates(activeProject?.id)
      .then((res) => {
        setAlertTemplates(res.data);
      })
      .catch((err) => logger.log('fetch templates error=>', err));

    // setAlertTemplates(StubOld);
  }, [activeProject]);

  const confirmDeleteWorkflow = (item) => {
    confirm({
      title: 'Do you really want to remove this workflow?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      onOk() {
        return removeSavedWorkflow(activeProject?.id, item?.id).then(
          (res) => {
            fetchSavedWorkflows(activeProject?.id);
            notification.success({
              message: 'Success',
              description: 'Deleted Workflow successfully',
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
      }
    });
  };

  const menu = (item) => (
    <Menu>
      <Menu.Item
        key='0'
        onClick={() => {
          setSelectedTemp(item?.alert_body);
          setAlertId(item?.id);
          setEditMode(true);
          setBuilderMode(true);
        }}
      >
        <a>Edit workflow</a>
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item
        key='1'
        onClick={() => {
          confirmDeleteWorkflow(item);
        }}
      >
        <a>
          <span style={{ color: 'red' }}>Remove workflow</span>
        </a>
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: '400px',
      render: (item) => (
        <Text
          type='title'
          level={7}
          truncate
          charLimit={50}
          extraClass='cursor-pointer m-0'
          onClick={() => {
            setSelectedTemp(item?.alert_body);
            setAlertId(item?.id);
            setEditMode(true);
            setBuilderMode(true);
          }}
        >
          {item?.title}
        </Text>
      )
    },
    {
      title: 'Last edited',
      dataIndex: 'last_edit',
      key: 'last_edit',
      render: (item) => (
        <Text type='title' level={7} truncate charLimit={50} extraClass='m-0'>
          {item}
        </Text>
      )
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (item) => (
        <div className='flex items-center'>
          {item?.status === 'paused' || item?.status === 'disabled' ? (
            <Badge
              className='fa-custom-badge fa-custom-badge--orange'
              status='processing'
              text='Paused'
            />
          ) : (
            <Badge
              className='fa-custom-badge fa-custom-badge--green'
              status='success'
              text='Published'
            />
          )}
          {item?.error && (
            <SVG name='InfoCircle' extraClass='ml-2' size={18} color='red' />
          )}
        </div>
      )
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      width: 75,
      render: (obj) => (
        <Dropdown
          trigger={['click']}
          overlay={menu(obj)}
          placement='bottomRight'
        >
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

  const newStep2Comp = (props) => (
    <div>
      <div className='p-6'>
        <div className='flex items-center p-4'>
          <Button
            type='text'
            icon={<SVG name='ArrowLeft' color='grey' size={4} />}
            onClick={() => {
              props.handleBack();
            }}
          >
            Back
          </Button>
        </div>

        <div className='flex items-center p-4'>
          <div className='p-2'>
            <img
              src={WorkflowHubspotThumbnail}
              style={{ 'max-height': '300px' }}
            />
          </div>

          <div className='pl-6'>
            <Text
              type='title'
              level={7}
              color='black'
              weight='bold'
              extraClass='m-0'
            >
              {' '}
              {props?.template?.title}
            </Text>
            <Text type='title' level={7} extraClass='mt-2'>
              {' '}
              {props?.template?.description}
            </Text>
            {props?.template?.categories.map((item) => (
              <Tag>{item}</Tag>
            ))}
          </div>
        </div>

        <div className='flex items-center p-4'>
          <Text type='title' level={7} color='grey' extraClass='mt-2'>
            {' '}
            {props?.template?.alert?.long_description}
          </Text>
        </div>
      </div>

      <div className='flex items-center justify-end p-4 border-top--thin-2'>
        <Button
          type='default'
          onClick={() => {
            props.onCancel();
          }}
        >
          Cancel
        </Button>

        <Button
          type='primary'
          className='ml-2'
          onClick={() => {
            setSelectedTemp(props?.template);
            props.onCancel();
            setBuilderMode(true);
          }}
        >
          Use this template
        </Button>
      </div>
    </div>
  );

  const onSearch = (e) => {
    const term = e.target.value;
    setSearchTerm(term);
    const searchResults = tableData?.filter((item) =>
      item?.name?.title?.toLowerCase().includes(term.toLowerCase())
    );
    setSearchTableData(searchResults);
  };

  const onRefresh = () => {
    setTableLoading(true);
    fetchSavedWorkflows(activeProject?.id).then(() => {
      setTableLoading(false);
    });
  };

  useEffect(() => {
    if (savedWorkflows) {
      const savedArr = [];
      savedWorkflows?.forEach((item, index) => {
        savedArr.push({
          key: index,
          name: item,
          last_edit: MomentTz(item.created_at).fromNow(),
          status: { status: item?.status, error: item?.last_fail_details },
          actions: item
        });
      });
      setTableData(savedArr);
    }
  }, [savedWorkflows]);

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          {!builderMode ? (
            <>
              <Row>
                <Col span={12}>
                  <Text
                    type='title'
                    level={3}
                    weight='bold'
                    extraClass='m-0'
                    id='fa-at-text--page-title'
                  >
                    Workflows
                  </Text>
                </Col>
                <Col span={12}>
                  <div className='flex justify-end'>
                    <Button
                      type='primary'
                      disabled={!alertTemplates}
                      onClick={() => {
                        // setBuilderMode(true)
                        dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
                      }}
                    >
                      <Space>
                        <SVG name='plus' size={16} color='white' />
                        New Workflow
                      </Space>
                    </Button>
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={24}>
                  <Text type='title' level={7} color='grey-2' extraClass='m-0'>
                    Set up automatic workflows for your platforms such as your
                    CRM.
                    <a
                      href='https://help.factors.ai/en/articles/7284705-alerts'
                      target='_blank'
                      rel='noreferrer'
                    >
                      {' '}
                      Learn more
                    </a>
                  </Text>

                  {tableData ? (
                    <div className='mt-8'>
                      <TableSearchAndRefresh
                        showSearch={showSearch}
                        setShowSearch={setShowSearch}
                        searchTerm={searchTerm}
                        setSearchTerm={setSearchTerm}
                        onSearch={onSearch}
                        onRefresh={onRefresh}
                        tableLoading={tableLoading}
                      />
                      <Table
                        className='fa-table--basic mt-2'
                        loading={tableLoading}
                        columns={columns}
                        dataSource={searchTerm ? searchTableData : tableData}
                        pagination
                      />
                    </div>
                  ) : (
                    <div className='flex flex-col items-center mt-10'>
                      <img src={WorkflowEmptyImg} width='350' />
                      <Text
                        type='title'
                        level={7}
                        color='grey'
                        extraClass='mt-2'
                      >
                        Lorem ipsum dolor sit amet consectetur. Morbi enim eget
                        egestas nulla aliquet sodales quisque
                      </Text>{' '}
                    </div>
                  )}

                  {alertTemplates && (
                    <ModalFlow
                      data={getAlertTemplatesTransformation(alertTemplates)}
                      visible={
                        dashboard_templates_modal_state.isNewDashboardTemplateModal
                      }
                      onCancel={() => {
                        dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
                      }}
                      Step2Screen={newStep2Comp}
                    />
                  )}
                </Col>
              </Row>{' '}
            </>
          ) : (
            <WorkflowBuilder
              setBuilderMode={setBuilderMode}
              selectedTemp={selectedTemp}
              alertId={alertId}
              editMode={editMode}
              setEditMode={setEditMode}
            />
          )}
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  savedWorkflows: state.workflows.savedWorkflows
});

export default connect(mapStateToProps, {
  fetchSavedWorkflows,
  fetchWorkflowTemplates,
  removeSavedWorkflow
})(Workflows);
