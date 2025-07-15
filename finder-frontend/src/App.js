import './App.css';
import React, { useState } from 'react';

function App() {
  const [domain, setDomain] = useState("");
  const [subdomains, setSubdomains] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

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
        <h1>Sub domain finder</h1>
        <p>Welcome! Enter a domain to find its subdomains.</p>
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
            style={{ fontSize: '1.5rem', padding: '0.5rem 2rem', cursor: 'pointer', backgroundColor: 'green', color: 'white', border: 'none', borderRadius: '4px' }}
          >
            Find Subdomains
          </button>
        </div>
        {loading && <p>Loading...</p>}
        {error && <p style={{color: 'red'}}>{error}</p>}
        {!loading && subdomains.length > 0 && (
          <p style={{padding: '10px'}}>Total Result = {subdomains.length}</p>
        )}
        <ul style={{ textAlign: 'left', maxWidth: '400px', margin: '0 auto' }}>
          {subdomains.sort().map((sub, idx) => (
            <li key={idx}>{sub}</li>
          ))}
        </ul>
      </header>
      <footer style={{ backgroundColor: 'black', color: 'white', padding: '1rem', marginTop: '2rem' }}>
        <p>© 2025 Sub Domain Finder. All rights reserved.</p>
      </footer>
    </div>
  );
}

export default App;
