import React from 'react';
import AppLayout from './Views/AppLayout';
import componentsLib from './Views/componentsLib';
import { HashRouter, Route, Switch } from 'react-router-dom';

function App() {
  return (
    <div className="App">
      <HashRouter>
        <Switch>
          <Route path="/" exact name="Home" component={AppLayout} />
          <Route path="/components" name="componentsLib" component={componentsLib} />
        </Switch>
      </HashRouter>
    </div>
  );
}

export default App;
