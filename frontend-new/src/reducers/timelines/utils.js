export const compareObjTimestampsDesc = (a, b) => b.timestamp - a.timestamp;

export const getAccountActivitiesWithEnableKeyConfig = (
  accountTimeline = [],
  disabledEvents = []
) => {
  const timelineArray = [];
  for (const user of accountTimeline) {
    const newOpts = (user.user_activities || []).map((activity) => {
      const isEnabled = !disabledEvents.includes(activity.display_name);
      return {
        ...activity,
        user: user.is_anonymous ? 'new_user' : user.user_id,
        id:user.user_id,
        enabled: isEnabled
      };
    });
    timelineArray.push(...newOpts);
  }
  return timelineArray;
};

export const formatAccountTimeline = (data, config) => {
  const milestones = data.milestones || {};
  const account_timeline = data.account_timeline || [];
  const account_activities = getAccountActivitiesWithEnableKeyConfig(
    account_timeline,
    config?.disabled_events
  );

  const anonymous_users = account_timeline.filter((user) => user.is_anonymous);
  const anonymous_user = anonymous_users.length
    ? [
        {
          title: 'New Users',
          subtitle: `${
            anonymous_users.length === 1
              ? '1 New User'
              : `${anonymous_users.length} New Users`
          }`,
          userId: 'new_user',
          isAnonymous: true
        }
      ]
    : [];

  const is_intent_user = account_timeline.find(
    (user) => user.user_name === 'Channel Activity'
  );
  const intent_user = is_intent_user
    ? [
        {
          title: is_intent_user.user_name,
          subtitle: is_intent_user.additional_prop,
          userId: is_intent_user.user_id,
          isAnonymous: is_intent_user.is_anonymous
        }
      ]
    : [];

  const non_anonymous_users = account_timeline
    .filter(
      (user) => !user.is_anonymous && user.user_name !== 'Channel Activity'
    )
    .sort((a, b) =>
      compareObjTimestampsDesc(a.user_activities[0], b.user_activities[0])
    )
    .map(
      ({
        user_name: title,
        additional_prop: subtitle,
        user_id: userId,
        is_anonymous: isAnonymous
      }) => ({ title, subtitle, userId, isAnonymous })
    );

  const account_events = account_activities
    .concat(
      Object.entries(milestones).map(([event_name, timestamp]) => ({
        event_name,
        timestamp,
        user: 'milestone'
      }))
    )
    .sort(compareObjTimestampsDesc);

  return {
    name: data.name,
    host: data.host_name,
    left_pane_props: data.left_pane_props,
    account_users: [...non_anonymous_users, ...anonymous_user, ...intent_user],
    account_events
  };
};

const addEnabledFlagToActivity = (activity, disabledEvents) => {
  const enabled = !disabledEvents.includes(activity.display_name);
  return { ...activity, enabled };
};

export const addEnabledFlagToActivities = (activities, disabledEvents) => {
  return (
    activities?.map((activity) =>
      addEnabledFlagToActivity(activity, disabledEvents)
    ) || []
  );
};

export const formatUsersTimeline = (data, config) => {
  const returnData = {
    title: data.is_anonymous ? 'New User' : data.name || data.user_id,
    subtitle: data.company || data.user_id,
    left_pane_props: data.left_pane_props,
    group_infos: data.group_infos,
    user_activities: []
  };
  const arrayMilestones = [
    ...Object.entries(data?.milestones || {}).map(([key, value]) => {
      return { event_name: key, timestamp: value, event_type: 'milestone' };
    })
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
          type: type,
          enabled: activeProps ? activeProps.includes(propName) : false
        };
      })
    : [];

  return userPropsWithEnableKey.sort((a, b) => b.enabled - a.enabled);
};
