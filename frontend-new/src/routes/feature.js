import { WhiteListedAccounts, TestEnvs } from 'Routes/constants';
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export const featureLock = (activeAgent, noEnvCheck = false) => {
  if(noEnvCheck){
    return WhiteListedAccounts.includes(activeAgent)
  }
  else{
    return (
      (window.document.domain === 'app.factors.ai' &&
        WhiteListedAccounts.includes(activeAgent)) ||
      TestEnvs.includes(window.document.domain)
    ); 
  }
};

export const AdminLock = (activeAgent) => {
  return activeAgent === 'solutions@factors.ai';
};

export const ScrollToTop = () => {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);

  return null;
};
