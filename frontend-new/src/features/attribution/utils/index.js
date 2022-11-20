export const convertToEventOptions = (eventNames, display_names = []) => {
  // Based on groups Currently clubbing all into one, untill we have backend support
  const options = [];

  Object.keys(eventNames).forEach((key) => {
    const icon = key.toLowerCase().split(' ').join('_');
    options.push({
      label: key,
      icon: icon,
      values: eventNames[key].map((v) => {
        const name = display_names[v];
        return [name ? name : v, v];
      })
    });
  });
  return options;
};
