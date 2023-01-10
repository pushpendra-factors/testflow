import React, { useState } from 'react';
import { FunnelsConversionDurationBlockComponent } from '../FunnelsConversionDurationBlock';

export default {
  title: 'Components/FunnelsConversionDurationBlock',
  component: FunnelsConversionDurationBlockComponent
};

export const Default = () => {
  const [state, setState] = useState({
    funnelConversionDurationNumber: 30,
    funnelConversionDurationUnit: 'D'
  });

  const onChange = (payload) => {
    setState((currState) => {
      return {
        ...currState,
        ...payload
      };
    });
  };

  return (
    <FunnelsConversionDurationBlockComponent
      funnelConversionDurationNumber={state.funnelConversionDurationNumber}
      funnelConversionDurationUnit={state.funnelConversionDurationUnit}
      onChange={onChange}
    />
  );
};
