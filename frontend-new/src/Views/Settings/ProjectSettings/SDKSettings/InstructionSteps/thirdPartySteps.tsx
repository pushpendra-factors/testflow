import React from 'react';
import CollapsibleContainer from 'Components/GenericComponents/CollapsibleContainer';
import Header from 'Components/GenericComponents/CollapsibleContainer/CollasibleHeader';
import ThirdPartyStepsBody from './thirdPartyStepsBody';

const ThirdPartySteps = () => (
  <CollapsibleContainer
    showBorder
    key='thirdparty'
    BodyComponent={<ThirdPartyStepsBody />}
    HeaderComponent={
      <Header
        title='Use a Customer Data Platform'
        description='Use an existing Customer Data Platform to bring in website data and events'
      />
    }
  />
);

export default ThirdPartySteps;
