/* eslint-disable */

// Data manipulation
export const convertToEventOptions = (eventNames) => {
  // Based on groups Currently clubbing all into one, untill we have backend support

  return [{
    label: 'Frequently Asked',
    icon: 'fav',
    values: eventNames
  }];
};
