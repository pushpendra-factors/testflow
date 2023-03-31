import { LoadingOutlined } from '@ant-design/icons';
import { Button, Row, message, Alert, Divider } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import React, { useCallback, useState } from 'react';
import { connect, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';
import { enableSlackIntegration } from 'Reducers/global';
import styles from './index.module.scss';

const HorizontalCard = ({
  title,
  description,
  icon,
  is_connected,
  onClickConnect
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const onClick = async () => {
    setIsLoading(true);
    if (onClickConnect) await onClickConnect();
    setIsLoading(false);
  };
  return (
    <Row className={styles['horizontalCard']}>
      <div className={styles['horizontalCardContent']}>
        <div className={styles['horizontalCardLeft']}>
          <div style={{ display: 'grid', placeContent: 'center' }}>{icon}</div>
          <div>
            <Text
              type={'title'}
              level={6}
              weight={'bold'}
              style={{ margin: 0 }}
            >
              {title}
            </Text>
            <div>{description}</div>
          </div>
        </div>
        <div className={styles['horizontalCardRight']}>
          <Button
            onClick={onClick}
            // icon={isLoading === true ? <LoadingOutlined /> : null}
          >
            {is_connected ? (
              <>
                <SVG name='Greentick' /> Already Connected
              </>
            ) : (
              <>{isLoading ? <LoadingOutlined /> : ''} Connect</>
            )}
          </Button>
        </div>
      </div>
    </Row>
  );
};
const OnBoard3 = ({ enableSlackIntegration }) => {
  const activeProject = useSelector((state) => state?.global?.active_project);
  const { int_slack } = useSelector(
    (state) => state?.global?.projectSettingsV1
  );
  const { int_hubspot, int_salesforce_enabled_agent_uuid } = useSelector(
    (state) => state?.global?.currentProjectSettings
  );
  const onConnectSlack = useCallback(() => {
    return new Promise((resolve, reject) => {
      enableSlackIntegration(activeProject.id, window.location.href)
        .then((r) => {
          if (r.status === 200) {
            window.open(r.data.redirectURL, '_self');
          }
          if (r.status >= 400) {
            message.error('Error fetching slack redirect url');
            reject();
          }
        })
        .catch((err) => {
          console.log('Slack error-->', err);
          reject();
        });
    });
  }, []);
  return (
    <div className={styles['onBoardContainer']}>
      <Alert
        className={styles['notification']}
        description={
          <div
            style={{
              display: 'flex',
              width: '100%',
              justifyContent: 'space-between'
            }}
          >
            <div>
              <Text
                type={'title'}
                level={6}
                weight={'bold'}
                style={{ margin: 0 }}
              >
                Necessary Integrations completed
              </Text>
              Awesome! All necessary integrations are now complete. You can
              integrate additional applications below or get started with your
              first Dashboard.
            </div>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <Button style={{ border: '1px solid #E5E5E5' }}>
                Go to Dashboard
              </Button>
            </div>
          </div>
        }
        icon={'🎉'}
        showIcon
      />
      {/* <SixSignal setIsActive={() => {}} kbLink={true} /> */}
      <div style={{ padding: '30px 0 20px 0' }}>
        <Text type={'title'} level={6} weight={'bold'}>
          Additional Integrations{' '}
          <span style={{ color: 'rgba(0, 0, 0, 0.45)' }}>(Optional)</span>
        </Text>{' '}
      </div>
      <HorizontalCard
        title={'Slack'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={<SVG name={'Slack'} size={40} extraClass={'inline mr-2 -mt-2'} />}
        is_connected={int_slack}
        onClickConnect={onConnectSlack}
      />
      <Divider style={{ margin: '5px 0' }} />
      <HorizontalCard
        title={'Hubspot'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={
          <SVG
            name={'Hubspot_ads'}
            size={40}
            extraClass={'inline mr-2 -mt-2'}
          />
        }
        is_connected={int_hubspot}
        onClickConnect={null}
      />
      <Divider style={{ margin: '5px 0' }} />
      <HorizontalCard
        title={'Salesforce'}
        description={
          'Get alerts when high-intent actions take place by your prospects. Close more deals by being closest to the action.'
        }
        icon={
          <SVG
            name={'Salesforce_ads'}
            size={40}
            extraClass={'inline mr-2 -mt-2'}
          />
        }
        is_connected={int_salesforce_enabled_agent_uuid}
        onClickConnect={null}
      />
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, { enableSlackIntegration })(OnBoard3);
