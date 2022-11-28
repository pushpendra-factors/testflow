import React, { useEffect, useState } from 'react';
import { Table, Button, Spin } from 'antd';
import Modal from 'antd/lib/modal/Modal';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import ContactDetails from './ContactDetails';
import {
  ProfileMapper,
  profileOptions,
  ReverseProfileMapper
} from '../../../utils/constants';
import FaSelect from '../../FaSelect';
import { getUserProperties } from '../../../reducers/coreQuery/middleware';
import PropertyFilter from './PropertyFilter';
import MomentTz from '../../MomentTz';
import {
  fetchDemoProject,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
} from '../../../reducers/global';
import ProfileBeforeIntegration from '../ProfileBeforeIntegration';
import { formatFiltersForPayload } from '../utils';
import {
  getProfileUsers,
  getProfileUserDetails
} from '../../../reducers/timelines/middleware';
import _ from 'lodash';

function UserProfiles({
  activeProject,
  contacts,
  getProfileUsers,
  getProfileUserDetails,
  getUserProperties,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  fetchDemoProject,
  currentProjectSettings
}) {
  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const columns = [
    {
      title: <div className={headerClassStr}>Identity</div>,
      dataIndex: 'identity',
      key: 'identity'
    },
    {
      title: <div className={headerClassStr}>Country</div>,
      dataIndex: 'country',
      key: 'country',
      render: (item) => item || '-'
    },
    {
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      render: (item) => MomentTz(item).format('DD MMMM YYYY, hh:mm:ss A')
    }
  ];
  const [isDDVisible, setDDVisible] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [demoProjectId, setDemoProjectId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeUser, setActiveUser] = useState({});
  const [filterPayload, setFilterPayload] = useState({
    source: 'web',
    filters: []
  });

  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector(
    (state) => state.dashboard
  );

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
    if (_.isEmpty(dashboards?.data)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
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
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null ||
    (integration?.int_client_six_signal_key || integration?.int_factors_six_signal_key);

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
      const opts = { ...filterPayload };
      opts.source = ProfileMapper[val[0]] || val[0];
      setFilterPayload(opts);
    }
    setDDVisible(false);
  };

  const setFilters = (filters) => {
    const opts = { ...filterPayload };
    opts.filters = filters;
    setFilterPayload(opts);
  };

  const clearFilters = () => {
    const opts = { ...filterPayload };
    opts.filters = [];
    setFilterPayload(opts);
  };

  useEffect(() => {
    const opts = { ...filterPayload };
    opts.filters = formatFiltersForPayload(filterPayload.filters);
    getProfileUsers(activeProject.id, opts);
  }, [filterPayload]);

  const selectUsers = () => (
    <div className="absolute top-0">
      {isDDVisible ? (
        <FaSelect
          options={[['All'], ...profileOptions.users]}
          onClickOutside={() => setDDVisible(false)}
          optionClick={(val) => onChange(val)}
        />
      ) : null}
    </div>
  );

  if (loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (isIntegrationEnabled || activeProject.id === demoProjectId) {
    return (
      <div className="fa-container mt-24 mb-12 min-h-screen">
        <Text type="title" level={3} weight="bold">
          User Profiles
        </Text>
        <div className="flex justify-between items-start my-4">
          <div className="flex items-start">
            <div className="relative mr-2">
              <Button
                className="fa-dd--custom-btn"
                type="text"
                icon={<SVG name="user_friends" size={16} />}
                onClick={() => setDDVisible(!isDDVisible)}
              >
                {ReverseProfileMapper[filterPayload.source]?.users || 'All'}
                <SVG name="caretDown" size={16} />
              </Button>
              {selectUsers()}
            </div>
            <div key={0} className="max-w-3xl">
              <PropertyFilter
                profileType="user"
                source={filterPayload.source}
                filters={filterPayload.filters}
                setFilters={setFilters}
                onFiltersLoad={[() => getUserProperties(activeProject.id)]}
              />
            </div>
          </div>
          {filterPayload.filters.length ? (
            <div>
              <Button
                className="fa-dd--custom-btn"
                type="text"
                icon={<SVG name="times_circle" size={16} />}
                onClick={clearFilters}
              >
                Clear Filters
              </Button>
            </div>
          ) : null}
        </div>
        {contacts.isLoading ? (
          <Spin size="large" className="fa-page-loader" />
        ) : (
          <div>
            <Table
              onRow={(user) => ({
                onClick: () => {
                  getProfileUserDetails(
                    activeProject.id,
                    user.identity,
                    user.is_anonymous,
                    currentProjectSettings.timelines_config
                  );
                  setActiveUser(user);
                  showModal();
                }
              })}
              className="fa-table--basic"
              dataSource={contacts.data}
              columns={columns}
              rowClassName="cursor-pointer"
              pagination={{ position: ['bottom', 'left'] }}
            />
          </div>
        )}

        <Modal
          title={null}
          visible={isModalVisible}
          className="fa-modal--full-width"
          footer={null}
          closable={null}
        >
          <ContactDetails user={activeUser} onCancel={handleCancel} />
        </Modal>
      </div>
    );
  }
  return <ProfileBeforeIntegration />;
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  contacts: state.timelines.contacts,
  userDetails: state.timelines.contactDetails,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileUsers,
      getProfileUserDetails,
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
