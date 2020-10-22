/* eslint-disable */

// Data manipulation
export const convertToEventOptions = (eventNames) => {
  // Based on groups Currently clubbing all into one, untill we have backend support
  const options = [];
  Object.keys(eventNames).forEach(key => {
    options.push({
      label: key,
      icon: 'fav',
      values: eventNames[key].map(v => [v])
    })
  })
  return options
};

export const convertPropsToOptions = (props) => {
  const options = [];
  Object.keys(props).forEach((type) => {
    props[type].forEach((val) => {
      options.push([val, type]);
    })
  })
  return options;
}
