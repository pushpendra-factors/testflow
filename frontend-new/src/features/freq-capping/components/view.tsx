import React, {
  Dispatch,
  SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useState
} from 'react';
import { Number, SVG, Text } from 'Components/factorsComponents';
import {
  Button,
  Col,
  Collapse,
  InputNumber,
  Radio,
  Row,
  Select,
  Switch,
  Table
} from 'antd';
import { Svg } from '@react-pdf/renderer';
import { useDispatch, useSelector } from 'react-redux';
import FaSelect from 'Components/GenericComponents/FaSelect';
import FaToggleBtn from 'Components/FaToggleBtn';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { computeFilterProperties } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { getGroupProperties, getGroups } from 'Reducers/coreQuery/middleware';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { FETCH_GROUPS_FULFILLED } from 'Reducers/types';
import { fetchGroupProperties, fetchGroups } from 'Reducers/coreQuery/services';
import { convertPropsToOptions, formatGroups } from 'Reducers/coreQuery/utils';
import {
  fetchGroupPropertiesAction,
  setGroupPropertiesNamesAction
} from 'Reducers/coreQuery/actions';
import { cloneDeep } from 'lodash';
import cx from 'classnames';
import { AdvanceRuleFilters, FrequencyCap } from '../types';
import styles from '../index.module.scss';

interface ViewProps {
  ruleToView: FrequencyCap | undefined;
  campaignConfig: object;
  setRuleToEdit: Dispatch<SetStateAction<FrequencyCap | undefined>>;
}

const { Panel } = Collapse;

