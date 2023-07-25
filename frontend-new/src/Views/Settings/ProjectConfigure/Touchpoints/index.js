import React, { useState, useEffect, useCallback } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import { getEventProperties } from 'Reducers/coreQuery/middleware';
import { fetchProjects, udpateProjectDetails } from 'Reducers/global';
import {
  Row,
  Col,
  Button,
  Tabs,
  Table,
  Dropdown,
  Menu,
  notification,
  Tooltip
} from 'antd';
import TouchpointView from './TouchPointView';
import MarketingInteractions from '../MarketingInteractions';
import FAFilterSelect from '../../../../components/FaFilterSelect';
import { setDisplayName } from 'Utils/dataFormatter';

import {
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';

import { MoreOutlined } from '@ant-design/icons';
import { OTPService } from '../../../../reducers/touchpoints/services';
import useService from '../../../../hooks/useService';
import { Otp } from '../../../../reducers/touchpoints/classes';
import { data } from 'autoprefixer';

const { TabPane } = Tabs;

const Touchpoints = ({
  activeProject,
  currentProjectSettings,
  getEventProperties,
  fetchProjects,
  udpateProjectDetails
}) => {
  const otpService = useService(activeProject.id, OTPService);
  const { eventPropNames } = useSelector((state) => state.coreQuery);

  const [tabNo, setTabNo] = useState('1');

  const [touchPointsData, setTouchPointsData] = useState([]);

  const [touchPointState, setTouchPointState] = useState({
    state: 'list',
    index: 0,
    loading: true
  });

  const columns = [
    {
      title: tabNo === '2' ? 'Hubspot Object' : 'Salesforce Object',
      dataIndex: 'filters',
      key: 'filters',
      render: (obj) => {
        return renderObjects(obj);
      }
    },
    {
      title: 'Property Mapping',
      dataIndex: 'properties_map',
      key: 'properties_map',
      render: (obj) => {
        return renderPropertyMap(obj);
      }
    },
    {
      title: '',
      dataIndex: 'id',
      key: 'id',
      render: (obj) => {
        return renderTableActions(obj);
      }
    }
  ];

  function callback(key) {
    setTabNo(key);
  }

  useEffect(() => {
    if (tabNo !== '1') {
      getOtpObjects();
    }
  }, [activeProject, tabNo]);

  const setSalesforceContactData = (data = []) => {
    const touchpointObjs = data.length
      ? [
          ...data
            .filter((dt) => dt.crm_type === 'salesforce')
            .map((rule, id) => ({
              ...rule,
              properties_map: setPropertyMapByDisplayName(rule.properties_map),
              index: id
            }))
        ]
      : [];
    setTouchPointsData(touchpointObjs);
    getEventProperties(activeProject.id, '$sf_campaign_member_updated');
    getEventProperties(activeProject.id, '$sf_campaign_member_created');
  };

  const setHubspotContactData = (data = []) => {
    const touchpointObjs = data.length
      ? [
          ...data
            .filter((dt) => dt.crm_type === 'hubspot')
            .map((rule, id) => ({
              ...rule,
              properties_map: setPropertyMapByDisplayName(rule.properties_map),
              index: id
            }))
        ]
      : [];
    setTouchPointsData(touchpointObjs);
    getEventProperties(activeProject.id, '$hubspot_contact_updated');
  };

  const setPropertyMapByDisplayName = (propertyMap) => {
    const propMap = { ...propertyMap };
    Object.keys(propMap).forEach((key) => {
      propMap[key].va = setDisplayName(eventPropNames, propMap[key].va);
    });
    return propMap;
  };

  const menu = (index) => {
    return (
      <Menu>
        <Menu.Item
          key='0'
          onClick={() => setTouchPointState({ state: 'edit', index: index })}
        >
          <a>Edit</a>
        </Menu.Item>
        <Menu.Item key='0' onClick={() => deleteTchPoint(index)}>
          <a>Delete</a>
        </Menu.Item>
      </Menu>
    );
  };

  const renderTableActions = (index) => {
    return (
      <Dropdown overlay={() => menu(index)} trigger={['hover']}>
        <Button type='text' icon={<MoreOutlined />} />
      </Dropdown>
    );
  };

  const renderObjects = (obj) => {
    const filters = [];
    obj?.forEach((filterObj, ind) => {
      if (filterObj.lop === 'AND') {
        filters.push({
          operator:
            filterObj.ty === 'datetime'
              ? reverseDateOperatorMap[filterObj.op]
              : reverseOperatorMap[filterObj.op],
          props: [
            filterObj.pr,
            filterObj.ty ? filterObj.ty : 'categorical',
            filterObj.en ? filterObj.en : 'event'
          ],
          values: [filterObj.va]
        });
      } else {
        filters[filters.length - 1].values.push(filterObj.va);
      }
    });
    return filters.map((filt) => (
      <div className={`mt-2 max-w-xl overflow-hidden`}>
        <FAFilterSelect
          filter={filt}
          disabled={true}
          applyFilter={() => {}}
        ></FAFilterSelect>
      </div>
    ));
  };

  const renderPropertyMap = (obj) => {
    return (
      <Col>
        {obj['$type'] && obj['$type']['va'] && (
          <Row>
            <Col span={8}>
              <Row className={'relative justify-between break-words'}>
                <Text
                  level={7}
                  type={'title'}
                  extraClass={'m-0'}
                  weight={'thin'}
                >
                  Type
                </Text>
                <SVG name={`ChevronRight`} />
              </Row>
            </Col>

            <Col className={`fa-truncate-150 break-words`}>
              <Text
                level={7}
                type={'title'}
                extraClass={'ml-4'}
                weight={'thin'}
              >
                {obj['$type']['va']}
              </Text>
            </Col>
          </Row>
        )}

        {obj['$source'] && obj['$source']['va'] && (
          <Row>
            <Col span={8}>
              <Row className={'relative justify-between break-words'}>
                <Text
                  level={7}
                  type={'title'}
                  extraClass={'m-0'}
                  weight={'thin'}
                >
                  Source
                </Text>
                <SVG name={`ChevronRight`} />
              </Row>
            </Col>

            <Col className={`fa-truncate-150 break-words`}>
              <Text
                level={7}
                type={'title'}
                extraClass={'ml-4'}
                weight={'thin'}
              >
                {obj['$source']['va']}
              </Text>
            </Col>
          </Row>
        )}

        {obj['$campaign'] && obj['$campaign']['va'] && (
          <Row>
            <Col span={8}>
              <Row className={'relative justify-between break-words'}>
                <Text
                  level={7}
                  type={'title'}
                  extraClass={'m-0'}
                  weight={'thin'}
                >
                  Campaign
                </Text>
                <SVG name={`ChevronRight`} />
              </Row>
            </Col>

            <Col className={`fa-truncate-150 break-words`}>
              <Text
                level={7}
                type={'title'}
                extraClass={'ml-4'}
                weight={'thin'}
              >
                {obj['$campaign']['va']}
              </Text>
            </Col>
          </Row>
        )}

        {obj['$channel'] && obj['$channel']['va'] && (
          <Row>
            <Col span={8}>
              <Row className={'relative justify-between break-words'}>
                <Text
                  level={7}
                  type={'title'}
                  extraClass={'m-0'}
                  weight={'thin'}
                >
                  Channel
                </Text>
                <SVG name={`ChevronRight`} />
              </Row>
            </Col>

            <Col className={`fa-truncate-150 break-words`}>
              <Text
                level={7}
                type={'title'}
                extraClass={'ml-4'}
                weight={'thin'}
              >
                {obj['$channel']['va']}
              </Text>
            </Col>
          </Row>
        )}
      </Col>
    );
  };

  const renderTitle = () => {
    let title = null;
    if (touchPointState.state === 'list') {
      title = (
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
          Touchpoints
        </Text>
      );
    }
    if (touchPointState.state === 'add') {
      title = (
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
          Add new Touchpoint
        </Text>
      );
    }

    if (touchPointState.state === 'edit') {
      title = (
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
          Edit Touchpoint
        </Text>
      );
    }
    return title;
  };

  const renderTitleActions = () => {
    let titleAction = null;
    if (touchPointState.state === 'list') {
      if (tabNo !== '1') {
        titleAction = (
          <Button
            size={'large'}
            onClick={() => {
              setTouchPointState({ state: 'add', index: 0 });
            }}
          >
            <SVG name={'plus'} extraClass={'mr-2'} size={16} />
            Add New
          </Button>
        );
      }
    }

    return titleAction;
  };

  const getOtpObjects = () => {
    const hsData = [];
    const sfData = [];
    setTouchPointState({ state: 'list', index: 0, loading: true });
    otpService.getTouchPoints().then((res) => {
      res?.data?.result?.forEach((data) => {
        if (data.crm_type === 'hubspot') {
          hsData.push(data);
        }
        if (data.crm_type === 'salesforce') {
          sfData.push(data);
        }
      });
      if (getCRMType() === 'hubspot') {
        setHubspotContactData(hsData);
      }
      if (getCRMType() === 'salesforce') {
        setSalesforceContactData(sfData);
      }
      setTouchPointState({ state: 'list', index: 0, loading: false });
    });
  };

  const deleteTchPoint = (index = 0) => {
    otpService.removeTouchPoint(index).then((res) => {
      getOtpObjects();
    });
    //Handle Error case
  };

  const setTouchPointObj = (tchObj, type) => {
    const touchPointObj = new Otp();
    touchPointObj.crm_type = type;
    touchPointObj.filters = tchObj.filters;
    touchPointObj.properties_map = tchObj.properties_map;
    touchPointObj.rule_type = tchObj.rule_type;
    touchPointObj.touch_point_time_ref = tchObj.touch_point_time_ref;
    return touchPointObj;
  };

  const onTchSave = (tchObj, index = -1) => {
    if (tabNo !== '1') {
      // Save OTP
      const otpObj = setTouchPointObj(tchObj, getCRMType());
      if (touchPointState.state === 'edit') {
        otpService.modifyTouchPoint(otpObj, index).then((res) => {
          getOtpObjects();
        });
      } else {
        otpService.createTouchPoint(otpObj).then((res) => {
          getOtpObjects();
        });
      }
    }
  };

  const onTchCancel = () => {
    setTouchPointState({ state: 'list', index: 0 });
  };

  const getCRMType = () => {
    if (tabNo === '2') return 'hubspot';
    if (tabNo === '3') return 'salesforce';
    if (tabNo === '1') return 'digital';
  };

  const renderTouchPointContent = () => {
    let touchPointContent = null;
    if (touchPointState.state === 'list') {
      touchPointContent = (
        <Tabs activeKey={`${tabNo}`} onChange={callback}>
          <TabPane tab='Digital Marketing' key='1'>
            <MarketingInteractions />
          </TabPane>

          <TabPane tab='Hubspot' key='2'>
            <div className={`mb-10 pl-4 mt-10`}>
              <Table
                className='fa-table--basic mt-4'
                columns={columns}
                dataSource={touchPointsData.filter(
                  (obj) => obj.crm_type === getCRMType()
                )}
                pagination={false}
                loading={touchPointState.loading}
              />
            </div>
          </TabPane>

          <TabPane tab='Salesforce' key='3'>
            <div className={`mb-10 pl-4 mt-10`}>
              <Table
                className='fa-table--basic mt-4'
                columns={columns}
                dataSource={touchPointsData.filter(
                  (obj) => obj.crm_type === getCRMType()
                )}
                pagination={false}
                loading={touchPointState.loading}
              />
            </div>
          </TabPane>
        </Tabs>
      );
    } else if (touchPointState.state === 'add') {
      touchPointContent = (
        <TouchpointView
          tchType={tabNo}
          rule={null}
          onSave={onTchSave}
          onCancel={onTchCancel}
        >
          {' '}
        </TouchpointView>
      );
    } else if (touchPointState.state === 'edit' && touchPointsData) {
      touchPointContent = (
        <TouchpointView
          tchType={tabNo}
          rule={touchPointsData.find((f) => f.id === touchPointState.index)}
          onSave={(obj) => onTchSave(obj, touchPointState.index)}
          onCancel={onTchCancel}
        >
          {' '}
        </TouchpointView>
      );
    }
    return touchPointContent;
  };

  return (
    <div className={'fa-container'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <Row>
            <Col span={12}>{renderTitle()}</Col>
            <Col span={12}>
              <div className={'flex justify-end'}>{renderTitleActions()}</div>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={24}>
              <Text
                type={'title'}
                level={7}
                color={'grey-2'}
                extraClass={'m-0'}
              >
                Effortlessly map and standardize your marketing parameters.
                Connect and align UTMs and other parameters used across your
                marketing efforts to a standardized set.
              </Text>
              <Text
                type={'title'}
                level={7}
                color={'grey-2'}
                extraClass={'m-0 mt-2'}
              >
                Query and filter by different parameter values within Factors,
                enabling seamless tracking and analysis of customer touchpoints
              </Text>
              <div className={'mt-6'}>{renderTouchPointContent()}</div>
            </Col>
          </Row>
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventProperties,
      fetchProjects,
      udpateProjectDetails
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(Touchpoints);
