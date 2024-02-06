import React, { useEffect } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const SixSignalRedirection = () => {
  const history = useHistory();
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  useEffect(() => {
    history.replace({
      pathname: PathUrls.VisitorIdentificationReport,
      search: `?${searchParams.toString()}`
    });
  }, []);
  return <></>;
};

export default SixSignalRedirection;
