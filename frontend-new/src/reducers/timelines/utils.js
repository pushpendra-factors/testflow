export const compareObjTimestampsDesc = (a, b) => {
  if (a.timestamp > b.timestamp) {
    return -1;
  }
  if (a.timestamp < b.timestamp) {
    return 1;
  }
  return 0;
};

export const formattedResponseData = (data) => {
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
  const nameArrayForAllUsers = data.account_timeline
    ?.sort((a, b) =>
      compareObjTimestampsDesc(a.user_activities[0], b.user_activities[0])
    )
    .map((user) => user.user_name);
  const timelineArray = [];
  data.account_timeline?.forEach((user) => {
    const newOpts = user.user_activities.map((activity) => ({
      ...activity,
      user: user.user_name,
      enabled: true
    }));
    timelineArray.push(...newOpts);
  });
  returnData.account_users = nameArrayForAllUsers;
  returnData.account_events = timelineArray.sort(compareObjTimestampsDesc);
  return returnData;
};
