import React, { useState, useCallback, useRef } from 'react';
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

  const handleQueryClick = useCallback((query) => {
    if (history.location.pathname === '/core-analytics') {
      setQueryToState(query);
    } else {
      history.push({
        pathname: '/core-analytics',
        state: { query, global_search: true }
      });
    }
  }, [setQueryToState, history]);

  return (
    <>
      {!visible ? (
        <Input
          ref={inputRef}
          size="large"
          placeholder="Lookup factors.ai"
          prefix={(
            <SVG name={'search'} size={24} color={'black'} />
          )}
          className={styles.searchBarBox}
          onFocus={handleFocus}
        />
      ) : null}
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
