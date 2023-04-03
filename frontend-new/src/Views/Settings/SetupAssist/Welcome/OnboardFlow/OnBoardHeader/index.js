import { ArrowLeftOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { Button, Row, Tooltip } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React, { useCallback, useEffect, useState } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import {
  BACK_STEP_ONBOARD_FLOW,
  NEXT_STEP_ONBOARD_FLOW,
  TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL
} from 'Reducers/types';
import logger from 'Utils/logger';
import styles from './index.module.scss';
import { udpateProjectSettings } from 'Reducers/global';
const RenderLogo = () => (
  <Button size='large' type='text' icon={<SVG size={32} name='Brand' />} />
);

const RenderTitle = (subTitle) => {
  return (
    <div className={styles['headerTitle']}>
      Website visitor identification : {subTitle}
    </div>
  );
};
const RenderStep = () => {
  const { currentStep } = useSelector((state) => state?.onBoardFlow);
  return (
    <div className={`${styles['headerTitle']}`}>
      <div className={`${styles['chip']}`}>{currentStep} of 3</div>
    </div>
  );
};
const AdditionalMenu = ({
  closeDrawer,
  setCurrentStep,
  stepDone,
  setStepDone,
  udpateProjectSettings
}) => {
  const history = useHistory();
  const dispatch = useDispatch();
  const handleCloseDrawer = useCallback(() => {
    // dispatch({ type: TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL });
    history.push('/welcome');
  }, []);
  const int_completed = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );

  const completeUserOnboard = () => {
    udpateProjectSettings(activeProject.id, {
      is_onboarding_completed: true
    });
  };

  const handleDoneDrawer = useCallback(() => {
    completeUserOnboard();
    // dispatch({ type: TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL });
    history.push('/');
  }, []);
  const {
    int_client_six_signal_key,
    int_factors_six_signal_key,
    int_clear_bit,
    is_deanonymization_requested
  } = useSelector((state) => state?.global?.currentProjectSettings);

  console.log(is_deanonymization_requested);
  const activeProject = useSelector((state) => state?.global?.active_project);
  const { steps, currentStep } = useSelector((state) => state?.onBoardFlow);
  const isNextBtnEnabled = () => {
    if (currentStep === 1) {
      return int_completed;
    } else if (currentStep === 2) {
      return (
        int_client_six_signal_key ||
        is_deanonymization_requested ||
        int_factors_six_signal_key ||
        int_clear_bit
      );
    } else if (currentStep === 3) {
      return steps.step3;
    }
    return false;
  };
  let isEnabled = !isNextBtnEnabled();
  return (
    <div className={styles['additionalMenu']}>
      {currentStep > 1 && currentStep !== null ? (
        <Button
          className={styles['btn']}
          onClick={() => {
            // dispatch({ type: BACK_STEP_ONBOARD_FLOW });
            console.log(
              '/welcome/visitoridentification/' + Number(currentStep - 1)
            );
            history.push(
              '/welcome/visitoridentification/' + Number(currentStep - 1)
            );
          }}
        >
          <ArrowLeftOutlined /> Back
        </Button>
      ) : (
        ''
      )}
      <Tooltip
        placement='bottom'
        title={
          isEnabled && currentStep === 1
            ? 'You have to verify the SDK to continue'
            : isEnabled && currentStep === 2
            ? "You have to enable yours or request Factor's 6Signal API Keys"
            : ''
        }
      >
        <Button
          className={styles['btn']}
          disabled={isEnabled}
          onClick={
            currentStep && currentStep >= 1 && currentStep < 3
              ? () => {
                  // dispatch({ type: NEXT_STEP_ONBOARD_FLOW });
                  // setCurrentStep((prev) => prev + 1);
                  history.push(
                    '/welcome/visitoridentification/' + Number(currentStep + 1)
                  );
                }
              : handleDoneDrawer
          }
          type='primary'
        >
          {currentStep && currentStep >= 1 && currentStep < 3 ? (
            <>
              {' '}
              Next <ArrowRightOutlined />
            </>
          ) : (
            'Done'
          )}
        </Button>
      </Tooltip>

      <div className={styles['closebtnc']}>
        <Button className={styles['btn']} onClick={handleCloseDrawer}>
          Close
        </Button>
      </div>
    </div>
  );
};
const OnBoardHeader = ({
  closeDrawer,
  setCurrentStep,
  stepDone,
  setStepDone,
  udpateProjectSettings
}) => {
  const history = useHistory();
  const dispatch = useDispatch();
  const currentStep = useSelector((state) => state?.onBoardFlow?.currentStep);
  const int_completed = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );
  const getTitle = function (step) {
    if (step === 1) {
      return 'Add the Javascript SDK';
    } else if (step === 2) {
      return 'Integrations to push';
    } else if (step === 3) {
      return 'Get alerts on Slack';
    } else {
      return '';
    }
  };
  return (
    <Row className={styles['headerContainer']}>
      <div>
        <Row>
          {RenderLogo()} {RenderTitle(getTitle(currentStep))} <RenderStep />{' '}
        </Row>
      </div>
      <div>
        <AdditionalMenu
          int_completed={int_completed}
          closeDrawer={closeDrawer}
          setCurrentStep={setCurrentStep}
          stepDone={stepDone}
          setStepDone={setStepDone}
          udpateProjectSettings={udpateProjectSettings}
        />
      </div>
    </Row>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, { udpateProjectSettings })(
  OnBoardHeader
);
