import './App.css';
import React, { useState } from 'react';

function App() {
  const [domain, setDomain] = useState("");
  const [subdomains, setSubdomains] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  // const [resultInfo, setResultInfo] = useState(null);

  const fetchSubdomains = async () => {
    setLoading(true);
    setError("");
    try {
      const response = await fetch('http://localhost:3000/find', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain })
      });
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setSubdomains(data);
    } catch (err) {
      setError("Could not fetch subdomains");
      setSubdomains([]);
    }
    setLoading(false);
  };

  // Handle Enter key in input
  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      fetchSubdomains();
    }
  };

  return (
    <div className="App">
      {/* Header */}
      <header className="App-header">
        <h1>🔍 Sub Domain Finder</h1>
        <p>Welcome! Enter a domain to find its subdomains using Subfinder and Sublist3r.</p>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '1rem' }}>
          <input
            type="text"
            placeholder="Enter a domain"
            value={domain}
            onChange={e => setDomain(e.target.value)}
            onKeyDown={handleKeyDown}
            style={{ fontSize: '1.5rem', padding: '0.5rem 1rem', width: '300px', marginRight: '1rem' }}
          />
          <button
            onClick={fetchSubdomains}
            disabled={loading}
            style={{ 
              fontSize: '1.5rem', 
              padding: '0.5rem 2rem', 
              cursor: loading ? 'not-allowed' : 'pointer', 
              backgroundColor: loading ? '#ccc' : '#28a745', 
              color: 'white', 
              border: 'none', 
              borderRadius: '4px' 
            }}
          >
            {loading ? '🔍 Scanning...' : 'Find Subdomains'}
          </button>
        </div>
        
        {loading && (
          <div style={{ textAlign: 'center', margin: '1rem 0' }}>
            <p>🔍 Scanning for subdomains...</p>
            <p style={{ fontSize: '0.9rem', color: '#666' }}>This may take a few minutes</p>
          </div>
        )}
        
        {error && (
          <div style={{ 
            color: 'red', 
            backgroundColor: '#f8d7da', 
            padding: '10px', 
            borderRadius: '4px', 
            margin: '10px 0' 
          }}>
            ❌ {error}
          </div>
        )}
        {!loading && subdomains.length > 0 && (
          <div style={{}}>
            <h4>Subdomains Found: {subdomains.length}</h4>
            <div style={{ }}>
              <ul style={{textAlign: "left", paddingLeft: "20px"}}>
                {subdomains.sort().map((sub, idx) => (
                  <li key={idx} >{sub} </li>))}
              </ul>
            </div>
          </div>
        )}
        
        {!loading && subdomains.length === 0 && !error && (
          <div style={{ 
            backgroundColor: '#fff3cd', 
            padding: '15px', 
            borderRadius: '4px', 
            margin: '10px 0',
            textAlign: 'center'
          }}>
            <p style={{ fontSize: '0.9rem', color: '#666' }}>
              Try a different domain or check if the domain exists
            </p>
          </div>
        )}
      </header>
      
      <footer style={{ backgroundColor: 'black', color: 'white', padding: '1rem', marginTop: '2rem' }}>
        <p>© 2025 Sub Domain Finder. All rights reserved.</p>
      </footer>
    </div>
  );
}

export default App;
