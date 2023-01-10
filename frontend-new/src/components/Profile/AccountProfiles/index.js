import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Spin, Popover, Tabs, notification } from 'antd';
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
  formatFiltersForPayload,
  getHost
} from '../utils';
import {
  getProfileAccounts,
  getProfileAccountDetails
} from '../../../reducers/timelines/middleware';
import {
  fetchProjectSettings,
  udpateProjectSettings
} from '../../../reducers/global';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { PropTextFormat } from 'Utils/dataFormatter';

function AccountProfiles({
  activeProject,
  accounts,
  accountDetails,
  fetchProjectSettings,
  udpateProjectSettings,
  currentProjectSettings,
  getProfileAccounts,
  getProfileAccountDetails,
  getGroupProperties
}) {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isDDVisible, setDDVisible] = useState(false);
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

  const displayFilterOpts = {
    All: 'All Accounts',
    $hubspot_company: 'Hubspot Companies',
    $salesforce_account: 'Salesforce Accounts'
  };

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
    getProfileAccounts(activeProject.id, opts);
  }, [activeProject, filterPayload]);

  const selectUsers = () => (
    <div className='absolute top-0'>
      {isDDVisible ? (
        <FaSelect
          options={enabledGroups()}
          onClickOutside={() => setDDVisible(false)}
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
        tab={<span className='fa-activity-filter--tabname'>Properties</span>}
        key='properties'
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

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex items-start'>
        <div className='relative mr-2'>
          <Button
            className='fa-dd--custom-btn'
            type='text'
            icon={<SVG name='user_friends' size={16} />}
            onClick={() => setDDVisible(!isDDVisible)}
          >
            {displayFilterOpts[filterPayload.source]}
            <SVG name='caretDown' size={16} />
          </Button>
          {selectUsers()}
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
  accountDetails: state.timelines.accountDetails,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccounts,
      getProfileAccountDetails,
      getGroupProperties,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
