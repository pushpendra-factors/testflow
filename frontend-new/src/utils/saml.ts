import { getBackendHost } from 'Views/Settings/ProjectSettings/IntegrationSettings/util';
import { message } from 'antd';
import logger from './logger';

const host = getBackendHost();

export function getSAMLValidateURL(email: string) {
  return `${host}/saml/login/validate?email=${encodeURIComponent(email)}`;
}

export function getSAMLRedirectURL(email: string) {
  return `${host}/saml/login?email=${encodeURIComponent(email)}`;
}

/*
    This function handles the logic to Redirection 
    with All the loading state
    1. Validation
    2. Redirection
*/
export function redirectSAMLProject(email: string, project: any = {}) {
  const messageLoadingHandle = message.loading('Checking SAML Account', 0);
  const url1 = getSAMLValidateURL(email);
  const url2 = getSAMLRedirectURL(email);
  fetch(url1)
    .then((res) => {
      // window.location.href = url2;
      if (res.status === 200) {
        messageLoadingHandle();
        message.success('Redirecting to your SAML');
        window.location.href = url2;
      } else {
        messageLoadingHandle();
        message.error("SAML Account Doesn't Exists");
      }
    })
    .catch((err) => {
      messageLoadingHandle();
      message.error("SAML Account Doesn't Exists");
      logger.log(err);
    });
}
