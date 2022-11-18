import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { useHistory } from 'react-router-dom';
import { Spin } from 'antd';
import { fetchProjectSettings } from 'Reducers/global';

import AttributionSetupDone from './AttributionSetupDone';
import AttributionSetupPending from './AttributionSetupPending';

function AttributionBaseComponent({ activeProject, fetchProjectSettings }) {
  const [loading, setLoading] = useState(true);
  const history = useHistory();

  useEffect(() => {
    const checkRedirection = async () => {
      const res = await fetchProjectSettings(activeProject?.id);
      if (res?.data?.attribution_config) {
        history.replace('/attribution/reports');
      }
      setLoading(false);
    };
    if (activeProject) {
      checkRedirection();
    }
  }, [activeProject]);

  if (loading)
    return (
      <div className='flex items-center justify-center h-full w-full'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );

  const setupDone = false;
  if (setupDone) return <AttributionSetupDone />;
  return <AttributionSetupPending />;
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectSettings
    },
    dispatch
  );

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(AttributionBaseComponent);
