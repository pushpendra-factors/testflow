import { useState, useEffect } from 'react';

const useService = (projectId, serviceClass) => {
  const [val, setVal] = useState(undefined);
  useEffect(() => {
    if(projectId){
        setVal(new serviceClass(null, projectId));
    }
  }, [projectId])
  return val;
};

export default useService;