import React, { useState, useEffect, useCallback } from 'react';
import { connect, useSelector, useDispatch } from "react-redux";
import {
    Row, Col, Select, Menu, Dropdown, Button, Form, Table, Input, message, Collapse, notification, Tooltip, Space, Checkbox
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, PlusOutlined } from '@ant-design/icons';
import _ from 'lodash';
import GroupSelect2 from '../../../../components/KPIComposer/GroupSelect2';
import FaSelect from 'Components/FaSelect';
import { createAlerts, fetchAlerts, deleteAlerts } from 'Reducers/global';
import ConfirmationModal from '../../../../components/ConfirmationModal';

const { Option } = Select;

const Alerts = ({
    activeProject,
    kpi,
    createAlerts,
    fetchAlerts,
    deleteAlerts,
    savedAlerts,
}) => {

    const [showForm, setShowForm] = useState(false);
    const [tableData, setTableData] = useState([]);
    const [tableLoading, setTableLoading] = useState(false);
    const [errorInfo, seterrorInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [viewMode, SetViewMode] = useState(false);
    const [viewAlertDetails, setAlertDetails] = useState(false);
    const [isDDVisible, setDDVisible] = useState(false);
    const [operSelectOpen, setOperSelectOpen] = useState(false);
    const [operatorState, setOperatorState] = useState(null);
    const [Value, setValue] = useState(null);
    const [ValueOpen, setValueOpen] = useState(false);
    const [emailEnabled, setEmailEnabled] = useState(false);
    const [slackEnabled, setSlackEnabled] = useState(false);
    const [queries, setQueries] = useState({});
    const [showCompareField, setShowCompareField] = useState(false);
    const [alertType, setAlertType] = useState(1);

    const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
    const [deleteApiCalled, setDeleteApiCalled] = useState(false);
    
    const [form] = Form.useForm();


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


    const SelectOperator = (val) => {
        setOperatorState(val[1]);
        setOperSelectOpen(false);
    }

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

    let kpiEvents = kpi?.config?.map((item) => {
        let metricsValues = item?.metrics?.map((value) => {
          if (value?.display_name) {
            return [value?.display_name, value?.name];
          } else {
            return [value, value];
          }
        });
        return {
          label: item.display_category,
          group: item.display_category,
          category: item.category,
          icon: 'custom_events',
          values: metricsValues,
        };
      });

      const onChangeDD = (value, group, category) => {
        const newEvent = { alias: '', label: '', filters: [], group: '' };
        newEvent.label = value[0];
        newEvent.metric = value[1];
        newEvent.group = group;
        if (category) {
          newEvent.category = category;
        }
        setDDVisible(false);
        setQueries(newEvent)
      };


        const triggerDropDown = () => {
            setDDVisible(true);
        };
    
      const selectEvents = () => {
        return (
          <>
            {isDDVisible ? (
              <div>
                <GroupSelect2
                  groupedProperties={kpiEvents ? kpiEvents : []}
                  placeholder='Select Event'
                  optionClick={(group, val, category) =>
                    onChangeDD(val, group, category)
                  }
                  onClickOutside={() => setDDVisible(false)}
                  allowEmpty={true}
                />
              </div>
            ) : null}
          </>
        );
      }; 
    

    const onFinish = data => {
        setLoading(true);
        // Putting All emails into single array
        let emails = [];
        if(emailEnabled) {
            if (data.emails) {
                emails = data.emails.map((item) => {
                    return item.email
                })
            }
            if (data.email) {
                emails.push(data.email)
            }
        }
        
        let payload = {
            "alert_type": alertType,
            "alert_description": {
              "name" : queries.metric,
              "query": {
                'ca': queries.category,
                'dc': queries.group,
                'me': [queries.metric],
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
        ['is less than', 'is_less_than'],
        ['is greater than', 'is_greater_than'],
        ['decreased by more than', 'decreased_by_more_than'],
        ['increased by more than', 'increased_by_more_than'],
        ['increased or decreased by more than', 'increased_or_decreased_by_more_than'],
        ['% has decreased by more than', '%_has_decreased_by_more_than'],
        ['% has increased by more than', '%_has_increased_by_more_than'],
        ['% has increased or decreased by more than', '%_has_increased_or_decreased_by_more_than'],
    ];

    return (
        <>
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
                        <Row className={'mt-4'}>
                            <Col>
                                <Form.Item
                                    name="query_type"
                                    className={'m-0'}
                                    // rules={[{ required: true, message: 'Please select KPI' }]}
                                >
                                    <Button
                                    className={`mr-2`}
                                    type='link'
                                    onClick={triggerDropDown}
                                    >
                                        {queries?.label ? queries?.label :'Select KPI'}
                                    </Button>
                                    {selectEvents()}
                                </Form.Item>
                            </Col>
                            <Col className={'ml-1'}>
                                <Form.Item
                                    name="operator"
                                    className={'m-0'}
                                    // rules={[{ required: true, message: 'Please select operator' }]}
                                >
                                    <Button
                                    className={`mr-2`}
                                    type='link'
                                    onClick={() => setOperSelectOpen(true)}
                                    >
                                    {operatorState ? operatorState.replace(/_/g, ' ') : 'Select Operator'}
                                    </Button>

                                    {operSelectOpen && (
                                    <FaSelect
                                        options={operatorOpts}
                                        optionClick={(val) => SelectOperator(val)}
                                        onClickOutside={() => setOperSelectOpen(false)}
                                    ></FaSelect>
                                    )}
                                </Form.Item>
                            </Col>
                            <Col className={'ml-1 w-24'}>
                                <Form.Item
                                    name="value"
                                    className={'m-0'}
                                    rules={[{ required: true, message: 'Please enter value' }]}
                                >
                                    {/* <Button
                                    className={`mr-2`}
                                    type='link'
                                    onClick={() => setValueOpen(!ValueOpen)}
                                    >
                                    {Value ? Value : 'Value'}
                                    </Button> */}

                                    {/* {ValueOpen && ( */}
                                        <Input className={'fa-input'} type={'number'} onChange={(e) => setValue(e.target.value)}/>
                                    {/* )} */}
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
                            {/* <Col span={4} className={'ml-4'}>
                                <Form.Item
                                    name="slack_enabled"
                                    className={'m-0'}
                                >
                                    <Checkbox defaultChecked={slackEnabled} onChange={(e) => setSlackEnabled(e.target.checked)}>Slack</Checkbox>
                                </Form.Item>
                            </Col> */}
                        </Row>
                        {emailEnabled && (
                        <Row className={'mt-4'}>
                            <Col span={10}>
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
                                <Col span={16}>
                                <Form.Item
                                    required={false}
                                    key={field.key}
                                >
                                <Row className={'mt-4'}>
                                    <Col span={15}>
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
                                <Col span={16} className={'mt-3'}>
                                {fields.length === 4 ? null: <Button type={'text'} icon={<PlusOutlined style={{color:'gray', fontSize:'18px'}} />} onClick={() => add()}>Add Email</Button>}
                                </Col>
                                </>
                                )}
                            </Form.List>
                        </Row>
                        )}

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
                        <Row className={'mt-4'}>
                            <Col>
                                    <Button
                                    className={`mr-2`}
                                    type='link'
                                    disabled={true}
                                    >
                                        {viewAlertDetails?.alert_description?.name}
                                    </Button>
                            </Col>
                            <Col className={'ml-1'}>
                                    <Button
                                    className={`mr-2`}
                                    type='link'
                                    disabled={true}
                                    >
                                        {(viewAlertDetails?.alert_description?.operator).replace(/_/g, ' ')}
                                    </Button>
                            </Col>
                            <Col className={'ml-1 w-24'}>
                                    <Input disabled={true} className={'fa-input'} type={'number'} value={viewAlertDetails?.alert_description?.value}/>
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
                            {/* <Col span={4} className={'ml-4'}>
                                    <Checkbox disabled={true} checked={viewAlertDetails?.alert_configuration?.slack_enabled}>Slack</Checkbox>
                            </Col> */}
                        </Row>
                        <Row className={'mt-4'}>
                            <Col span={10}>
                                {emailView()}
                            </Col>
                        </Row>

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
        </>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    savedAlerts: state.global.Alerts,
    kpi: state?.kpi,
});


export default connect(mapStateToProps, { createAlerts, fetchAlerts, deleteAlerts })(Alerts)