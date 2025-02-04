import { PathUrls } from 'Routes/pathUrls';
import { message } from 'antd';
import _ from 'lodash';

export const EMPTY_FUNCTION = () => {};

export const EMPTY_STRING = '';

export const EMPTY_OBJECT = {};

export const EMPTY_ARRAY = [];

export const isStringLengthValid = (str, length = 1) =>
  _.size(_.trim(str)) >= length;

export const groupFilters = (array, key) =>
  // Return the end result
  array.reduce((result, currentValue) => {
    // If an array already present for key, push it to the array. Else create an array and push the object
    (result[currentValue[key]] = result[currentValue[key]] || []).push(
      currentValue
    );
    // Return the current iteration `result` value, this will be taken as next iteration `result` value and accumulate
    return result;
  }, {}); // empty object is the initial value for result object

export const compareFilters = (a, b) => {
  if (a.ref < b.ref) return -1;
  if (a.ref > b.ref) return 1;
  return 0;
};

export const toCapitalCase = (str) => {
  const lower = str.toLowerCase();
  return lower.charAt(0).toUpperCase() + lower.slice(1);
};

export function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

export const abbreviateNumber = (n) => {
  if (n < 1e3) return n;
  if (n >= 1e3 && n < 1e6) return `${+(n / 1e3).toFixed(1)}K`;
  if (n >= 1e6 && n < 1e9) return `${+(n / 1e6).toFixed(1)}M`;
  if (n >= 1e9 && n < 1e12) return `${+(n / 1e9).toFixed(1)}B`;
  if (n >= 1e12) return `${+(n / 1e12).toFixed(1)}T`;
  return null;
};

export function generateRandomKey(length = 8) {
  var result = '';
  var characters =
    'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  var charactersLength = characters.length;
  for (var i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}

export function insertUrlParam(history, key, value) {
  if (history.pushState) {
    let searchParams = new URLSearchParams(window.location.search);
    searchParams.set(key, value);
    let newurl =
      window.location.protocol +
      '//' +
      window.location.host +
      window.location.pathname +
      '?' +
      searchParams.toString();
    window.history.pushState({ path: newurl }, '', newurl);
  }
}

export function isOnboarded(currentProjectSettings) {
  return (
    currentProjectSettings?.onboarding_steps?.more_info_form &&
    currentProjectSettings?.onboarding_steps?.project_created &&
    currentProjectSettings?.onboarding_steps?.sdk_setup &&
    currentProjectSettings?.onboarding_steps?.setup_completed &&
    currentProjectSettings?.onboarding_steps?.visitor_identification_setup
  );
}

export function getErrorMessage(resultState) {
  let errorMessage = '';

  if (resultState?.status === 500) {
    errorMessage =
      'The server encountered an internal error and could not complete your request';
  } else if (resultState?.status === 400) {
    errorMessage = '400 Bad Request. Please check your request parameters';
  } else if (resultState?.status === 404) {
    errorMessage = 'Resource Not Found! Please check your request.';
  } else {
    // check for no data found
    if (resultState?.status === 200) {
      if (!resultState?.data || resultState?.data?.length === 0) {
        errorMessage = 'No Data Found! Try a Different Time Range';
      } else if (resultState?.data?.metrics?.rows?.length === 0) {
        errorMessage = 'No Data Found! Try Changing Filters or Time Range';
      }
    } else {
      errorMessage =
        'We are facing trouble loading UI. Drop us a message on the in-app chat';
    }
  }

  return errorMessage;
}

export function getCookieValue(cookieName) {
  const name = cookieName + '=';
  const decodedCookie = decodeURIComponent(document.cookie);
  const cookieArray = decodedCookie.split(';');

  for (let i = 0; i < cookieArray.length; i++) {
    let cookie = cookieArray[i].trim();
    if (cookie.indexOf(name) === 0) {
      return cookie.substring(name.length, cookie.length);
    }
  }

  // Return null if the cookie with the specified name is not found
  return null;
}

export function getASCIISum(str) {
  let sum = 0;
  for (let i = 0; i < str.length; i++) {
    for (let j = 0; j < str[i].length; j++) {
      let char_code = str[i].charCodeAt(j);
      sum += char_code;
    }
  }
  return sum;
}

export async function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

export async function CopyTextToClipboard(text = '') {
  try {
    await navigator.clipboard.writeText(text);
    message.success('Successfully copied');
  } catch (err) {
    message.error('Failed to copy!');
  }
}

// This Selects Project which one to give Access
export function getAccessBasedOnLoginMethod(login_method = 0, projects = []) {
  switch (login_method) {
    case 1: {
      // Email CASE
      return projects.filter((e) => e.login_state === 1).length > 0;
    }
    case 3:
      // GOOGLE Case
      return (
        projects.filter((e) => e.login_state === 1 || e.login_state === 3)
          .length > 0
      );
    case 2: {
      return true;
    }
    default:
      return false;
  }
}

export function checkAccessBasedForCurrentProject(login_method, activeProject) {
  switch (login_method) {
    case 1:
      return activeProject?.login_method === 1;
    case 2:
      return activeProject?.login_method === 2;
    case 3:
      return (
        activeProject?.login_method === 1 || activeProject?.login_method === 3
      );
    case false:
      return false;

    default:
      return undefined;
  }
}

// THis checks if we have access to either of the prjects
// IF we don't this locks us to ProjectAuthChange Page
// And not Allows user to move to any other page
export function accessAuthCheck(login_method, projects) {
  switch (login_method) {
    case 1:
      return projects.find((e) => e.login_method === 1);
    case 2:
      return projects.find((e) => e.login_method === 2);
    case 3:
      return projects.find((e) => e.login_method === 1 || e.login_method === 3);
    case 0:
      return false;
    default:
      return undefined;
  }
}
export function haveRestrictionForSelectedProject(
  currentMethod,
  selectedMethod
) {
  switch (currentMethod) {
    case 1:
      return selectedMethod === 1;
    case 2:
      return false;
    case 3:
      return selectedMethod === 1 || selectedMethod === 3;
    default:
      return undefined;
  }
}
export const WHITELIST_INTERMEDIATE_STATES = new Set();
WHITELIST_INTERMEDIATE_STATES.add(PathUrls.ProjectChangeAuthentication);
WHITELIST_INTERMEDIATE_STATES.add(PathUrls.Onboarding);
