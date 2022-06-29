import React, { useEffect, useState } from 'react';
import { Table, Button, Spin, notification } from 'antd';
import { Text, SVG } from '../../factorsComponents';
import Modal from 'antd/lib/modal/Modal';
import ContactDetails from './ContactDetails';
import { connect } from 'react-redux';
import {
  fetchProfileUserDetails,
  fetchProfileUsers,
} from '../../../reducers/timeline';
import moment from 'moment';
import { bindActionCreators } from 'redux';
import {
  ProfileMapper,
  profileOptions,
  ReverseProfileMapper,
} from '../../../utils/constants';
import FaSelect from '../../FaSelect';
import { getUserProperties } from '../../../reducers/coreQuery/middleware';
import PropertyFilter from './PropertyFilter';
import { operatorMap } from '../../../Views/CoreQuery/utils';

function UserProfiles({
  activeProject,
  contacts,
  userDetails,
  fetchProfileUsers,
  fetchProfileUserDetails,
  getUserProperties,
}) {
  const columns = [
    {
      title: 'Identity',
      dataIndex: 'identity',
      key: 'identity',
    },
    {
      title: 'Country',
      dataIndex: 'country',
      key: 'country',
    },
    {
      title: 'Last Activity',
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      render: (item) => moment(item).format('DD MMMM YYYY, hh:mm:ss'),
    },
  ];
  const [usersLoading, setUsersLoading] = useState(true);
  const [timelineLoading, setTimelineLoading] = useState(true);
  const [isDDVisible, setDDVisible] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);

  const [filterPayload, setFilterPayload] = useState({
    source: 'web',
    filters: [],
  });

  useEffect(() => {
    getUserProperties(activeProject.id);
  }, [activeProject]);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };

  const onChange = (val) => {
    const opts = Object.assign({}, filterPayload);
    opts.source = ProfileMapper[val[0]] || val[0];
    setFilterPayload(opts);
    setDDVisible(false);
  };

  const setFilters = (filters) => {
    const opts = Object.assign({}, filterPayload);
    opts.filters = filters;
    setFilterPayload(opts);
  };
  const clearFilters = () => {
    const opts = Object.assign({}, filterPayload);
    opts.filters = [];
    setFilterPayload(opts);
  };

  const formatFiltersForPayload = (filters = []) => {
    const filterProps = [];
    filters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            en: 'user_g',
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? val : val,
          });
        });
      } else {
        filterProps.push({
          en: 'user_g',
          lop: 'AND',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? fil.values : fil.values,
        });
      }
    });
    return filterProps;
  };

  useEffect(() => {
    (async () => {
      const opts = Object.assign({}, filterPayload);
      opts.filters = formatFiltersForPayload(filterPayload.filters);
      try {
        await fetchProfileUsers(activeProject.id, opts);
        setUsersLoading(false);
      } catch (err) {
        notification.error({
          message: 'Error loading users.',
          description: getErrorMessage(err),
          duration: 3,
        });
      }
    })();
  }, [activeProject, filterPayload]);

  const selectUsers = () => {
    return (
      <div className='absolute top-0'>
        {isDDVisible ? (
          <FaSelect
            options={[['All'], ...profileOptions.users]}
            onClickOutside={() => setDDVisible(false)}
            optionClick={(val) => onChange(val)}
          ></FaSelect>
        ) : null}
      </div>
    );
  };

  return (
    <div className={'fa-container mt-24 mb-12 min-h-screen'}>
      <Text type={'title'} level={3} weight={'bold'}>
        User Profiles
      </Text>
      <div className='flex justify-between items-start my-4'>
        <div className='flex items-start'>
          <div className='relative mr-2'>
            {
              <Button
                className='fa-dd--custom-btn'
                type='text'
                icon={<SVG name='user_friends' size={16} />}
                onClick={() => setDDVisible(!isDDVisible)}
              >
                {ReverseProfileMapper[filterPayload.source]?.users || 'All'}
                <SVG name='caretDown' size={16} />
              </Button>
            }
            {selectUsers()}
          </div>
          <div key={0} className='max-w-3xl'>
            <PropertyFilter
              filters={filterPayload.filters}
              setFilters={setFilters}
              onFiltersLoad={[() => getUserProperties(activeProject.id)]}
            ></PropertyFilter>
          </div>
        </div>
        <div>
          <Button
            className='fa-dd--custom-btn'
            type='text'
            icon={<SVG name='times_circle' size={16} />}
            onClick={clearFilters}
          >
            Clear Filters
          </Button>
        </div>
      </div>
      {usersLoading ? (
        <Spin size={'large'} className={'fa-page-loader'} />
      ) : (
        <div>
          <Table
            onRow={(user) => {
              return {
                onClick: async () => {
                  try {
                    await fetchProfileUserDetails(
                      activeProject.id,
                      user.identity,
                      user.is_anonymous
                    );
                    setTimelineLoading(false);
                    showModal();
                  } catch (err) {
                    notification.error({
                      message: 'Error fetching User Details.',
                      description: getErrorMessage(err),
                      duration: 3,
                    });
                  }
                },
              };
            }}
            className='fa-table--basic'
            dataSource={contacts}
            columns={columns}
            rowClassName='cursor-pointer'
            pagination={{ position: ['bottom', 'left'] }}
          />
        </div>
      )}

      <Modal
        title={null}
        visible={isModalVisible}
        onCancel={handleCancel}
        className={'fa-modal--full-width'}
        footer={null}
        closable={null}
      >
        <ContactDetails
          onCancel={handleCancel}
          timelineLoading={timelineLoading}
          userDetails={userDetails}
        />
      </Modal>
    </div>
  );
}

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
      getUserProperties,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
