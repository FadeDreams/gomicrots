import React, { useState } from 'react';
import logo from './logo.svg';
import './App.css';

function App() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');

  const authenticate = () => {
    // Perform AJAX request using Fetch API (modern way)
    fetch('http://localhost:8082/authenticate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    })
      .then(response => {
        if (response.ok) {
          return response.json();
        } else {
          throw new Error('Authentication failed');
        }
      })
      .then(data => {
        alert('check xhr tools');
        // alert(data.message);
      })
      .catch(error => {
        alert('check xhr tools');
        // alert(error.message);
      });
  };

  const registerUser = () => {
    // Perform AJAX request for user registration
    fetch('http://localhost:8082/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password, firstName, lastName }),
    })
      .then(response => {
        if (response.ok) {
          return response.json();
        } else {
          throw new Error('Registration failed');
        }
      })
      .then(data => {
        alert('check xhr tools');
        // alert(data.message);
      })
      .catch(error => {
        alert('check xhr tools');
        // alert(error.message);
      });
  };

  const testLogHandler = () => {
    // Perform AJAX request for LogHandler Tester
    const xhr = new XMLHttpRequest();
    xhr.open('POST', 'http://localhost:8083/log', true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = function() {
      if (xhr.readyState === 4 && xhr.status === 200) {
        const response = JSON.parse(xhr.responseText);
        alert('check xhr tools');
        // alert(response);
      } else if (xhr.readyState === 4) {
        alert('check xhr tools');
        // alert('Failed to send log entry');
      }
    };

    const data = JSON.stringify({ level: 'data', message: 'xx' });
    xhr.send(data);
  };

  const pushToRabbitMQ = () => {
    // Perform AJAX request for Push to RabbitMQ
    const xhr = new XMLHttpRequest();
    xhr.open('POST', 'http://localhost:8081/push', true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = function() {
      if (xhr.readyState === 4 && xhr.status === 200) {
        const response = JSON.parse(xhr.responseText);
        // Uncomment the line below if you want to alert the response
        // alert(response.message);
        alert('check xhr tools');
      } else if (xhr.readyState === 4) {
        alert('check xhr tools');
        // Uncomment the line below if you want to alert on failure
        // alert('Failed to push to RabbitMQ');
      }
    };

    // Customize the payload as needed
    const data = JSON.stringify({ /* Your payload data here */ });
    xhr.send(data);
  };

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />

        <div className="mb-3">
          <label htmlFor="email" className="form-label">
            Email:
          </label>
          <input
            type="email"
            className="form-control"
            id="email"
            name="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="mb-3">
          <label htmlFor="password" className="form-label">
            Password:
          </label>
          <input
            type="password"
            className="form-control"
            id="password"
            name="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <div className="mb-3">
          <label htmlFor="firstName" className="form-label">
            First Name:
          </label>
          <input
            type="text"
            className="form-control"
            id="firstName"
            name="firstName"
            required
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
          />
        </div>
        <div className="mb-3">
          <label htmlFor="lastName" className="form-label">
            Last Name:
          </label>
          <input
            type="text"
            className="form-control"
            id="lastName"
            name="lastName"
            required
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
          />
        </div>

        <div>
          <button type="button" className="btn btn-primary" onClick={authenticate}>
            Login
          </button>
        </div>
        <div>
          <button type="button" className="btn btn-primary" onClick={registerUser}>
            Register
          </button>
          <button type="button" className="btn btn-secondary" onClick={testLogHandler}>
            LogHandler Tester
          </button>
          <button type="button" className="btn btn-success" onClick={pushToRabbitMQ}>
            Push to RabbitMQ
          </button>
        </div>
      </header>
    </div>
  );
}

export default App;
