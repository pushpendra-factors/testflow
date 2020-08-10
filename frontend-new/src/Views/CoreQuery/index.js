import React from 'react';
import Header from './header';
import QueryComposer from '../../components/QueryComposer';

function CoreQuery() {
    return (
        <>
            <Header />
            <div>
                <QueryComposer visible={true}></QueryComposer>
            </div>
        </>
    )
}

export default CoreQuery;