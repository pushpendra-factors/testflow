import React, { useState } from 'react';
import Header from './header';
import ResultsPage from './ResultsPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';

function CoreQuery() {
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [showResult, setShowResult] = useState(false);
    const [queries, setQueries] = useState([]);

    const addToQueries = (newEvent, index) => {
        const queryupdated = [...queries];
        if (queryupdated[index]) {
            queryupdated[index] = newEvent;
        } else {
            queryupdated.push(newEvent);
        }
        setQueries(queryupdated);
    }

    const runQuery = () => {
        setShowResult(true);
        closeDrawer();
    }

    const closeDrawer = () => {
        setDrawerVisible(false);
    }

    return (
        <>

            <QueryComposer
                drawerVisible={drawerVisible}
                queries={queries}
                onClose={closeDrawer}
                runQuery={runQuery}
                addEvent={addToQueries}
            >

            </QueryComposer>

            {
                showResult ? (<ResultsPage setDrawerVisible={setDrawerVisible} queries={queries.map(elem => elem.label)} />) : (<CoreQueryHome setDrawerVisible={setDrawerVisible} />)
            }

        </>
    )
}

export default CoreQuery;