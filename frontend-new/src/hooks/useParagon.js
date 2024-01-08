import { useCallback, useEffect, useState } from 'react';
import { paragon, SDK_EVENT } from '@useparagon/connect/dist/src/index';

if (typeof window !== 'undefined') {
  window.paragon = paragon;
}

export default function useParagon(paragonUserToken) {
  const [user, setUser] = useState(paragon.getUser());
  const [error, setError] = useState();

  const updateUser = useCallback(() => {
    const authedUser = paragon.getUser();
    if (authedUser.authenticated) {
      setUser({ ...authedUser });
    }
  }, []);

  // Listen for account state changes
  useEffect(() => {
    paragon.subscribe(SDK_EVENT.ON_INTEGRATION_INSTALL, updateUser);
    paragon.subscribe('onIntegrationUninstall', updateUser);
    return () => {
      paragon.unsubscribe('onIntegrationInstall', updateUser);
      paragon.unsubscribe('onIntegrationUninstall', updateUser);
    };
  }, []);

  useEffect(() => {
    if (!error) {
      paragon
        .authenticate('60ef58ab-2e11-4f44-a388-b81aafca37ad', paragonUserToken)
        .then(() => {
          const authedUser = paragon.getUser();
          if (authedUser.authenticated) {
            setUser(authedUser);
          }
        })
        .catch(setError);
    }
  }, [error, paragonUserToken]);

  return {
    paragon,
    user,
    error,
    updateUser
  };
}
