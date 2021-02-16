import React, { useState, useCallback, useRef, useEffect } from 'react';
import SearchModal from './SearchModal';
import { SVG } from '../factorsComponents';
import styles from './index.module.scss';
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


  useEffect(() => {
    document.onkeydown = keydown;
    function keydown(evt) {  
      // cmd+K to trigger global search
      if (evt.keyCode === 75) {   
        setVisible(true); 
      }
    } 
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
          ref={inputRef}
          size="large"
          placeholder="Search Factors ⌘K"
          prefix={(
            <SVG name={'search'} size={16} color={'grey'} />
          )}
          // className={styles.searchBarBox}
          className={'fa-global-search--input'}
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
