export const compareObjTimestampsDesc = (a, b) => {
  if (a.timestamp > b.timestamp) {
    return -1;
  }
  if (a.timestamp < b.timestamp) {
    return 1;
  }
  return 0;
};

export const getAccountActivitiesWithEnableKeyConfig = (
  accountTimeline,
  disabledEvents
) => {
  const timelineArray = [];
  accountTimeline?.forEach((user) => {
    const newOpts = user.user_activities.map((activity) => {
      let isEnabled = true;
      if (disabledEvents?.includes(activity.display_name)) {
        isEnabled = false;
      }
      return { ...activity, user: user.user_name, enabled: isEnabled };
    });
    timelineArray.push(...newOpts);
  });
  return timelineArray;
};

export const formatAccountTimeline = (data, config) => {
  const returnData = {
    name: data.name,
    host: data.host_name,
    left_pane_props: data.left_pane_props,
    account_users: [],
    account_events: []
  };
  returnData.account_users = data.account_timeline
    ?.sort((a, b) =>
      compareObjTimestampsDesc(a.user_activities[0], b.user_activities[0])
    )
    .map((user) => ({ title: user.user_name, subtitle: user.additional_prop }));
  returnData.account_events = getAccountActivitiesWithEnableKeyConfig(
    data?.account_timeline,
    config?.disabled_events
  ).sort(compareObjTimestampsDesc);
  return returnData;
};

export const getActivitiesWithEnableKeyConfig = (
  activities,
  disabledEvents = []
) =>
  activities?.map((activity) => {
    let isEnabled = true;
    if (disabledEvents?.includes(activity.display_name)) {
      isEnabled = false;
    }
    return { ...activity, enabled: isEnabled };
  });

export const formatUsersTimeline = (data, config) => {
  const returnData = {
    title: !data.is_anonymous ? data.name || data.user_id : 'Unidentified User',
    subtitle: data.company || data.user_id,
    left_pane_props: data.left_pane_props,
    group_infos: data.group_infos,
    user_activities: []
  };
  returnData.user_activities = getActivitiesWithEnableKeyConfig(
    data.user_activities,
    config?.disabled_events
  ).sort(compareObjTimestampsDesc);
  return returnData;
};

export const formatUserPropertiesToCheckList = (
  userProps,
  activeProps = []
) => {
  const userPropsWithEnableKey = userProps.map((userProp) => {
    const retObj = {
      display_name: userProp[0],
      prop_name: userProp[1],
      type: userProp[2],
      enabled: false
    };
    if (activeProps.includes(userProp[1])) {
      retObj.enabled = true;
    }
    return retObj;
  });
  return userPropsWithEnableKey.sort((a, b) => b.enabled - a.enabled);
};
