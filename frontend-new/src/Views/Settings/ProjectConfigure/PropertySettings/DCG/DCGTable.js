import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';

import { Text, SVG } from 'factorsComponents';
import { Modal, Col, Button, Tag, Table, Dropdown, Menu, message } from 'antd';
import { udpateProjectDetails } from 'Reducers/global';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import defaultRules from './defaultRules';
import _ from 'lodash';
import { DISPLAY_PROP } from 'Utils/constants';
import { reverseOperatorMap } from 'Utils/operatorMapping';

const { confirm } = Modal;

const DCGTable = ({
  activeProject,
  udpateProjectDetails,
  setShowModalVisible,
  setEditProperty,
  eventPropNames,
  enableEdit
}) => {
  const [DCGData, setDCGData] = useState([]);

  const [tableLoading, setTableLoading] = useState(false);

  useEffect(() => {
    setTableLoading(true);
    let ruleSet = null;

    if (activeProject?.channel_group_rules) {
      ruleSet = activeProject?.channel_group_rules;
    } else {
      ruleSet = defaultRules;
    }

    // if (_.isEmpty(activeProject?.channel_group_rules)) {
    //   ruleSet = defaultRules;
    // }

    if (ruleSet) {
      let DS = ruleSet?.map((item, index) => {
        return {
          key: index,
          channel: item.channel,
          conditions: item.conditions,
          actions: { index, item }
        };
      });
      setDCGData(DS);
      setTableLoading(false);
    } else {
      setTableLoading(false);
    }
  }, [activeProject]);

  const getBaseQueryfromResponse = (el) => {
    const filters = [];
    el.forEach((item) => {
      if (item.logical_operator === 'AND') {
        let conditionCamelCase = _.camelCase(item.condition);
        filters.push({
          operator: reverseOperatorMap[conditionCamelCase],
          props: [item.property, 'categorical', 'event'],
          values: [item.value]
        });
      } else {
        filters[filters.length - 1].values.push(item.value);
      }
    });

    return filters;
  };

  const matchEventName = (item) => {
    let findItem = eventPropNames?.[item];
    return findItem ? findItem : item;
  };

  const renderRow = (data) => {
    if (data) {
      let queryMap = getBaseQueryfromResponse(data);
      return (
        <div className={'w-full'} style={{ maxWidth: '550px' }}>
          {queryMap.map((item, index) => {
            return (
              <div className={'inline-flex items-center mb-2'} key={index}>
                {/* {
                    index != 0 && 
                  <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 mr-1"}>{item.logical_operator}</Text>
                  }
                  <Tag>{`${item.property} ${returnSymbols(item.condition)} ${item.value}`}</Tag> */}
                <Tag>{`${matchEventName(item?.props[0])} ${
                  item?.operator
                } ${_.join(
                  item?.values.map((vl) =>
                    DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl
                  ),
                  [', ']
                )}`}</Tag>
                {queryMap.length != index + 1 && (
                  <Text
                    type={'title'}
                    weight={'thin'}
                    color={'grey'}
                    level={8}
                    extraClass={'m-0 mr-1'}
                  >{`AND`}</Text>
                )}
              </div>
            );
          })}
        </div>
      );
    } else {
      null;
    }
  };
  const columns = [
    {
      title: 'Channel',
      dataIndex: 'channel',
      key: 'channel',
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Conditions',
      dataIndex: 'conditions',
      key: 'conditions',
      render: (item) => renderRow(item)
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (obj) => {
        if (enableEdit) {
          return null;
        }
        return (
          <div className={`flex justify-end`}>
            <Dropdown overlay={() => menu(obj)} trigger={['click']}>
              <Button size={'large'} type='text' icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        );
      }
    }
  ];

  const confirmRemove = (el) => {
    // activeProject?.channel_group_rules?.filter(item => item !== value)

    confirm({
      title: 'Do you want to remove this channel group?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        let updatedArr = activeProject?.channel_group_rules?.filter(
          (item, index) => {
            if (index != el.index) {
              return item;
            }
          }
        );

        udpateProjectDetails(activeProject.id, {
          channel_group_rules: updatedArr
        })
          .then(() => {
            message.success('Channel group removed!');
          })
          .catch((err) => {
            console.log('err->', err);
          });
      }
    });
  };

  const EditProperty = (obj) => {
    let queryMap = getBaseQueryfromResponse(obj?.item?.conditions);
    let finalData = {
      index: obj?.index,
      channel: obj?.item?.channel,
      conditions: queryMap
    };
    setEditProperty(finalData);
    setShowModalVisible(true);
  };

  const menu = (obj) => {
    return (
      <Menu>
        <Menu.Item key='0' onClick={() => EditProperty(obj)}>
          <a>Edit Property</a>
        </Menu.Item>
        <Menu.Item key='0' onClick={() => confirmRemove(obj)}>
          <a>Remove Property</a>
        </Menu.Item>
      </Menu>
    );
  };

  return (
    <>
      <Table
        className='fa-table--basic mt-4'
        columns={columns}
        dataSource={DCGData}
        pagination={false}
        loading={tableLoading}
      />
    </>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropNames: state.coreQuery.eventPropNames
});

export default connect(mapStateToProps, { udpateProjectDetails })(DCGTable);
