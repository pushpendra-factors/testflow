import React, { useEffect, useMemo, useState } from 'react';
import {
  Table,
  Button,
  Spin,
  Divider,
  notification,
  Popover,
  Tabs,
  Avatar,
  Input
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
import PropertyFilter from '../MyComponents/PropertyFilter';
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
  ALPHANUMSTR,
  DEFAULT_TIMELINE_CONFIG,
  formatEventsFromSegment,
  formatFiltersForPayload,
  formatPayloadForFilters,
  formatSegmentsObjToGroupSelectObj,
  getPropType,
  iconColors,
  propValueFormat
} from '../utils';
import {
  getProfileUsers,
  getProfileUserDetails,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId
} from '../../../reducers/timelines/middleware';
import _ from 'lodash';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import SegmentModal from './SegmentModal';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import EventsBlock from '../MyComponents/EventsBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';

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
  udpateProjectSettings,
  updateSegmentForId
}) {
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const { userPropNames } = useSelector((state) => state.coreQuery);

  const [listSearchItems, setListSearchItems] = useState([]);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchDDOpen, setSearchDDOpen] = useState(false);
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
  const [userValueOpts, setUserValueOpts] = useState({});

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
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    if (_.isEmpty(dashboards?.data)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
  }, [activeProject]);

  useEffect(() => {
    getUserProperties(activeProject.id);
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
    const tableProps = timelinePayload.segment_id
      ? activeSegment.query.table_props
      : currentProjectSettings.timelines_config?.user_config?.table_props;
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userProperties,
      tableProps
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [currentProjectSettings, userProperties, activeSegment, timelinePayload]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject]);

  const userOptions = [...profileOptions.users].map((item) => [
    item,
    ProfileMapper[item]
  ]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const getColumns = () => {
    const columns = [
      {
        title: <div className={headerClassStr}>Identity</div>,
        width: 280,
        dataIndex: 'identity',
        key: 'identity',
        fixed: 'left',
        ellipsis: true,
        render: (identity) => (
          <div className='flex items-center'>
            {identity.isAnonymous ? (
              <SVG
                name={`TrackedUser${identity.id.match(/\d/g)[0]}`}
                size={24}
              />
            ) : (
              <Avatar
                size={24}
                className='userlist-avatar'
                style={{
                  backgroundColor: `${
                    iconColors[
                      ALPHANUMSTR.indexOf(identity.id.charAt(0).toUpperCase()) %
                        8
                    ]
                  }`,
                  fontSize: '16px'
                }}
              >
                {identity.id.charAt(0).toUpperCase()}
              </Avatar>
            )}
            <span className='ml-2 truncate'>
              {identity.isAnonymous ? 'New User' : identity.id}
            </span>
          </div>
        )
      }
    ];
    const tableProps = timelinePayload.segment_id
      ? activeSegment.query.table_props
      : currentProjectSettings?.timelines_config?.user_config?.table_props;
    tableProps?.forEach((prop) => {
      const propDisplayName = userPropNames[prop]
        ? userPropNames[prop]
        : PropTextFormat(prop);
      const propType = getPropType(userProperties, prop);
      columns.push({
        title: (
          <Text
            type='title'
            level={7}
            color='grey-2'
            weight='bold'
            className='m-0'
            truncate
          >
            {propDisplayName}
          </Text>
        ),
        dataIndex: prop,
        key: prop,
        width: 300,
        render: (value) => (
          <Text type='title' level={7} className='m-0' truncate>
            {value ? propValueFormat(prop, value, propType) : '-'}
          </Text>
        )
      });
    });
    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 250,
      align: 'right',
      render: (item) => MomentTz(item).fromNow()
    });
    return columns;
  };

  const getTableData = (data) => {
    const tableData = data?.map((row) => {
      return {
        ...row,
        ...row?.table_props
      };
    });
    return tableData.sort(
      (a, b) =>
        parseInt((new Date(b.last_activity).getTime() / 1000).toFixed(0)) -
        parseInt((new Date(a.last_activity).getTime() / 1000).toFixed(0))
    );
  };

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };

  const onChange = (val) => {
    if (val[1] !== timelinePayload.source) {
      const opts = { ...timelinePayload };
      opts.source = val[1];
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
  }, [activeProject.id, timelinePayload, currentProjectSettings, segments]);

  const handleSaveSegment = (segmentPayload) => {
    createNewSegment(activeProject.id, segmentPayload)
      .then((response) => {
        if (response.type === 'SEGMENT_CREATION_FULFILLED') {
          notification.success({
            message: 'Success!',
            description: response?.payload?.message,
            duration: 3
          });
          setShowSegmentModal(false);
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
          options={[['All Users', 'All'], ...userOptions]}
          onClickOutside={() => setUserDDVisible(false)}
          optionClick={(val) => onChange(val)}
        />
      ) : null}
    </div>
  );

  const generateSegmentsList = () => {
    const segmentsList = [];
    if (timelinePayload.source === 'All') {
      Object.entries(segments)
        .filter((segment) =>
          Object.keys(ReverseProfileMapper).includes(segment[0])
        )
        .forEach(([group, vals]) => {
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

  const renderAdditionalActionsInSegment = () => (
    <div className='mb-2'>
      <Divider className='divider-margin' />
      <div className='flex items-center flex-col'>
        {timelinePayload.segment_id && (
          <Button
            size='large'
            type='text'
            className='w-full mb-2'
            onClick={clearSegment}
            icon={<SVG name='remove' />}
          >
            Clear Segment
          </Button>
        )}
        <Button
          type='link'
          size='large'
          className='w-full'
          icon={<SVG name='plus' color='purple' />}
          onClick={() => setShowSegmentModal(true)}
        >
          Add New Segment
        </Button>
      </div>
      <SegmentModal
        profileType='user'
        activeProject={activeProject}
        type={timelinePayload.source}
        typeOptions={userOptions}
        tableProps={
          currentProjectSettings.timelines_config?.user_config?.table_props
        }
        visible={showSegmentModal}
        segment={{}}
        onSave={handleSaveSegment}
        onCancel={() => setShowSegmentModal(false)}
        caller={'user_profiles'}
      />
    </div>
  );

  const selectSegment = () => (
    <div className='absolute top-8'>
      {isSegmentDDVisible ? (
        <GroupSelect2
          groupedProperties={generateSegmentsList()}
          placeholder='Search Segments'
          optionClick={onOptionClick}
          onClickOutside={() => setSegmentDDVisible(false)}
          additionalActions={renderAdditionalActionsInSegment()}
        />
      ) : null}
    </div>
  );

  const eventsList = (listEvents) => {
    const blockList = [];
    listEvents.forEach((event, index) => {
      blockList.push(
        <div key={index} className='m-0 mr-2 mb-2'>
          <EventsBlock
            index={index + 1}
            event={event}
            queries={listEvents}
            groupAnalysis={activeSegment?.query?.grpa}
            viewMode
          />
        </div>
      );
    });

    return (
      <div className='segment-query_block'>
        {blockList.length ? (
          <h2
            className={`title ${
              activeSegment?.query?.gup?.length ? '' : 'width-unset'
            }`}
          >
            Performed Events
          </h2>
        ) : null}
        <div className='content'>{blockList}</div>
      </div>
    );
  };

  const filtersList = (filters) => {
    return (
      <div className='segment-query_block'>
        <h2
          className={`title ${
            activeSegment?.query?.ewp?.length ? '' : 'width-unset'
          }`}
        >
          With Properties
        </h2>
        <div className='content'>
          <PropertyFilter
            filtersLimit={10}
            profileType='user'
            source={timelinePayload.source}
            filters={filters}
            viewMode
          ></PropertyFilter>
        </div>
      </div>
    );
  };

  const segmentInfo = () => {
    if (!activeSegment.query) return null;
    return (
      <div className='p-3'>
        {activeSegment.query.ewp && activeSegment.query.ewp.length
          ? eventsList(formatEventsFromSegment(activeSegment.query.ewp))
          : null}
        {activeSegment.query.gup && activeSegment.query.gup.length
          ? filtersList(formatPayloadForFilters(activeSegment.query.gup))
          : null}
        <h2 className='whitespace-no-wrap italic line-height-8 m-0 mr-2'>
          {`*Shows ${
            ReverseProfileMapper[activeSegment.type]?.users
          } from last 28 days.`}
        </h2>
      </div>
    );
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
      setCheckListUserProps(
        checkListProps.sort((a, b) => b.enabled - a.enabled)
      );
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum of 8 Table Properties Selection Allowed.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    if (timelinePayload?.segment_id?.length) {
      const query = { ...activeSegment.query };
      query.table_props = checkListUserProps
        .filter((item) => item.enabled === true)
        .map((item) => item?.prop_name);
      updateSegmentForId(activeProject.id, timelinePayload.segment_id, {
        query: { ...query }
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query }));
    } else {
      const config = { ...tlConfig };
      config.user_config.table_props = checkListUserProps
        .filter((item) => item.enabled === true)
        .map((item) => item?.prop_name);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...config }
      });
    }
    setShowPopOver(false);
  };

  const popoverContent = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <Tabs.TabPane
        tab={
          <span className='fa-activity-filter--tabname'>Table Properties</span>
        }
        key='props'
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

  const renderUserSelectDD = () => (
    <div className='relative mr-2'>
      <Button
        className='dropdown-btn'
        type='text'
        icon={<SVG name='user_friends' size={16} />}
        onClick={() => setUserDDVisible(!isUserDDVisible)}
      >
        {userOptions?.find(
          (item) => item[1] === timelinePayload?.source
        )?.[0] || 'All Users'}
        <SVG name='caretDown' size={16} />
      </Button>
      {selectUsers()}
    </div>
  );

  const renderSegmentSelect = () => (
    <div className='relative mr-2'>
      <Popover
        overlayClassName='fa-custom-popover'
        placement='bottomLeft'
        trigger={activeSegment.query ? 'hover' : ''}
        content={segmentInfo}
        mouseEnterDelay={0.5}
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
  );

  const renderPropertyFilter = () => (
    <div key={0} className='max-w-3xl'>
      <PropertyFilter
        profileType='user'
        source={timelinePayload.source}
        filters={timelinePayload.filters}
        setFilters={setFilters}
        onFiltersLoad={[() => getUserProperties(activeProject.id)]}
      />
    </div>
  );

  const renderClearFilterButton = () => (
    <Button
      className='dropdown-btn mr-2'
      type='text'
      icon={<SVG name='times_circle' size={16} />}
      onClick={clearFilters}
    >
      Clear Filters
    </Button>
  );

  useEffect(() => {
    fetchUserPropertyValues(activeProject.id, '$user_id')
      .then((res) => {
        setUserValueOpts({ ...res.data });
      })
      .catch((err) => {
        console.log(err);
        setUserValueOpts({});
      });
  }, [activeProject.id]);

  const onApplyClick = (val) => {
    const searchFilter = {
      props: ['$user_id', 'categorical', 'user'],
      operator: ['contains'],
      values: []
    };
    const payload = { ...timelinePayload };
    searchFilter.values.push(...val.map((vl) => JSON.parse(vl)[0]));
    payload.search_filter = formatFiltersForPayload([searchFilter]);
    setListSearchItems(searchFilter.values);
    setTimelinePayload(payload);
  };

  const searchUsers = () => (
    <div className='absolute top-0'>
      {searchDDOpen ? (
        <FaSelect
          multiSelect
          options={userValueOpts ? Object.entries(userValueOpts) : []}
          displayNames={userValueOpts}
          applClick={(val) => onApplyClick(val)}
          onClickOutside={() => setSearchDDOpen(false)}
          selectedOpts={listSearchItems}
          allowSearch
          placeholder='Search Users'
          style={{
            top: '-2px',
            left: '-60px',
            padding: 0
          }}
          posRight
        />
      ) : null}
    </div>
  );

  const onSearchClose = () => {
    setSearchBarOpen(false);
    setSearchDDOpen(false);
    if (timelinePayload?.search_filter?.length) {
      const payload = { ...timelinePayload };
      payload.search_filter = [];
      setListSearchItems([]);
      setTimelinePayload(payload);
    }
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
    setSearchDDOpen(true);
  };

  const renderSearchSection = () => (
    <div className='relative mr-2'>
      {searchBarOpen ? (
        <div className={'flex items-center justify-between'}>
          <Input
            size='large'
            value={listSearchItems ? listSearchItems.join(', ') : null}
            placeholder={'Search Users'}
            style={{ width: '240px', 'border-radius': '5px' }}
            prefix={<SVG name='search' size={16} color={'grey'} />}
            onClick={() => setSearchDDOpen(true)}
          />
          <Button className='search-btn' onClick={onSearchClose}>
            <SVG name={'close'} size={20} color={'grey'} />
          </Button>
        </div>
      ) : (
        <Button className='search-btn' onClick={onSearchOpen}>
          <SVG name={'search'} size={20} color={'grey'} />
        </Button>
      )}
      {searchUsers()}
    </div>
  );

  const renderTablePropsSelect = () => (
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
        icon={<SVG name='activity_filter' />}
        className='relative'
      >
        Configure
      </Button>
    </Popover>
  );

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>
        {renderUserSelectDD()}
        {renderSegmentSelect()}
        {renderPropertyFilter()}
      </div>
      <div className='flex items-center justify-between'>
        {timelinePayload.filters.length ? renderClearFilterButton() : null}
        {renderSearchSection()}
        {renderTablePropsSelect()}
      </div>
    </div>
  );

  const renderTable = () => (
    <div>
      <Table
        size='large'
        onRow={(user) => ({
          onClick: () => {
            getProfileUserDetails(
              activeProject.id,
              user.identity.id,
              user.identity.isAnonymous,
              currentProjectSettings.timelines_config
            );
            setActiveUser(user);
            showModal();
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(contacts.data)}
        columns={getColumns()}
        rowClassName='cursor-pointer'
        pagination={{ position: ['bottom', 'left'], defaultPageSize: '25' }}
        scroll={{
          x:
            currentProjectSettings?.timelines_config?.user_config?.table_props
              ?.length * 300
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
      udpateProjectSettings,
      updateSegmentForId
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
