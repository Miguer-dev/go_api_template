# Go API Template

This repository provides a comprehensive template for building a robust API in Go. The template includes essential features and configurations to accelerate your development process.

## Features

- **Request Limiting**: Configurable rate limiting based on the client's IP address to prevent abuse.
  
- **CORS Policy**: Pre-configured CORS policy to handle cross-origin resource sharing.

- **Request Metrics**: Integrated metrics for monitoring API requests and performance.

- **Graceful Shutdown**: Ensures all pending requests are completed before the server shuts down.

- **Panic Recovery**: Automatic recovery from panics in the main and secondary goroutines, ensuring the server remains operational.

- **SMTP Server Setup**: Pre-configured SMTP server with helper functions for sending emails, such as user activation or password resets.

- **JSON API Responses**: Standardized JSON format for all API responses.

- **User Management**:
  - **Registration**: User sign-up with email verification.
  - **Activation**: Secure token-based user account activation.
  - **Authentication**: Stateful token authentication to manage user sessions.
  - **Permissions**: Route protection based on user roles and permissions.

- **Token Management**:
  - User authentication tokens.
  - Activation tokens for user accounts.

- **PostgreSQL Integration**: Ready-to-use PostgreSQL database setup for handling all data storage needs.

- **CRUD Endpoints**: Pre-built Create, Read, Update, Delete endpoints with best practices.

- **Advanced Querying**:
  - **Filtering**: Easily filter data based on query parameters.
  - **Sorting**: Sort results on any field.
  - **Pagination**: Efficient pagination of results to handle large datasets.

- **Development Environment**:
  - **Air**: Hot reloading for a smooth development experience.
  
- **Makefile**:
  - **HELPERS**: Commands for common tasks.
  - **BUILD**: Build commands for different environments.
  - **DEVELOPMENT**: Commands tailored for development.
  - **PRODUCTION**: Commands for setting up and running production environments.
  - **QUALITY CONTROL**: Commands for linting, testing, and ensuring code quality.

- **Production Setup**:
  - **Bash Script**: Script to initialize and run the production server.
  - **Caddy Config**: Caddy server configuration for reverse proxying in production environments.

## Getting Started
1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/go-api-template.git
   ```
2. **Install dependencies**:
  ```bash
   make vendor
   ```
  This command will tidy and verify module dependencies and vendor them.
  
3. **Configure environment variables**:
- Duplicate .env.example and rename it to .env.
- Update the environment variables as needed.
4. **Apply database migrations**:
  ```bash
  make db/psql
  make db/migrations/up
  ```
  This command will apply all up database migrations.
  
5. **Run the API application**:
  ```bash
  air
  # or
  make run/api
  ```
## Based on "Let's Go Further"
This repository is based on the principles and content from the book **"Let's Go Further"** by Alex Edwards. The book provides an in-depth guide to building modern web applications in Go, and it serves as an excellent resource for understanding the concepts and patterns implemented in this template.

**Recommendation**: If you are serious about mastering Go and building scalable, maintainable web applications, I highly recommend reading **"Let's Go Further"**. The book offers invaluable insights and best practices that complement the functionality provided by this template.

## Contributing
Feel free to open issues or submit pull requests if you find any bugs or think of improvements.

## License
This project is licensed under the MIT License.
