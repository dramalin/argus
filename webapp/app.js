const { useState, useEffect } = React;

function App() {
  const [cpu, setCpu] = useState(null);
  const [memory, setMemory] = useState(null);
  const [network, setNetwork] = useState(null);
  const [processes, setProcesses] = useState([]);

  useEffect(() => {
    fetchData();
    const id = setInterval(fetchData, 5000);
    return () => clearInterval(id);
  }, []);

  function fetchData() {
    fetch('/api/cpu').then(r => r.json()).then(setCpu).catch(console.error);
    fetch('/api/memory').then(r => r.json()).then(setMemory).catch(console.error);
    fetch('/api/network').then(r => r.json()).then(setNetwork).catch(console.error);
    fetch('/api/process').then(r => r.json()).then(setProcesses).catch(console.error);
  }

  return React.createElement('div', null,
    React.createElement('h1', null, 'Argus Monitor'),
    React.createElement('pre', null, JSON.stringify({cpu, memory, network, processes}, null, 2))
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(React.createElement(App));
