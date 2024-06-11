import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { Text } from 'factorsComponents';
import { Row, Col, Button, Tabs, Select, message } from 'antd';
import {
  AttributionGroupOptions,
  DealOrOppurtunity,
  CompanyOrAccount
} from 'Utils/constants';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import SelectKPIBlock from './SelectKPIBlock';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../../reducers/global';

const { TabPane } = Tabs;
const { Option } = Select;

const defaultAttrConfigValue = {
  kpis_to_attribute: {
    user_kpi: [],
    sf_kpi: [],
    hs_kpi: []
  },
  attribution_window: 1,
  query_type: 'ConversionBased',
  enabled: true,
  user_kpi: true,
  hubspot_deals: false,
  salesforce_opportunities: false,
  hubspot_companies: true,
  salesforce_accounts: true,
  pre_compute_enabled: false
};

const AttributionSettings = ({
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings
}) => {
  const [attrConfig, setAttrConfig] = useState(defaultAttrConfigValue);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    if (currentProjectSettings?.attribution_config) {
      setAttrConfig(currentProjectSettings.attribution_config);
    }
  }, [currentProjectSettings]);

  const tabsArray = [
    ['Users', 'user_kpi'],
    ['Salesforce Opportunities', 'sf_kpi'],
    ['Hubspot Deals', 'hs_kpi']
  ];

  const [tabNo, setTabNo] = useState('0');

  function callback(key) {
    setTabNo(key);
  }

  const onWindowChange = (val) => {
    const opts = { ...attrConfig };
    opts.attribution_window = val;
    setAttrConfig(opts);
  };

  const onQueryTypeChange = (val) => {
    const opts = { ...attrConfig };
    opts.query_type = val;
    setAttrConfig(opts);
  };

  const renderEditActions = () => (
    <div className='flex justify-end'>
      <Button className='mr-2' size='large' onClick={onCancel}>
        Cancel
      </Button>
      <Button type='primary' size='large' onClick={onSave}>
        Save
      </Button>
    </div>
  );

  const onSave = () => {
    udpateProjectSettings(activeProject.id, {
      attribution_config: attrConfig
    })
      .then(() => {
        message.success('Project details updated!');
      })
      .catch((err) => {
        console.log('err->', err);
        message.error(err.data.error);
      });
  };

  const onCancel = () => {
    fetchProjectSettings(activeProject.id);
    setAttrConfig(
      currentProjectSettings?.attribution_config || defaultAttrConfigValue
    );
  };

  const kpiList = (header) => {
    const blockList = [];
    const value = attrConfig?.kpis_to_attribute[header] || [];

    value.forEach((ev, index) => {
      blockList.push(
        <div key={index}>
          <SelectKPIBlock
            header={header}
            index={index}
            ev={ev}
            value={value}
            attrConfig={attrConfig}
            setAttrConfig={setAttrConfig}
          />
        </div>
      );
    });

    if (value.length < 10) {
      blockList.push(
        <div key='init'>
          <SelectKPIBlock
            header={header}
            index={value.length + 1}
            value={value}
            attrConfig={attrConfig}
            setAttrConfig={setAttrConfig}
          />
        </div>
      );
    }

    return blockList;
  };

  const selectQueryType = () => {
    const queryTypes = [
      ['Conversion Time', 'ConversionBased'],
      ['Interaction Time', 'EngagementBased']
    ];
    return (
      <Select
        value={attrConfig?.query_type}
        style={{ width: 300 }}
        placement='bottomLeft'
        defaultValue='ConversionBased'
        onChange={(val) => {
          onQueryTypeChange(val);
        }}
      >
        {queryTypes.map((qType, index) => (
          <Option value={qType[1]}>{qType[0]}</Option>
        ))}
      </Select>
    );
  };

  const selectWindow = () => {
    const window = [1, 3, 7, 14, 20, 30, 60, 90, 180, 365];
    return (
      <Select
        value={attrConfig?.attribution_window}
        style={{ width: 300 }}
        placement='bottomLeft'
        onChange={(val) => {
          onWindowChange(val);
        }}
      >
        {window.map((days, index) => (
          <Option value={days}>
            {Number.isInteger(days) && `${days} ${days === 1 ? 'day' : 'days'}`}
          </Option>
        ))}
      </Select>
    );
  };

  const attributionWindow = () => (
    <div>
      <Row>
        <Col span={24}>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Attributions Window
          </Text>
        </Col>
        <Col span={18}>
          <Text type='title' level={7} extraClass='m-0'>
            Changing the attribution window will only apply going forward. These
            changes will be reflected in all reports within the Analytics
            property.
          </Text>
        </Col>
        <Col span={24}>
          <div className='mt-4'>{selectWindow()}</div>
        </Col>
      </Row>
    </div>
  );

  const renderAttributionQueryType = () => (
    <div>
      <Row>
        <Col span={24}>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Attributions Query Type
          </Text>
        </Col>
        <Col span={18}>
          {/* <Text type={'title'} level={7} extraClass={'m-0'}>
              Changing the attribution window will only apply going forward.
              These changes will be reflected in all reports within the
              Analytics property.
            </Text> */}
        </Col>
        <Col span={24}>
          <div className='mt-4'>{selectQueryType()}</div>
        </Col>
      </Row>
    </div>
  );

  const renderAttributionContent = () => (
    <Tabs activeKey={`${tabNo}`} onChange={callback}>
      {tabsArray.map((name, index) => (
        <TabPane tab={name[0]} key={index}>
          <div
            style={{
              border: '1px solid #f0f0f0',
              padding: '20px',
              paddingTop: 0
            }}
          >
            {kpiList(name[1])}
          </div>
        </TabPane>
      ))}
    </Tabs>
  );

  const onGroupAttributionChange = (val) => {
    const updatedAttrConfig = { ...attrConfig };
    if (val === CompanyOrAccount) {
      updatedAttrConfig.hubspot_companies = true;
      updatedAttrConfig.salesforce_accounts = true;
      updatedAttrConfig.hubspot_deals = false;
      updatedAttrConfig.salesforce_opportunities = false;
    } else if (val === DealOrOppurtunity) {
      updatedAttrConfig.hubspot_companies = false;
      updatedAttrConfig.salesforce_accounts = false;
      updatedAttrConfig.hubspot_deals = true;
      updatedAttrConfig.salesforce_opportunities = true;
    }
    setAttrConfig(updatedAttrConfig);
  };

  const getGroupAttributionValue = () => {
    if (
      attrConfig?.hubspot_companies === true &&
      attrConfig?.salesforce_accounts === true
    )
      return CompanyOrAccount;
    if (
      attrConfig?.hubspot_deals === true &&
      attrConfig?.salesforce_opportunities === true
    )
      return DealOrOppurtunity;
    return null;
  };

  const selectGroupAttribution = () => (
    <Select
      value={getGroupAttributionValue()}
      style={{ width: 300 }}
      placement='bottomLeft'
      onChange={(val) => {
        onGroupAttributionChange(val);
      }}
    >
      {AttributionGroupOptions.map((group) => (
        <Option value={group} key={group}>
          {group}
        </Option>
      ))}
    </Select>
  );

  const groupAttribution = () => (
    <div>
      <Row>
        <Col span={24}>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Group Attribution
          </Text>
        </Col>
        <Col span={18}>
          <Text type='title' level={7} extraClass='m-0'>
            This option allows you to attribute revenue to either all contacts
            associated with Deals / Opportunities Or all contacts associated
            with the Company/ Account. By default this is set to Company /
            Account. Pick one below.
          </Text>
        </Col>
        <Col span={24}>
          <div className='mt-4'>{selectGroupAttribution()}</div>
        </Col>
      </Row>
    </div>
  );

  return (
    <div className='fa-container'>
      <CommonSettingsHeader
        title='Attribution'
        description='Attribute revenue and conversions to the right channels, campaigns, and touchpoints using different models to identify effective strategies and maximize ROI.'
        actionsNode={renderEditActions()}
      />
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <Row className='flex items-center' />
          <Row>
            <div className='fa-warning' style={{ marginTop: -16 }}>
              This is configured at the time of initial setup. We don't support
              to change it at the moment. Please contact customer support for
              more details.
            </div>
          </Row>
          <Row>
            <Col span={24}>
              <Text type='title' level={6} weight='bold' extraClass='m-0'>
                KPI's to contribute
              </Text>
            </Col>
            <Col span={24}>
              <Text type='title' level={7} extraClass='m-0'>
                Select the KPI's to be considered as part of Attribution
                Reporting. You can select upto 5 KPI's to atrribution.
              </Text>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <div className='mt-6'>{renderAttributionContent()}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <div className='my-6'>{attributionWindow()}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <div className='my-6'>{renderAttributionQueryType()}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <div className='my-6'>{groupAttribution()}</div>
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

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings
})(AttributionSettings);
