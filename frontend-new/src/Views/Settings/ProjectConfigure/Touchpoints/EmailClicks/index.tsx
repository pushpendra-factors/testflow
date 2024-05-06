import React, { useEffect, useState } from 'react';
import {
  Row,
  Col,
  Modal,
  Input,
  Button,
  Form,
  Table,
  Tag,
  message,
  Divider,
  Select
} from 'antd';
import { connect } from 'react-redux';
import { Text, SVG } from 'factorsComponents';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { udpateProjectDetails } from 'Reducers/global';
import logger from 'Utils/logger.js';
import PlatformCard from './PlatformCard.tsx';

const { confirm } = Modal;

const EmailClicks = ({ activeProject, udpateProjectDetails }) => {
  const [newKey, setNewKey] = useState<null>(null);
  const [dataSource, setDataSource] = useState<null>(null);
  const [UTM, setUTM] = useState<object>({});
  const [tags, setTags] = useState<string[]>([]);
  const [selectedPlatform, setSelectedPlatform] = useState<object>({});

  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState<null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [visible, setVisible] = useState<boolean>(false);

  useEffect(() => {
    const DS: any = Object.keys(UTM)?.map((key, index) => ({
      key: index,
      parameter: 'User Email ID',
      actions: {
        key,
        tags: UTM[key]
      }
    }));
    setDataSource(DS?.filter((ds: any) => ds?.actions?.key === '$ep_email'));

    const data = UTM?.$ep_email?.map((item: string) =>
      item.startsWith('$qp_') ? item.slice(4) : item
    );
    setTags(data);
  }, [UTM]);

  useEffect(() => {
    if (activeProject) {
      if (activeProject?.interaction_settings) {
        setUTM(activeProject?.interaction_settings?.utm_mapping);
      }
    }
  }, [activeProject]);

  const onChange = () => {
    seterrorInfo(null);
  };
  const onReset = () => {
    seterrorInfo(null);
    setVisible(false);
    form.resetFields();
  };

  const onFinish = (values: any) => {
    let newArr: string[] = [];
    const UTM_Map = UTM;
    const stripLogic = (item: string) =>
      item.startsWith('$qp_') ? item : `$qp_${item}`;
    Object.keys(UTM_Map).map((item) => {
      if (newKey === item) {
        newArr = [...UTM_Map[newKey], stripLogic(values.utm_tag)];
      }
    });
    const updatedKey = { ...UTM, [newKey]: newArr };
    setLoading(true);
    udpateProjectDetails(activeProject.id, {
      interaction_settings: {
        utm_mapping: updatedKey
      }
    })
      .then(() => {
        message.success('UTM Tag added!');
        setVisible(false);
        onReset();
        setLoading(false);
      })
      .catch((err: any) => {
        logger.log(err);
        seterrorInfo(err.data.error);
        setLoading(false);
      });
  };

  const addUTM = (key: any) => {
    setNewKey(key);
    setVisible(true);
  };

  const removeUTM = (e, key: any, item: string) => {
    e.preventDefault();
    confirm({
      icon: <ExclamationCircleOutlined />,
      content: `Are you sure you want to delete?`,
      onOk() {
        const UTM_Map = UTM;
        let updatedKey = null;
        setLoading(true);
        Object.keys(UTM_Map).map((tag) => {
          if (key === tag) {
            const filteredItems = UTM_Map[key].filter(
              (arryItem: string) => arryItem !== item
            );
            updatedKey = { ...UTM, [key]: filteredItems };
          }
        });
        udpateProjectDetails(activeProject.id, {
          interaction_settings: {
            utm_mapping: updatedKey
          }
        })
          .then(() => {
            message.success('UTM Tag removed!');
            setLoading(false);
          })
          .catch((err: any) => {
            logger.log(err);
            message.error(err.data.error);
            setLoading(false);
          });
      }
    });
  };

  const columns = [
    {
      title: 'To identify',
      dataIndex: 'parameter',
      key: 'parameter'
    },
    {
      title: 'Check value of one of these UTM tag values',
      dataIndex: 'actions',
      key: 'actions',
      render: (text: any) => {
        const key = text?.key;

        if (text?.tags?.length === 0) {
          return (
            <Button
              onClick={() => {
                addUTM(key);
              }}
              style={{ transform: 'scale(0.8)' }}
              size='small'
              icon={<SVG name='plus' size={12} />}
            >
              Add
            </Button>
          );
        }

        const stripLogic = (item: string) =>
          item.startsWith('$qp_') ? item.slice(4) : item;
        return text?.tags.map((item: string, index: number) => {
          if (text?.tags?.length === index + 1) {
            return (
              <>
                <Tag
                  className='fa-tag--green'
                  color='green'
                  closable
                  onClose={(e) => removeUTM(e, key, item)}
                >
                  {stripLogic(item)}{' '}
                </Tag>
                <Button
                  onClick={() => {
                    addUTM(key);
                  }}
                  style={{ transform: 'scale(0.8)' }}
                  size='small'
                  icon={<SVG name='plus' size={12} />}
                >
                  Add
                </Button>
              </>
            );
          }

          return (
            <Tag
              className='fa-tag--green'
              color='green'
              closable
              onClose={(e) => removeUTM(e, key, item)}
            >
              {stripLogic(item)}
            </Tag>
          );
        });
      }
    }
  ];

  const selectOptions = [
    {
      value: 'hubspot',
      label: 'Hubspot'
    },
    {
      value: 'salesforceOutreach',
      label: 'Salesforce'
    },
    {
      value: 'salesforceEmailStudio',
      label: 'Salesforce email studio'
    },
    {
      value: 'apollo',
      label: 'Apollo'
    },
    {
      value: 'outreach',
      label: 'Outreach'
    }
  ];

  const handlePlatformChange = (value: string) => {
    setSelectedPlatform(value);
  };

  return (
    <>
      <div className='mb-10'>
        <Row>
          <Col>
            <Text type='title' level={7} color='grey' extraClass='m-0'>
              Identify the email address of people clicking on links inside your
              emails. You can attach specific UTM tags inside links to your
              website. Once a user visits your website using a link with such a
              tag added, their email will get identified.
            </Text>
          </Col>
        </Row>
        <Row className='mt-6'>
          <Col span={24}>
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              Map your UTM tags
            </Text>
            <Text type='title' level={7} color='grey' extraClass='m-0'>
              These are the tags that will be used to get the email of the
              person clicking the link in your emails.
            </Text>
          </Col>
        </Row>

        <Row className='m-0 mt-2'>
          <Col span={24}>
            <div className='mt-2'>
              <Table
                className='fa-table--basic mt-2'
                columns={columns}
                dataSource={dataSource}
                pagination={false}
              />
            </div>
          </Col>
        </Row>

        <Row className='mt-6'>
          <Col span={24}>
            <Text type='title' level={6} weight='bold' extraClass='m-0'>
              How to get contact’s email from links inside your emails
            </Text>
            <Text type='title' level={7} color='grey' extraClass='m-0'>
              Please select the platform you use to send out emails
            </Text>
          </Col>
        </Row>
        <Row className='mt-3'>
          <Col>
            <Select
              showSearch
              className='fa-select'
              style={{ minWidth: '200px' }}
              placeholder='Select Platform'
              optionFilterProp='children'
              onChange={(value) => handlePlatformChange(value)}
              labelInValue
              filterOption={(input, option) =>
                (option?.label ?? '')
                  .toLowerCase()
                  .includes(input.toLowerCase())
              }
              options={selectOptions}
            />
          </Col>
        </Row>
        <Row className='mt-4'>
          <Col span={24}>
            <PlatformCard tags={tags} selectedPlatform={selectedPlatform} />
          </Col>
        </Row>
      </div>

      <Modal
        visible={visible}
        zIndex={1020}
        onCancel={onReset}
        className='fa-modal--regular fa-modal--slideInDown'
        okText='Add'
        centered
        footer={null}
        transitionName=''
        maskTransitionName=''
      >
        <div className='p-4'>
          <Form
            form={form}
            onFinish={onFinish}
            className='w-full'
            onChange={onChange}
          >
            <Row>
              <Col span={24}>
                <Text type='title' level={5} weight='bold' extraClass='m-0'>
                  Add UTM Tag
                </Text>
              </Col>
            </Row>
            <Row>
              <Col span={24}>
                <Text type='title' level={7} color='grey' extraClass='m-0'>
                  Add a UTM parameter for identifying the email address of the
                  person clicking on links inside your emails.
                </Text>
              </Col>
            </Row>
            <Row className='mt-4'>
              <Col span={24}>
                <Form.Item
                  name='utm_tag'
                  rules={[
                    {
                      required: true,
                      message: 'Please enter UTM Tag'
                    }
                  ]}
                >
                  <Input
                    disabled={loading}
                    size='large'
                    className='fa-input w-full'
                    placeholder='Enter UTM Tag'
                  />
                </Form.Item>
              </Col>
            </Row>

            {errorInfo && (
              <Col span={24}>
                <div className='flex flex-col justify-center items-center mt-1'>
                  <Text type='title' color='red' size='7' className='m-0'>
                    {errorInfo}
                  </Text>
                </div>
              </Col>
            )}
            <Row className='mt-6'>
              <Col span={24}>
                <div className='flex justify-end'>
                  <Button size='large' onClick={onReset} className='mr-2'>
                    {' '}
                    Cancel{' '}
                  </Button>
                  <Button
                    loading={loading}
                    type='primary'
                    size='large'
                    htmlType='submit'
                  >
                    {' '}
                    Add{' '}
                  </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </div>
      </Modal>
    </>
  );
};

const mapStateToProps = (state: any) => ({
  activeProject: state.global.active_project,
  agents: state.agent.agents,
  projects: state.global.projects,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, { udpateProjectDetails })(EmailClicks);
