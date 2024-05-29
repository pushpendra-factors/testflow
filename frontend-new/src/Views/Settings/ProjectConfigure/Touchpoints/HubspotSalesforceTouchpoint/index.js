import React, { useState, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import { getEventPropertiesV2 } from 'Reducers/coreQuery/middleware';
import { udpateProjectDetails } from 'Reducers/global';
import { Row, Col, Button, Table, Dropdown, Menu } from 'antd';
import { setDisplayName } from 'Utils/dataFormatter';
import {
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';
import { MoreOutlined } from '@ant-design/icons';
import EmptyScreen from 'Components/EmptyScreen';
import FAFilterSelect from 'Components/FaFilterSelect';

import { OTPService } from 'Reducers/touchpoints/services';
import useService from 'hooks/useService';
import { Otp } from 'Reducers/touchpoints/classes';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';
import TouchpointView from '../TouchPointView';

const HubspotSalesforceTouchpoint = ({
  activeProject,
  getEventPropertiesV2,
  type
}) => {
  const otpService = useService(activeProject.id, OTPService);
  const { eventPropNames } = useSelector((state) => state.coreQuery);

  const tabNo = type === 'hubspot' ? '3' : '4';

  const [touchPointsData, setTouchPointsData] = useState([]);

  const [touchPointState, setTouchPointState] = useState({
    state: 'list',
    index: 0,
    loading: true
  });

  const columns = [
    {
      title: tabNo === '3' ? 'Hubspot Object' : 'Salesforce Object',
      dataIndex: 'filters',
      key: 'filters',
      render: (obj) => renderObjects(obj)
    },
    {
      title: 'Property Mapping',
      dataIndex: 'properties_map',
      key: 'properties_map',
      render: (obj) => renderPropertyMap(obj)
    },
    {
      title: '',
      dataIndex: 'id',
      key: 'id',
      render: (obj) => renderTableActions(obj)
    }
  ];

  useEffect(() => {
    getOtpObjects();
  }, [activeProject, otpService]);

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
    getEventPropertiesV2(activeProject.id, '$sf_campaign_member_updated');
    getEventPropertiesV2(activeProject.id, '$sf_campaign_member_created');
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
    getEventPropertiesV2(activeProject.id, '$hubspot_contact_updated');
  };

  const setPropertyMapByDisplayName = (propertyMap) => {
    const propMap = { ...propertyMap };
    Object.keys(propMap).forEach((key) => {
      propMap[key].va = setDisplayName(eventPropNames, propMap[key].va);
    });
    return propMap;
  };

  const menu = (index) => (
    <Menu>
      <Menu.Item
        key='0'
        onClick={() => setTouchPointState({ state: 'edit', index })}
      >
        <a>Edit</a>
      </Menu.Item>
      <Menu.Item key='0' onClick={() => deleteTchPoint(index)}>
        <a>Delete</a>
      </Menu.Item>
    </Menu>
  );

  const renderTableActions = (index) => (
    <Dropdown overlay={() => menu(index)} trigger={['hover']}>
      <Button type='text' icon={<MoreOutlined />} />
    </Dropdown>
  );

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
      <div className='mt-2 max-w-xl overflow-hidden'>
        <FAFilterSelect filter={filt} disabled applyFilter={() => {}} />
      </div>
    ));
  };

  const renderPropertyMap = (obj) => (
    <Col>
      {obj.$type && obj.$type.va && (
        <Row>
          <Col span={8}>
            <Row className='relative justify-between break-words'>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                Type
              </Text>
              <SVG name='ChevronRight' />
            </Row>
          </Col>

          <Col className='fa-truncate-150 break-words'>
            <Text level={7} type='title' extraClass='ml-4' weight='thin'>
              {obj.$type.va}
            </Text>
          </Col>
        </Row>
      )}

      {obj.$source && obj.$source.va && (
        <Row>
          <Col span={8}>
            <Row className='relative justify-between break-words'>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                Source
              </Text>
              <SVG name='ChevronRight' />
            </Row>
          </Col>

          <Col className='fa-truncate-150 break-words'>
            <Text level={7} type='title' extraClass='ml-4' weight='thin'>
              {obj.$source.va}
            </Text>
          </Col>
        </Row>
      )}

      {obj.$campaign && obj.$campaign.va && (
        <Row>
          <Col span={8}>
            <Row className='relative justify-between break-words'>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                Campaign
              </Text>
              <SVG name='ChevronRight' />
            </Row>
          </Col>

          <Col className='fa-truncate-150 break-words'>
            <Text level={7} type='title' extraClass='ml-4' weight='thin'>
              {obj.$campaign.va}
            </Text>
          </Col>
        </Row>
      )}

      {obj.$channel && obj.$channel.va && (
        <Row>
          <Col span={8}>
            <Row className='relative justify-between break-words'>
              <Text level={7} type='title' extraClass='m-0' weight='thin'>
                Channel
              </Text>
              <SVG name='ChevronRight' />
            </Row>
          </Col>

          <Col className='fa-truncate-150 break-words'>
            <Text level={7} type='title' extraClass='ml-4' weight='thin'>
              {obj.$channel.va}
            </Text>
          </Col>
        </Row>
      )}
    </Col>
  );

  const renderTitle = () => {
    let title = null;
    if (touchPointState.state === 'list') {
      title = (
        <Text type='title' level={7} extraClass='m-0' color='grey'>
          {tabNo === '3'
            ? 'Hubspot Offline Touchpoint'
            : 'Salesforce Offline Touchpoint'}
        </Text>
      );
    }
    if (touchPointState.state === 'add') {
      title = (
        <Text type='title' level={7} extraClass='m-0' color='grey'>
          Add new Touchpoint
        </Text>
      );
    }

    if (touchPointState.state === 'edit') {
      title = (
        <Text
          type='title'
          level={3}
          weight='bold'
          extraClass='m-0'
          id='fa-at-text--page-title'
        >
          Edit Touchpoint
        </Text>
      );
    }
    return title;
  };

  const renderTitleActions = () => {
    let titleAction = null;
    if (touchPointState.state === 'list') {
      titleAction = (
        <Button
          type='primary'
          onClick={() => {
            setTouchPointState({ state: 'add', index: 0 });
          }}
        >
          <SVG name='plus' extraClass='mr-2' size={16} color='white' />
          Add New
        </Button>
      );
    }

    return titleAction;
  };

  const getOtpObjects = () => {
    const hsData = [];
    const sfData = [];
    setTouchPointState({ state: 'list', index: 0, loading: true });
    otpService?.getTouchPoints().then((res) => {
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
    otpService?.removeTouchPoint(index).then((res) => {
      getOtpObjects();
    });
    // Handle Error case
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
    // Save OTP
    const otpObj = setTouchPointObj(tchObj, getCRMType());
    if (touchPointState.state === 'edit') {
      otpService?.modifyTouchPoint(otpObj, index).then((res) => {
        getOtpObjects();
      });
    } else {
      otpService?.createTouchPoint(otpObj).then((res) => {
        getOtpObjects();
      });
    }
  };

  const onTchCancel = () => {
    setTouchPointState({ state: 'list', index: 0 });
  };

  const getCRMType = () => {
    if (tabNo === '3') return 'hubspot';
    if (tabNo === '4') return 'salesforce';
  };

  const renderTouchPointContent = () => {
    let touchPointContent = null;
    if (touchPointState.state === 'list') {
      touchPointContent = (
        <div className='mb-10 mt-2'>
          {touchPointsData.filter((obj) => obj.crm_type === getCRMType())
            .length > 0 ? (
            <Table
              className='fa-table--basic'
              columns={columns}
              dataSource={touchPointsData.filter(
                (obj) => obj.crm_type === getCRMType()
              )}
              pagination={false}
              loading={touchPointState.loading}
            />
          ) : (
            <EmptyScreen
              learnMore='https://help.factors.ai/'
              loading={touchPointState.loading}
              title={`Record offline touchpoints from ${
                tabNo === '3' ? 'Hubspot' : 'Salesforce'
              } to attribute them accurately to leads, opportunities, and pipeline stages. Apply custom rules to define your touchpoints accurately.`}
            />
          )}
        </div>
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
    <div>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <Row>
            <Col span={20}>{renderTitle()}</Col>
            <Col span={4}>
              <div className='flex justify-end'>{renderTitleActions()}</div>
            </Col>
          </Row>

          <Row className='mt-4'>
            <Col span={24}>
              <div className='mt-6'>{renderTouchPointContent()}</div>
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
      getEventPropertiesV2,
      udpateProjectDetails
    },
    dispatch
  );

export default withFeatureLockHOC(
  connect(mapStateToProps, mapDispatchToProps)(HubspotSalesforceTouchpoint),
  {
    featureName: FEATURES.FEATURE_OFFLINE_TOUCHPOINTS,
    LockedComponent: (props) => (
      <CommonLockedComponent
        featureName={FEATURES.FEATURE_OFFLINE_TOUCHPOINTS}
        variant='tab'
        {...props}
      />
    )
  }
);
