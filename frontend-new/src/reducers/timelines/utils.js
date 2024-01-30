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
        id: user.user_id,
        enabled: isEnabled,
        isGroupEvent: user.user_name === 'group_user'
      };
    });
    timelineArray.push(...newOpts);
  }
  return timelineArray;
};

export const formatAccountTimeline = (data, config) => {
  const milestones = data.milestones || {};
  const accountTimeline = data.account_timeline || [];
  const accountActivities = getAccountActivitiesWithEnableKeyConfig(
    accountTimeline,
    config?.disabled_events
  );

  const anonymousUsers = accountTimeline.filter((user) => user.is_anonymous);
  const anonymousUser = anonymousUsers.length
    ? [
        {
          title: 'Anonymous Users',
          subtitle: `${
            anonymousUsers.length === 1
              ? '1 Anonymous User'
              : `${
                  anonymousUsers.length > 25 ? '25+' : anonymousUsers.length
                } Anonymous Users`
          }`,
          userId: 'new_user',
          isAnonymous: true
        }
      ]
    : [];

  const mapUser = ({
    user_name: title,
    additional_prop: subtitle,
    user_id: userId,
    is_anonymous: isAnonymous
  }) => ({
    title,
    subtitle,
    userId,
    isAnonymous
  });

  const intentUser = accountTimeline
    .filter((user) => user.user_name === 'group_user')
    .map(mapUser);

  const nonAnonymousUsers = accountTimeline
    .filter((user) => !user.is_anonymous && user.user_name !== 'group_user')
    .map(({ user_activities, ...rest }) => ({
      ...mapUser(rest),
      lastEventAt: user_activities?.sort(compareObjTimestampsDesc)?.[0]
        ?.timestamp
    }))
    .sort((a, b) => b.lastEventAt - a.lastEventAt);

  const milestoneEvents = Object.entries(milestones).map(
    ([event_name, timestamp]) => ({
      event_name,
      timestamp,
      user: 'milestone'
    })
  );

  const account_events = [...accountActivities, ...milestoneEvents].sort(
    compareObjTimestampsDesc
  );

  return {
    name: data.name,
    host: data.host_name,
    leftpane_props: data.leftpane_props,
    overview: data.overview,
    account_users: [...nonAnonymousUsers, ...anonymousUser, ...intentUser],
    account_events
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

  return userPropsWithEnableKey.sort((a, b) => b.enabled - a.enabled);
};
