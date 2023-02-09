import { useSelector } from 'react-redux';

const useAgentInfo = () => {
  const { isLoggedIn, agent_details } = useSelector((state) => state.agent);

  return { isLoggedIn, email: agent_details?.email, isAdmin: isLoggedIn };
};

export default useAgentInfo;
