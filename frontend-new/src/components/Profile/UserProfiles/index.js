import React, { useEffect, useState } from 'react';
import {
  Table,
  Button,
  Spin,
  Divider,
  notification,
  Popover,
  Tabs
} from 'antd';
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
  fetchBingAdsIntegration,
  udpateProjectSettings
} from '../../../reducers/global';
import ProfileBeforeIntegration from '../ProfileBeforeIntegration';
import {
  DEFAULT_TIMELINE_CONFIG,
  formatFiltersForPayload,
  formatPayloadForFilters,
  formatSegmentsObjToGroupSelectObj
} from '../utils';
import {
  getProfileUsers,
  getProfileUserDetails,
  createNewSegment,
  getSavedSegments
} from '../../../reducers/timelines/middleware';
import _ from 'lodash';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import SegmentModal from './SegmentModal';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { PropTextFormat } from 'Utils/dataFormatter';

function UserProfiles({
  activeProject,
  contacts,
  segments,
  createNewSegment,
  getSavedSegments,
  getProfileUsers,
  getProfileUserDetails,
  getUserProperties,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  fetchDemoProject,
  currentProjectSettings,
  udpateProjectSettings
}) {
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const userProperties = useSelector((state) => state.coreQuery.userProperties);

  const [isUserDDVisible, setUserDDVisible] = useState(false);
  const [isSegmentDDVisible, setSegmentDDVisible] = useState(false);
  const [showSegmentModal, setShowSegmentModal] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [demoProjectId, setDemoProjectId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeUser, setActiveUser] = useState({});
  const [activeSegment, setActiveSegment] = useState({});
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const [timelinePayload, setTimelinePayload] = useState({
    source: 'web',
    filters: []
  });

  useEffect(() => {
    if (currentProjectSettings?.timelines_config) {
      const timelinesConfig = {};
      timelinesConfig.disabled_events = [
        ...currentProjectSettings?.timelines_config?.disabled_events
      ];
      timelinesConfig.user_config = {
        ...DEFAULT_TIMELINE_CONFIG.user_config,
        ...currentProjectSettings?.timelines_config?.user_config
      };
      timelinesConfig.account_config = {
        ...DEFAULT_TIMELINE_CONFIG.account_config,
        ...currentProjectSettings?.timelines_config?.account_config
      };
      setTLConfig(timelinesConfig);
    }
  }, [currentProjectSettings]);

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
    integration?.int_client_six_signal_key ||
    integration?.int_factors_six_signal_key ||
    integration?.int_rudderstack;

  useEffect(() => {
    getUserProperties(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userProperties,
      currentProjectSettings.timelines_config?.user_config?.table_props
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [currentProjectSettings, userProperties]);

  useEffect(() => {
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const getColumns = () => {
    const columns = [
      {
        title: <div className={headerClassStr}>Identity</div>,
        width: 350,
        dataIndex: 'identity',
        key: 'identity',
        fixed: 'left',
        ellipsis: true
      }
    ];
    currentProjectSettings?.timelines_config?.user_config?.table_props?.forEach(
      (prop) => {
        columns.push({
          title: <div className={headerClassStr}>{PropTextFormat(prop)}</div>,
          dataIndex: prop,
          key: prop,
          width: 350,
          render: (item) => item || '-'
        });
      }
    );
    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      fixed: 'right',
      render: (item) => MomentTz(item).format('DD MMMM YYYY, hh:mm:ss A')
    });
    return columns;
  };

  const getTableData = (data) => {
    const tableData = data.map((row) => {
      return {
        ...row,
        ...row?.table_props
      };
    });
    return tableData;
  };

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };

  const onChange = (val) => {
    if ((ProfileMapper[val[0]] || val[0]) !== timelinePayload.source) {
      const opts = { ...timelinePayload };
      opts.source = ProfileMapper[val[0]] || val[0];
      setTimelinePayload(opts);
    }
    setUserDDVisible(false);
  };

  const setFilters = (filters) => {
    const opts = { ...timelinePayload };
    opts.filters = filters;
    setTimelinePayload(opts);
  };

  const clearFilters = () => {
    const opts = { ...timelinePayload };
    opts.filters = [];
    setTimelinePayload(opts);
  };

  useEffect(() => {
    const opts = { ...timelinePayload };
    opts.filters = formatFiltersForPayload(timelinePayload.filters);
    getProfileUsers(activeProject.id, opts);
  }, [timelinePayload, currentProjectSettings]);

  const handleSaveSegment = (segmentPayload) => {
    createNewSegment(activeProject.id, segmentPayload)
      .then((response) => {
        if (response.type === 'SEGMENT_CREATION_FULFILLED') {
          notification.success({
            message: 'Success!',
            description: response?.payload?.message,
            duration: 3
          });
          setSegmentDDVisible(false);
        }
      })
      .then(() => getSavedSegments(activeProject.id))
      .catch((err) => {
        notification.error({
          message: 'Error',
          description: 'Segment Creation Failed. Invalid Parameters.',
          duration: 3
        });
      });
  };

  const selectUsers = () => (
    <div className='absolute top-0'>
      {isUserDDVisible ? (
        <FaSelect
          options={[['All'], ...profileOptions.users]}
          onClickOutside={() => setUserDDVisible(false)}
          optionClick={(val) => onChange(val)}
        />
      ) : null}
    </div>
  );

  const generateSegmentsList = () => {
    const segmentsList = [];
    if (timelinePayload.source === 'All') {
      Object.entries(segments).forEach(([group, vals]) => {
        const obj = formatSegmentsObjToGroupSelectObj(group, vals);
        segmentsList.push(obj);
      });
    } else {
      const obj = formatSegmentsObjToGroupSelectObj(
        timelinePayload.source,
        segments[timelinePayload.source]
      );
      segmentsList.push(obj);
    }
    return segmentsList;
  };

  const onOptionClick = (_, data) => {
    const opts = { ...timelinePayload };
    opts.segment_id = data[1];
    setActiveSegment(data[2]);
    setTimelinePayload(opts);
    setSegmentDDVisible(false);
  };

  const clearSegment = () => {
    const opts = { ...timelinePayload };
    opts.segment_id = '';
    setActiveSegment({});
    setTimelinePayload(opts);
    setSegmentDDVisible(false);
  };

  const selectSegment = () => (
    <div className='absolute top-8'>
      {isSegmentDDVisible ? (
        <GroupSelect2
          groupedProperties={generateSegmentsList()}
          placeholder='Search Segments'
          optionClick={onOptionClick}
          onClickOutside={() => setSegmentDDVisible(false)}
          allowEmpty
          additionalActions={
            <>
              <Divider className='divider-margin' />
              <div className='flex items-center'>
                <Button
                  type='link'
                  size='large'
                  className='w-full'
                  icon={<SVG name='plus' color='purple' />}
                  onClick={() => setShowSegmentModal(true)}
                >
                  Add New Segment
                </Button>
                {timelinePayload.segment_id && (
                  <Button
                    type='primary'
                    size='large'
                    className='w-full ml-1'
                    onClick={clearSegment}
                    danger
                  >
                    Clear Segment
                  </Button>
                )}
              </div>
              <SegmentModal
                type={timelinePayload.source}
                visible={showSegmentModal}
                segment={{}}
                onSave={handleSaveSegment}
                onCancel={() => setShowSegmentModal(false)}
              />
            </>
          }
        />
      ) : null}
    </div>
  );

  const segmentInfo = () => {
    const cardContent = [];
    if (activeSegment.query) {
      if (activeSegment.query.gp) {
        const filters = formatPayloadForFilters(activeSegment.query.gp);
        cardContent.push(
          <div className='pointer-events-none mt-3'>
            <PropertyFilter
              mode='display'
              filtersLimit={10}
              profileType='user'
              source={timelinePayload.source}
              filters={filters}
            ></PropertyFilter>
          </div>
        );
      }
    }
    return cardContent;
  };

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListUserProps.filter((item) => item.enabled === true).length < 8
    ) {
      const checkListProps = [...checkListUserProps];
      const optIndex = checkListProps.findIndex(
        (obj) => obj.prop_name === option.prop_name
      );
      checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
      setCheckListUserProps(checkListProps);
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum Table Properties Selection Reached.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    const config = { ...tlConfig };
    config.user_config.table_props = checkListUserProps
      .filter((item) => item.enabled === true)
      .map((item) => item?.prop_name);
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...config }
    });
    setShowPopOver(false);
  };

  const popoverContent = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <Tabs.TabPane
        tab={<span className='fa-activity-filter--tabname'>Events</span>}
        key='events'
      >
        <SearchCheckList
          placeholder='Search Properties'
          mapArray={checkListUserProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
          showApply
          onApply={applyTableProps}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>
        <div className='relative mr-2'>
          <Button
            className='dropdown-btn'
            type='text'
            icon={<SVG name='user_friends' size={16} />}
            onClick={() => setUserDDVisible(!isUserDDVisible)}
          >
            {ReverseProfileMapper[timelinePayload.source]?.users || 'All'}
            <SVG name='caretDown' size={16} />
          </Button>
          {selectUsers()}
        </div>
        <div className='relative mr-2'>
          <Popover
            overlayClassName='fa-custom-popover'
            placement='bottomLeft'
            trigger={activeSegment.query ? 'hover' : ''}
            content={segmentInfo}
          >
            <Button
              className='dropdown-btn'
              type='text'
              onClick={() => setSegmentDDVisible(!isSegmentDDVisible)}
            >
              {Object.keys(activeSegment).length
                ? activeSegment.name
                : 'Select Segment'}
              <SVG name='caretDown' size={16} />
            </Button>
          </Popover>
          {selectSegment()}
        </div>
        <div key={0} className='max-w-3xl'>
          <PropertyFilter
            profileType='user'
            source={timelinePayload.source}
            filters={timelinePayload.filters}
            setFilters={setFilters}
            onFiltersLoad={[() => getUserProperties(activeProject.id)]}
          />
        </div>
      </div>
      <div className='flex items-center justify-between'>
        {timelinePayload.filters.length ? (
          <Button
            className='dropdown-btn'
            type='text'
            icon={<SVG name='times_circle' size={16} />}
            onClick={clearFilters}
          >
            Clear Filters
          </Button>
        ) : null}
        <Popover
          overlayClassName='fa-activity--filter'
          placement='bottomLeft'
          visible={showPopOver}
          onVisibleChange={(visible) => {
            setShowPopOver(visible);
          }}
          onClick={() => {
            setShowPopOver(true);
          }}
          trigger='click'
          content={popoverContent}
        >
          <Button
            size='large'
            className='fa-btn--custom mx-2 relative'
            // type='text'
          >
            <SVG name='activity_filter' />
          </Button>
        </Popover>
      </div>
    </div>
  );

  const renderTable = () => (
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
        className='fa-table--basic'
        dataSource={getTableData(contacts.data)}
        columns={getColumns()}
        rowClassName='cursor-pointer'
        pagination={{ position: ['bottom', 'left'] }}
        scroll={{
          x:
            currentProjectSettings?.timelines_config?.user_config?.table_props
              ?.length * 350
        }}
      />
      <div className='flex flex-row-reverse mt-4'></div>
    </div>
  );

  const renderContactDetailsModal = () => (
    <Modal
      title={null}
      visible={isModalVisible}
      className='fa-modal--full-width'
      footer={null}
      closable={null}
    >
      <ContactDetails user={activeUser} onCancel={handleCancel} />
    </Modal>
  );
  if (loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (isIntegrationEnabled || activeProject.id === demoProjectId) {
    return (
      <div className='list-container'>
        <Text type='title' level={3} weight='bold' extraClass='mt-12'>
          User Profiles
        </Text>
        {renderActions()}
        {contacts.isLoading ? (
          <Spin size='large' className='fa-page-loader' />
        ) : (
          renderTable()
        )}
        {renderContactDetailsModal()}
      </div>
    );
  }
  return <ProfileBeforeIntegration />;
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  contacts: state.timelines.contacts,
  segments: state.timelines.segments,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      createNewSegment,
      getProfileUsers,
      getProfileUserDetails,
      getSavedSegments,
      getUserProperties,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      fetchDemoProject,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
