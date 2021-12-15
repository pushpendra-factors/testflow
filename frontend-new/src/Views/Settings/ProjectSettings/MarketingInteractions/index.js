import React, { useEffect, useState } from 'react';
import {
    Row, Col, Switch, Modal, Input, Button, Form, Table, Tag, Space, message
} from 'antd';
import { connect } from 'react-redux';
import { Text, SVG } from 'factorsComponents';
import { scaleLinear } from 'd3-scale';
import { ExclamationCircleOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { udpateProjectDetails } from 'Reducers/global';

const { confirm } = Modal;

const UTM_Mapping = {
    "utm_mapping": {
      "$ad": [
        "$qp_utm_ad",
      ],
      "$term": [
        "$qp_utm_term",
      ],
      "$ad_id": [
        "$qp_utm_ad_id",
        "$qp_utm_adid",
        "$qp_utm_hsa_ad",
      ],
      "$gclid": [
        "$qp_gclid",
        "$qp_utm_gclid",
        "$qp_wbraid",
        "$qp_gbraid"
      ],
      "$fbclid": [
        "$qp_fbclid",
        "$qp_utm_fbclid"
      ],
      "$medium": [
        "$qp_utm_medium"
      ],
      "$source": [
        "$qp_utm_source"
      ],
      "$adgroup": [
        "$qp_utm_adgroup",
        "$qp_utm_ad_group"
      ],
      "$content": [
        "$qp_utm_content",
        "$qp_utm_utm_content"
      ],
      "$keyword": [
        "$qp_utm_keyword",
        "$qp_utm_key_word"
      ],
      "$campaign": [
        "$qp_utm_campaign",
        "$qp_utm_campaign_name"
      ],
      "$creative": [
        "$qp_utm_creative",
        "$qp_utm_creative_id",
        "$qp_utm_creativeid"
      ],
      "$adgroup_id": [
        "$qp_utm_adgroupid",
        "$qp_utm_adgroup_id",
        "$qp_utm_ad_group_id",
        "$qp_utm_hsa_grp",
      ],
      "$campaign_id": [
        "$qp_utm_campaignid",
        "$qp_utm_campaign_id",
        "$qp_utm_hsa_cam",
      ],
      "$keyword_match_type": [
        "$qp_utm_matchtype",
        "$qp_utm_match_type"
      ]
    }
  }
  

const MartInt = ({activeProject, udpateProjectDetails}) => {

    const [newKey, setNewKey] = useState(null);
    const [dataSource, setDataSource] = useState(null);
    // const [UTM, setUTM] = useState(UTM_Mapping?.utm_mapping);
    const [UTM, setUTM] = useState(false);

    const [form] = Form.useForm();
    const [errorInfo, seterrorInfo] = useState(null);
    const [success, setSuccess] = useState(false);
    const [loading, setLoading] = useState(false);
    const [visible, setVisible] = useState(false);

    const onFinish = values => {
        let newArr = [];
        const UTM_Map = UTM; 
        const stripLogic = (item) => item.startsWith("$qp_")? item : `$qp_${item}`;
         Object.keys(UTM_Map).map(function (item, index) {
            if (newKey == item) {
                newArr = [...UTM_Map[newKey], stripLogic(values.utm_tag)];
            }
        });
        let updatedKey = { ...UTM, [newKey]: newArr }; 
        setLoading(true);
        udpateProjectDetails(activeProject.id, {interaction_settings : {
            "utm_mapping": updatedKey
        }}).then(() => {
            message.success('UTM Tag added!');
            setVisible(false);
            onReset();
            setLoading(false);
          }).catch((err) => {
            console.log('err->', err);
            seterrorInfo(err.data.error); 
            setLoading(false);
          }); 
      };

    const addUTM = (key) => { 
        setNewKey(key)
        setVisible(true); 
    }

    const removeUTM = (e, key, item) => {
        e.preventDefault();
        confirm({
            icon: <ExclamationCircleOutlined />,
            content: `Are you sure you want to delete?`,
            onOk() { 
                let UTM_Map = UTM;
                let updatedKey = null;
                setLoading(true);
                Object.keys(UTM_Map).map(function (tag, index) {
                    if (key == tag) {
                        let filteredItems = UTM_Map[key].filter(arryItem => arryItem !== item)
                        updatedKey = { ...UTM, [key]: filteredItems };
                    } 
                }); 
                udpateProjectDetails(activeProject.id, {interaction_settings : {
                    "utm_mapping": updatedKey
                }}).then(() => {
                    message.success('UTM Tag removed!');
                    setLoading(false);
                  }).catch((err) => {
                    console.log('err->', err);
                    message.error(err.data.error);
                    setLoading(false);
                  });
            }
        });


    }

    const columns = [

        {
            title: 'For the following parameters... ',
            dataIndex: 'parameter',
            key: 'parameter',
        },
        {
            title: 'Track with these UTM tags, in this order of preference',
            dataIndex: 'actions',
            key: 'actions',
            render: (text) => {
                let key = text?.key;
                if (text?.tags?.length == 0) {
                    return <><Button onClick={() => { addUTM(key) }} style={{ transform: 'scale(0.8)' }} size={'small'} icon={<SVG name={'plus'} size={12} />}>Add</Button></>
                }
                else{
                    const stripLogic = (item) => item.startsWith("$qp_")? item.slice(4) : item;
                    return text?.tags.map((item, index) => {
                        if (text?.tags?.length == index + 1) {
                            return <><Tag className={'fa-tag--green'} color="green" closable={true} onClose={(e) => removeUTM(e, key, item)}>{stripLogic(item)} </Tag><Button onClick={() => { addUTM(key) }} style={{ transform: 'scale(0.8)' }} size={'small'} icon={<SVG name={'plus'} size={12} />}>Add</Button></>
                        } 
                        else {
                            return <Tag className={'fa-tag--green'} color="green" closable={true} onClose={(e) => removeUTM(e, key, item)} >{stripLogic(item)}</Tag>
                        }
                    }) 
                }
            }
        },
    ];


    useEffect(() => { 
            let DS = Object.keys(UTM)?.map(function (key, index) {
                return {
                    'key': index,
                    'parameter': key,
                    'actions': {
                        'key': key,
                        'tags': UTM[key]
                    }
                }
            });
            setDataSource(DS);  
    }, [UTM]);

    useEffect(() => {  
        if(activeProject){ 
                if(activeProject?.interaction_settings){
                    setUTM(activeProject?.interaction_settings?.utm_mapping) 
                }
                else{

                    setUTM(UTM_Mapping?.utm_mapping);
                }
            }
            else{
                setUTM(UTM_Mapping?.utm_mapping);
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


    return <>
        <div className={'mb-10 pl-4'}>
            {/* <Row>
                <Col span={24}>
                    <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Marketing Touchpoints</Text>
                    <Text type={'title'} level={7} extraClass={'m-0'}>Define how your online marketing efforts should be tracked.</Text>
                </Col>
            </Row> */}

            <Row className={'mt-4'}>
                <Col span={24}>
                    <div className={'mt-6'}>
                        <Table className="fa-table--basic mt-4"
                            columns={columns}
                            dataSource={dataSource}
                            pagination={false}
                        />
                    </div>
                </Col>
            </Row>


        </div>

        <Modal
        visible={visible}
        zIndex={1020}
        onCancel={onReset} 
        className={'fa-modal--regular fa-modal--slideInDown'}
        okText={'Add'}  
        centered={true}
        footer={null}
        transitionName=""
        maskTransitionName=""
      >
        <div className={'p-4'}>
          <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          >
            <Row>
              <Col span={24}>
                <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Add UTM Tag</Text>
              </Col>
            </Row>
            <Row className={'mt-4'}>
              <Col span={24}>
                {/* <Text type={'title'} level={7} extraClass={'m-0'}>UTM TAG</Text> */}
                <Form.Item
                    name="utm_tag"
                    rules={[
                      {
                        required: true,
                        message: 'Please enter UTM Tag'
                      }
                    ]}

                    >
                      <Input disabled={loading} size="large" className={'fa-input w-full'} placeholder="Enter UTM Tag" />
                      </Form.Item>
              </Col>
            </Row>
             
            {errorInfo && <Col span={24}>
                <div className={'flex flex-col justify-center items-center mt-1'} >
                    <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                </div>
            </Col>
            }
            <Row className={'mt-6'}>
              <Col span={24}>
                <div className={'flex justify-end'}>
                  <Button size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>
                  <Button loading={loading} type="primary" size={'large'} htmlType="submit"> Add </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </div>

      </Modal>


    </>
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    agents: state.agent.agents,
    projects: state.global.projects,
    currentAgent: state.agent.agent_details
});


export default connect(mapStateToProps, {udpateProjectDetails})(MartInt)