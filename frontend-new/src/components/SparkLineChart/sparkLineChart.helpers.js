export const displayQueryName = ({ query, eventNames }) => {
  const queryTitle = eventNames[query] || query;
  return queryTitle;
};
