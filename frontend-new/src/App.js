import React from 'react';
import AppLayout from './Views/AppLayout';
import { HashRouter, Route, Switch } from 'react-router-dom';

function App() {
  return (
    <div className="App">
      <HashRouter>
        <Switch>
          <Route path="/" name="Home" component={AppLayout} />
        </Switch>
      </HashRouter>
    </div>
  );
}

export default App;
