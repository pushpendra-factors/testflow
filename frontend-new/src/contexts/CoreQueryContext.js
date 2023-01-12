import React from 'react';
import { CORE_QUERY_INITIAL_STATE } from 'Views/CoreQuery/constants';

export const CoreQueryContext = React.createContext({
  coreQueryState: CORE_QUERY_INITIAL_STATE
});
