export const compareObjTimestampsDesc = (a, b) => b.timestamp - a.timestamp;

export const getAccountActivitiesWithEnableKey = (accountTimeline = []) => {
  const timelineArray = [];

  accountTimeline.forEach((user) => {
    const newOpts = (user.user_activities || []).map((event) => ({
      id: event.event_id,
      name: event.event_name,
      display_name: event.display_name,
      alias_name: event.alias_name,
      icon: event.icon,
      type: event.event_type,
      timestamp: event.timestamp,
      username: user.is_anonymous ? 'new_user' : user.user_name,
      user_id: user.user_id,
      is_group_user: user.user_name === 'group_user',
      is_anonymous_user: event.username === 'new_user',
      properties: event.properties || [],
      user_properties: user.properties || [],
      enabled: true
    }));
    timelineArray.push(...newOpts);
  });

  return timelineArray;
};

const mapUser = ({
  user_name: name,
  additional_prop: extraProp,
  user_id: id,
  is_anonymous: isAnonymous,
  user_properties: properties
}) => ({
  name,
  extraProp,
  id,
  isAnonymous,
  properties
});

export const formatAccountTimeline = (data) => {
  const milestones = data.milestones || {};
  const accountTimeline = data.account_timeline || [];
  const accountActivities = getAccountActivitiesWithEnableKey(accountTimeline);

  const accountUser = accountTimeline
    .filter((user) => user.user_name === 'group_user')
    .map(mapUser);

  const webUsers = accountTimeline
    .filter((user) => user.user_name !== 'group_user')
    .map(({ user_activities, ...rest }) => ({
      ...rest,
      lastEventAt: user_activities?.sort(compareObjTimestampsDesc)?.[0]
        ?.timestamp
    }))
    .sort((a, b) => b.lastEventAt - a.lastEventAt)
    .map(mapUser);

  const milestoneEvents = Object.entries(milestones).map(
    ([event_name, timestamp]) => ({
      name: event_name,
      timestamp,
      username: 'milestone',
      user_id: 'milestone'
    })
  );

  const events = [...accountActivities, ...milestoneEvents].sort(
    compareObjTimestampsDesc
  );

  return {
    name: data.name,
    domain: data.domain_name,
    leftpane_props: data.leftpane_props,
    overview: data.overview,
    users: [...webUsers, ...accountUser],
    events
  };
};

const addEnabledFlagToActivity = (activity, disabledEvents = []) => {
  const enabled = !disabledEvents.includes(activity.display_name);
  return { ...activity, enabled };
};

export const addEnabledFlagToActivities = (activities, disabledEvents) =>
  activities
    ?.map((activity) => addEnabledFlagToActivity(activity, disabledEvents))
    ?.sort((a, b) => b.enabled - a.enabled) || [];

export const formatUsersTimeline = (data, config) => {
  const returnData = {
    title: data.is_anonymous ? 'New User' : data.name || data.user_id,
    subtitle: data.company || data.user_id,
    leftpane_props: data.leftpane_props,
    account: data.account,
    user_activities: []
  };
  const arrayMilestones = [
    ...Object.entries(data?.milestones || {}).map(([key, value]) => ({
      event_name: key,
      timestamp: value,
      event_type: 'milestone'
    }))
  ];
  returnData.user_activities = addEnabledFlagToActivities(
    data.user_activities,
    config?.disabled_events
  )
    .concat(arrayMilestones)
    .sort(compareObjTimestampsDesc);
  return returnData;
};

export const formatUserPropertiesToCheckList = (
  userProps,
  activeProps = []
) => {
  const userPropsWithEnableKey = userProps
    ? userProps.map((userProp) => {
        const [displayName, propName, type] = userProp;
        return {
          display_name: displayName,
          prop_name: propName,
          type,
          enabled: activeProps ? activeProps.includes(propName) : false
        };
      })
    : [];
  // We can do it other way too.
  // by just reordering the elements of userPropsWithEnableKey => This way would be more better.
  // OR
  // we can just unshift the enabled rows in the correct order
  const matchedPropsInOrder = [];
  activeProps?.forEach((eachProp) => {
    const ele = userPropsWithEnableKey.find((e) => eachProp === e.prop_name);
    if (ele) {
      matchedPropsInOrder.push(ele);
    }
  });

  const result = [
    ...matchedPropsInOrder,
    ...userPropsWithEnableKey.filter((e) => !e.enabled)
  ];

  return result;
};
