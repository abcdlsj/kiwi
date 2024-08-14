import React, { useState, useEffect } from 'react';

function AppList({ onSelectApp }) {
  const [apps, setApps] = useState([]);

  useEffect(() => {
    fetchApps();
  }, []);

  const fetchApps = async () => {
    try {
      const response = await fetch('http://localhost:8080/apps');
      const data = await response.json();
      setApps(data.apps);
    } catch (error) {
      console.error('Error fetching apps:', error);
    }
  };

  return (
    <div className="mb-4">
      <h2 className="text-xl font-semibold mb-4 text-gray-700">Available Apps</h2>
      <ul className="space-y-2">
        {apps.map((app, index) => (
          <li key={index}>
            <button
              onClick={() => onSelectApp(app)}
              className="w-full text-left px-4 py-3 bg-gray-100 hover:bg-gray-200 rounded-lg transition duration-300 ease-in-out text-gray-700"
            >
              {app}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default AppList;