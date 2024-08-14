import React from 'react';
import ChatInterface from './components/ChatInterface';
import Header from './components/Header';

function App() {
  return (
    <div className="h-screen flex flex-col bg-gray-100 overflow-hidden">
      <Header />
      <main className="flex-grow overflow-hidden">
        <ChatInterface />
      </main>
    </div>
  );
}

export default App;