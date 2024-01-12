import { useCallback, useEffect, useState } from 'react';
import { paragon, SDK_EVENT } from '@useparagon/connect/dist/src/index';
import { message } from 'antd';

if (typeof window !== 'undefined') {
  window.paragon = paragon;
}

export default function useParagon(paragonUserToken) {
  const [token, setToken] = useState('');
  const [user, setUser] = useState(paragon.getUser());
  const [error, setError] = useState();
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    setToken(paragonUserToken);
  }, [paragonUserToken]);
  const updateUser = useCallback(() => {
    if (token) {
      const authedUser = paragon.getUser();
      if (authedUser.authenticated) {
        setUser({ ...authedUser });
      }
    }
  }, [token]);

  // Listen for account state changes
  useEffect(() => {
    paragon.subscribe(SDK_EVENT.ON_INTEGRATION_INSTALL, updateUser);
    paragon.subscribe('onIntegrationUninstall', updateUser);
    return () => {
      paragon.unsubscribe('onIntegrationInstall', updateUser);
      paragon.unsubscribe('onIntegrationUninstall', updateUser);
    };
  }, [token]);

  useEffect(() => {
    if (!error && token) {
      paragon
        .authenticate('60ef58ab-2e11-4f44-a388-b81aafca37ad', token)
        .then(() => {
          const authedUser = paragon.getUser();
          if (authedUser.authenticated) {
            setUser(authedUser);
          }
          setIsLoaded(true);
        })
        .catch((reason) => {
          setError(reason);
          setIsLoaded(true);
          message.error('Token Authentication Failed');
        });
    }
  }, [error, token]);

  return {
    paragon,
    user,
    error,
    updateUser,
    isLoaded
  };
}
