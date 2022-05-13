import React, { useState, useEffect, useCallback } from 'react';
import { connect, useSelector, useDispatch } from "react-redux";
import {
    Row, Col, Select, Menu, Dropdown, Button, Form, Table, Input, message, Collapse, notification, Tooltip, Space, Checkbox, Modal
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, PlusOutlined } from '@ant-design/icons';
import _ from 'lodash';
import GroupSelect2 from '../../../../components/KPIComposer/GroupSelect2';
import FaSelect from 'Components/FaSelect';
import { createAlerts, fetchAlerts, deleteAlerts } from 'Reducers/global';
import ConfirmationModal from '../../../../components/ConfirmationModal';
import QueryBlock from './QueryBlock';
import { deleteGroupByForEvent } from '../../../../reducers/coreQuery/middleware';
import { getEventsWithPropertiesKPI, getStateFromFilters } from './utils';
import { fetchSlackChannels, fetchProjectSettingsV1, enableSlackIntegration } from '../../../../reducers/global';
import SelectChannels from './SelectChannels';

const { Option } = Select;

const Alerts = ({
    activeProject,
    kpi,
    createAlerts,
    fetchAlerts,
    deleteAlerts,
    savedAlerts,
    agent_details,
    slack,
    fetchSlackChannels,
    fetchProjectSettingsV1,
    projectSettings,
    enableSlackIntegration,
}) => {

    const [showForm, setShowForm] = useState(false);
    const [tableData, setTableData] = useState([]);
    const [tableLoading, setTableLoading] = useState(false);
    const [errorInfo, seterrorInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [viewMode, SetViewMode] = useState(false);
    const [viewAlertDetails, setAlertDetails] = useState(false);
    const [operatorState, setOperatorState] = useState(null);
    const [Value, setValue] = useState(null);
    const [emailEnabled, setEmailEnabled] = useState(false);
    const [slackEnabled, setSlackEnabled] = useState(false);
    const [showCompareField, setShowCompareField] = useState(false);
    const [alertType, setAlertType] = useState(1);
    const [viewFilter, setViewFilter] = useState([]);
    const [channelOpts, setChannelOpts] = useState([]);
    const [selectedChannel, setSelectedChannel] = useState([]);
    const [showSelectChannelsModal, setShowSelectChannelsModal] = useState(false);
    const [viewSelectedChannels, setViewSelectedChannels] = useState([]);

    const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
    const [deleteApiCalled, setDeleteApiCalled] = useState(false);
    
    const [form] = Form.useForm();


    // KPI SELECTION
    const [queryType, setQueryType] = useState('kpi');
    const [queries, setQueries] = useState([]);
    const [selectedMainCategory, setSelectedMainCategory] = useState(false);
    const [KPIConfigProps, setKPIConfigProps] = useState([]);
    const [queryOptions, setQueryOptions] = useState({
      group_analysis: 'users',
      groupBy: [
        {
          prop_category: '', // user / event
          property: '', // user/eventproperty
          prop_type: '', // categorical  /numberical
          eventValue: '', // event name (funnel only)
          eventName: '', // eventName $present for global user breakdown
          eventIndex: 0,
        },
      ],
      globalFilters: [],
    });


    const confirmRemove = (id) => {
        return deleteAlerts(activeProject.id, id).then(res => {
            fetchAlerts(activeProject.id);
            notification.success({
                message: "Success",
                description: "Deleted Alert successfully ",
                duration: 5,
            })
        }, err => {
            notification.error({
                message: "Error",
                description: err.data,
                duration: 5,
            })
        });
      }

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
      }, [
          deleteWidgetModal
      ]);

    const menu = (item) => {
        return (
            <Menu>
                <Menu.Item key="0"
                     onClick={() => {
                        showDeleteWidgetModal(item.id);
                     }}
                 >
                    <a>Remove</a>
                </Menu.Item>
                <Menu.Item key="1"
                    onClick={() => {
                        SetViewMode(true)
                        setAlertDetails(item)
                    }}
                >
                    <a>View</a>
                </Menu.Item>
            </Menu>
        );
    };

    // if operatorstate is include 'by more than' in string then show compare field
    useEffect(() => {
        if (operatorState && operatorState.includes('by_more_than')) {
            setShowCompareField(true);
            setAlertType(2);
        } else {
            setShowCompareField(false)
            setAlertType(1);
        }
    }, [operatorState])

    useEffect(() => {
        if (viewAlertDetails?.alert_description?.query?.fil) {
           const filter = getStateFromFilters(viewAlertDetails.alert_description.query.fil);
           setViewFilter(filter);
        }
        if(viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups) {
            let obj = viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups
            for(let key in obj) {
                if(obj[key].length > 0) {
                    setViewSelectedChannels(obj[key]);
                }
            }
        }
    }, [viewAlertDetails])

    const columns = [

        {
            title: 'Alert',
            dataIndex: 'alert',
            key: 'alert',
            render: (text) => <Text type={'title'} level={7} truncate={true} charLimit={25}>{text}</Text>,
            // width: 100,
        },
        {
            title: 'Delivery Options',
            dataIndex: 'dop',
            key: 'dop',
            render: (text) => <Text type={'title'} level={7} truncate={true} charLimit={25}>{text}</Text>,
            // width: 200,
        },
        {
            title: '',
            dataIndex: 'actions',
            key: 'actions',
            align: 'right',
            width: 75,
            render: (obj) => (
                <Dropdown overlay={() => menu(obj)} trigger={['click']}>
                    <Button type="text" icon={<MoreOutlined rotate={90} style={{ color: 'gray', fontSize: '18px' }} />} />
                </Dropdown>
            )
        }
    ];

      const queryChange = (newEvent, index, changeType = 'add', flag = null) => {
        const queryupdated = [...queries];
        if (queryupdated[index]) {
          if (changeType === 'add') {
            if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
              deleteGroupByForEvent(newEvent, index);
            }
            queryupdated[index] = newEvent;
          } else {
            if (changeType === 'filters_updated') {
              // dont remove group by if filter is changed
              queryupdated[index] = newEvent;
            } else {
              deleteGroupByForEvent(newEvent, index);
              queryupdated.splice(index, 1);
            }
          }
        } else {
          if (flag) {
            Object.assign(newEvent, { pageViewVal: flag });
          }
          queryupdated.push(newEvent);
        }
        setQueries(queryupdated);
      };

      useEffect(() => {
        setSelectedMainCategory(queries[0]);
      }, [queries]);

      const queryList = () => {
        const blockList = [];
    
        queries.forEach((event, index) => {
          blockList.push(
            <div key={index} >
              <QueryBlock
                index={index + 1}
                queryType={queryType}
                event={event}
                queries={queries}
                eventChange={queryChange}
                selectedMainCategory={selectedMainCategory}
                setSelectedMainCategory={setSelectedMainCategory}
                KPIConfigProps={KPIConfigProps}
              />
            </div>
          );
        });
    
        if (queries.length < 1) {
          blockList.push(
            <div key={'init'}>
              <QueryBlock
                queryType={queryType}
                index={queries.length + 1}
                queries={queries}
                eventChange={queryChange}
                groupBy={queryOptions.groupBy}
                selectedMainCategory={selectedMainCategory}
                setSelectedMainCategory={setSelectedMainCategory}
                KPIConfigProps={KPIConfigProps}
              />
            </div>
          );
        }
    
        return blockList;
      };
    

    const onFinish = data => {
        setLoading(true);
        // Putting All emails into single array
        let emails = [];
        if(emailEnabled) {
            if (data.emails) {
                emails = data.emails.map((item) => {
                    return item.email;
                })
            }
            if (data.email) {
                emails.push(data.email);
            }
        }

        let slackChannels = {}
        if(slackEnabled) {
            const map = new Map();
            map.set(agent_details.uuid , selectedChannel);
            for (const [key, value] of map) {
                slackChannels = {...slackChannels, [key]: value}
            }
        }

        let payload = {
            "alert_type": alertType,
            "alert_description": {
              "name" : queries[0].metric,
              "query": {
                'ca': queries[0].category,
                'dc': queries[0].group,
                'fil': getEventsWithPropertiesKPI(queries[0].filters, queries[0]?.category),
                'me': [queries[0].metric],
                'pgUrl': queries[0]?.pageViewVal ? queries[0]?.pageViewVal : '',
                "tz": localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
              },
              "query_type": "kpi",
              "operator": operatorState,
              "value": Value,
              "date_range": data.date_range,
              'compared_to': data.compared_to,
            },
            "alert_configuration":{
              "email_enabled": emailEnabled ,
              "slack_enabled": slackEnabled,
              "emails": emails,
              "slack_channels_and_user_groups": slackChannels,
            }
          }
        
        createAlerts(activeProject.id, payload).then(res => {
            setLoading(false);
            fetchAlerts(activeProject.id);
            notification.success({
                message: "Alerts Saved",
                description: "New Alerts is created and saved successfully.",
            });
            form.resetFields();
            setShowForm(false);
        }).catch(err => {
            setLoading(false);
            notification.error({
                message: "Error",
                description: err?.data?.error,
            });
            console.log('create alerts error->', err)
        });
    };

    const emailView = () => {
        if (viewAlertDetails.alert_configuration.emails) {
            return viewAlertDetails.alert_configuration.emails.map((item, index) => {
                return (
                    <Input disabled={true} key={index} value={item} className={'fa-input'} size={'large'} placeholder={'yourmail@gmail.com'} />
                )
            })
        }
    }

    const onConnectSlack = () => {
        enableSlackIntegration(activeProject.id)
        .then((r) => {
            if (r.status == 200) {
                window.open(r.data.redirectURL, "_blank");
            }
            if (r.status >= 400) {
                message.error('Error fetching slack redirect url');
            }
        })
        .catch((err) => {
            console.log('Slack error-->', err);
        });
    }


    useEffect(() => {
        if (!savedAlerts) {
            setTableLoading(true);
            fetchAlerts(activeProject.id).then(() => {
                setTableLoading(false);
            })
        }
    }, [savedAlerts]);

    useEffect(() => {
        if (savedAlerts) {
            let savedArr = [];
            savedAlerts?.map((item, index) => {
                savedArr.push({
                    key: index,
                    alert: (item.alert_description.name + ' ' + item.alert_description.operator + ' ' + item.alert_description.value).replace(/_/g, ' '),
                    dop: (item.alert_configuration.email_enabled ? 'Email': '') + ' ' + (item.alert_configuration.slack_enabled ? 'Slack' : ''),
                    actions: item,
                });
            });
            setTableData(savedArr);
        }
    }, [savedAlerts]);

    const onChange = () => {
        seterrorInfo(null);
    };   

    const DateRangeTypes =[
        {value: 'last_week', label: 'Last week'},
        {value: 'last_month', label: 'Last month'},
        {value: 'last_quarter', label: 'Last quarter'},
      ];
    
    const DateRangeTypeSelect = (
        <Select className={'fa-select w-full'} size={'large'}
            options={DateRangeTypes}
            placeholder="Date range"
            showSearch
            
        >
        </Select>
    );

    const operatorOpts = [
        {label: 'is less than', value: 'is_less_than'},
        {label: 'is greater than', value: 'is_greater_than'},
        {label: 'decreased by more than', value: 'decreased_by_more_than'},
        {label: 'increased by more than', value: 'increased_by_more_than'},
        {label: 'increased or decreased by more than', value: 'increased_or_decreased_by_more_than'},
        {label: '% has decreased by more than', value: '%_has_decreased_by_more_than'},
        {label: '% has increased by more than', value: '%_has_increased_by_more_than'},
        {label: '% has increased or decreased by more than', value: '%_has_increased_or_decreased_by_more_than'},
    ];

    const selectOperator = (
        <Select className={'fa-select w-full'} size={'large'}
            options={operatorOpts}
            placeholder="Operator"
            showSearch
            onChange={(value) => {
                setOperatorState(value);
            }
            }
        >
        </Select>
    );

    useEffect(() => {
        fetchProjectSettingsV1(activeProject.id);
        fetchSlackChannels(activeProject.id);
    }, [activeProject, projectSettings?.int_slack, slackEnabled]);

    useEffect(() => {
        if (slack?.length > 0) {
            let tempArr = [];
            for (let i = 0; i < slack.length; i++) {
                tempArr.push({name: slack[i].name, id: slack[i].id, is_private: slack[i].is_private});
            }
            setChannelOpts(tempArr);
        }
    }, [activeProject, agent_details, slack]);

    const handleOk = () => {
        setSelectedChannel(selectedChannel);
        setShowSelectChannelsModal(false);
    }
    
    return (
        <div className={'fa-container mt-32 mb-12 min-h-screen'}>
          <Row gutter={[24, 24]} justify='center'>
            <Col span={18}>
            <div className={'mb-10 pl-4'}>

                {(!showForm && !viewMode) && <>
                    <Row>
                        <Col span={12}>
                            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Alerts</Text>
                        </Col>
                        <Col span={12}>
                            <div className={'flex justify-end'}>
                                <Button size={'large'} onClick={() => setShowForm(true)}><SVG name={'plus'} extraClass={'mr-1'} size={16} />New Alert</Button>
                            </div>
                        </Col>
                    </Row>
                    <Row className={'m-0'}>
                        <Col span={24}>
                            <div className={'m-0'}>
                                <Table className="fa-table--basic mt-8"
                                    columns={columns}
                                    dataSource={tableData}
                                    pagination={false}
                                    loading={tableLoading}
                                    tableLayout={'fixed'}
                                />
                            </div>
                        </Col>
                    </Row>
                </>
                }
                {(showForm && !viewMode) && <>


                    <Form
                        form={form}
                        onFinish={onFinish}
                        className={'w-full'}
                        onChange={onChange}
                        loading={true}
                    >
                        <Row>
                            <Col span={12}>
                                <Text type={'title'} level={3} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Create new alert</Text>
                            </Col>
                            <Col span={12}>
                                <div className={'flex justify-end'}>
                                    <Button size={'large'} disabled={loading} onClick={() => {
                                        setShowForm(false);
                                        setOperatorState('');
                                        setValue('');
                                        setQueries([]);
                                        setShowCompareField(false);
                                        setEmailEnabled(false);
                                        setSlackEnabled(false);
                                        form.resetFields();
                                    }}>Cancel</Button>
                                    <Button size={'large'} disabled={loading} loading={loading} className={'ml-2'} type={'primary'} htmlType="submit">Save</Button>
                                </div>
                            </Col>
                        </Row>
                        <Row className={'mt-8'}>
                            <Col span={18}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Notify me when</Text>
                            </Col>
                        </Row>
                        <Row className={'m-0'}>
                            <Col span={18}>
                                <Form.Item
                                    name="query_type"
                                    className={'m-0'}
                                >
                                    {queryList()}
                                </Form.Item>
                            </Col>
                        </Row>
                        <Row className={'mt-4'}>
                            <Col span={18}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Operator</Text>
                            </Col>
                        </Row>
                        <Row className={'mt-4'}>
                            <Col span={8} className={'m-0'}>
                                <Form.Item
                                    name="operator"
                                    className={'m-0'}
                                    rules={[{ required: true, message: 'Please select Operator' }]}
                                >
                                    {selectOperator}
                                </Form.Item>
                            </Col>
                            <Col span={8} className={'ml-4'}>
                                <Form.Item
                                    name="value"
                                    className={'m-0'}
                                    rules={[{ required: true, message: 'Please enter value' }]}
                                >
                                    <Input className={'fa-input'} size={'large'} type={'number'} placeholder={'Qualifier'} onChange={(e) => setValue(e.target.value)}/>
                                </Form.Item>
                            </Col>
                        </Row>

                        <Row className={'mt-4'}>
                            <Col span={8}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0 mb-1'}>In the period of</Text>
                                <Form.Item
                                    name="date_range"
                                    className={'m-0'}
                                    rules={[{ required: true, message: 'Please select Date range' }]}
                                >
                                    {DateRangeTypeSelect}
                                </Form.Item>
                            </Col>
                            {showCompareField && 
                            <Col span={8} className={'ml-4'}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0 mb-1'}>Compared to</Text>
                                <Form.Item
                                    name="compared_to"
                                    className={'m-0'}
                                    initialValue={'previous_period'}
                                    rules={[{ required: true, message: 'Please select Compare' }]}
                                >
                                    <Select className={'fa-select w-full'} size={'large'}
                                        placeholder="Compare"
                                        showSearch
                                        disabled={true}
                                    >
                                       <Option value="previous_period">Previous period</Option>
                                    </Select>
                                </Form.Item>
                            </Col>
                            }
                        </Row>

                        <Row className={'mt-2'}>
                            <Col span={24}>
                                <div className={'border-top--thin-2 pt-2 mt-2'} />
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Delivery options</Text>
                            </Col>
                        </Row>
                        
                        <Row className={'mt-2 ml-2'}>
                            <Col span={4}>
                                <Form.Item
                                    name="email_enabled"
                                    className={'m-0'}
                                >
                                    <Checkbox defaultChecked={emailEnabled} onChange={(e) => setEmailEnabled(e.target.checked)}>Email</Checkbox>
                                </Form.Item>
                            </Col>
                        </Row>
                        {emailEnabled && (
                        <Row className={'mt-4'}>
                            <Col span={8}>
                                <Form.Item
                                    label={null}
                                    name={'email'}
                                    validateTrigger={['onChange', 'onBlur']}
                                    rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                                >
                                <Input className={'fa-input'} size={'large'} placeholder={'yourmail@gmail.com'} />
                                </Form.Item>
                            </Col>
                            <Form.List
                            name="emails"
                            >
                            {(fields, { add, remove }) => (
                                <>
                                {fields.map((field, index) => (
                                <Col span={21}>
                                <Form.Item
                                    required={false}
                                    key={field.key}
                                >
                                <Row className={'mt-4'}>
                                    <Col span={9}>
                                        <Form.Item
                                            label={null}
                                            {...field}
                                            name={[field.name, 'email']}
                                            validateTrigger={['onChange', 'onBlur']}
                                            rules={[{ type: 'email', message: 'Please enter a valid e-mail' }, { required: true, message: 'Please enter email' }]} className={'m-0'}
                                        >
                                        <Input className={'fa-input'} size={'large'} placeholder={'yourmail@gmail.com'} />
                                        </Form.Item>
                                    </Col>
                                    {fields.length > 0 ? (
                                    <Col span={1} >
                                    <Button style={{backgroundColor:'white'}} className={'mt-1'} onClick={() => remove(field.name)}>
                                        <SVG
                                        name={'Trash'}
                                        size={20}
                                        color='gray'
                                        /></Button>
                                    </Col>
                                        ) : null}
                                </Row>
                                </Form.Item>
                                </Col>
                                ))}
                                <Col span={20} className={'mt-3'}>
                                {fields.length === 4 ? null: <Button type={'text'} icon={<PlusOutlined style={{color:'gray', fontSize:'18px'}} />} onClick={() => add()}>Add Email</Button>}
                                </Col>
                                </>
                                )}
                            </Form.List>
                        </Row>
                        )}
                        {/* <Row className={'mt-2 ml-2'}>
                            <Col className={'m-0'}>
                                <Form.Item
                                    name="slack_enabled"
                                    className={'m-0'}
                                >
                                    <Checkbox defaultChecked={slackEnabled} onChange={(e) => setSlackEnabled(e.target.checked)}>Slack</Checkbox>
                                </Form.Item>
                            </Col>
                        </Row>
                        {slackEnabled && !projectSettings.int_slack && (
                            <>
                                <Row className={'mt-2 ml-2'}>
                                    <Col span={10} className={'m-0'}>
                                        <Text type={'title'} level={6} color={'grey'} extraClass={'m-0'}>Slack is not integrated, Do you want to integrate with your slack account now?</Text>
                                    </Col>
                                </Row>
                                <Row className={'mt-2 ml-2'}>
                                    <Col span={10} className={'m-0'}>
                                        <Button onClick={onConnectSlack}><SVG name={'Slack'} />Connect to slack</Button>
                                    </Col>
                                </Row>
                            </>
                        )}
                        {slackEnabled && projectSettings.int_slack && (
                            <>
                                {selectedChannel.length > 0 && (
                                <Row className={'rounded-lg border-2 border-gray-200 mt-2 w-2/6'}>
                                    <Col className={'m-0'}>
                                        <Text type={'title'} level={6} color={'grey-2'} extraClass={'m-0 mt-2 ml-2'}>Selected Channels</Text>
                                        {selectedChannel.map((channel, index) => (
                                            <div key={index} >
                                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 ml-2 mt-1 mb-1'}>{'#'+ channel.name}</Text>
                                            </div>
                                        ))}
                                    </Col>
                                </Row>
                                )}
                                {!selectedChannel.length > 0 ? (
                                <Row className={'mt-2 ml-2'}>
                                    <Col span={10} className={'m-0'}>
                                        <Button type={'link'} onClick={() => setShowSelectChannelsModal(true)}>Select Channels</Button>
                                    </Col>
                                </Row>
                                ):
                                <Row className={'mt-2 ml-2'}>
                                    <Col span={10} className={'m-0'}>
                                        <Button type={'link'} onClick={() => setShowSelectChannelsModal(true)}>Manage Channels</Button>
                                    </Col>
                                </Row>
                                }
                            </>
                        )} */}

                    </Form>


                </>}

                {viewMode && <>

                    <Row>
                        <Col span={12}>
                            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>View Alert</Text>
                        </Col>
                        <Col span={12}>
                            <div className={'flex justify-end'}>
                                <Button size={'large'} disabled={loading} onClick={() => {
                                    SetViewMode(false);
                                }}>Back</Button>
                            </div>
                        </Col>
                    </Row>

                    <Row className={'mt-8'}>
                        <Col span={18}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Notify me when</Text>
                        </Col>
                    </Row>
                    <Row className={'m-0 mt-2'}>
                        <Col>
                            <Button
                            className={`mr-2`}
                            type='link'
                            disabled={true}
                            >
                                {viewAlertDetails?.alert_description?.name}
                            </Button>
                        </Col>
                        <Col>
                            {viewAlertDetails?.alert_description?.query?.pgUrl && (
                                <div>
                                    <span className={'mr-2'}>from</span>
                                    <Button
                                    className={`mr-2`}
                                    type='link'
                                    disabled={true}
                                    >
                                        {viewAlertDetails?.alert_description?.query?.pgUrl}
                                    </Button>
                                </div>
                            )}
                        </Col>
                    </Row>
                    {viewAlertDetails?.alert_description?.query?.fil?.length > 0 && (
                        <Row className={'mt-2'}>
                            <Col span={18}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0 my-1'}>Filters</Text>
                                {viewFilter.map((filter, index) => (
                                    <div key={index} className={'mt-1'}>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.extra[0]}
                                        </Button>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.operator}
                                        </Button>
                                        <Button
                                        className={`mr-2`}
                                        type='link'
                                        disabled={true}
                                        >
                                            {filter.values[0]}
                                        </Button>
                                    </div>
                                ))}
                            </Col>
                        </Row>
                    )}
                    <Row className={'mt-4'}>
                        <Col span={18}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Operator</Text>
                        </Col>
                    </Row>
                    <Row className={'mt-4'}>
                        <Col span={8} className={'ml-1'}>
                            <Input disabled={true} size="large"  className={'fa-input w-full'} value={(viewAlertDetails?.alert_description?.operator).replace(/_/g, ' ')} />
                        </Col>
                        <Col span={8} className={'ml-4 w-24'}>
                            <Input disabled={true} className={'fa-input'} size={'large'} type={'number'} value={viewAlertDetails?.alert_description?.value}/>
                        </Col>
                    </Row>

                    <Row className={'mt-4'}>
                        <Col span={8}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0 mb-1'}>In the period of</Text>
                            <Input disabled={true} size="large"  className={'fa-input w-full'} value={viewAlertDetails?.alert_description?.date_range} />
                        </Col>
                        {viewAlertDetails?.alert_description?.compared_to && (
                        <Col span={8} className={'ml-4'}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0 mb-1'}>Compared to</Text>
                            <Input disabled={true} size="large" className={'fa-input w-full'} value={viewAlertDetails?.alert_description?.compared_to}  />
                        </Col>
                        )}
                    </Row>

                    <Row className={'mt-2'}>
                        <Col span={24}>
                            <div className={'border-top--thin-2 pt-2 mt-2'} />
                            <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Delivery options</Text>
                        </Col>
                    </Row>
                    
                    <Row className={'mt-2 ml-2'}>
                        <Col span={4}>
                                <Checkbox disabled={true} checked={viewAlertDetails?.alert_configuration?.email_enabled}>Email</Checkbox>
                        </Col>
                    </Row>
                    <Row className={'mt-4'}>
                        <Col span={8}>
                            {emailView()}
                        </Col>
                    </Row>
                    {/* <Row className={'mt-2 ml-2'}>
                        <Col span={4}>
                                <Checkbox disabled={true} checked={viewAlertDetails?.alert_configuration?.slack_enabled}>Slack</Checkbox>
                        </Col>
                    </Row>
                    {viewAlertDetails?.alert_configuration?.slack_enabled && viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups && (
                    <Row className={'rounded-lg border-2 border-gray-200 mt-2 ml-2 w-2/6'}>
                        <Col className={'m-0'}>
                            <Text type={'title'} level={6} color={'grey-2'} extraClass={'m-0 mt-2 ml-2'}>Selected Channels</Text>
                            {viewSelectedChannels.map((channel, index) => (
                                <div key={index} >
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 ml-2 mt-1 mb-1'}>{'#'+ channel.name}</Text>
                                </div>
                            ))}
                        </Col>
                    </Row>
                    )} */}

                </>}

                <ConfirmationModal
                    visible={deleteWidgetModal ? true : false}
                    confirmationText="Do you really want to remove this alert?"
                    onOk={confirmDelete}
                    onCancel={showDeleteWidgetModal.bind(this, false)}
                    title="Remove Alerts"
                    okText="Confirm"
                    cancelText="Cancel"
                    confirmLoading={deleteApiCalled}
                />                
            </div>
            </Col>
          </Row>

            <Modal
                title={null}
                visible={showSelectChannelsModal}
                centered={true}
                zIndex={1005}
                width={700}
                onCancel={() => setShowSelectChannelsModal(false)}
                onOk={handleOk}
                className={'fa-modal--regular p-4 fa-modal--slideInDown'}
                closable={true}
                okText={'Save'}
                cancelText={'Close'}
                transitionName=''
                maskTransitionName=''
                okButtonProps={{ size: 'large' }}
                cancelButtonProps={{ size: 'large' }}
            >
                <div>
                <Row>
                    <Col span={24}>
                    <Text
                        type={'title'}
                        level={4}
                        weight={'bold'}
                        size={'grey'}
                        extraClass={'m-0'}
                    >
                        Select slack channels
                    </Text>
                    </Col>
                </Row>
                <Row>
                    <Col span={24}>
                        <SelectChannels
                            channelOpts={channelOpts}
                            selectedChannel={selectedChannel}
                            setSelectedChannel={setSelectedChannel}
                        />
                    </Col>
                </Row>
                </div>
            </Modal>
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    savedAlerts: state.global.Alerts,
    kpi: state?.kpi,
    agent_details: state.agent.agent_details,
    slack: state.global.slack,
    projectSettings: state.global.projectSettingsV1,
});


export default connect(mapStateToProps, { createAlerts, fetchAlerts, deleteAlerts, fetchSlackChannels, fetchProjectSettingsV1, enableSlackIntegration })(Alerts)