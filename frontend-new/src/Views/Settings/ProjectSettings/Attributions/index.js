import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { Text } from 'factorsComponents';
import { Row, Col, Button, Tabs, Select, message } from 'antd';
import SelectKPIBlock from './SelectKPIBlock';
import {
  udpateProjectSettings,
  fetchProjectSettings,
} from '../../../../reducers/global';

const { TabPane } = Tabs;
const { Option } = Select;

const Attributions = ({
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
}) => {
  const [attrConfig, setAttrConfig] = useState({
    kpis_to_attribute: {
      user_kpi: [],
      sf_kpi: [],
      hs_kpi: [],
    },
    attribution_window: 1,
    enabled: true,
  });

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
    ['Hubspot Deals', 'hs_kpi'],
  ];

  const [tabNo, setTabNo] = useState('0');
  const [edit, setEdit] = useState(false);

  function callback(key) {
    setTabNo(key);
  }

  const onWindowChange = (val) => {
    setEdit(true);
    const opts = Object.assign({}, attrConfig);
    opts.attribution_window = val;
    setAttrConfig(opts);
  };

  const renderEditActions = () => {
    return (
      <div className={`flex justify-end`}>
        <Button className={'mr-2'} size={'large'} onClick={onCancel}>
          Cancel
        </Button>
        <Button type={'primary'} size={'large'} onClick={onSave}>
          Save
        </Button>
      </div>
    );
  };

  const onSave = () => {
    udpateProjectSettings(activeProject.id, {
      attribution_config: attrConfig,
    })
      .then(() => {
        message.success('Project details updated!');
        setEdit(false);
      })
      .catch((err) => {
        console.log('err->', err);
        message.error(err.data.error);
      });
    setEdit(false);
  };

  const onCancel = () => {
    fetchProjectSettings(activeProject.id);
    setAttrConfig(currentProjectSettings.attribution_config);
    setEdit(false);
  };

  const kpiList = (header) => {
    const blockList = [];
    const value = attrConfig?.kpis_to_attribute[header]
      ? attrConfig?.kpis_to_attribute[header]
      : [];

    value.forEach((ev, index) => {
      blockList.push(
        <div key={index}>
          <SelectKPIBlock
            header={header}
            index={index}
            ev={ev}
            editMode={() => setEdit(true)}
            value={value}
            attrConfig={attrConfig}
            setAttrConfig={setAttrConfig}
          />
        </div>
      );
    });

    if (value.length < 5) {
      blockList.push(
        <div key={'init'}>
          <SelectKPIBlock
            header={header}
            index={value.length + 1}
            editMode={() => setEdit(true)}
            value={value}
            attrConfig={attrConfig}
            setAttrConfig={setAttrConfig}
          />
        </div>
      );
    }

    return blockList;
  };

  const selectWindow = () => {
    const window = [1, 3, 7, 14, 20, 30, 60, 90, -1];
    return (
      <Select
        value={attrConfig?.attribution_window}
        style={{ width: 300 }}
        placement='bottomLeft'
        onChange={(val) => {
          onWindowChange(val);
        }}
      >
        {window.map((days, index) => {
          return (
            <Option value={days}>
              {Number.isInteger(days) && days !== -1
                ? `${days} ${days === 1 ? 'day' : 'days'}`
                : days === -1
                ? 'Full User Journey'
                : days}
            </Option>
          );
        })}
      </Select>
    );
  };

  const attributionWindow = () => {
    return (
      <div style={{ height: '32rem' }}>
        <Row>
          <Col span={24}>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
              Attributions Window
            </Text>
          </Col>
          <Col span={18}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Changing the attribution window will only apply going forward.
              These changes will be reflected in all reports within the
              Analytics property.
            </Text>
          </Col>
          <Col span={24}>
            <div className={'mt-4'}>{selectWindow()}</div>
          </Col>
        </Row>
      </div>
    );
  };

  const renderAttributionContent = () => {
    return (
      <Tabs activeKey={`${tabNo}`} onChange={callback}>
        {tabsArray.map((name, index) => {
          return (
            <TabPane tab={name[0]} key={index}>
              <div
                style={{
                  border: '1px solid #f0f0f0',
                  padding: '20px',
                  paddingTop: 0,
                }}
              >
                {kpiList(name[1])}
              </div>
            </TabPane>
          );
        })}
      </Tabs>
    );
  };

  return (
    <div>
      <Row className={`flex items-center`}>
        <Col span={12}>
          <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 m-1'}>
            Attributions Configuration
          </Text>
        </Col>
        <Col span={12}>{edit ? renderEditActions() : null}</Col>
      </Row>
      <Row>
        <div className={'fa-warning'}>
          This is configured at the time of initial setup. We don’t support to
          change it at the moment. Please contact customer support for more
          details.
        </div>
      </Row>
      <Row>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
          KPI's to contribute
        </Text>
        <Text type={'title'} level={7} extraClass={'m-0'}>
          Select the KPI’s to be considered as part of Attribution Reporting.
          You can select upto 5 KPI’s to atrribution.
        </Text>
      </Row>
      <Row>
        <Col span={24}>
          <div className={'mt-6'}>{renderAttributionContent()}</div>
        </Col>
      </Row>
      <Row>
        <Col span={24}>
          <div className={'my-6'}>{attributionWindow()}</div>
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings,
})(Attributions);
