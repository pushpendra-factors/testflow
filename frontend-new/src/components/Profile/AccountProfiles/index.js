import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Spin } from 'antd';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import MomentTz from '../../MomentTz';
import AccountDetails from './AccountDetails';
import PropertyFilter from '../UserProfiles/PropertyFilter';
import { getGroupProperties } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import { formatFiltersForPayload, getHost } from '../utils';
import {
  getProfileAccounts,
  getProfileAccountDetails
} from '../../../reducers/timelines/middleware';

function AccountProfiles({
  activeProject,
  accounts,
  accountDetails,
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
  const groupState = useSelector((state) => state.groups);
  const groupOpts = groupState?.data;

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

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';
  const columns = [
    {
      title: <div className={headerClassStr}>Company Name</div>,
      dataIndex: 'account',
      key: 'account',
      render: (item) =>
        (
          <div className="flex items-center">
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
              alt=""
              width="20"
              height="20"
            />
            <span className="ml-2">{item.name}</span>
          </div>
        ) || '-'
    },
    {
      title: <div className={headerClassStr}>Region</div>,
      dataIndex: 'country',
      key: 'country',
      render: (item) => item || '-'
    },
    {
      title: <div className={headerClassStr}>Associated Contacts</div>,
      dataIndex: 'associated_contacts',
      key: 'associated_contacts',
      render: (item) => item || '-'
    },
    {
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'last_activity',
      key: 'last_activity',
      width: 300,
      render: (item) => MomentTz(item).format('DD MMMM YYYY, hh:mm:ss')
    }
  ];
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
    <div className="absolute top-0">
      {isDDVisible ? (
        <FaSelect
          options={enabledGroups()}
          onClickOutside={() => setDDVisible(false)}
          optionClick={(val) => onChange(val[1])}
        />
      ) : null}
    </div>
  );

  return (
    <div className="fa-container mt-24 mb-12 min-h-screen">
      <Text type="title" level={3} weight="bold">
        Account Profiles
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
              {displayFilterOpts[filterPayload.source]}
              <SVG name="caretDown" size={16} />
            </Button>
            {selectUsers()}
          </div>
          <div key={0} className="max-w-3xl">
            <PropertyFilter
              profileType="account"
              source={filterPayload.source}
              filters={filterPayload.filters}
              setFilters={setFilters}
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
      {accounts.isLoading ? (
        <Spin size="large" className="fa-page-loader" />
      ) : accounts.data.length ? (
        <div>
          <Table
            onRow={(user) => ({
              onClick: () => {
                getProfileAccountDetails(activeProject.id, user.identity);
                showModal();
              }
            })}
            className="fa-table--basic"
            dataSource={accounts.data}
            columns={columns}
            rowClassName="cursor-pointer"
            pagination={{ position: ['bottom', 'left'] }}
          />
        </div>
      ) : (
        <Text type="title" level={6} extraClass="mt-20 italic">
          There are currently no Accounts available for this project.
        </Text>
      )}
      <Modal
        title={null}
        visible={isModalVisible}
        className="fa-modal--full-width"
        footer={null}
        closable={null}
      >
        <AccountDetails
          onCancel={handleCancel}
          accountDetails={accountDetails}
        />
      </Modal>
    </div>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  accounts: state.timelines.accounts,
  accountDetails: state.timelines.accountDetails
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccounts,
      getProfileAccountDetails,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
