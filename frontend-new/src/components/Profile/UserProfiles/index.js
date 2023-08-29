import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Table,
  Button,
  Spin,
  notification,
  Popover,
  Tabs,
  Avatar,
  Input
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import FaSelect from '../../FaSelect';
import { getUserPropertiesV2 } from '../../../reducers/coreQuery/middleware';
import PropertyFilter from '../MyComponents/PropertyFilter';
import MomentTz from '../../MomentTz';
import {
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  udpateProjectSettings
} from '../../../reducers/global';
import {
  ALPHANUMSTR,
  DEFAULT_TIMELINE_CONFIG,
  EngagementTag,
  formatFiltersForPayload,
  getPropType,
  iconColors,
  propValueFormat,
  sortStringColumn,
  sortNumericalColumn,
  formatReqPayload
} from '../utils';
import {
  getProfileUsers,
  getProfileUserDetails,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId
} from '../../../reducers/timelines/middleware';
import _ from 'lodash';
import SegmentModal from './SegmentModal';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
import ProfilesWrapper from '../ProfilesWrapper';
import { getUserOptions } from './userProfiles.helpers';
import {
  selectActiveSegment,
  selectSegmentModalState,
  selectTimelinePayload
} from 'Reducers/userProfilesView/selectors';
import {
  setTimelinePayloadAction,
  setActiveSegmentAction,
  setSegmentModalStateAction
} from 'Reducers/userProfilesView/actions';
import { useHistory } from 'react-router-dom';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import UpgradeModal from '../UpgradeModal';
import RangeNudge from 'Components/GenericComponents/RangeNudge';
import { showUpgradeNudge } from 'Views/Settings/ProjectSettings/Pricing/utils';
import CommonBeforeIntegrationPage from 'Components/GenericComponents/CommonBeforeIntegrationPage';

const userOptions = getUserOptions();

