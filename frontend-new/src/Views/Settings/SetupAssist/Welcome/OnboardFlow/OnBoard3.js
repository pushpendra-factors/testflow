import { Button, Row, message } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';
import { connect, useSelector } from 'react-redux';
import { enableSlackIntegration } from 'Reducers/global';
import styles from './index.module.scss';
const OnBoard3 = ({ enableSlackIntegration }) => {
  const activeProject = useSelector((state) => state?.global?.active_project);
  const int_slack = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_slack
  );
  const onConnectSlack = () => {
    enableSlackIntegration(activeProject.id)
      .then((r) => {
        if (r.status === 200) {
          window.open(r.data.redirectURL, '_blank');
        }
        if (r.status >= 400) {
          message.error('Error fetching slack redirect url');
        }
      })
      .catch((err) => {
        console.log('Slack error-->', err);
      });
  };
  return (
    <div className={styles['onBoardContainer']}>
      {/* <SixSignal setIsActive={() => {}} kbLink={true} /> */}
      <div>
        <Text type={'title'} level={6} weight={'bold'}>
          Get key alerts on slack
        </Text>
        <Row className={styles['horizontalCard']}>
          <div className={styles['horizontalCardContent']}>
            <div className={styles['horizontalCardLeft']}>
              <div style={{ display: 'grid', placeContent: 'center' }}>
                {' '}
                <SVG
                  name={'Slack'}
                  size={40}
                  extraClass={'inline mr-2 -mt-2'}
                />
              </div>
              <div>
                <Text
                  type={'title'}
                  level={6}
                  weight={'bold'}
                  style={{ margin: 0 }}
                >
                  Slack
                </Text>
                <div>
                  Get alerts when high-intent actions take place by your
                  prospects. Close more deals by being closest to the action.
                </div>
              </div>
            </div>
            <div className={styles['horizontalCardRight']}>
              <Button onClick={int_slack ? null : onConnectSlack}>
                {int_slack ? (
                  <>
                    <SVG name='Greentick' /> Already Connected
                  </>
                ) : (
                  'Connect'
                )}
              </Button>
            </div>
          </div>
        </Row>
      </div>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, { enableSlackIntegration })(OnBoard3);
