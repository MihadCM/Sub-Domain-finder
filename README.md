# Sub Domain Finder

A comprehensive subdomain enumeration tool that combines Subfinder and Sublist3r results and serves them via a web interface.

## 🏗️ Architecture

### Backend (Go + Fiber)
- **Location**: `finder-backend/`
- **Port**: 3000
- **Technology**: Go with Fiber web framework
- **Function**: Runs subdomain enumeration and serves results via API

### Frontend (React)
- **Location**: `finder-frontend/`
- **Port**: 3001 (default React port)
- **Technology**: React.js
- **Function**: Provides user interface to interact with the backend

## 🔄 Data Flow

### How `combined_subs` Variable is Passed to Frontend

1. **User Input**: User enters a domain in the React frontend
2. **API Request**: Frontend sends POST request to `http://localhost:3000/find`
3. **Backend Processing**: 
   - Runs Subfinder: `subfinder -d domain -silent`
   - Runs Sublist3r: `python sublist3r.py -d domain`
   - Combines results: `combined := append(subfinderResults, sublist3rFiltered...)`
   - Deduplicates: `combined_subs := deduplicate(combined)`
4. **API Response**: Backend returns `combined_subs` as JSON array
5. **Frontend Display**: React receives and displays the subdomains

### API Endpoints

- `POST /find` - Main endpoint for subdomain enumeration
- `GET /api/subdomains/:domain` - Alternative GET endpoint
- `GET /health` - Health check endpoint

## 🚀 Quick Start

### 1. Start the Backend
```bash
cd finder-backend
go mod tidy
go run .
```

### 2. Start the Frontend
```bash
cd finder-frontend
npm install
npm start
```

### 3. Access the Application
- Frontend: http://localhost:3001
- Backend API: http://localhost:3000

## 📊 Example API Response

```json
[
  "sub1.example.com",
  "sub2.example.com",
  "sub3.example.com",
  "www.example.com",
  "mail.example.com"
]
```

## 🔧 Configuration

### Backend Paths
Update these paths in `finder-backend/main.go`:
```go
pythonPath := "/Users/mihad/Desktop/SecIq/Sublist3r/venv/bin/python"
sublist3rPath := "/Users/mihad/Desktop/SecIq/Sublist3r"
```

### Frontend API URL
The frontend is configured to call `http://localhost:3000/find` by default.

## 🛠️ Dependencies

### Backend Dependencies
- Go 1.21+
- Fiber v2.52.0
- Subfinder (external tool)
- Sublist3r (external tool)

### Frontend Dependencies
- React 19.1.0
- React DOM 19.1.0
- React Scripts 5.0.1

## 🔍 Features

- **Combined Enumeration**: Uses both Subfinder and Sublist3r
- **Deduplication**: Removes duplicate subdomains
- **Real-time Results**: Shows results as they're found
- **Error Handling**: Graceful error handling and user feedback
- **CORS Enabled**: Frontend can communicate with backend
- **Responsive UI**: Modern, user-friendly interface

## 📝 How It Works

1. **User enters domain** in the React frontend
2. **Frontend sends request** to Go backend via Fiber API
3. **Backend runs Subfinder** to find subdomains
4. **Backend runs Sublist3r** to find additional subdomains
5. **Backend combines results** using the `combined_subs` variable
6. **Backend returns JSON** array of unique subdomains
7. **Frontend displays results** in a clean, organized list

The `combined_subs` variable contains the final deduplicated list of all subdomains found by both tools, which is then sent to the frontend for display. 