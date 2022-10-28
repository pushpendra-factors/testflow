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
    industry: data.industry,
    country: data.country,
    number_of_employees: data.number_of_employees,
    number_of_users: data.number_of_users,
    account_users: [],
    account_events: []
  };
  returnData.account_users = data.account_timeline
    ?.sort((a, b) =>
      compareObjTimestampsDesc(a.user_activities[0], b.user_activities[0])
    )
    .map((user) => user.user_name);
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
    title: !data.is_anonymous ? data.name || '-' : 'Unidentified User',
    subtitle: data.company || data.user_id,
    email: data.email,
    country: data.country,
    web_session_count: data.web_sessions_count,
    number_of_page_views: data.number_of_page_views,
    time_spent_on_site: data.time_spent_on_site,
    group_infos: data.group_infos,
    user_activities: []
  };
  returnData.user_activities = getActivitiesWithEnableKeyConfig(
    data.user_activities,
    config?.disabled_events
  ).sort(compareObjTimestampsDesc);
  return returnData;
};
