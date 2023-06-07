import { useEffect } from 'react';

const useKey = (targetKey, callback) => {
  useEffect(() => {
    // Event handler for keydown event
    const handleKeyDown = (event) => {
      if (event.key === targetKey) {
        callback();
      }
    };

    // Add event listener
    document.addEventListener('keydown', handleKeyDown);

    // Clean up the event listener
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [targetKey, callback]);
};

export default useKey;
