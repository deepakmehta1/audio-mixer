# Golang Project Setup

This document will guide you through installing Golang, setting up environment variables, managing dependencies using Go Modules, and running the application.

## 1. Install Golang

### Using Homebrew (Recommended)
1. Open your Terminal.
2. Install Golang with Homebrew:
   ```bash
   brew install go
   ```
3. Verify the installation:
   ```bash
   go version
   ```
### Clone and setup the audio-mixer project
1. Clone the repository:
    ```bash
   git clone https://github.com/deepakmehta1/audio-mixer.git
   cd audio-mixer
   ```
2. Create the Environment Variables File:
   Create a .env file at the root of your project with the following content:
   ```bash
   PORT=8080
   YOUTUBE_API_KEY=API_KEY
   HLS_BASE_URL=http://localhost:8080/hls/
   ```
3. To download and install the dependencies listed in your code, run::
   ```bash
   go mod tidy
   ```
4. Running the Application
    ```bash
   go run cmd/server/main.go
   ```
