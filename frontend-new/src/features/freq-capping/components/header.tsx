import React, {
  Dispatch,
  SetStateAction,
  useCallback,
  useEffect,
  useReducer,
  useState
} from 'react';
import { Button, Input, Row, Space, Spin } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory, useParams } from 'react-router-dom';
import { AppContentHeader } from 'Views/AppContentHeader';
import SVG from 'Components/factorsComponents/SVG';
import { Text } from 'Components/factorsComponents';
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons';
import { PathUrls } from 'Routes/pathUrls';
import { cloneDeep } from 'lodash';
import { ComponentStates, FrequencyCap } from '../types';

interface FreqCapHeaderType {
  state: ComponentStates;
  ruleToView: FrequencyCap | undefined;
  isRuleEdited: boolean;
  setRuleToEdit: Dispatch<SetStateAction<FrequencyCap | undefined>> | undefined;
  publishChanges: () => any;
}
export const FreqCapHeader = ({
  state,
  ruleToView,
  isRuleEdited,
  setRuleToEdit,
  publishChanges
}: FreqCapHeaderType) => {
  const history = useHistory();
  const { rule_id } = useParams();

  const setHeaderContent = (textValue: string) => {
    const editedRule = cloneDeep(ruleToView);
    editedRule?.display_name = textValue;
    setRuleToEdit(editedRule);
  };

  const headerContent = () => {
    if (state === ComponentStates.LIST || state === ComponentStates.EMPTY) {
      return (
        <Row>
          {/* <SVG name='userLock' size={24} /> */}
          <Text type='title' level={7} weight='bold' extraClass='ml-1'>
            Account Level Frequency Capping (ALFC)
          </Text>
        </Row>
      );
    }
    if (state === ComponentStates.VIEW) {
      return (
        <Row>
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
            extraClass='ml-1'
          >
            {ruleToView?.display_name || `Untitled Frequency cap rule`}
          </Text>
        </Row>
      );
    }
  };

  const actions = () => {
    if (state === ComponentStates.LIST || state === ComponentStates.EMPTY) {
      return (
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
      );
    }
    if (state === ComponentStates.VIEW) {
      return (
        <Button
          type='primary'
          id='fa-at-btn--new-report'
          onClick={() => publishChanges()}
          disabled={!isRuleEdited && rule_id !== 'new'}
        >
          <Space>Publish</Space>
        </Button>
      );
    }
  };
  return <AppContentHeader heading={headerContent()} actions={actions()} />;
};