function UserProfiles({
  activeProject,
  contacts,
  segments,
  createNewSegment,
  getSavedSegments,
  getProfileUsers,
  getProfileUserDetails,
  getUserPropertiesV2,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  currentProjectSettings,
  udpateProjectSettings,
  updateSegmentForId
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const userPropertiesV2 = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const { userPropNames } = useSelector((state) => state.coreQuery);
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const activeSegment = useSelector((state) => selectActiveSegment(state));
  const showSegmentModal = useSelector((state) =>
    selectSegmentModalState(state)
  );
  const { sixSignalInfo } = useSelector((state) => state.featureConfig);

  const [listSearchItems, setListSearchItems] = useState([]);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchDDOpen, setSearchDDOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const [userValueOpts, setUserValueOpts] = useState({});
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;
  const { isFeatureLocked: isEngagementLocked } = useFeatureLock(
    FEATURES.FEATURE_ENGAGEMENT
  );

  useEffect(() => {
    if (!timelinePayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues =
        timelinePayload?.search_filter?.map((vl) => vl?.va) || [];
      setListSearchItems(_.uniq(listValues));
      setSearchBarOpen(true);
    }
  }, [timelinePayload?.search_filter]);

  const setTimelinePayload = useCallback(
    (payload) => {
      dispatch(setTimelinePayloadAction(payload));
    },
    [dispatch]
  );

  const setActiveSegment = useCallback(
    (segmentPayload) => {
      dispatch(setActiveSegmentAction(segmentPayload));
    },
    [dispatch]
  );

  const setShowSegmentModal = useCallback(
    (value) => {
      dispatch(setSegmentModalStateAction(value));
    },
    [dispatch]
  );

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
  }, [currentProjectSettings?.timelines_config]);

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
    getUserPropertiesV2(activeProject.id);
  }, [activeProject?.id]);

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
    const tableProps = timelinePayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings.timelines_config?.user_config?.table_props;
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userPropertiesModified,
      tableProps
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [
    currentProjectSettings,
    userPropertiesV2,
    activeSegment,
    timelinePayload
  ]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject.id]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const { tableProperties, tableColumns } = useMemo(() => {
    const columns = [
      {
        title: <div className={headerClassStr}>Identity</div>,
        width: 280,
        dataIndex: 'identity',
        key: 'identity',
        fixed: 'left',
        ellipsis: true,
        sorter: (a, b) => sortStringColumn(a.identity.id, b.identity.id),
        render: (identity) => (
          <div className='flex items-center'>
            {identity.isAnonymous ? (
              <SVG
                name={`TrackedUser${identity.id?.match(/\d/)?.[0] || 0}`}
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
    // Engagement Column
    const engagementExists = contacts.data?.find(
      (item) =>
        item.engagement &&
        (item.engagement !== undefined || item.engagement !== '')
    );
    if (engagementExists && !isEngagementLocked) {
      columns.push({
        title: <div className={headerClassStr}>Engagement</div>,
        width: 150,
        dataIndex: 'engagement',
        key: 'engagement',
        fixed: 'left',
        defaultSortOrder: 'descend',
        sorter: {
          compare: (a, b) => sortNumericalColumn(a.score, b.score),
          multiple: 1
        },
        render: (status) =>
          status ? (
            <div
              className='engagement-tag'
              style={{ '--bg-color': EngagementTag[status]?.bgColor }}
            >
              <img
                src={`../../../assets/icons/${EngagementTag[status]?.icon}.svg`}
                alt=''
              />
              <Text type='title' level={7} extraClass='m-0'>
                {status}
              </Text>
            </div>
          ) : (
            '-'
          )
      });
    }

    const tableProps = timelinePayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings?.timelines_config?.user_config?.table_props;

    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    tableProps
      ?.filter((entry) => entry !== '' && entry !== undefined)
      ?.forEach((prop) => {
        const propDisplayName = userPropNames[prop]
          ? userPropNames[prop]
          : PropTextFormat(prop);
        const propType = getPropType(userPropertiesModified, prop);
        columns.push({
          title: (
            <Text
              type='title'
              level={7}
              color='grey-2'
              weight='bold'
              extraClass='m-0'
              truncate
              charLimit={25}
            >
              {propDisplayName}
            </Text>
          ),
          dataIndex: prop,
          key: prop,
          width: 260,
          sorter: (a, b) =>
            propType === 'numerical'
              ? sortNumericalColumn(a[prop], b[prop])
              : sortStringColumn(a[prop], b[prop]),
          render: (value) => (
            <Text type='title' level={7} extraClass='m-0' truncate>
              {value ? propValueFormat(prop, value, propType) : '-'}
            </Text>
          )
        });
      });

    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'lastActivity',
      key: 'lastActivity',
      width: 200,
      align: 'right',
      sorter: {
        compare: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
        multiple: 2
      },
      render: (item) => MomentTz(item).fromNow()
    });
    return { tableProperties: tableProps, tableColumns: columns };
  }, [contacts?.data, currentProjectSettings, timelinePayload, activeSegment]);

  const getTableData = (data) => {
    const sortedData = data.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData.map((row) => ({
      ...row,
      ...row?.tableProps
    }));
  };

  const setFilters = (filters) => {
    const opts = { ...timelinePayload };
    opts.filters = filters;
    setTimelinePayload(opts);
    setActiveSegment(activeSegment);
    getUsers(opts);
  };

  const clearFilters = () => {
    const opts = { ...timelinePayload };
    opts.filters = [];
    setTimelinePayload(opts);
    setActiveSegment(activeSegment);
    getUsers(opts);
  };

  const getUsers = (payload) => {
    if (payload.source && payload.source !== '') {
      const formatPayload = { ...payload };
      formatPayload.filters = formatFiltersForPayload(payload?.filters) || [];
      const reqPayload = formatReqPayload(formatPayload, activeSegment);
      getProfileUsers(activeProject.id, reqPayload, activeAgent);
    }
  };

  useEffect(() => {
    getUsers(timelinePayload);
  }, [timelinePayload.source, timelinePayload.segment_id]);

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
          // setSegmentDDVisible(false);
        }
      })
      .then(() => getSavedSegments(activeProject.id))
      .catch((err) => {
        notification.error({
          message: 'Error',
          description:
            err?.data?.error || 'Segment Creation Failed. Invalid Parameters.',
          duration: 3
        });
      });
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
        description: 'Maximum of 8 Table Properties Selection Allowed.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    if (timelinePayload?.segment_id?.length) {
      const updatedQuery = { ...activeSegment.query };
      updatedQuery.table_props =
        checkListUserProps
          ?.filter((item) => item.enabled === true)
          ?.map((item) => item?.prop_name)
          ?.filter((entry) => entry !== '' && entry !== undefined) || [];
      updateSegmentForId(activeProject.id, timelinePayload.segment_id, {
        query: { ...updatedQuery }
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query: updatedQuery }))
        .finally(() => getUsers(timelinePayload));
    } else {
      const config = { ...tlConfig };
      config.user_config.table_props = checkListUserProps
        ?.filter((item) => item.enabled === true)
        ?.map((item) => item?.prop_name);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...config }
      }).then(() => getUsers(timelinePayload));
    }
    setShowPopOver(false);
  };

  const handleDisableOptionClick = () => {
    setIsUpgradeModalVisible(true);
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
          showDisabledOption={isEngagementLocked}
          // disabledOptions={['Engagement', 'Engaged Channels']}
          handleDisableOptionClick={handleDisableOptionClick}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const renderPropertyFilter = () => (
    <div key={0} className='max-w-3xl'>
      <PropertyFilter
        profileType='user'
        source={timelinePayload.source}
        filters={timelinePayload.filters}
        setFilters={setFilters}
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
      props: ['', '$user_id', 'categorical', 'user'],
      operator: ['contains'],
      values: []
    };
    const payload = { ...timelinePayload };
    searchFilter.values.push(...val.map((vl) => JSON.parse(vl)[0]));
    payload.search_filter = formatFiltersForPayload([searchFilter]);
    setListSearchItems(searchFilter.values);
    setTimelinePayload(payload);
    setActiveSegment(activeSegment);
    getUsers(payload);
  };

  const searchUsers = () => (
    <div className='absolute top-0'>
      {searchDDOpen ? (
        <FaSelect
          multiSelect
          options={
            userValueOpts
              ? Object.keys(userValueOpts).map((value) => [value])
              : []
          }
          displayNames={userValueOpts}
          applClick={(val) => onApplyClick(val)}
          onClickOutside={() => setSearchDDOpen(false)}
          selectedOpts={listSearchItems}
          allowSearch
          placeholder='Search Users'
          style={{
            top: '-8px',
            right: 0,
            padding: '8px 8px 12px',
            overflowX: 'hidden'
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
      setActiveSegment(activeSegment);
      getUsers(payload);
    }
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
    setSearchDDOpen(true);
  };

  const renderSearchSection = () => (
    <div className='relative'>
      {searchBarOpen ? (
        <div className={'flex items-center justify-between'}>
          {!searchDDOpen && (
            <Input
              size='large'
              value={listSearchItems ? listSearchItems.join(', ') : null}
              placeholder={'Search Users'}
              style={{ width: '240px', 'border-radius': '5px' }}
              prefix={<SVG name='search' size={16} color={'grey'} />}
              onClick={() => setSearchDDOpen(true)}
            />
          )}
          <Button type='text' className='search-btn' onClick={onSearchClose}>
            <SVG name={'close'} size={20} color={'grey'} />
          </Button>
        </div>
      ) : (
        <Button type='text' className='search-btn' onClick={onSearchOpen}>
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
        Edit Columns
      </Button>
    </Popover>
  );

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>{renderPropertyFilter()}</div>
      <div className='inline-flex gap--6'>
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
            history.push(
              `/profiles/people/${btoa(user.identity.id)}?is_anonymous=${
                user.identity.isAnonymous
              }`
            );
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(contacts.data)}
        columns={tableColumns}
        rowClassName='cursor-pointer'
        pagination={{ position: ['bottom', 'left'], defaultPageSize: '25' }}
        scroll={{
          x: tableProperties?.length * 250
        }}
      />
      <div className='flex flex-row-reverse mt-4'></div>
    </div>
  );

  if (loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (isIntegrationEnabled) {
    return (
      <ProfilesWrapper>
        {showUpgradeNudge(
          sixSignalInfo?.usage || 0,
          sixSignalInfo?.limit || 0,
          currentProjectSettings
        ) && (
          <div className='mb-4'>
            <RangeNudge
              title='Accounts Identified'
              amountUsed={sixSignalInfo?.usage || 0}
              totalLimit={sixSignalInfo?.limit || 0}
            />
          </div>
        )}

        <Text type='title' level={3} weight='bold' extraClass='mb-0'>
          User Profiles
        </Text>
        {renderActions()}
        {contacts.isLoading ? (
          <Spin size='large' className='fa-page-loader' />
        ) : (
          renderTable()
        )}
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
        <UpgradeModal
          visible={isUpgradeModalVisible}
          variant='account'
          onCancel={() => setIsUpgradeModalVisible(false)}
        />
      </ProfilesWrapper>
    );
  }
  return <CommonBeforeIntegrationPage />;
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
      getUserPropertiesV2,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      udpateProjectSettings,
      updateSegmentForId
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
