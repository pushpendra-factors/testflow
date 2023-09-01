import { useSelector } from 'react-redux';

const useAgentInfo = () => {
  const { isLoggedIn, agent_details, agents } = useSelector(
    (state) => state.agent
  );
  let isAdmin = false;
  let isAgentInvited = false;
  if (agents && Array.isArray(agents)) {
    agents.forEach((agent) => {
      if (agent?.email === agent_details?.email) {
        if (agent?.role == 2) {
          isAdmin = true;
        }
        if (agent?.invited_by) {
          isAgentInvited = true;
        }
      }
    });
  }
  return {
    isLoggedIn,
    email: agent_details?.email,
    isAdmin,
    firstName: agent_details?.first_name,
    lastName: agent_details?.last_name,
    isAgentInvited
  };
};

export default useAgentInfo;
