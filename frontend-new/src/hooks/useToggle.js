import { useState, useCallback } from 'react';

const useToggle = (initialValue) => {
  const [val, setVal] = useState(initialValue);

  const handleToggle = useCallback(() => {
    setVal((curr) => !curr);
  }, []);

  return [val, handleToggle];
};

export default useToggle;
