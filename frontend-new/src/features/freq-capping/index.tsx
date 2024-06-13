import React, { useCallback, useEffect, useReducer, useState } from 'react';
import { Spin, notification } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory, useParams } from 'react-router-dom';
import { fetchCampaignConfig } from 'Reducers/coreQuery/services';
import { convertCampaignConfig } from 'Reducers/coreQuery/utils';
import { getCampaignConfigAction } from 'Reducers/coreQuery/actions';
import logger from 'Utils/logger';
import {
  getEventPropertiesV2,
  getGroupProperties,
  getGroups,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { cloneDeep, isEqual } from 'lodash';
import { AppContentHeader } from 'Views/AppContentHeader';
import { PathUrls } from 'Routes/pathUrls';
import FrequencyCappingList from './components/list';
import FrequencyCappingView from './components/view';
import { ComponentStates, FrequencyCap, RuleQueryParams } from './types';
import {
  getLinkedinFreqCapRuleConfig,
  getLinkedinFreqCapRules,
  publishLinkedinFreqCapRules,
  updateLinkedinFreqCapRules
} from './state/service';
import { FreqCapHeader } from './components/header';

const FrequencyCapping = () => {
  const [componentState, setComponentState] = useState(ComponentStates.LOADING);
  const [freqCapRules, setFreqCapRules] = useState<Array<FrequencyCap>>([]);
  const [selectedRule, setSelectedRule] = useState<FrequencyCap | undefined>();
  const [isRuleEdited, setIsRuleEdited] = useState(false);
  const [campaignConfig, setCampaignConfig] = useState<object>({
    campaign: [],
    campaign_group: []
  });

  const dispatch = useDispatch();
  const history = useHistory();

  const { active_project } = useSelector((state: any) => state.global);

  const { rule_id } = useParams<RuleQueryParams>();

  useEffect(() => {
    // getUserPropertiesV2(active_project.id);
    getGroups(active_project);
    // Config call
    fetchConfig();

    // Fetch call
    fetchFreqCapRules();
    setComponentState(ComponentStates.LIST);
  }, []);

  useEffect(() => {
    if (rule_id && rule_id !== 'new') {
      setRuleBasedOnRuleId();
      setComponentState(ComponentStates.VIEW);
    }
    if (rule_id === 'new') {
      const ruleToView = new FrequencyCap(active_project.id);
      setSelectedRule(ruleToView);
      setComponentState(ComponentStates.VIEW);
    }
  }, [rule_id, freqCapRules]);

  useEffect(() => {
    if (selectedRule) {
      const ruleToView = freqCapRules.find((rule) => rule.id === rule_id);
      const ruleEdited = isEqual(ruleToView, selectedRule);
      setIsRuleEdited(!ruleEdited);
    } else {
      setIsRuleEdited(false);
    }
  }, [selectedRule]);

  const setRuleBasedOnRuleId = () => {
    const ruleToView = freqCapRules.find((rule) => rule.id === rule_id);
    if (ruleToView) {
      setSelectedRule(ruleToView);
    }
  };

  const fetchConfig = async () => {
    const configResponse = await getLinkedinFreqCapRuleConfig(
      active_project?.id
    );
    if (configResponse?.status === 200) {
      const campaingConf = cloneDeep(campaignConfig);
      configResponse?.data?.forEach((conf) => {
        if (!conf.deleted) {
          campaingConf[conf.type].push(conf);
        }
      });
      setCampaignConfig(campaingConf);
    }
  };

  const fetchFreqCapRules = async () => {
    const response = await getLinkedinFreqCapRules(active_project?.id);
    if (response?.status === 200) {
      setFreqCapRules(response.data);
    }
  };

  const publishFreqCalRules = async () => {
    if (selectedRule.id) {
      const response = await updateLinkedinFreqCapRules(
        active_project.id,
        selectedRule
      );
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
    } else {
      const response = await publishLinkedinFreqCapRules(
        active_project.id,
        selectedRule
      );

      if (response?.status === 200) {
        notification.success({
          message: 'Success',
          description: 'Rule Successfully Created!',
          duration: 3
        });
        fetchFreqCapRules();
        history.replace(`${PathUrls.FreqCap}/${response.data.id}`);
      } else {
        notification.error({
          message: 'Failed!',
          description: response?.status,
          duration: 3
        });
      }
    }
  };

  const renderList = () => (
    <div>
      <FreqCapHeader
        state={componentState}
        ruleToView={undefined}
        setRuleToEdit={undefined}
      />
      <FrequencyCappingList
        freqCapRules={freqCapRules}
        deleteCallBack={fetchFreqCapRules}
      />
    </div>
  );

  const renderView = () => (
    <div>
      <FreqCapHeader
        state={componentState}
        ruleToView={selectedRule}
        setRuleToEdit={setSelectedRule}
        isRuleEdited={isRuleEdited}
        publishChanges={() => {
          publishFreqCalRules();
        }}
      />
      <FrequencyCappingView
        campaignConfig={campaignConfig}
        ruleToView={selectedRule}
        setRuleToEdit={setSelectedRule}
      />
    </div>
  );

  switch (componentState) {
    case ComponentStates.LOADING: {
      return <Spin size='large' className='fa-page-loader' />;
      break;
    }
    case ComponentStates.LIST: {
      return renderList();
      break;
    }
    case ComponentStates.VIEW: {
      return renderView();
      break;
    }
    default:
      // default Error state goes here
      return null;
      break;
  }
};

export default FrequencyCapping;
