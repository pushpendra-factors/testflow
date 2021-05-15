/* eslint-disable */

// Data manipulation
export const convertToEventOptions = (eventNames, display_names = []) => {
  // Based on groups Currently clubbing all into one, untill we have backend support
  const options = [];
  
  Object.keys(eventNames).forEach(key => {
    options.push({
      label: key,
      icon: 'fav',
      values: eventNames[key].map(v => {
        const name = display_names[v];
        return [name? name : v, v]
      })
    })
  })
  return options
};

export const convertPropsToOptions = (props, display_names = []) => {
  const options = [];
  Object.keys(props).forEach((type) => {
    props[type].forEach((val) => {
      options.push([display_names[val]? display_names[val] : val, val, type]);
    })
  })
  return options;
}

const convertToChannelOptions = (objects) => {
  const opts = [];
  objects.forEach((obj, i) => {
    let lbl = obj.name.replace('_', ' ');
    const vals = obj.properties.map(v => [v.name, v.type])
    
    opts.push({
      label: lbl,
      icon: obj.name,
      values: vals
    });
  })
  return opts;
}

export const convertCampaignConfig = (data) => {
    const confg = {
      metrics: [],
      properties: []
    };


    confg.metrics = data.select_metrics;
    confg.properties = convertToChannelOptions(data.object_and_properties);

    return confg;
 }
