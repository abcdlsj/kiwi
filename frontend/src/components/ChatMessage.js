import React from 'react';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';

SyntaxHighlighter.registerLanguage('yaml', yaml);

function ChatMessage({ role, content, code, options, onDeploy }) {
  const isUser = role === 'user';
  return (
    <div className={`mb-4 ${isUser ? 'text-right' : 'text-left'}`}>
      <div
        className={`inline-block p-4 rounded-lg ${
          isUser ? 'bg-blue-100 text-blue-900' : 'bg-gray-100 text-gray-800'
        }`}
      >
        <p className="mb-2">{content}</p>
        {code && (
          <SyntaxHighlighter language="json" style={docco} className="rounded-md">
            {code}
          </SyntaxHighlighter>
        )}
        {options && (
          <div className="mt-2">
            <button
              onClick={onDeploy}
              className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 transition duration-300 ease-in-out"
            >
              Deploy
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default ChatMessage;