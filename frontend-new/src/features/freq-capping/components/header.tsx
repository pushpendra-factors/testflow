import React, {
  Dispatch,
  SetStateAction,
  useCallback,
  useEffect,
  useReducer,
  useState
} from 'react';
import {
  Button,
  Dropdown,
  Input,
  Menu,
  Modal,
  Row,
  Space,
  Spin,
  Switch,
  notification
} from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory, useParams } from 'react-router-dom';
import { AppContentHeader } from 'Views/AppContentHeader';
import SVG from 'Components/factorsComponents/SVG';
import { Text } from 'Components/factorsComponents';
import {
  ArrowLeftOutlined,
  ExclamationCircleOutlined,
  PlusOutlined
} from '@ant-design/icons';
import { PathUrls } from 'Routes/pathUrls';
import { cloneDeep } from 'lodash';
import { ComponentStates, FrequencyCap } from '../types';
import {
  deleteLinkedinFreqCapRules,
  updateLinkedinFreqCapRules
} from '../state/service';
import styles from '../index.module.scss';

interface FreqCapHeaderType {
  state: ComponentStates;
  ruleToView?: FrequencyCap | undefined;
  isRuleEdited?: boolean;
  setRuleToEdit?:
    | Dispatch<SetStateAction<FrequencyCap | undefined>>
    | undefined;
  publishChanges?: () => any;
  fetchFreqCapRules?: () => any;
  setRuleBasedOnRuleId?: () => any;
}
export const FreqCapHeader = ({
  state,
  ruleToView,
  isRuleEdited,
  setRuleToEdit,
  publishChanges,
  fetchFreqCapRules,
  setRuleBasedOnRuleId
}: FreqCapHeaderType) => {
  const history = useHistory();
  const { rule_id } = useParams();

  const setHeaderContent = (textValue: string) => {
    const editedRule = cloneDeep(ruleToView);
    editedRule?.display_name = textValue;
    setRuleToEdit(editedRule);
  };

  const { confirm } = Modal;

  const headerContent = () => {
    if (state === ComponentStates.LIST || state === ComponentStates.EMPTY) {
      return (
        <Row className='items-start'>
          {/* <SVG name='userLock' size={24} /> */}
          <Text type='title' level={6} weight='bold' extraClass='ml-1 mb-0'>
            Account Level Frequency Capping (ALFC)
          </Text>
        </Row>
      );
    }
    if (state === ComponentStates.VIEW) {
      return (
        <Row className={`items-start ${styles['header-title']}`}>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => history.replace(`${PathUrls.FreqCap}`)}
            className='mr-4'
          />
          {/* <Input
            size='large'
            style={{ width: '300px' }}
            placeholder='Untitled Workflow '
            value={ruleToView?.display_name}
            onChange={(e) => setHeaderContent(e.target.value)}
            className='fa-input '
          /> */}
          <Text
            editable={{ onChange: setHeaderContent }}
            type='title'
            level={4}
            weight='bold'
            extraClass='ml-1 mb-0'
          >
            {ruleToView?.display_name || `Untitled Frequency cap rule`}
          </Text>
        </Row>
      );
    }
  };

  const updateChanges = async (rule: FrequencyCap) => {
    const response = await updateLinkedinFreqCapRules(rule.project_id, rule);
    if (response?.status === 200) {
      notification.success({
        message: 'Success',
        description: 'Rule Successfully Updated!',
        duration: 3
      });
      fetchFreqCapRules();
    } else {
      notification.error({
        message: 'Failed!',
        description: response?.status,
        duration: 3
      });
    }
  };

  const toggleStatus = () => {
    const title =
      ruleToView?.status === 'active'
        ? 'Pause this feature?'
        : 'Are you sure you want to publish';

    const content =
      ruleToView?.status === 'active'
        ? 'It will be temporarilyIt will be temporarily unavailable until you resume it.'
        : 'Once published, it will be active and running.';
    const okText =
      ruleToView?.status === 'active' ? 'Pause' : 'Confirm Publish';
    confirm({
      title,
      icon: <ExclamationCircleOutlined />,
      content,
      okText,
      onOk: () => {
        const editedRule = cloneDeep(ruleToView);
        editedRule?.status =
          ruleToView?.status === 'active' ? 'paused' : 'active';

        updateChanges(editedRule);
        // setRuleToEdit(editedRule);
      }
    });
  };

  const makeACopy = () => {
    const editedCopy = cloneDeep(ruleToView);
    history.replace(`${PathUrls.FreqCap}/new`);
  };

  const deleteRule = async () => {
    const response = await deleteLinkedinFreqCapRules(
      ruleToView?.project_id,
      rule_id
    );
    if (response?.status === 200) {
      notification.success({
        message: 'Success',
        description: 'Rule Successfully Deleted!',
        duration: 3
      });

      fetchFreqCapRules();
      history.replace(`${PathUrls.FreqCap}`);
    }
  };

  const getMoreOptions = () => (
    <Menu style={{ minWidth: '200px', padding: '10px' }}>
      <Menu.Item
        icon={
          <SVG
            name='trash'
            extraClass='self-center'
            style={{ marginRight: '10px' }}
          />
        }
        style={{ display: 'flex', padding: '10px', margin: '5px' }}
        key='delete'
        onClick={() => {
          deleteRule();
        }}
      >
        <span style={{ paddingLeft: '5px' }}>Delete</span>
      </Menu.Item>
    </Menu>
  );

  const actions = () => {
    if (state === ComponentStates.LIST || state === ComponentStates.EMPTY) {
      return (
        <Row>
          <Button
            type='primary'
            id='fa-at-btn--new-report'
            onClick={() => history.replace(`${PathUrls.FreqCap}/new`)}
          >
            <Space>
              <SVG name='plus' size={16} color='white' />
              Add New Rule
            </Space>
          </Button>
        </Row>
      );
    }
    if (state === ComponentStates.VIEW) {
      if (rule_id === 'new') {
        return (
          <Row className='items-center'>
            <Button
              type='primary'
              id='fa-at-btn--new-report'
              onClick={() => publishChanges()}
              disabled={
                !ruleToView?.display_name ||
                (ruleToView?.object_type !== 'account' &&
                  ruleToView.object_ids.length === 0)
              }
            >
              <Space>Publish</Space>
            </Button>

            <Dropdown
              disabled
              placement='bottomRight'
              overlay={getMoreOptions()}
              trigger={['click']}
            >
              <Button
                type='text'
                size='large'
                className='fa-btn--custom ml-2'
                disabled
              >
                <SVG name='more' />
              </Button>
            </Dropdown>
          </Row>
        );
      }
      if (rule_id !== 'new' && !isRuleEdited) {
        return (
          <Row className='items-center'>
            <Switch
              checkedChildren='Active'
              unCheckedChildren='Paused'
              onChange={toggleStatus}
              checked={ruleToView?.status === 'active'}
            />
            <Dropdown
              placement='bottomRight'
              overlay={getMoreOptions()}
              trigger={['click']}
            >
              <Button
                type='text'
                size='large'
                className='fa-btn--custom ml-1'
                disabled
              >
                <SVG name='more' />
              </Button>
            </Dropdown>
          </Row>
        );
      }
      if (rule_id !== 'new' && isRuleEdited) {
        return (
          <Row className='items-center'>
            <Button className='mr-1' onClick={() => setRuleBasedOnRuleId()}>
              <Space>Discard Changes</Space>
            </Button>
            <Button type='primary' onClick={() => publishChanges()}>
              <Space>Publish</Space>
            </Button>
          </Row>
        );
      }
    }
  };
  return <AppContentHeader heading={headerContent()} actions={actions()} />;
};
