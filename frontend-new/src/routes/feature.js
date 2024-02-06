import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { WhiteListedAccounts, TestEnvs, SolutionsAccountId } from './constants';

export const featureLock = (activeAgent) =>
  (window.document.domain === 'app.factors.ai' &&
    WhiteListedAccounts.includes(activeAgent)) ||
  TestEnvs.includes(window.document.domain);

export const AdminLock = (activeAgent) => activeAgent === SolutionsAccountId;

export const ScrollToTop = () => {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);

  return null;
};
