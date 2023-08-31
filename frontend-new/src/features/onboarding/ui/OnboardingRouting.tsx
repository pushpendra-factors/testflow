import { AdminLock } from 'Routes/feature';
import { PathUrls } from 'Routes/pathUrls';
import useAgentInfo from 'hooks/useAgentInfo';
import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { OnboardingStepsConfig, SETUP_COMPLETED } from './types';

const OnboardingRouting = () => {
  const { isAgentInvited, email, isLoggedIn } = useAgentInfo();
  const { agent_details, agents } = useSelector((state) => state.agent);
  const { projects, currentProjectSettings, active_project } = useSelector(
    (state) => state.global
  );

  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;
  const history = useHistory();
  const location = useLocation();

  useEffect(() => {
    let routeFlag = false;
    let routePath = '';

    if (
      !isLoggedIn ||
      (currentProjectSettings?.project_id && active_project?.id
        ? currentProjectSettings.project_id != active_project.id
        : false)
    ) {
      return;
    } else if (!projects || projects?.length === 0) {
      //if no projects are available
      routeFlag = true;
      routePath = PathUrls.Onboarding;
    } else if (currentProjectSettings && !onboarding_steps?.[SETUP_COMPLETED]) {
      routeFlag = true;
      routePath = PathUrls.Onboarding;
    } else if (agents && agents?.length > 0 && agent_details) {
      if (isAgentInvited) {
        if (!agent_details?.is_form_filled && !AdminLock(email)) {
          //render invited user form
          routeFlag = true;
          routePath = `${PathUrls.Onboarding}?target=invited_user`;
        }
      } else {
        if (!agent_details?.is_onboarding_flow_seen) {
          // handle onboarding
          routeFlag = true;
          routePath = PathUrls.Onboarding;
        }
      }
    }

    if (location.pathname !== PathUrls.Onboarding && routeFlag && routePath) {
      history.push(routePath);
    }
  }, [
    agents,
    agent_details,
    isAgentInvited,
    location?.pathname,
    email,
    projects,
    onboarding_steps,
    currentProjectSettings,
    isLoggedIn,
    active_project?.id
  ]);
  return <></>;
};

export default OnboardingRouting;
