import React, { useEffect, useState } from 'react';
import { Row, Col, Table } from 'antd';
import People from './People.json';
import { Text } from '../../factorsComponents';
import Modal from 'antd/lib/modal/Modal';
import ContactDetails from './ContactDetails';
import { connect, useSelector } from 'react-redux';
import {
  fetchProfileUserDetails,
  fetchProfileUsers,
} from '../../../reducers/timeline';
import moment from 'moment';
import { bindActionCreators } from 'redux';

const UserProfiles = ({
  activeProject,
  contacts,
  userDetails,
  fetchProfileUsers,
  fetchProfileUserDetails,
}) => {
  const columns = [
    {
      title: 'Identity',
      dataIndex: 'identity',
      key: 'identity',
      width: 300,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Groups',
      dataIndex: 'groups',
      key: 'groups',
    },
    {
      title: 'Last Activity',
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      render: (item) => moment(item).format('DD MMMM YYYY, hh:mm:ss'),
    },
  ];

  const [isModalVisible, setIsModalVisible] = useState(false);

  useEffect(() => {
    fetchProfileUsers(activeProject.id);
  }, [activeProject]);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };

  return (
    <div className={'fa-container mt-24 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <Col span={24}>
            <Text
              type={'title'}
              level={3}
              weight={'bold'}
              extraClass={'m-0 mb-4'}
            >
              User Profiles
            </Text>
          </Col>
          <Col span={24}>
            <Table
              onRow={(user) => {
                return {
                  onClick: () => {
                    fetchProfileUserDetails(
                      activeProject.id,
                      user.identity,
                      user.is_anonymous
                    );
                    showModal();
                  },
                };
              }}
              className='fa-table--basic'
              dataSource={contacts}
              columns={columns}
              rowClassName='cursor-pointer'
              pagination={{ position: ['bottom', 'left'] }}
            />
          </Col>
        </Col>
      </Row>
      <Modal
        title={null}
        visible={isModalVisible}
        onCancel={handleCancel}
        className={'fa-modal--full-width'}
        footer={null}
        closable={null}
      >
        <ContactDetails onCancel={handleCancel} userDetails={userDetails} />
      </Modal>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  contacts: state.timeline.contacts,
  userDetails: state.timeline.contactDetails,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProfileUsers,
      fetchProfileUserDetails,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
