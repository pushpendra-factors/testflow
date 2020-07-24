import React from 'react';
import Sidebar from './components/Sidebar';

function App() {

  return (
    <section className="min-h-screen">
      <Sidebar />
      <section className="overflow-x-hidden p-4 min-h-screen">
        Content
        </section>
    </section>
  );
}

export default App;
