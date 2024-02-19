import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { Text, SVG } from 'factorsComponents';
import { Modal, Col, Button, Tag, Table, Dropdown, Menu, message } from 'antd';
import { udpateProjectDetails } from 'Reducers/global';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import defaultRules from './defaultRules';
import _, { isEqual } from 'lodash';
import { DISPLAY_PROP } from 'Utils/constants';
import { reverseOperatorMap } from 'Utils/operatorMapping';
import styles from './index.module.scss';
import { ReactSortable } from 'react-sortablejs';
import cx from 'classnames';
import RouterPrompt from 'Components/GenericComponents/RouterPrompt';

const { confirm } = Modal;

const DCGTable = ({
  activeProject,
  udpateProjectDetails,
  setShowModalVisible,
  setEditProperty,
  eventPropNames,
  userPropNames,
  enableEdit
}) => {
  const [DCGData, setDCGData] = useState([]);
  const [initialDCGData, setInitialDCGData] = useState([]);
  const [showBottomButtons, setShowBottomButtons] = useState(false);
  const [tableLoading, setTableLoading] = useState(false);

  useEffect(() => {
    setTableLoading(true);

    if (activeProject?.channel_group_rules) {
      const ruleSet = activeProject?.channel_group_rules;

      const transformedData = ruleSet?.map((item, index) => ({
        key: index,
        channel: item?.channel,
        conditions: item?.conditions,
        actions: { index, item }
      }));
      setInitialDCGData(transformedData);
      setDCGData(transformedData);
    } else {
      setInitialDCGData([]);
      setDCGData([]);
    }

    setTableLoading(false);
  }, [activeProject]);

  const getBaseQueryFromResponse = (el) => {
    const filters = [];
    el.forEach((item, i) => {
      const conditionCamelCase = _.camelCase(item.condition);

      if (item.logical_operator === 'AND') {
        filters.push({
          operator: reverseOperatorMap[conditionCamelCase] || item.condition,
          props: [
            'event',
            item.property,
            item.condition == 'between' ? 'datetime' : 'categorical',
            ,
            'event'
          ],
          values: [item.value],
          ref: filters.length
        });
      } else if (filters.length > 0) {
        if (item.property == filters[filters.length - 1].props[1]) {
          filters[filters.length - 1].values.push(item.value);
        } else
          filters.push({
            operator: reverseOperatorMap[conditionCamelCase] || item.condition,
            props: [
              'event',
              item.property,
              item.condition == 'between' ? 'datetime' : 'categorical',
              'event'
            ],
            values: [item.value],
            ref: filters.length - 1
          });
      }
    });
    return filters;
  };
  const mapQueryByRefs = (queryMap) => {
    // TmpType <Key, Array<EachQueryFilter>>
    let tmp = {};

    queryMap.forEach((eachQueryFilter) => {
      if (eachQueryFilter.ref in tmp) {
        // OR Filter
        tmp[eachQueryFilter.ref].push(eachQueryFilter);
      } else {
        // ref doesn't exists so, its a new ADD Fitler
        tmp[eachQueryFilter.ref] = [eachQueryFilter];
      }
    });
    return tmp;
  };
  const matchEventName = (item) => {
    let findItem = eventPropNames?.[item] || userPropNames?.[item];
    return findItem ? findItem : item;
  };

  const renderRow = (data) => {
    if (!data) {
      return null; // Return early if data is falsy
    }

    const queryMap = mapQueryByRefs(getBaseQueryFromResponse(data));

    return (
      <div className='w-full' style={{ maxWidth: '550px' }}>
        {Object.keys(queryMap).map((eachKey, index) => (
          <>
            {queryMap[eachKey].map((eachFilter, eachIndex) => {
              return (
                <div className='inline-flex items-center mb-2' key={eachIndex}>
                  {eachFilter.props.length > 0 ? (
                    <Button type='default'>
                      <Text
                        type='title'
                        weight='thin'
                        color='grey'
                        level={8}
                        truncate
                      >
                        {`${matchEventName(eachFilter.props[1])} ${
                          eachFilter.operator
                        } ${_.join(
                          eachFilter.values.map((vl) =>
                            DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl
                          ),
                          ', '
                        )}`}
                      </Text>
                    </Button>
                  ) : (
                    <div className={styles.internal}>
                      <Text type='title' weight='thin' color='grey' level={8}>
                        {`${eachFilter.operator} ${_.join(
                          eachFilter.values.map((vl) =>
                            DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl
                          ),
                          ', '
                        )}`}
                      </Text>
                    </div>
                  )}
                  {queryMap[eachKey].length > 1 &&
                    eachIndex < queryMap[eachKey].length - 1 && (
                      <Text
                        type='title'
                        weight='thin'
                        color='grey'
                        level={8}
                        extraClass='m-0 mr-1 ml-1'
                      >
                        OR
                      </Text>
                    )}{' '}
                  {index < Object.keys(queryMap).length - 1 &&
                    eachIndex == queryMap[eachKey].length - 1 && (
                      <Text
                        type='title'
                        weight='thin'
                        color='grey'
                        level={8}
                        extraClass='m-0 mr-1 ml-1'
                      >
                        AND
                      </Text>
                    )}
                </div>
              );
            })}
          </>
        ))}
      </div>
    );
  };

  const columns = [
    {
      key: 'sort',
      render: () => (
        <div className={cx(styles.dcgTable__additional_actions)}>
          <SVG name='drag' className={styles.dragIcon} />
        </div>
      )
    },

    {
      title: 'Channel',
      dataIndex: 'channel',
      key: 'channel',
      render: (text) => <span className='capitalize'>{text}</span>
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
          <div className='flex justify-end'>
            <Dropdown overlay={() => menu(obj)} trigger={['click']}>
              <Button size='large' type='text' icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        );
      }
    }
  ];

  const confirmRemove = (el) => {
    confirm({
      title: 'Do you want to remove this channel group?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        const updatedArr = (activeProject?.channel_group_rules || []).filter(
          (item, index) => index !== el.index
        );

        udpateProjectDetails(activeProject.id, {
          channel_group_rules: updatedArr
        })
          .then(() => {
            message.success('Channel group removed!');
          })
          .catch((err) => {
            console.error('Error:', err);
          });
      }
    });
  };

  const EditProperty = (obj) => {
    let queryMap = getBaseQueryFromResponse(obj?.item?.conditions);
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

  const handleMoveRow = (modifiedData) => {
    if (!isEqual(DCGData, modifiedData)) {
      setDCGData(modifiedData);
      setShowBottomButtons(true);
    }
  };

  const handleCancel = () => {
    setDCGData(initialDCGData);
    setShowBottomButtons(false);
  };
  const handleSave = () => {
    let updatedArr = DCGData.filter((item) => {
      if (item.channel !== 'Internal') {
        return item;
      }
    });

    udpateProjectDetails(activeProject.id, {
      channel_group_rules: updatedArr
    })
      .then(() => {
        message.success('Channel Groups Orders Changed!');
      })
      .catch((err) => {
        console.log('err->', err);
      });
  };

  const SortableTable = ({ dataSource, columns, ...otherProps }) => {
    return (
      <ReactSortable
        list={dataSource || []}
        setList={handleMoveRow}
        animation={150}
        tag={'tbody'}
        className='ant-table-tbody'
      >
        {dataSource?.map((item, index) => (
          <tr
            key={item.key}
            className={cx(
              styles.dcgTable__table_row,
              'ant-table-row ant-table-row-level-0'
            )}
          >
            {columns?.map((column) => (
              <td key={column.key} className='ant-table-cell'>
                {column.render
                  ? column.render(item[column.dataIndex])
                  : item[column.dataIndex]}
              </td>
            ))}
          </tr>
        ))}
      </ReactSortable>
    );
  };
  return (
    <div>
      <Text
        type='paragraph'
        mini={6}
        weight={'thin'}
        color={'#3E516C'}
        extraClass={'mt-2'}
      >
        These rules are checked sequentially from top to bottom to assign
        channel.
      </Text>
      <Table
        className='fa-table--basic mt-6'
        columns={columns}
        dataSource={DCGData}
        pagination={false}
        loading={tableLoading}
        components={{
          body: {
            wrapper: (props) => (
              <SortableTable
                {...props}
                dataSource={DCGData}
                columns={columns}
              />
            )
          }
        }}
      />
      {showBottomButtons && (
        <div className={`flex justify-between ${styles.dcgTable__changesCard}`}>
          <Text type={'title'} level={7} extraClass={'m-0'}>
            Order of checking for conditions changed. Do you wish to save this
            new order?
          </Text>
          <div className='flex flex-row gap-4'>
            <Button onClick={handleCancel}>Discard Changes</Button>
            <Button className={'ml-2'} type={'primary'} onClick={handleSave}>
              Save Changes
            </Button>
          </div>
        </div>
      )}
      <RouterPrompt
        when={showBottomButtons}
        title='You have unsaved changes on this page. Would you like to discard the changes?'
        cancelText='Cancel'
        okText='Discard Changes'
        onOK={() => true}
        onCancel={() => false}
      />
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropNames: state.coreQuery.eventPropNames,
  userPropNames: state.coreQuery.userPropNames
});

export default connect(mapStateToProps, { udpateProjectDetails })(DCGTable);
