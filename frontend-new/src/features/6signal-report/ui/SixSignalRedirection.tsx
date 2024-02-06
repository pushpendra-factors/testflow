import React, { useEffect } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { APP_LAYOUT_ROUTES } from 'Routes/constants';

const SixSignalRedirection = () => {
  const history = useHistory();
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  useEffect(() => {
    history.replace({
      pathname: APP_LAYOUT_ROUTES.VisitorIdentificationReport.path,
      search: `?${searchParams.toString()}`
    });
  }, []);
  return <></>;
};

export default SixSignalRedirection;
