import React, { useEffect } from 'react';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { Spin } from 'antd';
import { isEmpty } from 'lodash';

import AttributionSetupDone from './AttributionSetupDone';
import AttributionSetupPending from './AttributionSetupPending';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';

function AttributionBaseComponent({
  currentProjectSettings,
  currentProjectSettingsLoading
}) {
  const history = useHistory();

  useEffect(() => {
    if (
      !currentProjectSettingsLoading &&
      currentProjectSettings &&
      !isEmpty(currentProjectSettings)
    ) {
      if (currentProjectSettings?.attribution_config) {
        history.replace(ATTRIBUTION_ROUTES.reports);
      }
    }
  }, [currentProjectSettings, currentProjectSettingsLoading]);

  if (currentProjectSettingsLoading)
    return (
      <div className='flex items-center justify-center h-full w-full'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  // TBD attribution precompute screen
  // const setupDone = false;
  // if (setupDone) return <AttributionSetupDone />;
  return <AttributionSetupPending />;
}

const mapStateToProps = (state) => ({
  currentProjectSettings: state.global.currentProjectSettings,
  currentProjectSettingsLoading: state.global.currentProjectSettingsLoading
});

export default connect(mapStateToProps, null)(AttributionBaseComponent);
