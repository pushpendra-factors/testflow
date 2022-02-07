import React, { useState } from 'react';
import { connect } from 'react-redux';
import { useEffect } from 'react';
import { Button, message, notification } from 'antd';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { useSelector } from 'react-redux';
import { udpateProjectSettings } from 'Reducers/global';
import factorsai from 'factorsai';

const RevealIntegration = ({
  udpateProjectSettings,
  setIsActive,
  kbLink = false,
  active = false,
}) => {
  const [loading, setLoading] = useState(false);

  const { active_project: activeProject, currentProjectSettings } = useSelector(
    (state) => state.global
  );

  useEffect(() => {
    if (currentProjectSettings?.int_clear_bit) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const enableClearbitReveal = () => {
    setLoading(true);

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION',{'name': 'reveal','activeProjectID': activeProject.id});

    udpateProjectSettings(activeProject.id, { int_clear_bit: true })
      .then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Clearbit Reveal integration enabled!');
        }, 500);
        setIsActive(true);
      })
      .catch((err) => {
        setLoading(false);
        message.error(`${err?.data?.error}`);
        setIsActive(false);
      });
  };

  const onDisconnect = () => {
    setLoading(true);
    udpateProjectSettings(activeProject.id, { int_clear_bit: false })
      .then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Clearbit Reveal integration disabled!');
        }, 500);
        setIsActive(false);
      })
      .catch((err) => {
        message.error(`${err?.data?.error}`);
        setLoading(false);
      });
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            subtitle={'Facing issues with Clearbit Reveal integrations'}
          />
        }
        onError={FaErrorLog}
      >
        <div className={'mt-4 flex'}>
          {active ? (
            <Button loading={loading} onClick={onDisconnect}>
              Disable
            </Button>
          ) : (
            <Button
              type={'primary'}
              loading={loading}
              onClick={enableClearbitReveal}
            >
              Enable Now
            </Button>
          )}
          {kbLink && (
            <a className={'ant-btn ml-2 '} target={'_blank'} href={kbLink}>
              View documentation
            </a>
          )}
        </div>
      </ErrorBoundary>
    </>
  );
};

export default connect(null, { udpateProjectSettings })(RevealIntegration);
