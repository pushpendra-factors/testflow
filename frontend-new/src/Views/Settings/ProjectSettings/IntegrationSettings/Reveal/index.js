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
  const currentAgent = useSelector((state) => state.agent.agent_details);

  useEffect(() => {
    if (currentProjectSettings?.int_clear_bit) {
      setIsActive(true);
    }
  }, [currentProjectSettings]);

  const sendSlackNotification = () => {
    let webhookURL = 'https://hooks.slack.com/services/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';
    let data = {
        "text": `User ${currentAgent.email} from Project "${activeProject.name}" Activated Integration: Reveal`,
        "username" : "Signup User Actions",
        "icon_emoji" : ":golf:"
    }
    let params = {
        method: 'POST',
        body: JSON.stringify(data)
    }

    fetch(webhookURL, params)
    .then((response) => response.json())
    .then((response) => {
        console.log(response);
    })
    .catch((err) => {
        console.log('err',err);
    });
  }

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
        sendSlackNotification();
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
