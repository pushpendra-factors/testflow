import { useSelector } from 'react-redux';

const useAgentInfo = () => {
  const { isLoggedIn, agent_details, agents } = useSelector(
    (state) => state.agent
  );
  let isAdmin = false;
  if (agents && Array.isArray(agents)) {
    agents.forEach((agent) => {
      if (agent?.email === agent_details?.email && agent?.role == 2) {
        isAdmin = true;
      }
    });
  }
  return { isLoggedIn, email: agent_details?.email, isAdmin };
};

export default useAgentInfo;
