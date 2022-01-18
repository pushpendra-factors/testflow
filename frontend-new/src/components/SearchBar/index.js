import React, { useState, useCallback, useRef } from 'react';
import SearchModal from './SearchModal';
import { SVG } from '../factorsComponents';
import { Input } from 'antd';
import { useHistory } from 'react-router-dom';

function SearchBar({ setQueryToState }) {
  const inputRef = useRef(null);
  const [visible, setVisible] = useState(false);
  const history = useHistory();

  const handleFocus = useCallback(() => {
    document.activeElement.blur();
    setVisible(true);
  }, []);

  const handleClose = useCallback(() => {
    setVisible(false);
  }, []);


  const handleQueryClick = useCallback((query) => {
    if (history.location.pathname === '/analyse') {
      setQueryToState(query);
    } else {
      history.push({
        pathname: '/analyse',
        state: { query, global_search: true }
      });
    }
  }, [setQueryToState, history]);

  return (
    <> 
        <Input
          data-tour = 'step-2'
          ref={inputRef}
          size="large"
          placeholder="Lookup factors.ai"
          prefix={(
            <SVG name={'search'} size={16} color={'grey'} />
          )}
          className={'fa-global-search--input-new'}
          onFocus={handleFocus}
        /> 
      <SearchModal
        visible={visible}
        setVisible={setVisible}
        handleClose={handleClose}
        handleQueryClick={handleQueryClick}
      />
    </>

  );
}

export default SearchBar;
