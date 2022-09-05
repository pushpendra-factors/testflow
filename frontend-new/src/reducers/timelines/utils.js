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
  let returnData = {
    name: data.name,
    industry: data.industry,
    country: data.country,
    number_of_employees: data.number_of_employees,
    number_of_users: data.number_of_users,
    account_users: [],
    account_events: [],
  };
  const nameArrayForAllUsers = data.account_timeline?.map((data) => {
    return data.user_name;
  });
  const timelineArray = [];
  data.account_timeline?.forEach((user) => {
    const newOpts = user.user_activities.map((data) => {
      return { ...data, user: user.user_name, enabled: true };
    });
    timelineArray.push(...newOpts);
  });
  returnData.account_users = nameArrayForAllUsers;
  returnData.account_events = timelineArray.sort(compareObjTimestampsDesc);
  return returnData;
};
