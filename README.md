# Notes App

This is a simple notes application that allows users to create, view, and filter notes. The application is built using Go and uses an SQLite database to store the notes.

## Prerequisites

Before compiling the application, make sure you have the following installed:

- Go programming language, tested on GO version 1.22
- MinGW-w64 C++ compiler (for SQLite library compilation)

### Installing MinGW-w64 on Windows

To install MinGW-w64 on Windows one can use Chocolatey:

```powershell
choco install mingw
```

## Building the Application

To build the application, follow these steps:

### Building 32-bit version

To build the 32-bit version of the application, run the following command:

```powershell
$env:GOOS="windows"; $env:GOARCH="386"; go build -o notes-x32.exe .\main.go
```

### Building 64-bit version

To build the 64-bit version of the application, run the following command:

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o notes-x64.exe .\main.go
```

## Running the Application

To run the application without building an executable, use the following command:

```powershell
go run main.go
```

The compilation has been tested on Windows 10 and Windows 11.

## Accessing the Application

Once the application is running, you can access it by opening a web browser and navigating to `http://localhost:8080`.
