import React, { useEffect, useState } from 'react';
import { Widget } from '@typeform/embed-react';
import useMobileView from 'hooks/useMobileView';
import useAgentInfo from 'hooks/useAgentInfo';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { udpateProjectSettings } from 'Reducers/global';
import {
  CommonStepsProps,
  MORE_INFO_FORM,
  OnboardingStepsConfig,
  SETUP_COMPLETED
} from '../../types';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { Spin, notification } from 'antd';
import { updateAgentInfo } from 'Reducers/agentActions';
import logger from 'Utils/logger';

const TypeformDetails = ({
  variant = 'admin',
  incrementStepCount,
  udpateProjectSettings,
  updateAgentInfo
}: TypeformDetailsPropsType) => {
  const [showForm, setShowForm] = useState(true);
  const [loading, setLoading] = useState(false);
  const isMobileView = useMobileView();
  const history = useHistory();
  const { active_project, projectSettingsV1, currentProjectSettings } =
    useSelector((state: any) => state.global);
  const { agent_details } = useSelector((state) => state.agent);

  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;
  const { email, firstName, lastName } = useAgentInfo();

  const handleSubmit = async () => {
    try {
      setLoading(true);
      setShowForm((state) => !state);

      if (variant === 'admin') {
        //handle admin submission
        let updatedOnboardingConfig = {
          [MORE_INFO_FORM]: true,
          [SETUP_COMPLETED]: true
        };
        if (onboarding_steps) {
          updatedOnboardingConfig = {
            ...onboarding_steps,
            ...updatedOnboardingConfig
          };
        }
        await udpateProjectSettings(active_project.id, {
          onboarding_steps: updatedOnboardingConfig
        });
        await updateAgentInfo({ is_onboarding_flow_seen: true });
        setLoading(false);
        incrementStepCount();
      }
      if (variant === 'invitedUser') {
        //handle invited user form submission
        await updateAgentInfo({ is_form_filled: true });
        setLoading(false);
        history.push(PathUrls.ProfileAccounts);
      }
    } catch (error) {
      setLoading(false);

      logger.error('Error in verifying SDK', error);
      notification.error({
        message: 'Error',
        description: 'Error in submitting details!',
        duration: 3
      });
    }
  };

  const renderTypeformWidget = () => (
    <Widget
      id={BUILD_CONFIG.typeformId}
      onSubmit={handleSubmit}
      style={{ width: '100%', height: isMobileView ? 500 : '100%' }}
      hidden={{
        email: email,
        first_name: firstName,
        last_name: lastName,
        project_name: active_project?.name,
        project_id: active_project?.id,
        timezone: active_project?.time_zone,
        sdk_integrated: projectSettingsV1?.int_completed ? 'true' : 'false',
        is_user_invited: variant === 'invitedUser' ? 'true' : 'false'
      }}
    />
  );

  useEffect(() => {
    if (variant === 'admin' && onboarding_steps?.[MORE_INFO_FORM]) {
      incrementStepCount();
      return;
    }
    if (variant === 'invitedUser' && agent_details?.is_form_filled) {
      history.push(PathUrls.ProfileAccounts);
    }
  }, [onboarding_steps, variant, agent_details, incrementStepCount]);

  if (loading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  }

  return <>{showForm && <>{renderTypeformWidget()}</>}</>;
};

type TypeformDetailsComponentProps = {
  variant: 'invitedUser' | 'admin';
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ udpateProjectSettings, updateAgentInfo }, dispatch);
const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type TypeformDetailsPropsType = ReduxProps &
  CommonStepsProps &
  TypeformDetailsComponentProps;

export default connector(TypeformDetails);
