import { useEffect, useState } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import useQuery from './useQuery';

const useTabs = (defaultActiveTab: string) => {
  const [activeKey, setActiveKey] = useState(defaultActiveTab);
  const routerQuery = useQuery();
  const history = useHistory();
  const location = useLocation();
  const paramActiveTab = routerQuery.get('activeTab');
  const handleTabChange = (key: string) => {
    setActiveKey(key);
    history.replace(`${location.pathname}?activeTab=${key}`);
  };
  useEffect(() => {
    if (!paramActiveTab) {
      history.replace(`${location.pathname}?activeTab=${activeKey}`);
    }
    if (paramActiveTab && activeKey !== paramActiveTab) {
      setActiveKey(paramActiveTab);
    }
  }, [paramActiveTab, activeKey]);

  return { activeKey, handleTabChange };
};

export default useTabs;
