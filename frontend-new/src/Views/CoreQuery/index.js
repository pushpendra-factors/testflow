import React from 'react';
import Header from './header';
import Content from './Content';
import QueryComposer from '../../components/QueryComposer';

function CoreQuery() {
    return (
        <>
            <Header />
            <div>
                <QueryComposer visible={true}></QueryComposer>
            </div>
            <Content />
        </>
    )
}

export default CoreQuery;