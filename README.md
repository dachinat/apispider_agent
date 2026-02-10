[![Sponsor with PayPal](https://apispider.com/paypal.png)](https://paypal.me/dachina) [![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/L4L01T5K7F)


# APISpider – Agent

<h2 align="center">
  <img height="175" src="https://apispider.com/logo-full.png" alt="APISpider logo">
</h2>

<h3 align="center">
  APISpider agent - a local proxy for a web app to test private APIs
</h3>

This repository contains the **a lightweight APISpider local proxy** that allows APISpider web app to test localhost and private APIs. You can also check out the sibling repositories:

[APISpider frontend](https://github.com/dachinat/apispider_frontend)

[APISpider backend](https://github.com/dachinat/apispider_backend)


---

## Tech Stack

- Language: Go (v1.22+)

---

## Getting Started

### Why do I need this?

When using ApiSpider from a hosted server (e.g., apispider.com), you can't test APIs on your localhost or private network. This agent runs on your local machine and acts as a bridge between the web app and your local/private APIs.

### Features

- Test localhost APIs (http://localhost:3000, http://127.0.0.1:8080, etc.)
- Test private network APIs (http://192.168.1.100)
- No CORS issues
- Automatic response decompression (gzip, deflate)
- Lightweight (~4MB binary)
- Cross-platform (Windows, macOS, Linux)

### Installation

#### Option 1: Download Binary (recommended)
1. Download the latest release for your OS
2. Run the executable
3. Agent starts on http://localhost:8889

#### Option 2: Build from Source
```bash
cd apispider-agent
go build -o apispider-agent
./apispider-agent
```

### Usage

1. **Start the agent:**
   
   ```bash
   ./apispider-agent
   ```

2. **Open ApiSpider web app:**
   - The app will automatically detect the running agent
   - You'll see a green "Agent Connected" indicator

3. **Test localhost APIs:**
   - Enter: `http://localhost:3000/api/users`
   - Click Send
   - Request goes through the local agent!

### How It Works

```
Browser (apispider.com)
    ↓ HTTP Request
Local Agent (localhost:8889)
    ↓ Execute Request
Your localhost API (localhost:3000)
    ↓ Response
Browser receives response
```

### Security

- Agent only accepts requests from your browser
- No data is sent to external servers
- All requests are executed locally on your machine
- Open source - inspect the code!

### Endpoints

- `GET /`
- `GET /health` - Health check (used by frontend to detect agent)
- `POST /execute` - Execute HTTP request

### Port

Default port: `8889`

To change the port, modify the `port` variable in `main.go` and rebuild.

### Troubleshooting

**Agent not detected:**
- Make sure agent is running
- Check http://localhost:8889/health in your browser
- Restart the agent

**Request fails:**
- Check if your local API is running
- Verify the URL is correct
- Check agent logs for errors

### Building for Distribution

```bash
# Build for current OS
go build -o apispider-agent

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o apispider-agent-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o apispider-agent-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o apispider-agent-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o apispider-agent-windows-amd64.exe
```


## Contributing

Contributions are welcome and appreciated

If you’d like to contribute:

1. Fork the repository
2. Create a new branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Commit your changes (`git commit -m "Add my feature"`)
5. Push to the branch (`git push origin feature/my-feature`)
6. Open a Pull Request

Please try to:
- Follow the existing code style and project structure
- Keep commits focused and descriptive
- Test your changes before submitting

If you’re unsure where to start, feel free to open an issue or discussion.

---

## Support the Project

If you find APISpider useful, consider supporting the project  
[Support on Ko-fi](https://ko-fi.com/L4L01T5K7F)
