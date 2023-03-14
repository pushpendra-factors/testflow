import { Modal } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import OnBoard1 from './OnBoard1';
import OnBoard2 from './OnBoard2';
import OnBoard3 from './OnBoard3';
import OnBoardHeader from './OnBoardHeader';

const OnBoard = () => {
  // const location = useLocation();
  // const history = useHistory();
  const { isWebsiteVisitorIdentificationVisible, currentStep } = useSelector(
    (state) => state.onBoardFlow
  );

  return (
    <div>
      <Modal
        title={<OnBoardHeader />}
        visible={isWebsiteVisitorIdentificationVisible}
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
