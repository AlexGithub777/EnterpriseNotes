# Quick Start Guide

This guide outlines the steps and details required to install the Enterprise Notes application on a new server.

## Prerequisites

Before you begin, make sure you have the following prerequisites:

1. **Operating System**: Ensure that your server is running a supported operating system. Your application should work on Linux, Windows, and other common platforms.

2. **Go Installation**: You need to have Go installed on your server. You can download it from [Go Downloads](https://golang.org/dl/). Make sure to set up your Go environment correctly by configuring the `GOPATH` and `GOBIN` variables.

3. **Database**: Install and configure PostgreSQL on your server. You can download PostgreSQL from [PostgreSQL Downloads](https://www.postgresql.org/download/). Set up a PostgreSQL database with the necessary user and database name.

## Installation Steps

1. **Clone the Repository**:

    - SSH: Run the following command to clone your application repository:
        ```
        git clone git@github.com/AlexGithub777/notes.git
        ```
    - HTTPS: If you're using HTTPS, use this command instead:
        ```
        git clone https://github.com/AlexGithub777/notes.git
        ```

2. **Navigate to the Application Directory**: (cd your-repo)

3. **Database Configuration**:

-   Edit the `dbsetup.go` file to configure your PostgreSQL database connection settings. Modify the following constants to match your PostgreSQL setup:

    ```go
    const (
        host     = "localhost"
        port     = 5432
        user     = "yourdbuser"
        password = "yourdbpassword"
        dbname   = "yourdbname"
    )
    ```

4. **Build the Application**:

-   Build the application using the `go build` command. You have two options:
    -   Without vendored packages (if your dependencies are not vendored):
        ```
        ./buildpkg.cmd
        ```
    -   With vendored packages:
        ```
        ./buildvendor.cmd
        ```

5. **Run the Application**:

-   Start the application by running the `notes` executable:
    ```
    ./notes
    ```
-   The application will start, and you'll see log output indicating that it's listening on a port, typically 8080.

6. **Access the Application**:

-   Open a web browser and navigate to your server's IP address or domain name with the specified port (e.g., `http://your-server-ip:8080`).

7. **Testing**:

-   Test the application by creating, listing, and managing notes.

8. **Security**: Ensure that your server is properly secured, especially if it's accessible from the internet. You may need to configure firewalls, apply security updates, and implement authentication mechanisms to protect your application and database.
