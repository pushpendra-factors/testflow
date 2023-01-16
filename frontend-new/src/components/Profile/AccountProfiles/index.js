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
  formatFiltersForPayload,
  formatPayloadForFilters,
  formatSegmentsObjToGroupSelectObj,
  getHost
} from '../utils';
import {
  getProfileAccounts,
  getProfileAccountDetails,
  createNewSegment,
  getSavedSegments
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
  getGroupProperties
}) {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isAccountDDVisible, setAccountDDVisible] = useState(false);
  const [isSegmentDDVisible, setSegmentDDVisible] = useState(false);
  const [showSegmentModal, setShowSegmentModal] = useState(false);
  const [activeSegment, setActiveSegment] = useState({});
  const [filterPayload, setFilterPayload] = useState({
    source: 'All',
    filters: []
  });
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
    currentProjectSettings?.timelines_config?.account_config?.table_props?.forEach(
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
    if (val !== filterPayload.source) {
      const opts = { ...filterPayload };
      opts.source = val;
      setFilterPayload(opts);
    }
    setAccountDDVisible(false);
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
    getProfileAccounts(activeProject.id, opts);
  }, [activeProject, filterPayload]);

  const selectUsers = () => (
    <div className='absolute top-0'>
      {isAccountDDVisible ? (
        <FaSelect
          options={enabledGroups()}
          onClickOutside={() => setAccountDDVisible(false)}
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

    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      listProperties,
      currentProjectSettings.timelines_config?.account_config?.table_props
    );
    setCheckListAccountProps(userPropsWithEnableKey);
  }, [currentProjectSettings, groupProperties]);

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
      setCheckListAccountProps(checkListProps);
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
    config.account_config.table_props = checkListAccountProps
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
    if (filterPayload.source === 'All') {
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
        filterPayload.source,
        segments[filterPayload.source]
      );
      segmentsList.push(obj);
    }
    return segmentsList;
  };

  const onOptionClick = (_, data) => {
    const opts = { ...filterPayload };
    opts.segment_id = data[1];
    setActiveSegment(data[2]);
    setFilterPayload(opts);
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
    const opts = { ...filterPayload };
    opts.segment_id = '';
    setActiveSegment({});
    setFilterPayload(opts);
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
                {filterPayload.segment_id && (
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
                profileType='account'
                type={filterPayload.source}
                typeOptions={enabledGroups().filter(
                  (group) => group[1] !== 'All'
                )}
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
              profileType='account'
              source={filterPayload.source}
              filters={filters}
            ></PropertyFilter>
          </div>
        );
      }
    }
    return cardContent;
  };

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>
        <div className='relative mr-2'>
          <Button
            className='dropdown-btn'
            type='text'
            icon={<SVG name='user_friends' size={16} />}
            onClick={() => setAccountDDVisible(!isAccountDDVisible)}
          >
            {displayFilterOpts[filterPayload.source]}
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
            profileType='account'
            source={filterPayload.source}
            filters={filterPayload.filters}
            setFilters={setFilters}
          />
        </div>
      </div>
      <div className='flex items-center justify-between'>
        {filterPayload.filters.length ? (
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
        className='fa-table--basic'
        dataSource={getTableData(accounts.data)}
        columns={getColumns()}
        rowClassName='cursor-pointer'
        pagination={{ position: ['bottom', 'left'] }}
        scroll={{
          x:
            currentProjectSettings?.timelines_config?.account_config
              ?.table_props?.length * 350
        }}
        footer={() => (
          <div className='text-right'>
            <a className='font-size--small' href='https://clearbit.com'>
              Logos provided by Clearbit
            </a>
          </div>
        )}
      />
      <div className='flex flex-row-reverse mt-4'></div>
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
        <Text type='title' level={6} extraClass='mt-20 italic'>
          There are currently no Accounts available for this project.
        </Text>
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
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
