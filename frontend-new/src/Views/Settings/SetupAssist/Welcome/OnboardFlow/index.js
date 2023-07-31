import { Modal } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import { JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION } from 'Reducers/types';
import OnBoard1 from './OnBoard1';
import OnBoard2 from './OnBoard2';
import OnBoard3 from './OnBoard3';
import OnBoardHeader from './OnBoardHeader';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';

const OnBoard = () => {
  const location = useLocation();
  const history = useHistory();
  const dispatch = useDispatch();
  const { step } = useParams();
  const {
    int_client_six_signal_key,
    int_clear_bit,
    is_deanonymization_requested
  } = useSelector((state) => state?.global?.currentProjectSettings);
  const int_completed = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );
  const { isWebsiteVisitorIdentificationVisible, currentStep, steps } =
    useSelector((state) => state.onBoardFlow);
  const { isFeatureConnected: isFactorsDeanonymizationConnected } =
    useFeatureLock(FEATURES.INT_FACTORS_DEANONYMISATION);
  const checkIsValid = (step) => {
    if (step == 1) {
      return int_completed;
    } else if (step == 2) {
      return (
        int_client_six_signal_key ||
        is_deanonymization_requested ||
        int_clear_bit ||
        isFactorsDeanonymizationConnected
      );
    } else if (step == 3) {
      return steps.step3;
    }
    return false;
  };
  useEffect(() => {
    if (is_deanonymization_requested) {
      history.push('/welcome/visitoridentification/3');
    }
  }, []);
  useEffect(() => {
    if (step == '1' || step == '2' || step == '3') {
      if (step == '1') {
        dispatch({
          type: JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION,
          payload: Number(step)
        });
      } else if (step == '2') {
        if (checkIsValid(1)) {
          dispatch({
            type: JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION,
            payload: Number(step)
          });
        } else {
          history.push('/welcome/visitoridentification/1');
        }
      } else if (step == '3') {
        let t1 = checkIsValid(1);
        let t2 = checkIsValid(2);
        if ((t1 && t2) || (t1 === undefined && t2 === undefined)) {
          dispatch({
            type: JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION,
            payload: Number(step)
          });
        } else {
          history.push('/welcome/visitoridentification/1');
        }
      }
    } else {
      history.push('/welcome/visitoridentification/1');
    }
  }, [step]);

  return (
    <div>
      <Modal
        title={<OnBoardHeader />}
        visible={true}
        footer={null}
        centered={false}
        mask={false}
        closable={false}
        className='fa-modal--full-width'
      >
        {currentStep === 1 ? (
          <OnBoard1 />
        ) : currentStep === 2 ? (
          <OnBoard2 />
        ) : currentStep === 3 ? (
          <OnBoard3 />
        ) : (
          'Some error occured'
        )}
      </Modal>
    </div>
  );
};

export default OnBoard;
