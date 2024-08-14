import React, { useState } from 'react';
import ChatMessage from './ChatMessage';
import AppList from './AppList';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';

SyntaxHighlighter.registerLanguage('yaml', yaml);

function ChatInterface() {
  const [messages, setMessages] = useState([
    { role: 'assistant', content: 'Welcome to the Kiwi Deployment Assistant. Please select an app to deploy from the list.' }
  ]);
  const [input, setInput] = useState('');
  const [selectedApp, setSelectedApp] = useState(null);
  const [deployOptions, setDeployOptions] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (input.trim() === '') return;

    const newMessages = [...messages, { role: 'user', content: input }];
    setMessages(newMessages);
    setInput('');

    try {
      const response = await fetch(`http://localhost:8080/apps/${selectedApp}/detact`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: input }),
      });
      const data = await response.json();

      console.log(data)
      
      const aiResponse = { 
        role: 'assistant', 
        content: data.message, 
        options: data.options,
        code: JSON.stringify(data.options, null, 2)
      };
      setMessages([...newMessages, aiResponse]);
      setDeployOptions(data.options);
    } catch (error) {
      console.error('Error detecting deployment options:', error);
      const errorResponse = { role: 'assistant', content: 'Sorry, there was an error processing your request. Please try again.' };
      setMessages([...newMessages, errorResponse]);
    }
  };

  const handleDeploy = async () => {
    if (!deployOptions) return;

    try {
      const response = await fetch(`http://localhost:8080/apps/${selectedApp}/deploy`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(deployOptions),
      });
      const data = await response.json();

      console.log(data)
      
      const deployResponse = { role: 'assistant', content: data.message };
      setMessages([...messages, deployResponse]);
      setDeployOptions(null);
    } catch (error) {
      console.error('Error deploying app:', error);
      const errorResponse = { role: 'assistant', content: 'Sorry, there was an error deploying the app. Please try again.' };
      setMessages([...messages, errorResponse]);
    }
  };

  const handleSelectApp = async (app) => {
    setSelectedApp(app);
    try {
      const response = await fetch(`http://localhost:8080/apps/${app}/template`);
      const data = await response.json();
      const yamlContent = JSON.stringify(data, null, 2);
      const templateMessage = { 
        role: 'assistant', 
        content: `Here's the template for ${app}:`,
        code: yamlContent
      };
      setMessages([...messages, templateMessage]);
    } catch (error) {
      console.error('Error fetching app template:', error);
    }
  };

  return (
    <div className="h-full max-w-6xl mx-auto bg-white shadow-lg flex">
      <div className="w-1/3 p-4 bg-gray-50 border-r border-gray-200 overflow-y-auto">
        <AppList onSelectApp={handleSelectApp} />
      </div>
      <div className="w-2/3 flex flex-col">
        <div className="flex-grow overflow-y-auto p-4 space-y-4">
          {messages.map((message, index) => (
            <ChatMessage 
              key={index} 
              role={message.role} 
              content={message.content} 
              code={message.code}
              options={message.options}
              onDeploy={handleDeploy}
            />
          ))}
        </div>
        <form onSubmit={handleSubmit} className="p-4 bg-gray-50 border-t border-gray-200">
          <div className="flex">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              className="flex-grow px-4 py-2 border border-gray-300 rounded-l-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Type your message..."
              disabled={!selectedApp}
            />
            <button
              type="submit"
              className="px-6 py-2 bg-blue-600 text-white rounded-r-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-300 ease-in-out"
              disabled={!selectedApp}
            >
              Send
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default ChatInterface;