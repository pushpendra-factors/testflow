import { ArrowLeftOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { Button, Row } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React, { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import {
  BACK_STEP_ONBOARD_FLOW,
  NEXT_STEP_ONBOARD_FLOW,
  TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL
} from 'Reducers/types';
import styles from './index.module.scss';
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
const AdditionalMenu = (
  int_completed,
  closeDrawer,
  setCurrentStep,
  stepDone,
  setStepDone
) => {
  const history = useHistory();
  const dispatch = useDispatch();
  const handleCloseDrawer = useCallback(() => {
    dispatch({ type: TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL });
    history.push('/');
  }, []);
  const { int_client_six_signal_key, int_factors_six_signal_key } = useSelector(
    (state) => state?.global?.currentProjectSettings
  );
  const { steps, currentStep } = useSelector((state) => state?.onBoardFlow);
  const isNextBtnEnabled = () => {
    if (currentStep === 1) {
      return steps.step1;
    } else if (currentStep === 2) {
      return steps.step2 || int_client_six_signal_key;
    } else if (currentStep === 3) {
      return steps.step3;
    }
    return false;
  };

  return (
    <div className={styles['additionalMenu']}>
      {currentStep > 1 && currentStep !== null ? (
        <Button
          className={styles['btn']}
          onClick={() => {
            dispatch({ type: BACK_STEP_ONBOARD_FLOW });
            // setCurrentStep((prev) => prev - 1);
            // history.go(-1);
            // history.push('/visitoridentification/' + Number(step - 1));
          }}
        >
          <ArrowLeftOutlined /> Back
        </Button>
      ) : (
        ''
      )}

      <Button
        className={styles['btn']}
        disabled={!isNextBtnEnabled()}
        onClick={
          currentStep && currentStep >= 1 && currentStep < 3
            ? () => {
                dispatch({ type: NEXT_STEP_ONBOARD_FLOW });
                // setCurrentStep((prev) => prev + 1);
                // history.push(
                //   '/visitoridentification/' + Number(currentStep + 1)
                // );
              }
            : handleCloseDrawer
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

      <div className={styles['closebtnc'] + ' ' + styles['btn']}>
        <Button onClick={handleCloseDrawer}>Close</Button>
      </div>
    </div>
  );
};
const OnBoardHeader = ({
  closeDrawer,
  setCurrentStep,
  stepDone,
  setStepDone
}) => {
  const history = useHistory();
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
          {RenderLogo()} {RenderTitle(getTitle(currentStep))} <RenderStep />
        </Row>
      </div>
      <div>
        <AdditionalMenu
          int_completed={int_completed}
          closeDrawer={closeDrawer}
          setCurrentStep={setCurrentStep}
          stepDone={stepDone}
          setStepDone={setStepDone}
        />
      </div>
    </Row>
  );
};

export default OnBoardHeader;
