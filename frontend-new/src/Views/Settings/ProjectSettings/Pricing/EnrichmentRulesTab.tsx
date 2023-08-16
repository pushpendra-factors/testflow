import { Text } from 'Components/factorsComponents';
import { Alert, Divider, Radio, RadioChangeEvent, Modal } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { SixSignalConfigType } from '../IntegrationSettings/SixSignalFactors/types';
import EnrichFeature from '../IntegrationSettings/SixSignalFactors/EnrichFeature';
import { bindActionCreators } from 'redux';
import { connect, ConnectedProps } from 'react-redux';

import { udpateProjectSettings } from 'Reducers/global';
import logger from 'Utils/logger';
import { isEmpty } from 'lodash';

const { confirm } = Modal;

const EnrichmentRulesTab = ({
  udpateProjectSettings
}: EnrichmentRulesPropsType) => {
  const { isAdmin } = useAgentInfo();
  const six_signal_config: SixSignalConfigType = useSelector(
    (state) => state?.global?.currentProjectSettings?.six_signal_config
  );
  const active_project = useSelector((state) => state.global.active_project);

  const [enrichmentType, setEnrichmentType] = useState<boolean | null>(null);

  const handleEnrichmentChange = (e: RadioChangeEvent) => {
    if (e.target.value === false) {
      if (!six_signal_config || isEmpty(six_signal_config)) {
        setEnrichmentType(false);
      } else {
        confirm({
          title: 'Confirmation',
          content: `Are you sure you want to remove the Enrichment Rules?`,
          async onOk() {
            try {
              await udpateProjectSettings(active_project?.id, {
                six_signal_config: {}
              });
            } catch (error) {
              logger.error('Error in updating project settings', error);
            }
          },
          onCancel() {
            // Reset the switch value to the previous one
          }
        });
      }
    } else {
      setEnrichmentType(e.target.value);
    }
  };

  useEffect(() => {
    if (!six_signal_config || isEmpty(six_signal_config)) {
      setEnrichmentType(false);
    } else {
      setEnrichmentType(true);
    }
  }, [six_signal_config]);

  return (
    <div className='py-4' style={{ pointerEvents: !isAdmin ? 'none' : 'all' }}>
      {!isAdmin && (
        <div className='my-8'>
          <Alert
            message={
              <Text type={'paragraph'} mini color='character-title'>
                Only admin has access to edit this function. To make more
                modifications, get in touch with admin.
              </Text>
            }
            type='info'
            showIcon
          />
        </div>
      )}
      <div className='mb-6'>
        <Text
          type={'title'}
          level={4}
          weight={'bold'}
          extraClass={'m-0 mb-2'}
          color='character-primary'
        >
          Set up rules for visitor identification
        </Text>
        <Text
          type={'title'}
          level={7}
          extraClass={'m-0'}
          color='character-secondary'
        >
          You can choose identify all accounts that visit your website or set
          custom rules to identify only some of them.
        </Text>
      </div>
      <Divider />
      <div className='mb-8'>
        <Radio.Group onChange={handleEnrichmentChange} value={enrichmentType}>
          <Radio value={false}>Identify all accounts</Radio>
          <Radio value={true}>Set custom rules</Radio>
        </Radio.Group>
      </div>
      {enrichmentType === false && (
        <Text
          type={'title'}
          level={6}
          extraClass='m-0'
          color='character-primary'
        >
          Identify all accounts that visit your website. This ensures that you
          don’t miss out on any account. This affects your monthly quota of
          accounts.
        </Text>
      )}
      {enrichmentType && (
        <>
          <div className='mt-4'>
            <EnrichFeature
              type='page'
              title='Identify accounts who visited specific pages'
              subtitle='Include or exclude pages to only identify accounts that visit the pages you care about'
              actionButtonText='Select pages'
            />
          </div>
          <div className='mt-4'>
            <EnrichFeature
              type='country'
              title='Identify accounts only from selected countries/region'
              subtitle='Include or exclude countries to only identify accounts from the geography you care about'
              actionButtonText='Select Countries '
            />
          </div>
        </>
      )}
    </div>
  );
};
const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ udpateProjectSettings }, dispatch);
const connector = connect(null, mapDispatchToProps);
type EnrichmentRulesPropsType = ConnectedProps<typeof connector>;

export default connector(EnrichmentRulesTab);
