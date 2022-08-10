import React, { useEffect, useState } from 'react';
import { Table, Button, Spin, notification } from 'antd';
import { Text, SVG } from '../../factorsComponents';
import Modal from 'antd/lib/modal/Modal';
import ContactDetails from './ContactDetails';
import { connect, useSelector } from 'react-redux';
import {
  fetchProfileUserDetails,
  fetchProfileUsers,
} from '../../../reducers/timeline';
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
import MomentTz from '../../MomentTz';
import { fetchDemoProject, fetchProjectSettingsV1, fetchProjectSettings, fetchMarketoIntegration, fetchBingAdsIntegration } from '../../../reducers/global';
import ProfileBeforeIntegration from '../ProfileBeforeIntegration';

function UserProfiles({
  activeProject,
  contacts,
  userDetails,
  fetchProfileUsers,
  fetchProfileUserDetails,
  getUserProperties,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  fetchDemoProject
}) {
  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const columns = [
    {
      title: <div className={headerClassStr}>Identity</div>,
      dataIndex: 'identity',
      key: 'identity',
    },
    {
      title: <div className={headerClassStr}>Country</div>,
      dataIndex: 'country',
      key: 'country',
      render: (item) => item || '-',
    },
    {
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      render: (item) => MomentTz(item).format('DD MMMM YYYY, hh:mm:ss'),
    },
  ];
  const [usersLoading, setUsersLoading] = useState(true);
  const [isDDVisible, setDDVisible] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [demoProjectId, setDemoProjectId] = useState(null);
  const [loading, setLoading] = useState(true);

  const [filterPayload, setFilterPayload] = useState({
    source: 'web',
    filters: [],
  });

  const integration = useSelector((state) => state.global.currentProjectSettings);
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);

  useEffect(() => {
    fetchDemoProject()
      .then((res) => {
        setDemoProjectId(res.data[0]);
      })
      .catch((err) => {
        console.log(err.data.error);
      });
  }, [activeProject]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    fetchBingAdsIntegration(activeProject.id);
    fetchMarketoIntegration(activeProject.id);
  }, [activeProject]);


  const isIntegrationEnabled =
  integration?.int_segment ||
  integration?.int_adwords_enabled_agent_uuid ||
  integration?.int_linkedin_agent_uuid ||
  integration?.int_facebook_user_id ||
  integration?.int_hubspot ||
  integration?.int_salesforce_enabled_agent_uuid ||
  integration?.int_drift ||
  integration?.int_google_organic_enabled_agent_uuid ||
  integration?.int_clear_bit ||
  integrationV1?.int_completed ||
  bingAds?.accounts ||
  marketo?.status || integrationV1?.int_slack || integration?.lead_squared_config !== null;


  useEffect(() => {
    getUserProperties(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };

  const onChange = (val) => {
    if ((ProfileMapper[val[0]] || val[0]) !== filterPayload.source) {
      setUsersLoading(true);
      const opts = Object.assign({}, filterPayload);
      opts.source = ProfileMapper[val[0]] || val[0];
      setFilterPayload(opts);
    }
    setDDVisible(false);
  };

  const setFilters = (filters) => {
    setUsersLoading(true);
    const opts = Object.assign({}, filterPayload);
    opts.filters = filters;
    setFilterPayload(opts);
  };

  const clearFilters = () => {
    setUsersLoading(true);
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

  if (loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (isIntegrationEnabled || activeProject.id === demoProjectId) {
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
          {filterPayload.filters.length ? (
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
          ) : null}
        </div>
        {usersLoading ? (
          <Spin size={'large'} className={'fa-page-loader'} />
        ) : (
          <div>
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
          </div>
        )
      }

      <Modal
        title={null}
        visible={isModalVisible}
        className={'fa-modal--full-width'}
        footer={null}
        closable={null}
      >
        <ContactDetails onCancel={handleCancel} userDetails={userDetails} />
      </Modal>
    </div>
  );
  } else {
    return (
      <>
        <ProfileBeforeIntegration />
      </>
    );
  }

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
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      fetchDemoProject
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
