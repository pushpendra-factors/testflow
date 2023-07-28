/* eslint-disable */

import _ from 'lodash';

// Data manipulation
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

export const convertPropsToOptions = (props, display_names = []) => {
  const options = [];
  Object.keys(props).forEach((type) => {
    props[type].forEach((val) => {
      options.push([
        display_names[val] ? display_names[val] : val,
        val,
        type === 'unknown' ? 'categorical' : type
      ]);
    });
  });
  return options;
};

export const convertEventsPropsToOptions = (props, display_names = []) => {
  const options = {};
  Object.keys(props).forEach((type) => {
    const categoryOptions = props[type];
    Object.keys(categoryOptions).forEach((group) => {
      const groupOptions = categoryOptions[group];
      if (!options[group]) {
        options[group] = [];
      }
      groupOptions.forEach((val) => {
        options[group].push([
          display_names[val] ? display_names[val] : val,
          val,
          type === 'unknown' ? 'categorical' : type
        ]);
      });
    });
  });
  return options;
};
export const convertUserPropsToOptions = (
  props,
  display_names = [],
  disabledEventValues = []
) => {
  const userOptions = {};
  const eventUserOptions = {};
  Object.keys(props).forEach((type) => {
    const categoryOptions = props[type];
    Object.keys(categoryOptions).forEach((group) => {
      const groupOptions = categoryOptions[group];
      if (!userOptions[group]) {
        userOptions[group] = [];
      }
      if (!eventUserOptions[group]) {
        eventUserOptions[group] = [];
      }
      groupOptions.forEach((val) => {
        userOptions[group].push([
          display_names[val] ? display_names[val] : val,
          val,
          type
        ]);
        if (!disabledEventValues.includes(val)) {
          eventUserOptions[group].push([
            display_names[val] ? display_names[val] : val,
            val,
            type
          ]);
        }
      });
    });
  });
  return { userOptions, eventUserOptions };
};

export const convertCustomEventCategoryToOptions = (data) => {
  const mainItem = data.properties;
  const keys = Object.keys(mainItem);
  const finalArr = keys.map((type, index) => {
    let arr = mainItem[type].map((item) => {
      return [_.startCase(item), item, type];
    });
    return arr;
  });
  return _.flatten(finalArr);
};

const convertToChannelOptions = (objects) => {
  const opts = [];
  objects.forEach((obj, i) => {
    let lbl = obj.name.replace('_', ' ');
    const vals = obj.properties.map((v) => [v.name, v.type]);

    opts.push({
      label: lbl,
      icon: obj.name,
      values: vals
    });
  });
  return opts;
};

export const convertCampaignConfig = (data) => {
  const confg = {
    metrics: [],
    properties: []
  };

  confg.metrics = data.select_metrics;
  confg.properties = convertToChannelOptions(data.object_and_properties);

  return confg;
};

export const DEFAULT_TOUCHPOINTS = [
  {
    label: 'Campaign',
    key: 'Campaign'
  },
  {
    label: 'Source',
    key: 'Source'
  },
  {
    label: 'AdGroup',
    key: 'AdGroup'
  },
  {
    label: 'Keyword',
    key: 'Keyword'
  },
  {
    label: 'Channel',
    key: 'ChannelGroup'
  },
  {
    label: 'Landing Page',
    key: 'LandingPage'
  },
  {
    label: 'All Page Views',
    key: 'AllPageView'
  }
];

export function getTouchPointLabel(key) {
  return (
    DEFAULT_TOUCHPOINTS.find((touchPoint) => touchPoint.key === key)?.label ||
    ''
  );
}
