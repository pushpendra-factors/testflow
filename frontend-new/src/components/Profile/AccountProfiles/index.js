import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Modal,
  Spin,
  Popover,
  Tabs,
  notification,
  Divider
} from 'antd';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import MomentTz from '../../MomentTz';
import AccountDetails from './AccountDetails';
import PropertyFilter from '../UserProfiles/PropertyFilter';
import { getGroupProperties } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import {
  DEFAULT_TIMELINE_CONFIG,
  displayFilterOpts,
  formatEventsFromSegment,
  formatFiltersForPayload,
  formatPayloadForFilters,
  formatSegmentsObjToGroupSelectObj,
  getHost,
  propValueFormat
} from '../utils';
import {
  getProfileAccounts,
  getProfileAccountDetails,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId
} from '../../../reducers/timelines/middleware';
import {
  fetchProjectSettings,
  udpateProjectSettings
} from '../../../reducers/global';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import SegmentModal from '../UserProfiles/SegmentModal';
import EventsBlock from '../UserProfiles/EventsBlock';

function AccountProfiles({
  activeProject,
  accounts,
  segments,
  createNewSegment,
  getSavedSegments,
  accountDetails,
  fetchProjectSettings,
  udpateProjectSettings,
  currentProjectSettings,
  getProfileAccounts,
  getProfileAccountDetails,
  getGroupProperties,
  updateSegmentForId
}) {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isGroupDDVisible, setGroupDDVisible] = useState(false);
  const [isSegmentDDVisible, setSegmentDDVisible] = useState(false);
  const [showSegmentModal, setShowSegmentModal] = useState(false);
  const [activeSegment, setActiveSegment] = useState({});
  const [accountPayload, setAccountPayload] = useState({
    source: 'All',
    filters: []
  });
  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const groupState = useSelector((state) => state.groups);
  const groupOpts = groupState?.data;
  const [activeModalKey, setActiveModalKey] = useState('');
  const [showPopOver, setShowPopOver] = useState(false);
  const [checkListAccountProps, setCheckListAccountProps] = useState([]);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);

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

  const enabledGroups = () => {
    const groups = [['All Accounts', 'All']];
    groupOpts?.forEach((elem) => {
      if (
        elem.name === '$hubspot_company' ||
        elem.name === '$salesforce_account'
      ) {
        groups.push([displayFilterOpts[elem.name], elem.name]);
      }
    });
    return groups;
  };

  useEffect(() => {
    getGroupProperties(activeProject.id, '$hubspot_company');
    getGroupProperties(activeProject.id, '$salesforce_account');
  }, [activeProject.id]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const getColumns = () => {
    const columns = [
      {
        title: <div className={headerClassStr}>Company Name</div>,
        dataIndex: 'account',
        key: 'account',
        width: 300,
        fixed: 'left',
        ellipsis: true,
        render: (item) =>
          (
            <div className='flex items-center'>
              <img
                src={`https://logo.clearbit.com/${getHost(item.host)}`}
                onError={(e) => {
                  if (
                    e.target.src !==
                    'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg'
                  ) {
                    e.target.src =
                      'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg';
                  }
                }}
                alt=''
                width='20'
                height='20'
              />
              <span className='ml-2'>{item.name}</span>
            </div>
          ) || '-'
      }
    ];
    const tableProps = accountPayload.segment_id
      ? activeSegment.query.table_props
      : currentProjectSettings?.timelines_config?.account_config?.table_props;
    tableProps?.forEach((prop) => {
      const propDisplayName = groupPropNames[prop]
        ? groupPropNames[prop]
        : PropTextFormat(prop);
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
        render: (item) => (
          <Text type='title' level={7} className='m-0' truncate>
            {propValueFormat(prop, item) || '-'}
          </Text>
        )
      });
    });
    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 250,
      fixed: 'right',
      align: 'right',
      render: (item) => MomentTz(item).fromNow()
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
    if (val !== accountPayload.source) {
      const opts = { ...accountPayload };
      opts.source = val;
      setAccountPayload(opts);
    }
    setGroupDDVisible(false);
  };

  const setFilters = (filters) => {
    const opts = { ...accountPayload };
    opts.filters = filters;
    setAccountPayload(opts);
  };

  const clearFilters = () => {
    const opts = { ...accountPayload };
    opts.filters = [];
    setAccountPayload(opts);
  };

  useEffect(() => {
    const opts = { ...accountPayload };
    opts.filters = formatFiltersForPayload(accountPayload.filters);
    getProfileAccounts(activeProject.id, opts);
  }, [activeProject, currentProjectSettings, accountPayload, segments]);

  const selectGroup = () => (
    <div className='absolute top-0'>
      {isGroupDDVisible ? (
        <FaSelect
          options={enabledGroups()}
          onClickOutside={() => setGroupDDVisible(false)}
          optionClick={(val) => onChange(val[1])}
        />
      ) : null}
    </div>
  );

  useEffect(() => {
    const listProperties = [
      ...(groupProperties.$hubspot_company
        ? groupProperties.$hubspot_company
        : []),
      ...(groupProperties.$salesforce_account
        ? groupProperties.$salesforce_account
        : [])
    ];
    const tableProps = accountPayload.segment_id
      ? activeSegment.query.table_props
      : currentProjectSettings.timelines_config?.account_config?.table_props;
    const accountPropsWithEnableKey = formatUserPropertiesToCheckList(
      listProperties,
      tableProps
    );
    setCheckListAccountProps(accountPropsWithEnableKey);
  }, [currentProjectSettings, groupProperties, accountPayload]);

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListAccountProps.filter((item) => item.enabled === true).length < 8
    ) {
      const checkListProps = [...checkListAccountProps];
      const optIndex = checkListProps.findIndex(
        (obj) => obj.prop_name === option.prop_name
      );
      checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
      setCheckListAccountProps(
        checkListProps.sort((a, b) => b.enabled - a.enabled)
      );
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum Table Properties Selection Reached.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    if (accountPayload?.segment_id?.length) {
      const query = { ...activeSegment.query };
      query.table_props = checkListAccountProps
        .filter((item) => item.enabled === true)
        .map((item) => item?.prop_name);
      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
        query: { ...query }
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query }));
    } else {
      const config = { ...tlConfig };
      config.account_config.table_props = checkListAccountProps
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
          mapArray={checkListAccountProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
          showApply
          onApply={applyTableProps}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const generateSegmentsList = () => {
    const segmentsList = [];
    if (accountPayload.source === 'All') {
      Object.entries(segments)
        .filter((segment) =>
          ['$hubspot_company', '$salesforce_account'].includes(segment[0])
        )
        .forEach(([group, vals]) => {
          const obj = formatSegmentsObjToGroupSelectObj(group, vals);
          segmentsList.push(obj);
        });
    } else {
      const obj = formatSegmentsObjToGroupSelectObj(
        accountPayload.source,
        segments[accountPayload.source]
      );
      segmentsList.push(obj);
    }
    return segmentsList;
  };

  const onOptionClick = (_, data) => {
    const opts = { ...accountPayload };
    opts.segment_id = data[1];
    setActiveSegment(data[2]);
    setAccountPayload(opts);
    setSegmentDDVisible(false);
  };

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

  const clearSegment = () => {
    const opts = { ...accountPayload };
    opts.segment_id = '';
    setActiveSegment({});
    setAccountPayload(opts);
    setSegmentDDVisible(false);
  };

  const renderAdditionalActionsInSegment = () => (
    <div className='mb-2'>
      <Divider className='divider-margin' />
      <div className='flex items-center flex-col'>
        {accountPayload.segment_id && (
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
        profileType='account'
        activeProject={activeProject}
        type={accountPayload.source}
        typeOptions={enabledGroups().filter((group) => group[1] !== 'All')}
        visible={showSegmentModal}
        segment={{}}
        onSave={handleSaveSegment}
        onCancel={() => setShowSegmentModal(false)}
        caller={'account_profiles'}
        tableProps={
          currentProjectSettings.timelines_config?.account_config?.table_props
        }
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
          allowEmpty
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
            displayMode
          />
        </div>
      );
    });

    return (
      <div className='flex items-start'>
        {blockList.length ? (
          <h2 className='whitespace-no-wrap line-height-8 m-0 mr-2'>
            Performed Events
          </h2>
        ) : null}
        <div className='flex flex-wrap flex-col'>{blockList}</div>
      </div>
    );
  };

  const filtersList = (filters) => {
    return (
      <div className='flex items-start'>
        <h2 className='whitespace-no-wrap line-height-8 m-0 mr-2'>
          Properties
        </h2>
        <div className='flex flex-wrap flex-col'>
          <PropertyFilter
            filtersLimit={10}
            profileType='user'
            source={accountPayload.source}
            filters={filters}
            displayMode
          ></PropertyFilter>
        </div>
      </div>
    );
  };

  const segmentInfo = () => {
    if (activeSegment.query) {
      return (
        <div className='p-3'>
          {activeSegment.query.ewp && activeSegment.query.ewp.length
            ? eventsList(formatEventsFromSegment(activeSegment.query.ewp))
            : null}
          {activeSegment.query.gup && activeSegment.query.gup.length
            ? filtersList(formatPayloadForFilters(activeSegment.query.gup))
            : null}
        </div>
      );
    }
    return null;
  };

  const renderGroupSelectDD = () => (
    <div className='relative mr-2'>
      <Button
        className='dropdown-btn'
        type='text'
        icon={<SVG name='user_friends' size={16} />}
        onClick={() => setGroupDDVisible(!isGroupDDVisible)}
      >
        {displayFilterOpts[accountPayload.source]}
        <SVG name='caretDown' size={16} />
      </Button>
      {selectGroup()}
    </div>
  );

  const renderSegmentSelect = () => (
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
  );

  const renderPropertyFilter = () => (
    <div key={0} className='max-w-3xl'>
      <PropertyFilter
        profileType='account'
        source={accountPayload.source}
        filters={accountPayload.filters}
        setFilters={setFilters}
      />
    </div>
  );

  const renderClearFilterButton = () => (
    <Button
      className='dropdown-btn'
      type='text'
      icon={<SVG name='times_circle' size={16} />}
      onClick={clearFilters}
    >
      Clear Filters
    </Button>
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
      <Button size='large' className='fa-btn--custom mx-2 relative'>
        <SVG name='activity_filter' />
      </Button>
    </Popover>
  );

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>
        {renderGroupSelectDD()}
        {renderSegmentSelect()}
        {renderPropertyFilter()}
      </div>
      <div className='flex items-center justify-between'>
        {accountPayload.filters.length ? renderClearFilterButton() : null}
        {renderTablePropsSelect()}
      </div>
    </div>
  );

  const renderTable = () => (
    <div>
      <Table
        onRow={(account) => ({
          onClick: () => {
            getProfileAccountDetails(
              activeProject.id,
              account.identity,
              currentProjectSettings?.timelines_config
            );
            setActiveModalKey(account.identity);
            showModal();
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(accounts.data)}
        columns={getColumns()}
        rowClassName='cursor-pointer'
        pagination={{ position: ['bottom', 'left'], defaultPageSize: '25' }}
        scroll={{
          x:
            currentProjectSettings?.timelines_config?.account_config
              ?.table_props?.length * 300
        }}
        footer={() => (
          <div className='text-right'>
            <a
              className='font-size--small'
              href='https://clearbit.com'
              target='_blank'
            >
              Logos provided by Clearbit
            </a>
          </div>
        )}
      />
    </div>
  );

  const renderAccountDetailsModal = () => (
    <Modal
      title={null}
      visible={isModalVisible}
      className='fa-modal--full-width'
      footer={null}
      closable={null}
    >
      <AccountDetails
        accountId={activeModalKey}
        onCancel={handleCancel}
        accountDetails={accountDetails}
      />
    </Modal>
  );

  return (
    <div className='list-container'>
      <Text type='title' level={3} weight='bold' extraClass='mt-12'>
        Account Profiles
      </Text>
      {renderActions()}
      {accounts.isLoading ? (
        <Spin size='large' className='fa-page-loader' />
      ) : accounts.data.length ? (
        renderTable()
      ) : (
        <div className='ant-empty ant-empty-normal'>
          <div className='ant-empty-image'>
            <SVG name='nodata' size={150} />
          </div>
          <div className='ant-empty-description'>No Accounts Found</div>
        </div>
      )}
      {renderAccountDetailsModal()}
    </div>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  accounts: state.timelines.accounts,
  segments: state.timelines.segments,
  accountDetails: state.timelines.accountDetails,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccounts,
      getProfileAccountDetails,
      createNewSegment,
      getSavedSegments,
      getGroupProperties,
      fetchProjectSettings,
      udpateProjectSettings,
      updateSegmentForId
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
