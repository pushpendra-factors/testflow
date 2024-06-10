import { SVG, Text } from 'Components/factorsComponents';
import { Button, Dropdown, Menu, Row, Table, notification, Col } from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import EmptyScreen from 'Components/EmptyScreen';
import { ConnectedProps, connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  deleteSmartProperty,
  fetchSmartProperties
} from 'Reducers/settings/middleware';
import { MoreOutlined } from '@ant-design/icons';
import ConfirmationModal from 'Components/ConfirmationModal';
import logger from 'Utils/logger';
import SmartProperties from '../../PropertySettings/SmartProperties';

const CampaignGroup = ({
  fetchSmartProperties,
  deleteSmartProperty,
  smartProperties,
  activeProject
}: CampaignGroupProps) => {
  const [selectedProperty, setSelectedProperty] = useState(null);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [showPropertyForm, setShowPropertyForm] = useState(false);
  const [tableLoading, setTableLoading] = useState(false);

  const [smartPropData, setSmartPropData] = useState([]);

  const editProp = (obj) => {
    setSelectedProperty(obj);
    setShowPropertyForm(true);
  };
  const menu = (obj) => (
    <Menu style={{ display: 'block' }}>
      <Menu.Item key='0' onClick={() => showDeleteWidgetModal(obj.id)}>
        Remove Property
      </Menu.Item>
      <Menu.Item key='0' onClick={() => editProp(obj)}>
        Edit Property
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: 'Display name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <span className='capitalize'>{text}</span>
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (text) => (
        <span className='capitalize'>{text ? text.replace('_', ' ') : ''}</span>
      )
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (obj) => (
        <div className='flex justify-end'>
          <Dropdown
            overlay={() => menu(obj)}
            trigger={['click']}
            placement='bottomRight'
          >
            <Button size='large' type='text' icon={<MoreOutlined />} />
          </Dropdown>
        </div>
      )
    }
  ];

  const confirmRemove = (id) =>
    deleteSmartProperty(activeProject.id, id).then(
      (res) => {
        fetchSmartProperties(activeProject.id);
        notification.success({
          message: 'Success',
          description: 'Deleted property successfully ',
          duration: 5
        });
      },
      (err) => {
        notification.error({
          message: 'Error',
          description: err.data.error,
          duration: 5
        });
      }
    );

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
    } catch (err) {
      logger.error(err);
      logger.error(err?.response);
    }
  }, [deleteWidgetModal]);

  useEffect(() => {
    const properties = [];
    smartProperties.forEach((prop) => {
      // harcoded type
      properties.push({
        name: prop.name,
        type: prop.type_alias,
        actions: prop
      });
    });
    setSmartPropData(properties);
  }, [smartProperties]);

  useEffect(() => {
    if (activeProject?.id) {
      setTableLoading(true);
      fetchSmartProperties(activeProject.id).then(() => {
        setTableLoading(false);
      });
    }
  }, [activeProject]);
  return (
    <div className='mb-4'>
      {!showPropertyForm && (
        <Row>
          <Col span={20}>
            <div>
              <Text type='title' level={7} extraClass='m-0' color='grey'>
                Organize your campaigns and ad-groups efficiently by grouping
                them based on relevant criteria for streamlined management and
                analysis.{' '}
                {/* <a
                  href='https://help.factors.ai/en/articles/7284109-custom-properties'
                  target='_blank'
                  rel='noreferrer'
                >
                  Learn more
                </a> */}
              </Text>
            </div>
          </Col>
          <Col span={4}>
            <div className='flex justify-end'>
              <Button
                onClick={() => {
                  setShowPropertyForm(true);
                }}
                type='primary'
              >
                <SVG name='plus' size={16} color='white' />
                Add New
              </Button>
            </div>
          </Col>
        </Row>
      )}

      {showPropertyForm ? (
        <SmartProperties
          smartProperty={selectedProperty}
          setShowSmartProperty={(showVal) => {
            setShowPropertyForm(showVal);
            setSelectedProperty(null);
            fetchSmartProperties(activeProject.id);
          }}
        />
      ) : !showPropertyForm && smartPropData && smartPropData.length > 0 ? (
        <Table
          className='fa-table--basic mt-4'
          columns={columns}
          dataSource={smartPropData}
          pagination={false}
          loading={tableLoading}
        />
      ) : (
        <EmptyScreen
          learnMore='https://help.factors.ai/en/articles/7284109-custom-properties'
          loading={tableLoading}
          title={`Group ad campaigns by themes to understand how your investment in these advertising campaigns are paying off. Such as grouping campaigns containing ‘search’ in their name into 'Search Campaigns'.`}
        />
      )}
      <ConfirmationModal
        visible={!!deleteWidgetModal}
        confirmationText='Do you really want to remove this property?'
        onOk={confirmDelete}
        onCancel={() => showDeleteWidgetModal(false)}
        title='Remove Property'
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteApiCalled}
      />
    </div>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchSmartProperties,
      deleteSmartProperty
    },
    dispatch
  );

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  smartProperties: state.settings.smartProperties
});
const connector = connect(mapStateToProps, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;

type CampaignGroupProps = ReduxProps;

export default connector(CampaignGroup);
