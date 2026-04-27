const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const fs = require('fs');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;
const DB_FILE = path.join(__dirname, 'database.json');

app.use(cors());
app.use(bodyParser.json());
app.use(express.static(path.join(__dirname, '../public')));

function readDatabase() {
  try {
    if (fs.existsSync(DB_FILE)) {
      const data = fs.readFileSync(DB_FILE, 'utf8');
      return JSON.parse(data);
    }
  } catch (e) {
    console.error('Error reading database:', e);
  }
  return { contacts: [] };
}

function writeDatabase(data) {
  fs.writeFileSync(DB_FILE, JSON.stringify(data, null, 2));
}

app.get('/api/contacts', (req, res) => {
  const db = readDatabase();
  res.json(db.contacts);
});

app.post('/api/contacts', (req, res) => {
  const { email, name, company } = req.body;
  
  if (!email) {
    return res.status(400).json({ error: 'Email is required' });
  }

  const db = readDatabase();
  const newContact = {
    id: Date.now(),
    email,
    name: name || '',
    company: company || '',
    createdAt: new Date().toISOString()
  };
  
  db.contacts.push(newContact);
  writeDatabase(db);
  
  res.status(201).json(newContact);
});

app.get('/api/logs/demo', (req, res) => {
  const logLevels = ['INFO', 'WARN', 'ERROR', 'DEBUG'];
  const messages = [
    'Request processed successfully',
    'Database connection established',
    'User authentication failed',
    'Cache miss for key: user_session_123',
    'API rate limit exceeded',
    'Memory usage at 78%',
    'Scheduled job completed',
    'New user registered',
    'Payment transaction initiated',
    'Connection timeout on external service'
  ];
  
  const logs = [];
  for (let i = 0; i < 15; i++) {
    const level = logLevels[Math.floor(Math.random() * logLevels.length)];
    const message = messages[Math.floor(Math.random() * messages.length)];
    const timestamp = new Date(Date.now() - Math.random() * 60000).toISOString();
    
    logs.push({
      id: i + 1,
      timestamp,
      level,
      message,
      service: ['auth-service', 'payment-api', 'user-service', 'data-pipeline'][Math.floor(Math.random() * 4)]
    });
  }
  
  logs.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
  res.json(logs);
});

app.listen(PORT, () => {
  console.log(`LogSwarm server running on http://localhost:${PORT}`);
});