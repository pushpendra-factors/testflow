import { WhiteListedAccounts, TestEnvs } from 'Routes/constants';

export const featureLock = (activeAgent) => {
    return (
      (window.document.domain === 'app.factors.ai' &&
        WhiteListedAccounts.includes(activeAgent)) ||
        TestEnvs.includes(window.document.domain)
    );
};