const FrequencyCappingView = ({
  ruleToView,
  campaignConfig,
  setRuleToEdit
}: ViewProps) => {
  const dispatch = useDispatch();
  const { active_project } = useSelector((state: any) => state.global);
  const userProperties = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const availableGroups = useSelector((state) => state.coreQuery.groups);

  const [objectIds, setObjectIds] = useState([]);

  useEffect(() => {
    getGroupsDispatch(active_project.id);
  }, []);

  const getGroupsDispatch = async (projectID) => {
    const response = await fetchGroups(projectID);
    const data = formatGroups(response.data);

    dispatch({
      type: FETCH_GROUPS_FULFILLED,
      payload: data
    });
  };

  useEffect(() => {
    if (
      ruleToView?.object_type &&
      campaignConfig &&
      ruleToView.object_type !== 'account'
    ) {
      setObjectIds(campaignConfig[ruleToView?.object_type]);
    }
  }, [campaignConfig, ruleToView?.object_type]);

  const getGroupPropsFromAPI = useCallback(
    async (groupId) => {
      if (!groupProperties[groupId]) {
        const response = await fetchGroupProperties(active_project.id, groupId);
        const options = convertPropsToOptions(
          response.data?.properties,
          response.data?.display_names
        );

        dispatch(
          setGroupPropertiesNamesAction(groupId, response.data?.display_names)
        );
        dispatch(fetchGroupPropertiesAction(options, groupId));
      }
    },
    [active_project.id, groupProperties]
  );

  useEffect(() => {
    getGroupPropsFromAPI(GROUP_NAME_DOMAINS);
    Object.keys(availableGroups?.all_groups || {}).forEach((group) => {
      getGroupPropsFromAPI(group);
    });
  }, [active_project.id, availableGroups]);

  const selectObjectType = (event) => {
    const editedRule = cloneDeep(ruleToView);
    editedRule?.object_type = event.target.value;
    editedRule?.object_ids = [];
    setRuleToEdit(editedRule);
  };

  const selectObjectIds = (ids) => {
    const editedRule = cloneDeep(ruleToView);
    editedRule?.object_ids = ids;
    setRuleToEdit(editedRule);
  };

  const renderTitle = () => (
    <div>
      <Row className='flex justify-between'>
        <Text
          color='character-primary'
          level={5}
          weight='bold'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-title'
        >
          At what level you want to set capping
        </Text>

        <Button disabled>Log</Button>
      </Row>
      <Row>
        <Text
          color='character-secondary'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-desc'
        >
          Choose to cap impression at campaign level or at a campaign group
          level
        </Text>
      </Row>
    </div>
  );

  const renderObjectSelection = () => (
    <div>
      <Row className='mt-8'>
        <Text
          extraClass='mb-0'
          color='character-secondary'
          type='title'
          weight='thin'
        >
          Select Level
        </Text>
      </Row>
      <Row>
        <Radio.Group value={ruleToView?.object_type}>
          <Radio.Button
            onChange={selectObjectType}
            className={styles['left-tab-button']}
            value='campaign'
          >
            <div className='flex gap-x-1 justify-center items-center h-full'>
              <Svg size={16} name='eye' />
              <Text
                level={7}
                color={
                  ruleToView?.object_type === 'campaign'
                    ? 'brand-color-6'
                    : 'black'
                }
                type='title'
                extraClass='mb-0'
              >
                Campaign
              </Text>
            </div>
          </Radio.Button>
          <Radio.Button
            onChange={selectObjectType}
            className={}
            value='campaign_group'
          >
            <div className='flex gap-x-1 justify-center items-center h-full'>
              <Svg size={16} name='eye' />
              <Text
                level={7}
                color={
                  ruleToView?.object_type === 'campaign_group'
                    ? 'brand-color-6'
                    : 'black'
                }
                type='title'
                extraClass='mb-0'
              >
                Campaign Group
              </Text>
            </div>
          </Radio.Button>
          <Radio.Button
            onChange={selectObjectType}
            className={styles['right-tab-button']}
            value='account'
          >
            <div className='flex gap-x-1 justify-center items-center h-full'>
              <Svg size={16} name='eye' />
              <Text
                level={7}
                color={
                  ruleToView?.object_type === 'account'
                    ? 'brand-color-6'
                    : 'black'
                }
                type='title'
                extraClass='mb-0'
              >
                All Ad Accounts
              </Text>
            </div>
          </Radio.Button>
        </Radio.Group>
      </Row>
    </div>
  );

  const renderObjectIdsOptions = () =>
    objectIds.map((obj: any) => <Option value={obj.id}>{obj.name}</Option>);

  const renderObjectIdSelection = () => {
    if (ruleToView?.object_type === 'account') return null;
    const placeHolder =
      ruleToView?.object_type === 'campaign' ? 'Campaigns' : 'Campaign Groups';
    return (
      <div className='w-full pb-2 mt-8'>
        <Row className=''>
          <Text
            extraClass='mb-0'
            color='character-secondary'
            type='title'
            weight='thin'
          >
            Select {placeHolder}
          </Text>
        </Row>

        <Row className='mt-2'>
          <Select
            mode='multiple'
            allowClear
            value={ruleToView?.object_ids}
            className='w-full'
            placeholder={`Select ${placeHolder}`}
            onChange={(e) => selectObjectIds(e)}
          >
            {renderObjectIdsOptions()}
          </Select>
        </Row>
      </div>
    );
  };

  const renderDefaultCapTable = () => {
    const setTotalImpression = (num) => {
      const editedRule = cloneDeep(ruleToView);
      editedRule?.impression_threshold = num;
      setRuleToEdit(editedRule);
    };

    const setTotalClicks = (num) => {
      const editedRule = cloneDeep(ruleToView);
      editedRule?.click_threshold = num;
      setRuleToEdit(editedRule);
    };

    const defaultColumns = [
      {
        title: 'Default Criteria',
        dataIndex: 'default_criteria',
        key: 'default_criteria',
        width: '700px',
        render: () => (
          <Button disabled className='btn-total-round'>
            For all accounts
          </Button>
        )
      },
      {
        title: 'Total Clicks Cap',
        dataIndex: 'total_clicks',
        key: 'total_clicks',
        render: (item: number) => (
          <InputNumber value={item} onChange={setTotalClicks} />
        )
      },
      {
        title: 'Total Impressions Cap',
        dataIndex: 'total_impression',
        key: 'total_impression',
        render: (item: number) => (
          <InputNumber value={item} onChange={setTotalImpression} />
        )
      }
    ];

    const tableData = [
      {
        default_criteria: undefined,
        total_clicks: ruleToView?.click_threshold,
        total_impression: ruleToView?.impression_threshold
      }
    ];

    return (
      <Row className='mt-4'>
        <Table
          className='fa-table--basic mt-6'
          columns={defaultColumns}
          dataSource={tableData}
          pagination={false}
          loading={false}
          tableLayout='fixed'
        />
      </Row>
    );
  };

  const toggleAdvanceRulesContainer = () =>
    setRuleToEdit({
      ...ruleToView,
      is_advanced_rule_enabled: !ruleToView?.is_advanced_rule_enabled
    });

  const renderCollapseIcon = () => (
    <Switch
      checked={ruleToView?.is_advanced_rule_enabled}
      onClick={toggleAdvanceRulesContainer}
    />
  );

  const renderAdvanceCapHeader = () => (
    <div>
      <Row>
        <Text
          color='character-primary'
          level={5}
          weight='bold'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-title'
        >
          Set Advanced custom rules
        </Text>
      </Row>
      <Row>
        <Text
          color='character-secondary'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-desc'
        >
          Choose to cap impression at campaign level or at a campaign group
          level
        </Text>
      </Row>
    </div>
  );

  const mainFilterProps = useMemo(
    () =>
      computeFilterProperties({
        userProperties,
        groupProperties,
        availableGroups: availableGroups?.all_groups,
        profileType: 'account'
      }),
    [userProperties, groupProperties, availableGroups]
  );

  const handleInsertFilter = (filterState, index) => {};

  const handleCloseFilter = () => {};

  const handleDeleteFilter = () => {};

  const showFilterDropdown = () => {
    const editedRule = cloneDeep(ruleToView);
    editedRule?.advanced_rules.push([
      { click_threshold: 0, impression_threshold: 0, filters: [] }
    ]);
    setRuleToEdit(editedRule);
  };

  const renderAdvanceCapTableFooter = () => (
    <Button
      className={cx('flex items-center gap-x-2', styles['add-filter-button'])}
      type='text'
      onClick={showFilterDropdown}
    >
      <SVG name='plus' color='#00000073' />
      <Text
        type='title'
        color='character-title'
        extraClass='mb-0'
        weight='medium'
      >
        Add filter
      </Text>
    </Button>
  );

  const renderAdvanceCapRules = () => {
    const columns = [
      {
        title: 'Accounts that match',
        dataIndex: 'filters',
        key: 'filters',
        width: '700px',
        render: (filters: any, index) => (
          <FilterWrapper
            viewMode={false}
            projectID={active_project?.id}
            index={index}
            filterProps={mainFilterProps}
            minEntriesPerGroup={3}
            insertFilter={handleInsertFilter}
            closeFilter={handleCloseFilter}
            deleteFilter={handleDeleteFilter}
            showInList={false}
          />
        )
      },
      {
        title: 'Total Clicks Cap',
        dataIndex: 'total_clicks',
        key: 'total_clicks',
        render: (item: number) => <InputNumber value={item} />
      },
      {
        title: 'Total Impressions Cap',
        dataIndex: 'total_impression',
        key: 'total_impression',
        render: (item: number) => <InputNumber value={item} />
      }
    ];

    const tableData = ruleToView?.advanced_rules.map(
      (rule: AdvanceRuleFilters) => ({
        filters: rule.filters,
        total_clicks: rule.click_threshold,
        total_impression: rule.impression_threshold
      })
    );

    return (
      <Row>
        <Collapse
          className={`w-full ${styles['advance-rules-cap']}`}
          onChange={() => {}}
          expandIconPosition='right'
          expandIcon={renderCollapseIcon}
          defaultActiveKey={ruleToView?.is_advanced_rule_enabled ? 1 : 0}
        >
          <Panel key={1} header={renderAdvanceCapHeader()}>
            <Row className='mt-4'>
              <Select disabled defaultValue='segment'>
                <Option value='segment'>Based on segment membership</Option>
              </Select>
            </Row>

            <Row className={`mt-4 ${styles['cap-table']}`}>
              <Table
                className='fa-table--basic mt-6'
                columns={columns}
                dataSource={tableData}
                bordered={false}
                pagination={false}
                loading={false}
                tableLayout='fixed'
                footer={renderAdvanceCapTableFooter}
              />
            </Row>
          </Panel>
        </Collapse>
      </Row>
    );
  };

  const renderCappingConditions = () => (
    <div className='mt-8'>
      <Row className='flex justify-between'>
        <Col>
          <Text
            color='character-primary'
            level={5}
            weight='bold'
            extraClass='mb-0'
            type='title'
            id='fa-at-text--draft-title'
          >
            Set capping condition
          </Text>
        </Col>
        <div className='flex'>
          <Text
            color='character-secondary'
            level={7}
            extraClass='mb-0 mr-2'
            type='title'
            id='fa-at-text--draft-desc'
          >
            Capping interval
          </Text>

          <Select disabled defaultValue='monthly'>
            <Option value='monthly'>Monthly</Option>
          </Select>
        </div>
      </Row>
      <Row>
        <Text
          color='character-secondary'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-desc'
        >
          You can either set a cap for all accounts or add advanced capping
          criteria.
        </Text>
      </Row>
      {renderDefaultCapTable()}
    </div>
  );

  return (
    <div>
      {renderTitle()}
      {renderObjectSelection()}
      {renderObjectIdSelection()}
      {renderCappingConditions()}
      {renderAdvanceCapRules()}
    </div>
  );
};

export default FrequencyCappingView;
